package handlers

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	learning_payload "sag-wiki/app/learning/models/payload"
	learning_service "sag-wiki/app/learning/service"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
)

type GuideHandler struct {
	db    *database.DatabaseService
	guide *learning_service.GuideService
}

func NewGuideHandler(db *database.DatabaseService) *GuideHandler {
	return &GuideHandler{
		db:    db,
		guide: learning_service.NewGuideService(db),
	}
}

func (h *GuideHandler) Get(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	objectiveID := strings.TrimSpace(c.Params("objectiveId"))
	contentHash := strings.TrimSpace(c.Query("content_hash"))
	if objectiveID == "" {
		return response.BadRequestCtx(c, "缺少小目标 ID")
	}
	if contentHash == "" {
		return response.BadRequestCtx(c, "缺少内容 hash")
	}

	obj, err := h.db.Objectives.FindOne(c.Context(), objectiveID)
	if err != nil {
		return response.NotFoundCtx(c, "小目标不存在")
	}
	if obj.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	result, err := h.guide.FindCached(c.Context(), obj, contentHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.SuccessCtx[any](c, nil)
		}
		log.Printf("❌ 读取导学缓存失败: user_id=%s objective_id=%s content_hash=%s err=%v", userID, objectiveID, contentHash, err)
		return response.InternalServerCtx(c, "读取导学缓存失败")
	}

	return response.SuccessCtx(c, result)
}

func (h *GuideHandler) Generate(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req learning_payload.GenerateGuideRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	if strings.TrimSpace(req.ObjectiveID) == "" {
		return response.BadRequestCtx(c, "缺少小目标 ID")
	}
	if strings.TrimSpace(req.Markdown) == "" {
		return response.BadRequestCtx(c, "缺少 Markdown 学习资料")
	}

	obj, err := h.db.Objectives.FindOne(c.Context(), req.ObjectiveID)
	if err != nil {
		return response.NotFoundCtx(c, "小目标不存在")
	}
	if obj.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	result, err := h.guide.Generate(c.Context(), obj, req.Markdown)
	if err != nil {
		return response.InternalServerCtx(c, "导学生成失败,请重试")
	}

	return response.SuccessCtx(c, result)
}
