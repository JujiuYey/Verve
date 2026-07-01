package handlers

import (
	"github.com/gofiber/fiber/v2"

	"verve/common/response"
	"verve/infrastructure/database"
)

// 学习画像处理器
type ProfileHandler struct {
	db *database.DatabaseService
}

func NewProfileHandler(db *database.DatabaseService) *ProfileHandler {
	return &ProfileHandler{db: db}
}

// 获取某目标的学习画像
func (h *ProfileHandler) Get(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	goalID := c.Params("id")

	// 校验目标归属
	goal, err := h.db.Goals.FindOne(c.Context(), goalID)
	if err != nil {
		return response.NotFoundCtx(c, "学习目标不存在")
	}
	if goal.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	profile, err := h.db.Profiles.FindByGoal(c.Context(), goalID)
	if err != nil {
		// 尚未生成画像
		return response.SuccessCtx[any](c, nil)
	}
	return response.SuccessCtx(c, profile)
}
