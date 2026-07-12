package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/gofiber/fiber/v2"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	learning_repo "verve/app/learning/repository"
	learning_service "verve/app/learning/service"
	rag_db "verve/app/rag/models/db"
	rag_payload "verve/app/rag/models/payload"
	rag_service "verve/app/rag/service"
	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
)

type sessionStore interface {
	FindOne(ctx context.Context, id string) (*learning_db.LearningSession, error)
	Create(ctx context.Context, session *learning_db.LearningSession) error
	Update(ctx context.Context, session *learning_db.LearningSession) error
}

type documentStore interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Document, error)
}

type accessibleDocumentStore interface {
	FindAccessible(ctx context.Context, userID, id string) (*wiki_db.Document, error)
}

type folderStore interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Folder, error)
}

type messageStore interface {
	FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningMessage, error)
}

type reviewStore interface {
	FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningExplanationReview, error)
	FindByTurn(ctx context.Context, turnID string) (*learning_db.LearningExplanationReview, error)
}

type listenerTurnStore interface {
	BeginListenerTurn(ctx context.Context, sessionID, requestID, explanation string) (*learning_repo.BeginTurnResult, error)
	CompleteListenerTurn(ctx context.Context, sessionID, turnID, assistantContent string, review *learning_db.LearningExplanationReview) error
	RetryFailedTurn(ctx context.Context, turnID string) error
	FailTurn(ctx context.Context, turnID, errorCode, errorMessage string) error
}

type explanationMemoryStore interface {
	FindDocumentItems(ctx context.Context, userID, documentID string, limit int) ([]*learning_db.LearningMemoryItem, error)
	RecordExplanationReview(ctx context.Context, userID string, session *learning_db.LearningSession, review *learning_db.LearningExplanationReview) error
}

type turnSubmitter interface {
	Submit(context.Context, string, string, learning_service.TurnRequest) (*learning_payload.TimelineItem, error)
}

type timelineStore interface {
	FindBySession(context.Context, string) ([]*learning_payload.TimelineItem, error)
}

type SessionHandlerDependencies struct {
	Sessions    sessionStore
	Documents   accessibleDocumentStore
	Turns       listenerTurnStore
	Messages    messageStore
	Reviews     reviewStore
	Reviewer    learning_service.FeynmanReviewer
	Memory      explanationMemoryStore
	TurnService turnSubmitter
	Timeline    timelineStore
}

type SessionHandler struct {
	deps SessionHandlerDependencies
}

func NewSessionHandler(db *database.DatabaseService, minio *storage.MinIOService, retriever *rag_service.Retriever) *SessionHandler {
	source := &feynmanDocumentSource{
		documents: db.Documents,
		folders:   db.Folders,
		files:     minio,
		retriever: retriever,
		chunks:    db.RAGChunks,
	}
	memory := learning_service.NewMemoryService(db)
	reviewer := learning_service.NewFeynmanReviewService(source)
	turnService := learning_service.NewTurnService(db.Sessions, db.Turns, db.Timeline, map[string]learning_service.AgentProcessor{
		learning_db.LearningAgentListener: learning_service.NewListenerProcessor(reviewer, db.Reviews, memory),
		learning_db.LearningAgentTeacher:  learning_service.NewTeacherProcessor(learning_service.NewTeacherService(source), db.Reviews, memory),
		learning_db.LearningAgentCurator:  learning_service.NewCuratorProcessor(learning_service.NewCuratorService(source, db.ChangeRequests)),
	}, memory)
	return NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions: db.Sessions, Documents: source, Turns: db.Turns, Messages: db.Messages, Reviews: db.Reviews,
		Reviewer: reviewer, Memory: memory, TurnService: turnService, Timeline: db.Timeline,
	})
}

func NewSessionHandlerWithDependencies(deps SessionHandlerDependencies) *SessionHandler {
	return &SessionHandler{deps: deps}
}

