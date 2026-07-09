package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	wiki_payload "verve/app/wiki/models/payload"
	wiki_repo "verve/app/wiki/repository"
	"verve/common/response"
	"verve/infrastructure/database"
)

type AgentInstanceHandler struct {
	agents  *wiki_repo.AgentInstanceRepository
	folders wiki_repo.FolderRepository
}

func NewAgentInstanceHandler(dbService *database.DatabaseService) *AgentInstanceHandler {
	return &AgentInstanceHandler{
		agents:  wiki_repo.NewAgentInstanceRepository(dbService.GetDB()),
		folders: wiki_repo.NewFolderRepository(dbService.GetDB()),
	}
}

func (h *AgentInstanceHandler) FindByRoot(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	rootFolderID := strings.TrimSpace(c.Query("root_folder_id"))
	if rootFolderID == "" {
		return response.BadRequestCtx(c, "缺少知识库根目录")
	}
	instance, err := h.agents.FindByRoot(c.Context(), userID, rootFolderID)
	if err != nil {
		return response.NotFoundCtx(c, "Wiki Agent 尚未创建")
	}
	return response.SuccessCtx(c, instance)
}

func (h *AgentInstanceHandler) Ensure(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	var req wiki_payload.EnsureAgentInstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, "参数错误: "+err.Error())
	}
	rootFolderID := strings.TrimSpace(req.RootFolderID)
	if rootFolderID == "" {
		return response.BadRequestCtx(c, "缺少知识库根目录")
	}
	rootFolder, err := h.folders.FindOne(c.Context(), rootFolderID)
	if err != nil {
		return response.NotFoundCtx(c, "知识库根目录不存在")
	}
	if rootFolder.ParentID != nil && strings.TrimSpace(*rootFolder.ParentID) != "" {
		return response.BadRequestCtx(c, "只能为知识库根目录创建 Agent")
	}
	if rootFolder.UserID != nil && *rootFolder.UserID != "" && *rootFolder.UserID != userID {
		return response.NotFoundCtx(c, "知识库根目录不存在")
	}
	instance, err := h.agents.EnsureByRoot(c.Context(), userID, rootFolder, req.Name, req.Description)
	if err != nil {
		return response.InternalServerCtx(c, "创建 Wiki Agent 失败: "+err.Error())
	}
	return response.SuccessCtx(c, instance)
}

func currentUserID(c *fiber.Ctx) (string, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", false
	}
	userIDStr, ok := userID.(string)
	if !ok || strings.TrimSpace(userIDStr) == "" {
		return "", false
	}
	return userIDStr, true
}
