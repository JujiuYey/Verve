package handlers

import (
	"strings"

	system_db "verve/app/system/models/db"
	system_payload "verve/app/system/models/payload"
	system_repository "verve/app/system/repository"
	"verve/common/response"
	"verve/infrastructure/database"

	"github.com/gofiber/fiber/v2"
)

// ModelHandler 已启用模型处理器
type ModelHandler struct {
	repo system_repository.ModelConfigRepository
}

func NewModelHandler(dbService *database.DatabaseService) *ModelHandler {
	repo := system_repository.NewModelConfigRepository(dbService.GetDB())
	return &ModelHandler{repo: repo}
}

func (h *ModelHandler) FindModels(c *fiber.Ctx) error {
	models, err := h.repo.FindModels(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "获取模型列表失败")
	}
	return response.SuccessCtx(c, models)
}

func (h *ModelHandler) CreateModel(c *fiber.Ctx) error {
	var req system_payload.CreateAIModelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if strings.TrimSpace(req.PlatformID) == "" || strings.TrimSpace(req.ModelName) == "" {
		return response.BadRequestCtx(c, "平台 ID 和模型名称不能为空")
	}

	model := &system_db.SysModel{
		PlatformID:  strings.TrimSpace(req.PlatformID),
		ModelName:   strings.TrimSpace(req.ModelName),
		DisplayName: strings.TrimSpace(req.DisplayName),
		Status:      "active",
	}

	if err := h.repo.CreateModel(c.Context(), model); err != nil {
		return response.InternalServerCtx(c, "创建模型失败")
	}
	return response.SuccessCtx(c, model)
}

func (h *ModelHandler) UpdateModel(c *fiber.Ctx) error {
	id := c.Params("id")
	var req system_payload.UpdateAIModelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if req.Status != nil && *req.Status != "active" && *req.Status != "inactive" {
		return response.BadRequestCtx(c, "Invalid status, must be 'active' or 'inactive'")
	}

	model, err := h.repo.UpdateModel(c.Context(), id, system_repository.ModelUpdate{
		Status:      req.Status,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		return response.InternalServerCtx(c, "更新模型失败")
	}
	return response.SuccessCtx(c, model)
}

func (h *ModelHandler) DeleteModel(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.DeleteModel(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "删除模型失败")
	}
	return response.SuccessCtx(c, "模型删除成功")
}