func (h *SessionHandler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var req learning_payload.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	req.DocumentID = strings.TrimSpace(req.DocumentID)
	if req.DocumentID == "" {
		return response.BadRequestCtx(c, "document_id 不能为空")
	}
	if _, err := h.deps.Documents.FindAccessible(c.Context(), userID, req.DocumentID); err != nil {
		switch {
		case errors.Is(err, errDocumentForbidden):
			return response.ForbiddenCtx(c, "无权访问")
		case errors.Is(err, sql.ErrNoRows):
			return response.NotFoundCtx(c, "文档不存在")
		default:
			return response.InternalServerCtx(c, "读取文档失败")
		}
	}
	session := &learning_db.LearningSession{UserID: userID, DocumentID: req.DocumentID, Status: "active"}
	if err := h.deps.Sessions.Create(c.Context(), session); err != nil {
		return response.InternalServerCtx(c, "创建会话失败")
	}
	return response.SuccessCtx(c, fiber.Map{"session_id": session.ID})
}

func (h *SessionHandler) FindOne(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	session, err := h.ownedSession(c.Context(), userID, c.Params("id"))
	if err != nil {
		return writeSessionLookupError(c, err)
	}
	messages, err := h.deps.Messages.FindBySession(c.Context(), session.ID)
	if err != nil {
		return response.InternalServerCtx(c, "加载会话消息失败")
	}
	reviews, err := h.deps.Reviews.FindBySession(c.Context(), session.ID)
	if err != nil {
		return response.InternalServerCtx(c, "加载解释记录失败")
	}
	timeline := make([]*learning_payload.TimelineItem, 0)
	if h.deps.Timeline != nil {
		timeline, err = h.deps.Timeline.FindBySession(c.Context(), session.ID)
		if err != nil {
			return response.InternalServerCtx(c, "加载学习时间线失败")
		}
	}
	return response.SuccessCtx(c, fiber.Map{"session": session, "messages": messages, "reviews": reviews, "timeline": timeline})
}

func (h *SessionHandler) SubmitTurn(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if h.deps.TurnService == nil {
		return response.InternalServerCtx(c, "学习轮次服务未配置")
	}
	var request learning_service.TurnRequest
	if err := c.BodyParser(&request); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	item, err := h.deps.TurnService.Submit(c.Context(), userID, c.Params("id"), request)
	if err != nil {
		return writeTurnError(c, err)
	}
	return response.SuccessCtx(c, item)
}

