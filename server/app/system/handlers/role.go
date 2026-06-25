package handlers

import (
	"github.com/gofiber/fiber/v2"

	system_db "sag-wiki/app/system/models/db"
	system_payload "sag-wiki/app/system/models/payload"
	system_repo "sag-wiki/app/system/repository"
	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
)

// 角色处理器
type RoleHandler struct {
	repo system_repo.RoleRepository
}

// 创建角色处理器
func NewRoleHandler(dbService *database.DatabaseService) *RoleHandler {
	return &RoleHandler{
		repo: system_repo.NewRoleRepository(dbService.GetDB()),
	}
}

// 获取角色列表
func (h *RoleHandler) FindPage(c *fiber.Ctx) error {
	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	req.Validate()
	offset := req.GetOffset()

	roles, total, err := h.repo.FindPage(c.Context(), offset, req.PageSize)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch roles")
	}

	return response.PaginateCtx(c, roles, total, req.Page, req.PageSize)
}

// 获取角色详情
func (h *RoleHandler) FindOne(c *fiber.Ctx) error {
	id := c.Params("id")

	role, err := h.repo.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "Role not found")
	}

	return response.SuccessCtx(c, role)
}

// 创建角色
func (h *RoleHandler) Create(c *fiber.Ctx) error {
	var req system_payload.CreateRoleRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	role := &system_db.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.repo.Create(c.Context(), role); err != nil {
		return response.InternalServerCtx(c, "Failed to create role")
	}

	return response.SuccessCtx(c, role)
}

// 更新角色
func (h *RoleHandler) Update(c *fiber.Ctx) error {
	var req system_payload.UpdateRoleRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	// 检查角色是否存在
	role, err := h.repo.FindOne(c.Context(), req.ID)
	if err != nil {
		return response.NotFoundCtx(c, "Role not found")
	}

	// 更新字段
	role.Name = req.Name
	role.Description = req.Description

	if err := h.repo.Update(c.Context(), role); err != nil {
		return response.InternalServerCtx(c, "Failed to update role")
	}

	return response.SuccessCtx(c, role)
}

// 删除角色
func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.Delete(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "Failed to delete role")
	}

	return response.SuccessMsgCtx(c, "Role deleted successfully")
}
