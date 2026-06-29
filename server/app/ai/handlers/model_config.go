package handlers

import (
	"fmt"
	"log"
	"strings"

	"sag-wiki/app/ai/models/db"
	"sag-wiki/app/ai/models/payload"
	"sag-wiki/app/ai/repository"
	"sag-wiki/app/ai/service"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"

	"github.com/gofiber/fiber/v2"
)

type ModelConfigHandler struct {
	repo    repository.ModelConfigRepository
	syncSvc *service.ModelSyncService
}

func NewModelConfigHandler(dbService *database.DatabaseService) *ModelConfigHandler {
	repo := repository.NewModelConfigRepository(dbService.GetDB())
	return &ModelConfigHandler{
		repo:    repo,
		syncSvc: service.NewModelSyncService(repo),
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
	if !isValidModelType(req.ModelType) {
		return response.BadRequestCtx(c, "Invalid model_type, must be 'chat', 'embedding' or 'rerank'")
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
	if !isValidModelType(req.ModelType) {
		return response.BadRequestCtx(c, "Invalid model_type, must be 'chat', 'embedding' or 'rerank'")
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

func (h *ModelConfigHandler) FindPlatforms(c *fiber.Ctx) error {
	platforms, err := h.repo.FindPlatforms(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "获取模型平台失败")
	}
	return response.SuccessCtx(c, platforms)
}

func (h *ModelConfigHandler) CreatePlatform(c *fiber.Ctx) error {
	var req payload.CreateAIPlatformRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if !isSupportedProviderType(req.ProviderType) {
		return response.BadRequestCtx(c, "当前仅支持 OpenAI 兼容接口")
	}
	if strings.TrimSpace(req.Name) == "" {
		return response.BadRequestCtx(c, "平台名称不能为空")
	}
	if strings.TrimSpace(req.BaseURL) == "" {
		return response.BadRequestCtx(c, "API 地址不能为空")
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return response.BadRequestCtx(c, "API Key 不能为空")
	}

	platform := &db.SysModelPlatform{
		Name:             strings.TrimSpace(req.Name),
		ProviderType:     req.ProviderType,
		DefaultBaseURL:   strings.TrimSpace(req.DefaultBaseURL),
		BaseURL:          strings.TrimSpace(req.BaseURL),
		APIKeyCiphertext: strings.TrimSpace(req.APIKey),
		ModelListPath:    req.ModelListPath,
		AuthScheme:       req.AuthScheme,
		DocsURL:          req.DocsURL,
		APIKeyURL:        req.APIKeyURL,
	}
	if platform.DefaultBaseURL == "" {
		platform.DefaultBaseURL = platform.BaseURL
	}

	if err := h.repo.CreatePlatform(c.Context(), platform); err != nil {
		return response.InternalServerCtx(c, "创建模型平台失败")
	}
	return response.SuccessCtx(c, platform)
}

func (h *ModelConfigHandler) UpdatePlatformConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	var req payload.UpdateAIPlatformConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if strings.TrimSpace(req.BaseURL) == "" {
		return response.BadRequestCtx(c, "API 地址不能为空")
	}

	platform, err := h.repo.UpdatePlatformConfig(
		c.Context(),
		id,
		strings.TrimSpace(req.BaseURL),
		req.APIKey,
		req.ClearAPIKey,
	)
	if err != nil {
		return response.InternalServerCtx(c, "保存模型平台配置失败")
	}
	return response.SuccessCtx(c, platform)
}

func (h *ModelConfigHandler) DeletePlatform(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.DeletePlatform(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "删除模型平台失败")
	}
	return response.SuccessCtx(c, "模型平台删除成功")
}

func (h *ModelConfigHandler) SyncModels(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.syncSvc.SyncModels(c.Context(), id)
	if err != nil {
		return response.InternalServerCtx(c, "同步模型失败: "+err.Error())
	}
	return response.SuccessCtx(c, result)
}

func (h *ModelConfigHandler) FindModels(c *fiber.Ctx) error {
	models, err := h.repo.FindModels(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "获取模型列表失败")
	}
	return response.SuccessCtx(c, models)
}

func (h *ModelConfigHandler) CreateModel(c *fiber.Ctx) error {
	var req payload.CreateAIModelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if strings.TrimSpace(req.PlatformID) == "" || strings.TrimSpace(req.ModelName) == "" {
		return response.BadRequestCtx(c, "平台 ID 和模型名称不能为空")
	}
	if !isValidModelType(req.ModelType) {
		return response.BadRequestCtx(c, "Invalid model_type, must be 'chat', 'embedding' or 'rerank'")
	}

	model := &db.SysModel{
		PlatformID:   strings.TrimSpace(req.PlatformID),
		ModelName:    strings.TrimSpace(req.ModelName),
		DisplayName:  strings.TrimSpace(req.DisplayName),
		ModelType:    req.ModelType,
		Capabilities: req.Capabilities,
		Source:       req.Source,
		Status:       "active",
		IsDefault:    req.IsDefault,
		Temperature:  req.Temperature,
		TopP:         req.TopP,
		MaxTokens:    req.MaxTokens,
		TopK:         req.TopK,
	}

	if err := h.repo.CreateModel(c.Context(), model); err != nil {
		return response.InternalServerCtx(c, "创建模型失败")
	}
	return response.SuccessCtx(c, model)
}

func (h *ModelConfigHandler) UpdateModel(c *fiber.Ctx) error {
	id := c.Params("id")
	var req payload.UpdateAIModelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if req.Status != nil && *req.Status != "active" && *req.Status != "inactive" {
		return response.BadRequestCtx(c, "Invalid status, must be 'active' or 'inactive'")
	}

	model, err := h.repo.UpdateModel(c.Context(), id, repository.ModelUpdate{
		Status:       req.Status,
		DisplayName:  req.DisplayName,
		Capabilities: req.Capabilities,
	})
	if err != nil {
		return response.InternalServerCtx(c, "更新模型失败")
	}
	return response.SuccessCtx(c, model)
}

func (h *ModelConfigHandler) DeleteModel(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.DeleteModel(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "删除模型失败")
	}
	return response.SuccessCtx(c, "模型删除成功")
}

func isValidModelType(modelType string) bool {
	return modelType == db.ModelTypeChat || modelType == db.ModelTypeEmbedding || modelType == "rerank"
}

func isSupportedProviderType(providerType string) bool {
	return providerType == "" || providerType == "openai_compatible"
}