func (h *SessionHandler) Review(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	session, err := h.ownedSession(c.Context(), userID, c.Params("id"))
	if err != nil {
		return writeSessionLookupError(c, err)
	}
	if session.Status != "active" {
		return c.Status(fiber.StatusConflict).JSON(response.FailWithCode(fiber.StatusConflict, "会话已结束"))
	}
	var req learning_payload.ReviewExplanationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	req.RequestID = strings.TrimSpace(req.RequestID)
	if req.RequestID == "" {
		return response.BadRequestCtx(c, "request_id 不能为空")
	}
	req.Explanation = strings.TrimSpace(req.Explanation)
	if req.Explanation == "" {
		return response.BadRequestCtx(c, "explanation 不能为空")
	}
	if h.deps.TurnService != nil {
		item, submitErr := h.deps.TurnService.Submit(c.Context(), userID, session.ID, learning_service.TurnRequest{
			RequestID: req.RequestID, AgentType: learning_db.LearningAgentListener, Content: req.Explanation,
		})
		if submitErr != nil {
			return writeTurnError(c, submitErr)
		}
		if item == nil || item.Artifact == nil || item.Artifact.Type != learning_payload.ArtifactExplanationReview {
			return response.InternalServerCtx(c, "读取解释审阅失败")
		}
		review, ok := item.Artifact.Data.(*learning_db.LearningExplanationReview)
		if !ok {
			return response.InternalServerCtx(c, "解释审阅格式无效")
		}
		return response.SuccessCtx(c, reviewResultFromModel(review))
	}
	if h.deps.Turns == nil {
		return response.InternalServerCtx(c, "学习轮次存储未配置")
	}
	started, err := h.deps.Turns.BeginListenerTurn(c.Context(), session.ID, req.RequestID, req.Explanation)
	if err != nil {
		return response.InternalServerCtx(c, "创建学习轮次失败")
	}
	if started == nil || started.Turn == nil {
		return response.InternalServerCtx(c, "创建学习轮次失败")
	}
	turn := started.Turn
	if !started.Created {
		switch turn.Status {
		case learning_db.LearningTurnCompleted:
			review, findErr := h.deps.Reviews.FindByTurn(c.Context(), turn.ID)
			if findErr != nil {
				return response.InternalServerCtx(c, "读取已完成审阅失败")
			}
			return response.SuccessCtx(c, reviewResultFromModel(review))
		case learning_db.LearningTurnProcessing:
			return c.Status(fiber.StatusConflict).JSON(response.FailWithCode(fiber.StatusConflict, "请求正在处理中"))
		case learning_db.LearningTurnFailed:
			if err := h.deps.Turns.RetryFailedTurn(c.Context(), turn.ID); err != nil {
				return response.InternalServerCtx(c, "重试学习轮次失败")
			}
		default:
			return response.InternalServerCtx(c, "学习轮次状态无效")
		}
	}
	prior, err := h.deps.Reviews.FindBySession(c.Context(), session.ID)
	if err != nil {
		h.failTurn(c.Context(), turn.ID, "listener_context_failed", err)
		return response.InternalServerCtx(c, "加载先前解释失败")
	}
	memoryItems := make([]*learning_db.LearningMemoryItem, 0)
	if h.deps.Memory != nil {
		loaded, memoryErr := h.deps.Memory.FindDocumentItems(c.Context(), userID, session.DocumentID, 20)
		if memoryErr != nil {
			log.Printf("load explanation memory failed, continuing without memory: session_id=%s user_id=%s document_id=%s err=%v", session.ID, userID, session.DocumentID, memoryErr)
		} else if loaded != nil {
			memoryItems = loaded
		}
	}
	reviewResult, err := h.deps.Reviewer.Review(c.Context(), learning_service.FeynmanReviewRequest{
		UserID: userID, DocumentID: session.DocumentID, Explanation: req.Explanation,
		PriorTurns: prior, MemoryItems: memoryItems,
	})
	if err != nil {
		h.failTurn(c.Context(), turn.ID, "listener_failed", err)
		log.Printf("Feynman review failed: session_id=%s user_id=%s document_id=%s err=%v", session.ID, userID, session.DocumentID, err)
		return response.InternalServerCtx(c, "审阅解释失败,请重试")
	}
	if reviewResult == nil {
		h.failTurn(c.Context(), turn.ID, "listener_empty_result", errors.New("reviewer returned no result"))
		log.Printf("Feynman review returned no result: session_id=%s user_id=%s document_id=%s", session.ID, userID, session.DocumentID)
		return response.InternalServerCtx(c, "审阅解释失败,请重试")
	}
	review := reviewModelFromResult(reviewResult)
	assistantContent, err := json.Marshal(reviewResult)
	if err != nil {
		h.failTurn(c.Context(), turn.ID, "listener_encode_failed", err)
		return response.InternalServerCtx(c, "记录解释审阅失败")
	}
	if err := h.deps.Turns.CompleteListenerTurn(c.Context(), session.ID, turn.ID, string(assistantContent), review); err != nil {
		h.failTurn(c.Context(), turn.ID, "listener_persist_failed", err)
		return response.InternalServerCtx(c, "记录解释审阅失败")
	}
	if h.deps.Memory != nil {
		if err := h.deps.Memory.RecordExplanationReview(c.Context(), userID, session, review); err != nil {
			log.Printf("record explanation memory failed: session_id=%s user_id=%s document_id=%s err=%v", session.ID, userID, session.DocumentID, err)
		}
	}
	return response.SuccessCtx(c, reviewResult)
}

