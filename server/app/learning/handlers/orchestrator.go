package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"

	learning_payload "verve/app/learning/models/payload"
	learning_service "verve/app/learning/service"
	"verve/common/response"
	"verve/infrastructure/database"
)

type OrchestratorHandler struct {
	orchestrator *learning_service.LearningOrchestratorService
}

func NewOrchestratorHandler(db *database.DatabaseService) *OrchestratorHandler {
	return &OrchestratorHandler{
		orchestrator: learning_service.NewLearningOrchestratorService(db),
	}
}

func (h *OrchestratorHandler) Orchestrate(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未授权")
	}

	var req learning_payload.OrchestrateLearningRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.BadRequestCtx(c, err.Error())
		}
	}

	result, err := h.orchestrator.Orchestrate(c.Context(), userID, req.Intent)
	if err != nil {
		log.Printf("❌ 学习调度失败: user_id=%s intent_chars=%d err=%v", userID, len(req.Intent), err)
		return response.InternalServerCtx(c, "学习调度失败")
	}
	return response.SuccessCtx(c, result)
}
