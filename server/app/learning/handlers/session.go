package handlers

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/gofiber/fiber/v2"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	learning_service "verve/app/learning/service"
	rag_db "verve/app/rag/models/db"
	rag_payload "verve/app/rag/models/payload"
	rag_service "verve/app/rag/service"
	wiki_db "verve/app/wiki/models/db"
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

type messageStore interface {
	FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningMessage, error)
}

type reviewStore interface {
	Create(ctx context.Context, review *learning_db.LearningExplanationReview) error
	FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningExplanationReview, error)
}

type explanationMemoryRecorder interface {
	RecordExplanationReview(ctx context.Context, userID string, session *learning_db.LearningSession, review *learning_db.LearningExplanationReview) error
}

type SessionHandlerDependencies struct {
	Sessions  sessionStore
	Documents documentStore
	Messages  messageStore
	Reviews   reviewStore
	Reviewer  learning_service.FeynmanReviewer
	Memory    explanationMemoryRecorder
}

type SessionHandler struct {
	deps SessionHandlerDependencies
}

func NewSessionHandler(db *database.DatabaseService, minio *storage.MinIOService, retriever *rag_service.Retriever) *SessionHandler {
	source := &feynmanDocumentSource{
		documents: db.Documents,
		files:     minio,
		retriever: retriever,
		chunks:    db.RAGChunks,
	}
	return NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions: db.Sessions, Documents: db.Documents, Messages: db.Messages, Reviews: db.Reviews,
		Reviewer: learning_service.NewFeynmanReviewService(source),
		Memory:   learning_service.NewMemoryService(db),
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
	if _, err := h.deps.Documents.FindOne(c.Context(), req.DocumentID); err != nil {
		return response.NotFoundCtx(c, "文档不存在")
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
	return response.SuccessCtx(c, fiber.Map{"session": session, "messages": messages, "reviews": reviews})
}

func (h *SessionHandler) Review(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	session, err := h.ownedSession(c.Context(), userID, c.Params("id"))
	if err != nil {
		return writeSessionLookupError(c, err)
	}
	var req learning_payload.ReviewExplanationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	req.Explanation = strings.TrimSpace(req.Explanation)
	if req.Explanation == "" {
		return response.BadRequestCtx(c, "explanation 不能为空")
	}
	prior, err := h.deps.Reviews.FindBySession(c.Context(), session.ID)
	if err != nil {
		return response.InternalServerCtx(c, "加载先前解释失败")
	}
	reviewResult, err := h.deps.Reviewer.Review(c.Context(), session.DocumentID, req.Explanation, prior)
	if err != nil {
		log.Printf("Feynman review failed: session_id=%s user_id=%s document_id=%s err=%v", session.ID, userID, session.DocumentID, err)
		return response.InternalServerCtx(c, "审阅解释失败,请重试")
	}
	if reviewResult == nil {
		log.Printf("Feynman review returned no result: session_id=%s user_id=%s document_id=%s", session.ID, userID, session.DocumentID)
		return response.InternalServerCtx(c, "审阅解释失败,请重试")
	}
	review := reviewModelFromResult(session, userID, req.Explanation, reviewResult)
	if err := h.deps.Reviews.Create(c.Context(), review); err != nil {
		return response.InternalServerCtx(c, "记录解释审阅失败")
	}
	if h.deps.Memory != nil {
		if err := h.deps.Memory.RecordExplanationReview(c.Context(), userID, session, review); err != nil {
			log.Printf("record explanation memory failed: session_id=%s user_id=%s document_id=%s err=%v", session.ID, userID, session.DocumentID, err)
		}
	}
	return response.SuccessCtx(c, reviewResult)
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
	errSessionNotFound  = errors.New("session not found")
	errSessionForbidden = errors.New("session forbidden")
)

func (h *SessionHandler) ownedSession(ctx context.Context, userID, id string) (*learning_db.LearningSession, error) {
	session, err := h.deps.Sessions.FindOne(ctx, id)
	if err != nil {
		return nil, errSessionNotFound
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
	return response.NotFoundCtx(c, "会话不存在")
}

func reviewModelFromResult(session *learning_db.LearningSession, userID, explanation string, result *learning_service.FeynmanReview) *learning_db.LearningExplanationReview {
	return &learning_db.LearningExplanationReview{
		SessionID: session.ID, DocumentID: session.DocumentID, UserID: userID, Explanation: explanation,
		HeardSummary: result.HeardSummary, ClearPoints: result.ClearPoints,
		ConfusingPoints: result.ConfusingPoints, Misconceptions: result.Misconceptions,
		FollowUpQuestion: result.FollowUpQuestion, ExplanationSummary: result.ExplanationSummary,
		ReadyToWrapUp: result.ReadyToWrapUp, ContextSufficient: result.ContextSufficient,
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
	files     documentContentStore
	retriever documentRetriever
	chunks    documentChunkStore
}

func (s *feynmanDocumentSource) LoadDocument(ctx context.Context, documentID string) (*wiki_db.Document, string, error) {
	doc, err := s.documents.FindOne(ctx, documentID)
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