func (h *SessionHandler) failTurn(ctx context.Context, turnID, errorCode string, cause error) {
	if h.deps.Turns == nil || strings.TrimSpace(turnID) == "" || cause == nil {
		return
	}
	if err := h.deps.Turns.FailTurn(ctx, turnID, errorCode, cause.Error()); err != nil {
		log.Printf("mark learning turn failed: turn_id=%s code=%s err=%v", turnID, errorCode, err)
	}
}

func (h *SessionHandler) Complete(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	session, err := h.ownedSession(c.Context(), userID, c.Params("id"))
	if err != nil {
		return writeSessionLookupError(c, err)
	}
	reviews, err := h.deps.Reviews.FindBySession(c.Context(), session.ID)
	if err != nil {
		return response.InternalServerCtx(c, "加载解释记录失败")
	}
	summaryParts := make([]string, 0, len(reviews))
	for _, review := range reviews {
		if review != nil && strings.TrimSpace(review.ExplanationSummary) != "" {
			summaryParts = append(summaryParts, strings.TrimSpace(review.ExplanationSummary))
		}
	}
	summary := strings.Join(summaryParts, "\n")
	now := time.Now()
	session.Status = "completed"
	session.EndedAt = &now
	session.Summary = &summary
	if err := h.deps.Sessions.Update(c.Context(), session); err != nil {
		return response.InternalServerCtx(c, "结束会话失败")
	}
	return response.SuccessCtx(c, fiber.Map{"summary": summary})
}

var (
	errSessionNotFound   = errors.New("session not found")
	errSessionForbidden  = errors.New("session forbidden")
	errDocumentForbidden = errors.New("document forbidden")
)

func (h *SessionHandler) ownedSession(ctx context.Context, userID, id string) (*learning_db.LearningSession, error) {
	session, err := h.deps.Sessions.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errSessionNotFound
		}
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session.UserID != userID {
		return nil, errSessionForbidden
	}
	return session, nil
}

func writeSessionLookupError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errSessionForbidden) {
		return response.ForbiddenCtx(c, "无权访问")
	}
	if errors.Is(err, errSessionNotFound) {
		return response.NotFoundCtx(c, "会话不存在")
	}
	return response.InternalServerCtx(c, "读取会话失败")
}

func writeTurnError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, learning_service.ErrTurnProcessing):
		return c.Status(fiber.StatusConflict).JSON(response.FailWithCode(fiber.StatusConflict, "turn_processing"))
	case errors.Is(err, learning_repo.ErrTurnRequestConflict):
		return c.Status(fiber.StatusConflict).JSON(response.FailWithCode(fiber.StatusConflict, "request_conflict"))
	case errors.Is(err, learning_service.ErrIndexNotReady):
		return c.Status(fiber.StatusConflict).JSON(response.FailWithCode(fiber.StatusConflict, "index_not_ready"))
	case errors.Is(err, learning_service.ErrTurnSessionForbidden):
		return response.ForbiddenCtx(c, "无权访问")
	case errors.Is(err, learning_service.ErrTurnSessionCompleted):
		return c.Status(fiber.StatusConflict).JSON(response.FailWithCode(fiber.StatusConflict, "session_completed"))
	case errors.Is(err, learning_service.ErrUnsupportedAgent):
		return response.BadRequestCtx(c, "agent_type 无效")
	case errors.Is(err, learning_service.ErrInvalidTurnRequest):
		return response.BadRequestCtx(c, "request_id 和 content 不能为空")
	case errors.Is(err, wiki_repo.ErrChangeRequestForbidden):
		return response.ForbiddenCtx(c, "无权替代该文档变更申请")
	case errors.Is(err, sql.ErrNoRows):
		return response.NotFoundCtx(c, "会话不存在")
	default:
		return response.InternalServerCtx(c, "处理学习轮次失败")
	}
}

