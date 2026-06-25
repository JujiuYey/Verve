package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	system_repo "sag-wiki/app/system/repository"
	wiki_db "sag-wiki/app/wiki/models/db"
	wiki_repo "sag-wiki/app/wiki/repository"
	"sag-wiki/infrastructure/database"

	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
)

// FolderTreeNode 树形节点结构
type FolderTreeNode struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	ParentID    *string           `json:"parent_id,omitempty"`
	UserID      *string           `json:"user_id,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	HasChildren bool              `json:"hasChildren"`
	Children    []*FolderTreeNode `json:"children"`
}

// 文件夹处理器
type FolderHandler struct {
	repo           wiki_repo.FolderRepository
	permissionRepo wiki_repo.FolderPermissionRepository
	userRepo       *system_repo.UserRepository
	deptRepo       system_repo.DepartmentRepository
	db             *bun.DB
}

// 创建文件夹处理器
func NewFolderHandler(dbService *database.DatabaseService) *FolderHandler {
	db := dbService.GetDB()
	return &FolderHandler{
		repo:           wiki_repo.NewFolderRepository(db),
		permissionRepo: wiki_repo.NewFolderPermissionRepository(db),
		userRepo:       system_repo.NewUserRepository(db),
		deptRepo:       system_repo.NewDepartmentRepository(db),
		db:             db,
	}
}

// 获取文件夹列表（分页）
func (h *FolderHandler) FindPage(c *fiber.Ctx) error {
	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	req.Validate()
	offset := req.GetOffset()

	filters := make(map[string]interface{})
	if parentID := c.Query("parent_id"); parentID != "" {
		filters["parent_id"] = parentID
	} else {
		filters["root"] = true
	}

	folders, total, err := h.repo.Page(c.Context(), offset, req.PageSize, filters)
	if err != nil {
		return response.InternalServerCtx(c, "获取文件夹列表失败")
	}

	return response.PaginateCtx(c, folders, total, req.Page, req.PageSize)
}

// 获取文件夹列表（不分页）
func (h *FolderHandler) FindList(c *fiber.Ctx) error {
	filters := make(map[string]interface{})
	if parentID := c.Query("parent_id"); parentID != "" {
		filters["parent_id"] = parentID
	} else {
		filters["root"] = true
	}

	folders, err := h.repo.List(c.Context(), filters)
	if err != nil {
		return response.InternalServerCtx(c, "获取文件夹列表失败")
	}

	return response.SuccessCtx(c, folders)
}

// 获取文件夹树形结构
func (h *FolderHandler) GetTree(c *fiber.Ctx) error {
	folders, err := h.repo.GetAll(c.Context())
	if err != nil {
		return response.InternalServerCtx(c, "获取文件夹树失败")
	}

	// 构建树形结构
	tree := h.buildTree(folders)

	return response.SuccessCtx(c, tree)
}

// buildTree 构建树形结构
func (h *FolderHandler) buildTree(folders []*wiki_db.Folder) []*FolderTreeNode {
	// 创建 ID 到文件夹的映射
	folderMap := make(map[string]*wiki_db.Folder)
	for _, f := range folders {
		folderMap[f.ID] = f
	}

	// 创建节点映射
	nodeMap := make(map[string]*FolderTreeNode)
	for _, f := range folders {
		nodeMap[f.ID] = &FolderTreeNode{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Description,
			ParentID:    f.ParentID,
			UserID:      f.UserID,
			CreatedAt:   f.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   f.UpdatedAt.Format("2006-01-02 15:04:05"),
			HasChildren: false,
			Children:    make([]*FolderTreeNode, 0),
		}
	}

	// 构建父子关系
	var roots []*FolderTreeNode
	for _, f := range folders {
		node := nodeMap[f.ID]
		if f.ParentID != nil && *f.ParentID != "" {
			// 有父节点，添加到父节点的 children
			if parentNode, ok := nodeMap[*f.ParentID]; ok {
				parentNode.Children = append(parentNode.Children, node)
				parentNode.HasChildren = true
			} else {
				// 父节点不存在，当作根节点
				roots = append(roots, node)
			}
		} else {
			// 根节点
			roots = append(roots, node)
		}
	}

	return roots
}

// 获取文件夹详情
func (h *FolderHandler) FindOne(c *fiber.Ctx) error {
	id := c.Params("id")

	folder, err := h.repo.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "文件夹不存在")
	}

	return response.SuccessCtx(c, folder)
}

// 创建文件夹
func (h *FolderHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		ParentID    *string `json:"parent_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	userIDStr := userID.(string)

	folder := &wiki_db.Folder{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		UserID:      &userIDStr,
		CreatedBy:   &userIDStr,
		UpdatedBy:   &userIDStr,
	}

	if err := h.repo.Create(c.Context(), folder); err != nil {
		return response.InternalServerCtx(c, "创建文件夹失败: "+err.Error())
	}

	return response.SuccessCtx(c, folder)
}

