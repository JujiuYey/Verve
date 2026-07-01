package handlers

import (
	"strings"

	system_db "sag-wiki/app/system/models/db"
	system_payload "sag-wiki/app/system/models/payload"
	system_repository "sag-wiki/app/system/repository"
	system_service "sag-wiki/app/system/service"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"

	"github.com/gofiber/fiber/v2"
)

// PlatformHandler 模型平台处理器
type PlatformHandler struct {
	repo    system_repository.ModelConfigRepository
	syncSvc *system_service.ModelSyncService
}

func NewPlatformHandler(dbService *database.DatabaseService) *PlatformHandler {
	repo := system_repository.NewModelConfigRepository(dbService.GetDB())
	return &PlatformHandler{
		repo:    repo,
		syncSvc: system_service.NewModelSyncService(repo),
	}
}

func (h *PlatformHandler) FindPlatforms(c *fiber.Ctx) error {
	platforms, err := h.repo.FindPlatforms(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "获取模型平台失败")
	}
	return response.SuccessCtx(c, platforms)
}

func (h *PlatformHandler) CreatePlatform(c *fiber.Ctx) error {
	var req system_payload.CreateAIPlatformRequest
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

	platform := &system_db.SysModelPlatform{
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

func (h *PlatformHandler) UpdatePlatformConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	var req system_payload.UpdateAIPlatformConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if strings.TrimSpace(req.BaseURL) == "" {
		return response.BadRequestCtx(c, "API 地址不能为空")
	}

	platform, err := h.repo.UpdatePlatformConfig(c.Context(), id, strings.TrimSpace(req.BaseURL), req.APIKey, req.ClearAPIKey)
	if err != nil {
		return response.InternalServerCtx(c, "保存模型平台配置失败")
	}
	return response.SuccessCtx(c, platform)
}

func (h *PlatformHandler) DeletePlatform(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.DeletePlatform(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "删除模型平台失败")
	}
	return response.SuccessCtx(c, "模型平台删除成功")
}

func (h *PlatformHandler) SyncModels(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.syncSvc.SyncModels(c.Context(), id)
	if err != nil {
		return response.InternalServerCtx(c, "同步模型失败: "+err.Error())
	}
	return response.SuccessCtx(c, result)
}

func isSupportedProviderType(providerType string) bool {
	return providerType == "" || providerType == "openai_compatible"
}