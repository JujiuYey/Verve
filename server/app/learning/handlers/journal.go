package handlers

import (
	"github.com/gofiber/fiber/v2"

	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
)

// 学习日志处理器
type JournalHandler struct {
	db *database.DatabaseService
}

func NewJournalHandler(db *database.DatabaseService) *JournalHandler {
	return &JournalHandler{db: db}
}

// 学习日志分页(仅本人,按日期倒序)
func (h *JournalHandler) FindPage(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	req.Validate()

	journals, total, err := h.db.Journals.FindByUser(c.Context(), userID, req.GetOffset(), req.PageSize)
	if err != nil {
		return response.InternalServerCtx(c, "获取学习日志失败")
	}
	return response.PaginateCtx(c, journals, total, req.Page, req.PageSize)
}
