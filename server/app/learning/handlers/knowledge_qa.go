package handlers

import (
	"bufio"
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	learning_service "verve/app/learning/service"
	rag_service "verve/app/rag/service"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
)

type knowledgeQAProcessor interface {
	Prepare(ctx context.Context, request learning_service.KnowledgeQARequest) (*learning_service.PreparedKnowledgeQA, error)
	Generate(ctx context.Context, prepared *learning_service.PreparedKnowledgeQA) (*learning_service.KnowledgeQAAnswer, error)
}

type KnowledgeQAHandler struct {
	service knowledgeQAProcessor
}

func NewKnowledgeQAHandler(db *database.DatabaseService, retriever *rag_service.Retriever) *KnowledgeQAHandler {
	var resolver llm.AgentModelResolver
	if db != nil {
		resolver = db.ModelConfigs
	}
	return newKnowledgeQAHandler(learning_service.NewKnowledgeQAService(
		retriever,
		learning_service.NewMemoryService(db),
		resolver,
	))
}

func newKnowledgeQAHandler(service knowledgeQAProcessor) *KnowledgeQAHandler {
	return &KnowledgeQAHandler{service: service}
}

func (h *KnowledgeQAHandler) Ask(c *fiber.Ctx) error {
	var request learning_service.KnowledgeQARequest
	if err := c.BodyParser(&request); err != nil {
		return response.BadRequestCtx(c, "请求体格式错误")
	}
	if h == nil || h.service == nil {
		return response.InternalServerCtx(c, "知识问答服务不可用")
	}

	prepared, err := h.service.Prepare(c.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, learning_service.ErrKnowledgeQAQuestionRequired):
			return response.BadRequestCtx(c, "问题不能为空")
		case errors.Is(err, learning_service.ErrKnowledgeQAQuestionTooLong):
			return response.BadRequestCtx(c, "问题不能超过 2000 个字符")
		default:
			log.Printf("knowledge QA retrieval failed: %v", err)
			return response.InternalServerCtx(c, "检索 Wiki 失败，请重试")
		}
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		if len(prepared.Sources) > 0 {
			_ = writeSSEEvent(w, SSESources, map[string]interface{}{"sources": prepared.Sources})
		}
		if prepared.ImmediateAnswer == nil {
			_ = writeSSEEvent(w, SSEStatus, map[string]interface{}{"phase": "generating"})
		}
		answer, generationErr := h.service.Generate(c.Context(), prepared)
		if generationErr != nil {
			log.Printf("knowledge QA generation failed: %v", generationErr)
			_ = writeSSEEvent(w, SSEError, map[string]interface{}{"message": "生成回答失败，请重试"})
		} else {
			_ = writeSSEEvent(w, SSEAnswer, map[string]interface{}{
				"knowledgeAnswer": answer.KnowledgeAnswer,
				"learningAdvice":  answer.LearningAdvice,
			})
		}
		_, _ = w.WriteString("data: [DONE]\n\n")
		_ = w.Flush()
	}))
	return nil
}
