package handlers

import (
	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"

	"github.com/gofiber/fiber/v2"
)

// 练习记录处理器
type ExerciseHandler struct {
	db *database.DatabaseService
}

func NewExerciseHandler(db *database.DatabaseService) *ExerciseHandler {
	return &ExerciseHandler{db: db}
}

// 练习记录分页(仅本人,按创建时间倒序)
func (h *ExerciseHandler) FindPage(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	req.Validate()

	exercises, total, err := h.db.Exercises.FindByUser(c.Context(), userID, req.GetOffset(), req.PageSize)
	if err != nil {
		return response.InternalServerCtx(c, "获取练习记录失败")
	}
	return response.PaginateCtx(c, exercises, total, req.Page, req.PageSize)
}
