package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	learning_db "verve/app/learning/models/db"
	learning_service "verve/app/learning/service"
	learning_tools "verve/app/learning/tools"
	wiki_db "verve/app/wiki/models/db"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
)

type CoachHandler struct {
	db *database.DatabaseService
}

// NewCoachHandler 构造陪练 handler,持有 db 供 runtime context 与 tools 共用。
func NewCoachHandler(db *database.DatabaseService) *CoachHandler {
	return &CoachHandler{db: db}
}

// Chat 处理 POST /api/learning/coach/chat 陪练对话请求(SSE)。
//
// 流程:解析 message → 构建 runtime context → 初始化 CoachTools 与 agent →
// 通过 eino ADK Runner 流式产出 → 写入 SSE 帧;最后解析 <ACTION> 标签
// 追加一条 action 事件,并以 [DONE] 收尾。
func (h *CoachHandler) Chat(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var req struct {
		Message string `json:"message"`
	}
	_ = c.BodyParser(&req)
	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = "继续学习"
	}

	runtimeContext, err := h.buildRuntimeContext(c.Context(), userID)
	if err != nil {
		log.Printf("❌ 构建学习调度上下文失败: user_id=%s err=%v", userID, err)
		return response.InternalServerCtx(c, "构建学习上下文失败")
	}

	tools := learning_tools.NewCoachTools(h.db, userID)
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
		if action := learning_service.ParseCoachAction(content); action != nil {
			_ = writeSSEEvent(w, SSEAction, map[string]interface{}{"action": action})
		}
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		_ = w.Flush()
	}))
	return nil
}

// buildRuntimeContext 聚合陪练 agent 所需的运行时上下文:用户文件夹(已过滤)、
// 文档、最近 objective、每个文件夹的 LearningProfile、最近 journal。
func (h *CoachHandler) buildRuntimeContext(ctx context.Context, userID string) (learning_service.CoachRuntimeContext, error) {
	folders, err := h.db.Folders.List(ctx, map[string]interface{}{})
	if err != nil {
		return learning_service.CoachRuntimeContext{}, err
	}
	folders = filterFoldersForUser(folders, userID)

	docs, err := h.db.Documents.List(ctx, "", "")
	if err != nil {
		return learning_service.CoachRuntimeContext{}, err
	}

	objectives, err := h.db.Objectives.FindRecentByUser(ctx, userID, 50)
	if err != nil {
		return learning_service.CoachRuntimeContext{}, err
	}

	journals, _, err := h.db.Journals.FindByUser(ctx, userID, 0, 10)
	if err != nil {
		return learning_service.CoachRuntimeContext{}, err
	}

	profiles := make([]*learning_db.LearningProfile, 0)
	seenFolders := make(map[string]bool)
	for _, folder := range folders {
		if seenFolders[folder.ID] {
			continue
		}
		seenFolders[folder.ID] = true
		profile, err := h.db.Profiles.FindByFolder(ctx, folder.ID)
		if err == nil {
			profiles = append(profiles, profile)
		} else if err != sql.ErrNoRows {
			return learning_service.CoachRuntimeContext{}, err
		}
	}

	return learning_service.CoachRuntimeContext{
		UserID:     userID,
		Folders:    folders,
		Documents:  docs,
		Objectives: objectives,
		Profiles:   profiles,
		Journals:   journals,
	}, nil
}

// filterFoldersForUser 过滤出归属当前用户的文件夹;未归属(folder.UserID 为空)
// 视为公开资源一并保留。
func filterFoldersForUser(folders []*wiki_db.Folder, userID string) []*wiki_db.Folder {
	result := make([]*wiki_db.Folder, 0, len(folders))
	for _, folder := range folders {
		if folder.UserID != nil && *folder.UserID != "" && *folder.UserID != userID {
			continue
		}
		result = append(result, folder)
	}
	return result
}
