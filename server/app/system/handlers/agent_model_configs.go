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

type AgentModelConfigHandler struct {
	repo system_repository.ModelConfigRepository
}

func NewAgentModelConfigHandler(dbService *database.DatabaseService) *AgentModelConfigHandler {
	return &AgentModelConfigHandler{repo: system_repository.NewModelConfigRepository(dbService.GetDB())}
}

func (h *AgentModelConfigHandler) FindConfigs(c *fiber.Ctx) error {
	configs, err := h.repo.FindAgentModelConfigs(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "获取 Agent 模型配置失败")
	}
	return response.SuccessCtx(c, configs)
}

func (h *AgentModelConfigHandler) UpsertConfig(c *fiber.Ctx) error {
	agentKey := strings.TrimSpace(c.Params("agentKey"))
	sceneKey := strings.TrimSpace(c.Params("sceneKey"))
	if agentKey == "" || sceneKey == "" {
		return response.BadRequestCtx(c, "Agent 和场景不能为空")
	}

	var req system_payload.UpsertAgentModelConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	if strings.TrimSpace(req.ModelID) == "" {
		return response.BadRequestCtx(c, "模型不能为空")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	config, err := h.repo.UpsertAgentModelConfig(c.Context(), &system_db.AgentModelConfig{
		AgentKey: agentKey,
		SceneKey: sceneKey,
		ModelID:  req.ModelID,
		Params:   req.Params,
		Enabled:  enabled,
	})
	if err != nil {
		return response.InternalServerCtx(c, "保存 Agent 模型配置失败")
	}
	return response.SuccessCtx(c, config)
}
