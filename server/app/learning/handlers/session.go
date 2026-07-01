package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	learning_service "verve/app/learning/service"
	learning_tools "verve/app/learning/tools"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
)

// 学习会话处理器
type SessionHandler struct {
	db       *database.DatabaseService
	examiner *learning_service.ExaminerService
}

func NewSessionHandler(db *database.DatabaseService) *SessionHandler {
	return &SessionHandler{
		db:       db,
		examiner: learning_service.NewExaminerService(db),
	}
}

// 开始一节(关联某个小目标)
func (h *SessionHandler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req learning_payload.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	obj, err := h.db.Objectives.FindOne(c.Context(), req.ObjectiveID)
	if err != nil {
		return response.NotFoundCtx(c, "小目标不存在")
	}
	if obj.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	session := &learning_db.LearningSession{
		UserID:      userID,
		ObjectiveID: obj.ID,
		Status:      "active",
	}
	if err := h.db.Sessions.Create(c.Context(), session); err != nil {
		return response.InternalServerCtx(c, "创建会话失败")
	}

	return response.SuccessCtx(c, fiber.Map{"session_id": session.ID})
}

// 会话详情(含历史消息)
func (h *SessionHandler) FindOne(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	session, err := h.db.Sessions.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "会话不存在")
	}
	if session.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	messages, _ := h.db.Messages.FindBySession(c.Context(), id)
	return response.SuccessCtx(c, fiber.Map{"session": session, "messages": messages})
}

// Chat 陪练对话(SSE,Tutor)
func (h *SessionHandler) Chat(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	session, err := h.db.Sessions.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "会话不存在")
	}
	if session.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	var req struct {
		Message string `json:"message"`
	}
	_ = c.BodyParser(&req)

	// 落用户消息
	if req.Message != "" {
		_ = h.db.Messages.Create(c.Context(), &learning_db.LearningMessage{
			SessionID: id,
			Role:      "user",
			Content:   req.Message,
		})
	}

	// 当前小目标作为上下文
	obj, err := h.db.Objectives.FindOne(c.Context(), session.ObjectiveID)
	if err != nil {
		return response.InternalServerCtx(c, "小目标不存在")
	}

	// 构造 Tutor agent(带读工具)
	tools := learning_tools.NewLearningTools(h.db.Objectives, h.db.Exercises)
	agent, err := llm.NewTutorAgent(c.Context(), tools)
	if err != nil {
		return response.InternalServerCtx(c, "Tutor 初始化失败: "+err.Error())
	}
	runner := adk.NewRunner(c.Context(), adk.RunnerConfig{EnableStreaming: true, Agent: agent})
	iter := runner.Query(c.Context(), buildTutorQuery(obj, req.Message))

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	db := h.db
	sessionID := id
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		content := writeLearningSSE(w, iter)
		// 落 assistant 消息(用独立 context,避免请求结束被取消)
		if strings.TrimSpace(content) != "" {
			agentType := "tutor"
			_ = db.Messages.Create(context.Background(), &learning_db.LearningMessage{
				SessionID: sessionID,
				Role:      "assistant",
				AgentType: &agentType,
				Content:   content,
			})
		}
	}))
	return nil
}

// Exercise 提交练习验证(Examiner)
func (h *SessionHandler) Exercise(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	session, err := h.db.Sessions.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "会话不存在")
	}
	if session.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	var req learning_payload.SubmitExerciseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	obj, err := h.db.Objectives.FindOne(c.Context(), session.ObjectiveID)
	if err != nil {
		return response.InternalServerCtx(c, "小目标不存在")
	}

	result, err := h.examiner.Judge(c.Context(), obj, req.Type, req.Prompt, req.UserAnswer)
	if err != nil {
		log.Printf("❌ Examiner 判定失败: session_id=%s user_id=%s folder_id=%s objective_id=%s objective_title=%q type=%s prompt_chars=%d answer_chars=%d err=%v",
			id,
			userID,
			stringValue(obj.SourceFolderID),
			obj.ID,
			obj.Title,
			req.Type,
			len(req.Prompt),
			len(req.UserAnswer),
			err,
		)
		return response.InternalServerCtx(c, "判定失败,请重试")
	}

	// 落练习记录
	ua, verdict, ma, fb := req.UserAnswer, result.Verdict, result.MasteryAfter, result.Feedback
	_ = h.db.Exercises.Create(c.Context(), &learning_db.LearningExercise{
		SessionID:    id,
		ObjectiveID:  obj.ID,
		UserID:       userID,
		Type:         req.Type,
		Prompt:       req.Prompt,
		UserAnswer:   &ua,
		Verdict:      &verdict,
		MasteryAfter: &ma,
		Feedback:     &fb,
	})

	// 更新掌握层级
	obj.MasteryLevel = result.MasteryAfter
	_ = h.db.Objectives.Update(c.Context(), obj)
	h.syncLearningExaminerState(c.Context(), userID, obj, result)

	return response.SuccessCtx(c, fiber.Map{
		"verdict":             result.Verdict,
		"mastery_after":       result.MasteryAfter,
		"feedback":            result.Feedback,
		"objective_id":        obj.ID,
		"evidence":            result.Evidence,
		"weak_points":         result.WeakPoints,
		"next_recommendation": result.NextRecommendation,
		"review_required":     result.ReviewRequired,
	})
}

