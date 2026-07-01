package handlers

import (
	"github.com/gofiber/fiber/v2"

	"verve/common/response"
	"verve/infrastructure/database"
)

type ObjectiveHandler struct {
	db *database.DatabaseService
}

func NewObjectiveHandler(db *database.DatabaseService) *ObjectiveHandler {
	return &ObjectiveHandler{db: db}
}

func (h *ObjectiveHandler) FindOne(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	objective, err := h.db.Objectives.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "学习小节不存在")
	}
	if objective.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	return response.SuccessCtx(c, objective)
}
