package handlers

import (
	learning_db "verve/app/learning/models/db"
	"verve/common/pagination"
	"verve/common/response"
	"verve/infrastructure/database"

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
	objectiveID := c.Query("objective_id")

	var exercises []*learning_db.LearningExercise
	var total int
	var err error
	if objectiveID != "" {
		exercises, total, err = h.db.Exercises.FindByUserAndObjective(c.Context(), userID, objectiveID, req.GetOffset(), req.PageSize)
	} else {
		exercises, total, err = h.db.Exercises.FindByUser(c.Context(), userID, req.GetOffset(), req.PageSize)
	}
	if err != nil {
		return response.InternalServerCtx(c, "获取练习记录失败")
	}
	return response.PaginateCtx(c, exercises, total, req.Page, req.PageSize)
}
