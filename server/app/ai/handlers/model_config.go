package handlers

import (
	"fmt"
	"log"
	"sag-wiki/app/ai/models/db"
	"sag-wiki/app/ai/models/payload"
	"sag-wiki/app/ai/repository"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"

	"github.com/gofiber/fiber/v2"
)

type ModelConfigHandler struct {
	repo repository.ModelConfigRepository
}

func NewModelConfigHandler(dbService *database.DatabaseService) *ModelConfigHandler {
	return &ModelConfigHandler{
		repo: repository.NewModelConfigRepository(dbService.GetDB()),
	}
}

// 获取模型配置列表
func (h *ModelConfigHandler) FindList(c *fiber.Ctx) error {
	configs, err := h.repo.FindList(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch model configs")
	}

	return response.SuccessCtx(c, configs)
}

// 创建模型配置
func (h *ModelConfigHandler) Create(c *fiber.Ctx) error {
	var req payload.CreateModelConfigRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	// 验证模型类型
	if req.ModelType != db.ModelTypeChat && req.ModelType != db.ModelTypeEmbedding {
		return response.BadRequestCtx(c, "Invalid model_type, must be 'chat' or 'embedding'")
	}

	modelConfig := &db.ModelConfig{
		Vendor:      req.Vendor,
		Name:        req.Name,
		APIKey:      req.APIKey,
		BaseURL:     req.BaseURL,
		ModelType:   req.ModelType,
		Model:       req.Model,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		TopK:        req.TopK,
		IsActive:    req.IsActive,
		IsDefault:   req.IsDefault,
	}

	if err := h.repo.Create(c.Context(), modelConfig); err != nil {
		return response.InternalServerCtx(c, "Failed to create model config")
	}

	return response.SuccessCtx(c, modelConfig)
}

// 更新模型配置
func (h *ModelConfigHandler) Update(c *fiber.Ctx) error {
	var req payload.UpdateModelConfigRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	existingConfig, err := h.repo.FindOne(c.Context(), req.ID)
	if err != nil {
		return response.NotFoundCtx(c, "Model config not found")
	}

	// 验证模型类型
	if req.ModelType != db.ModelTypeChat && req.ModelType != db.ModelTypeEmbedding {
		return response.BadRequestCtx(c, "Invalid model_type, must be 'chat' or 'embedding'")
	}

	existingConfig.Vendor = req.Vendor
	existingConfig.Name = req.Name
	existingConfig.APIKey = req.APIKey
	existingConfig.BaseURL = req.BaseURL
	existingConfig.ModelType = req.ModelType
	existingConfig.Model = req.Model
	existingConfig.Temperature = req.Temperature
	existingConfig.TopP = req.TopP
	existingConfig.MaxTokens = req.MaxTokens
	existingConfig.TopK = req.TopK
	existingConfig.IsActive = req.IsActive
	existingConfig.IsDefault = req.IsDefault

	if err := h.repo.Update(c.Context(), existingConfig); err != nil {
		return response.InternalServerCtx(c, "Failed to update model config")
	}

	return response.SuccessCtx(c, existingConfig)
}

// 删除模型配置
func (h *ModelConfigHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.Delete(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "Failed to delete model config")
	}

	return response.SuccessCtx(c, "Model config deleted successfully")
}

// 设置默认模型配置
func (h *ModelConfigHandler) SetDefault(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.SetDefault(c.Context(), id); err != nil {
		log.Printf("SetDefault error: %v", err)
		return response.InternalServerCtx(c,
			fmt.Sprintf("设置默认模型配置失败: %v", err))
	}

	return response.SuccessCtx(c, "Default model config set successfully")
}