func reviewModelFromResult(result *learning_service.FeynmanReview) *learning_db.LearningExplanationReview {
	return &learning_db.LearningExplanationReview{
		HeardSummary: result.HeardSummary, ClearPoints: result.ClearPoints,
		ConfusingPoints: result.ConfusingPoints, Misconceptions: result.Misconceptions,
		FollowUpQuestion: result.FollowUpQuestion, ExplanationSummary: result.ExplanationSummary,
		ReadyToWrapUp: result.ReadyToWrapUp, ContextSufficient: result.ContextSufficient,
	}
}

func reviewResultFromModel(review *learning_db.LearningExplanationReview) *learning_service.FeynmanReview {
	if review == nil {
		return nil
	}
	return &learning_service.FeynmanReview{
		HeardSummary: review.HeardSummary, ClearPoints: review.ClearPoints,
		ConfusingPoints: review.ConfusingPoints, Misconceptions: review.Misconceptions,
		FollowUpQuestion: review.FollowUpQuestion, ExplanationSummary: review.ExplanationSummary,
		ReadyToWrapUp: review.ReadyToWrapUp, ContextSufficient: review.ContextSufficient,
	}
}

type documentContentStore interface {
	GetFileContent(ctx context.Context, objectName string) (string, error)
}

type documentRetriever interface {
	SearchDocument(ctx context.Context, documentID, query string, limit int) ([]rag_payload.SearchResult, error)
}

type documentChunkStore interface {
	FindNeighbors(ctx context.Context, documentID string, indexes []int, radius int) ([]*rag_db.WikiChunk, error)
}

type feynmanDocumentSource struct {
	documents documentStore
	folders   folderStore
	files     documentContentStore
	retriever documentRetriever
	chunks    documentChunkStore
}

func (s *feynmanDocumentSource) FindAccessible(ctx context.Context, userID, documentID string) (*wiki_db.Document, error) {
	doc, err := s.documents.FindOne(ctx, documentID)
	if err != nil {
		return nil, err
	}
	folder, err := s.folders.FindOne(ctx, doc.FolderID)
	if err != nil {
		return nil, err
	}
	if folder.UserID != nil && strings.TrimSpace(*folder.UserID) != "" && *folder.UserID != userID {
		return nil, errDocumentForbidden
	}
	return doc, nil
}

func (s *feynmanDocumentSource) LoadDocument(ctx context.Context, userID, documentID string) (*wiki_db.Document, string, error) {
	doc, err := s.FindAccessible(ctx, userID, documentID)
	if err != nil {
		return nil, "", err
	}
	content, err := s.files.GetFileContent(ctx, doc.FilePath)
	if err != nil {
		return nil, "", err
	}
	return doc, content, nil
}

func (s *feynmanDocumentSource) SearchDocument(ctx context.Context, documentID, query string, limit int) ([]rag_payload.SearchResult, error) {
	return s.retriever.SearchDocument(ctx, documentID, query, limit)
}

func (s *feynmanDocumentSource) FindNeighbors(ctx context.Context, documentID string, indexes []int, radius int) ([]*rag_db.WikiChunk, error) {
	return s.chunks.FindNeighbors(ctx, documentID, indexes, radius)
}

// writeLearningSSEContent is retained for the Coach stream.
func writeLearningSSEContent(w *bufio.Writer, iter *adk.AsyncIterator[*adk.AgentEvent]) string {
	var content strings.Builder
	state := &thinkParseState{}
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			_ = writeSSEEvent(w, SSEError, map[string]interface{}{"content": event.Err.Error()})
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
					_ = writeSSEEvent(w, SSEError, map[string]interface{}{"content": err.Error()})
					return content.String()
				}
				if chunk != nil {
					dispatchMessageChunk(w, chunk, event.AgentName, &content, state)
				}
			}
		} else if mo.Message != nil {
			dispatchMessageChunk(w, mo.Message, event.AgentName, &content, state)
		}
	}
	return content.String()
}