func (h *SessionHandler) syncLearningExaminerState(ctx context.Context, userID string, obj *learning_db.LearningObjective, result *learning_service.JudgeResult) {
	if result == nil {
		return
	}
	folderID := stringValue(obj.SourceFolderID)
	if folderID == "" {
		return
	}

	if profile, err := h.db.Profiles.FindByFolder(ctx, folderID); err == nil {
		learning_service.MergeLearningProfileState(profile, obj, result)
		_ = h.db.Profiles.Update(ctx, profile)
	} else if err == sql.ErrNoRows {
		profile := &learning_db.LearningProfile{
			UserID:   userID,
			FolderID: folderID,
		}
		learning_service.MergeLearningProfileState(profile, obj, result)
		_ = h.db.Profiles.Create(ctx, profile)
	}

	learned := obj.Title
	evidence := result.Evidence
	weakPoints := strings.Join(result.WeakPoints, "、")
	nextStep := result.NextRecommendation
	_ = h.db.Journals.Create(ctx, &learning_db.LearningJournal{
		UserID:     userID,
		FolderID:   folderID,
		Date:       time.Now(),
		Learned:    &learned,
		Evidence:   &evidence,
		WeakPoints: &weakPoints,
		NextStep:   &nextStep,
	})
}

// Complete 结束本节(标记完成、推进进度、写日志)
func (h *SessionHandler) Complete(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	session, err := h.db.Sessions.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "会话不存在")
	}
	if session.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	now := time.Now()
	session.Status = "completed"
	session.EndedAt = &now
	_ = h.db.Sessions.Update(c.Context(), session)

	obj, err := h.db.Objectives.FindOne(c.Context(), session.ObjectiveID)
	if err != nil {
		return response.InternalServerCtx(c, "小目标不存在")
	}
	obj.Status = "completed"
	_ = h.db.Objectives.Update(c.Context(), obj)

	// 推进到下一个小目标
	var next *learning_db.LearningObjective
	if obj.SourceFolderID != nil {
		objectives, err := h.db.Objectives.FindByFolder(c.Context(), *obj.SourceFolderID)
		if err == nil {
		for _, o := range objectives {
			if o.OrderIndex > obj.OrderIndex {
				next = o
				break
			}
		}
	}
	}
	if next != nil {
		next.Status = "active"
		_ = h.db.Objectives.Update(c.Context(), next)
	}

	// 写一条简单日志(同日同目标的唯一冲突忽略)
	learned := obj.Title
	if obj.SourceFolderID != nil {
		_ = h.db.Journals.Create(c.Context(), &learning_db.LearningJournal{
			UserID:   userID,
			FolderID: *obj.SourceFolderID,
			Date:     now,
			Learned:  &learned,
		})
	}

	resp := fiber.Map{"summary": "已完成:" + obj.Title}
	if next != nil {
		resp["next_objective"] = fiber.Map{"id": next.ID, "title": next.Title}
	}
	return response.SuccessCtx(c, resp)
}

// 构造注入当前小目标 + 用户消息的查询
func buildTutorQuery(obj *learning_db.LearningObjective, message string) string {
	var sb strings.Builder
	sb.WriteString("当前学习小目标:")
	sb.WriteString(obj.Title)
	if obj.Detail != nil && *obj.Detail != "" {
		sb.WriteString("\n要点:")
		sb.WriteString(*obj.Detail)
	}
	if strings.TrimSpace(message) != "" {
		sb.WriteString("\n\n学习者说:")
		sb.WriteString(message)
	} else {
		sb.WriteString("\n\n请开始这一节,先用一个小问题诊断学习者的基础。")
	}
	return sb.String()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// 把 agent 事件流写成 SSE,并返回累积的 assistant 文本(对齐现有 chat.go)
func writeLearningSSE(w *bufio.Writer, iter *adk.AsyncIterator[*adk.AgentEvent]) string {
	var content strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
			_ = w.Flush()
			break
		}
		if event.Err != nil {
			data, _ := json.Marshal(map[string]string{"type": "error", "content": event.Err.Error()})
			writeSSEData(w, data)
			break
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		mo := event.Output.MessageOutput

		if mo.MessageStream != nil {
			for {
				chunk, err := mo.MessageStream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					data, _ := json.Marshal(map[string]string{"type": "error", "content": err.Error()})
					writeSSEData(w, data)
					return content.String()
				}
				if chunk == nil {
					continue
				}
				if chunk.Role != schema.Tool {
					content.WriteString(chunk.Content)
				}
				data, _ := json.Marshal(map[string]interface{}{
					"type":    "stream_chunk",
					"content": chunk.Content,
					"agent":   event.AgentName,
				})
				writeSSEData(w, data)
			}
		} else if mo.Message != nil {
			if mo.Message.Role != schema.Tool {
				content.WriteString(mo.Message.Content)
			}
			data, _ := json.Marshal(map[string]interface{}{
				"type":    "message",
				"content": mo.Message.Content,
				"agent":   event.AgentName,
			})
			writeSSEData(w, data)
		}
	}
	return content.String()
}

func writeSSEData(w *bufio.Writer, data []byte) {
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n\n"))
	_ = w.Flush()
}