// 更新文件夹
func (h *FolderHandler) Update(c *fiber.Ctx) error {
	var req struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		ParentID    *string `json:"parent_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	folder, err := h.repo.FindOne(c.Context(), req.ID)
	if err != nil {
		return response.NotFoundCtx(c, "文件夹不存在")
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	userIDStr := userID.(string)

	folder.Name = req.Name
	folder.Description = req.Description
	folder.ParentID = req.ParentID
	folder.UpdatedBy = &userIDStr

	if err := h.repo.Update(c.Context(), folder); err != nil {
		return response.InternalServerCtx(c, "更新文件夹失败: "+err.Error())
	}

	return response.SuccessCtx(c, folder)
}

// 删除文件夹
func (h *FolderHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.repo.Delete(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "删除文件夹失败: "+err.Error())
	}

	return response.SuccessMsgCtx(c, "文件夹删除成功")
}

// SelectedItem 表示选中的用户或部门
type SelectedItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "user" or "department"
}

// 获取文件夹权限
func (h *FolderHandler) GetPermissions(c *fiber.Ctx) error {
	id := c.Params("id")

	permissions, err := h.permissionRepo.FindByFolderID(c.Context(), id)
	if err != nil {
		return response.InternalServerCtx(c, "获取权限失败: "+err.Error())
	}

	// 转换为 SelectedItem 格式
	var items []SelectedItem
	for _, perm := range permissions {
		if perm.UserID != nil {
			user, err := h.userRepo.FindOne(c.Context(), *perm.UserID)
			if err == nil {
				name := user.Username
				if user.FullName != nil && *user.FullName != "" {
					name = *user.FullName
				}
				items = append(items, SelectedItem{
					ID:   user.ID,
					Name: name,
					Type: "user",
				})
			}
		} else if perm.DepartmentID != nil {
			dept, err := h.deptRepo.FindOne(c.Context(), *perm.DepartmentID)
			if err == nil {
				items = append(items, SelectedItem{
					ID:   dept.ID,
					Name: dept.Name,
					Type: "department",
				})
			}
		}
	}

	return response.SuccessCtx(c, items)
}

// SetPermissionsRequest 设置权限请求
type SetPermissionsRequest struct {
	Items []SelectedItem `json:"items"`
}

// 设置文件夹权限
func (h *FolderHandler) SetPermissions(c *fiber.Ctx) error {
	id := c.Params("id")

	var req SetPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	// 使用事务
	err := h.db.RunInTx(c.Context(), nil, func(ctx context.Context, tx bun.Tx) error {
		// 删除现有权限
		if err := h.permissionRepo.DeleteByFolderID(ctx, id); err != nil {
			return err
		}

		// 批量插入新权限
		var permissions []*wiki_db.FolderPermission
		for _, item := range req.Items {
			perm := &wiki_db.FolderPermission{
				FolderID:       id,
				PermissionType: wiki_db.PermissionRead,
			}
			if item.Type == "user" {
				perm.UserID = &item.ID
			} else if item.Type == "department" {
				perm.DepartmentID = &item.ID
			}
			permissions = append(permissions, perm)
		}

		if err := h.permissionRepo.BatchCreate(ctx, permissions); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return response.InternalServerCtx(c, "设置权限失败: "+err.Error())
	}

	return response.SuccessMsgCtx(c, "权限设置成功")
}
