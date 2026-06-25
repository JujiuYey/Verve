package handlers

import (
	"github.com/gofiber/fiber/v2"

	system_db "sag-wiki/app/system/models/db"
	system_payload "sag-wiki/app/system/models/payload"
	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"

	system_repo "sag-wiki/app/system/repository"
)

// 部门处理器
type DepartmentHandler struct {
	repo     system_repo.DepartmentRepository
	userRepo *system_repo.UserRepository
}

// 创建部门处理器
func NewDepartmentHandler(dbService *database.DatabaseService) *DepartmentHandler {
	db := dbService.GetDB()
	return &DepartmentHandler{
		repo:     system_repo.NewDepartmentRepository(db),
		userRepo: system_repo.NewUserRepository(db),
	}
}

// 获取部门列表
func (h *DepartmentHandler) FindPage(c *fiber.Ctx) error {
	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	req.Validate()
	offset := req.GetOffset()

	departments, total, err := h.repo.FindPage(c.Context(), offset, req.PageSize)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch departments")
	}

	return response.PaginateCtx(c, departments, total, req.Page, req.PageSize)
}

// 获取部门选项
func (h *DepartmentHandler) FindOptions(c *fiber.Ctx) error {
	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	req.Validate()

	departments, err := h.repo.FindOptions(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch departments")
	}

	return response.SuccessCtx(c, departments)
}

// 获取部门详情
func (h *DepartmentHandler) FindOne(c *fiber.Ctx) error {
	id := c.Params("id")

	department, err := h.repo.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "Department not found")
	}

	return response.SuccessCtx(c, department)
}

// 创建部门
func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	var req system_payload.CreateDepartmentRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	department := &system_db.Department{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
	}

	if err := h.repo.Create(c.Context(), department); err != nil {
		return response.InternalServerCtx(c, "Failed to create department")
	}

	return response.SuccessCtx(c, department)
}

// 更新部门
func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	var req system_payload.UpdateDepartmentRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	// 检查部门是否存在
	department, err := h.repo.FindOne(c.Context(), req.ID)
	if err != nil {
		return response.NotFoundCtx(c, "Department not found")
	}

	// 更新字段
	department.Name = req.Name
	department.Description = req.Description
	department.ParentID = req.ParentID

	if err := h.repo.Update(c.Context(), department); err != nil {
		return response.InternalServerCtx(c, "Failed to update department")
	}

	return response.SuccessCtx(c, department)
}

// 删除部门
func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.Delete(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "Failed to delete department")
	}

	return response.SuccessMsgCtx(c, "Department deleted successfully")
}

// 获取部门树
func (h *DepartmentHandler) FindTree(c *fiber.Ctx) error {
	departments, err := h.repo.FindTree(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch department tree")
	}

	return response.SuccessCtx(c, departments)
}

// 获取部门下的用户列表
func (h *DepartmentHandler) FindUsers(c *fiber.Ctx) error {
	id := c.Params("id")

	users, err := h.userRepo.FindByDepartmentID(c.Context(), id)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch users")
	}

	return response.SuccessCtx(c, users)
}

// 搜索用户和部门
func (h *DepartmentHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return response.BadRequestCtx(c, "Search query is required")
	}

	departments, err := h.repo.SearchDepartments(c.Context(), query)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to search departments")
	}

	users, err := h.userRepo.SearchUsers(c.Context(), query)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to search users")
	}

	result := map[string]interface{}{
		"departments": departments,
		"users":       users,
	}

	return response.SuccessCtx(c, result)
}
