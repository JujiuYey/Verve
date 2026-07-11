package handlers

import (
	"bufio"
	"context"
	"errors"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	learning_db "verve/app/learning/models/db"
	learning_service "verve/app/learning/service"
	learning_tools "verve/app/learning/tools"
	rag_service "verve/app/rag/service"
	wiki_db "verve/app/wiki/models/db"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
)

type CoachHandler struct {
	db        *database.DatabaseService
	retriever *rag_service.Retriever
}

// NewCoachHandler 构造陪练 handler,持有 db 供 runtime context 与 tools 共用。
func NewCoachHandler(db *database.DatabaseService, retriever *rag_service.Retriever) *CoachHandler {
	return &CoachHandler{db: db, retriever: retriever}
}

// Chat 处理 POST /api/learning/coach/chat 陪练对话请求(SSE)。
//
// 流程:解析 message → 构建 runtime context → 初始化 CoachTools 与 agent →
// 通过 eino ADK Runner 流式产出 → 写入 SSE 帧;最后解析 <ACTION> 标签
// 追加一条 action 事件,并以 [DONE] 收尾。
func (h *CoachHandler) Chat(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var req struct {
		Message      string `json:"message"`
		RootFolderID string `json:"root_folder_id"`
	}
	_ = c.BodyParser(&req)
	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = "继续学习"
	}

	runtimeContext, err := h.buildRuntimeContext(c.Context(), userID, strings.TrimSpace(req.RootFolderID))
	if err != nil {
		log.Printf("❌ 构建学习调度上下文失败: user_id=%s err=%v", userID, err)
		return response.InternalServerCtx(c, "构建学习上下文失败")
	}

	tools := learning_tools.NewCoachTools(h.db, h.retriever, userID)
	agent, err := llm.NewCoachAgent(c.Context(), tools)
	if err != nil {
		return response.InternalServerCtx(c, "学习 agent 初始化失败: "+err.Error())
	}

	runner := adk.NewRunner(c.Context(), adk.RunnerConfig{EnableStreaming: true, Agent: agent})
	iter := runner.Query(c.Context(), learning_service.BuildCoachQuery(runtimeContext, message))

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		content := writeLearningSSEContent(w, iter)
		_ = writeScopedCoachAction(w, content, runtimeContext.Documents)
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		_ = w.Flush()
	}))
	return nil
}

func writeScopedCoachAction(w *bufio.Writer, content string, documents []*wiki_db.Document) error {
	action := learning_service.ParseCoachActionForDocuments(content, documents)
	if action == nil {
		return nil
	}
	return writeSSEEvent(w, SSEAction, map[string]interface{}{"action": action})
}

// buildRuntimeContext 聚合陪练 agent 所需的运行时上下文:用户文件夹(已过滤)、
// 文档和学习记忆。
func (h *CoachHandler) buildRuntimeContext(ctx context.Context, userID string, rootFolderID string) (learning_service.CoachRuntimeContext, error) {
	if h == nil || h.db == nil {
		return learning_service.CoachRuntimeContext{}, errors.New("learning coach database is not configured")
	}

	folders, err := h.db.Folders.List(ctx, map[string]interface{}{})
	if err != nil {
		return learning_service.CoachRuntimeContext{}, err
	}
	folders = learning_service.FilterCoachFoldersForUser(folders, userID)
	if rootFolderID != "" && !learning_service.CoachFolderIsAccessible(folders, rootFolderID) {
		return learning_service.CoachRuntimeContext{}, errors.New("requested Wiki root folder is not accessible")
	}

	rootFolderName := ""
	if rootFolderID != "" {
		if !learning_service.CoachFolderIsAccessible(folders, rootFolderID) {
			return learning_service.CoachRuntimeContext{}, errors.New("Wiki root folder is not accessible")
		}
		if rootFolder, err := h.db.Folders.FindOne(ctx, rootFolderID); err == nil {
			rootFolderName = rootFolder.Name
		}
	}

	docs, err := h.db.Documents.List(ctx, "", "")
	if err != nil {
		return learning_service.CoachRuntimeContext{}, err
	}
	docs = learning_service.FilterCoachDocumentsByFolders(docs, folders)
	if rootFolderID != "" {
		docs, err = h.filterDocumentsByRoot(ctx, docs, rootFolderID)
		if err != nil {
			return learning_service.CoachRuntimeContext{}, err
		}
	}

	memoryItems := make([]*learning_db.LearningMemoryItem, 0)
	if h.db.Memories != nil {
		memoryItems, err = learning_service.NewMemoryService(h.db).FindCoachItems(ctx, userID, rootFolderID, 20)
		if err != nil {
			return learning_service.CoachRuntimeContext{}, err
		}
	}

	return learning_service.CoachRuntimeContext{
		UserID:         userID,
		RootFolderID:   rootFolderID,
		RootFolderName: rootFolderName,
		Folders:        folders,
		Documents:      docs,
		MemoryItems:    memoryItems,
	}, nil
}

func (h *CoachHandler) filterDocumentsByRoot(ctx context.Context, docs []*wiki_db.Document, rootFolderID string) ([]*wiki_db.Document, error) {
	folderIDs, err := h.db.Folders.GetAllSubFolderIDs(ctx, rootFolderID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(folderIDs))
	for _, folderID := range folderIDs {
		allowed[folderID] = true
	}
	result := make([]*wiki_db.Document, 0, len(docs))
	for _, doc := range docs {
		if allowed[doc.FolderID] {
			result = append(result, doc)
		}
	}
	return result, nil
}
