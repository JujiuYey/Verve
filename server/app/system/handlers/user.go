package handlers

import (
	"fmt"
	"path/filepath"

	"github.com/gofiber/fiber/v2"

	system_db "sag-wiki/app/system/models/db"
	system_payload "sag-wiki/app/system/models/payload"
	system_repo "sag-wiki/app/system/repository"
	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
	"sag-wiki/infrastructure/storage"
	"sag-wiki/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// 用户处理器
type UserHandler struct {
	repo         *system_repo.UserRepository
	minioService *storage.MinIOService
}

// 创建用户处理器
func NewUserHandler(dbService *database.DatabaseService, minioService *storage.MinIOService) *UserHandler {
	return &UserHandler{
		repo:         system_repo.NewUserRepository(dbService.GetDB()),
		minioService: minioService,
	}
}

// 获取用户列表
func (h *UserHandler) FindPage(c *fiber.Ctx) error {
	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	req.Validate()
	offset := req.GetOffset()

	search := c.Query("search")

	users, total, err := h.repo.FindPage(c.Context(), offset, req.PageSize, search)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch users")
	}

	// 不返回密码，并添加部门和角色信息
	userResponses := make([]map[string]interface{}, len(users))
	for i, user := range users {
		// 获取用户的部门和角色
		departments, primaryDepartmentID, _ := h.repo.GetUserDepartments(c.Context(), user.ID)
		roles, _ := h.repo.GetUserRolesWithDetails(c.Context(), user.ID)

		userResponses[i] = map[string]interface{}{
			"id":                    user.ID,
			"username":              user.Username,
			"email":                 user.Email,
			"full_name":             user.FullName,
			"avatar":                user.Avatar,
			"status":                user.Status,
			"created_at":            user.CreatedAt,
			"updated_at":            user.UpdatedAt,
			"departments":           departments,
			"roles":                 roles,
			"primary_department_id": primaryDepartmentID,
		}
	}

	return response.PaginateCtx(c, userResponses, total, req.Page, req.PageSize)
}

// 获取用户详情
func (h *UserHandler) FindOne(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := h.repo.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "User not found")
	}

	// 获取用户的部门和角色
	_, primaryDepartmentID, err := h.repo.GetUserDepartments(c.Context(), user.ID)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch user departments")
	}

	roles, err := h.repo.GetUserRolesWithDetails(c.Context(), user.ID)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to fetch user roles")
	}

	// 🔑 从 roles 中提取 roleIDs
	var roleIDs []string
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}

	// 不返回密码
	userResponse := system_payload.UserInfo{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		FullName:            user.FullName,
		Avatar:              user.Avatar,
		Status:              user.Status,
		Roles:               roleIDs,
		PrimaryDepartmentID: primaryDepartmentID,
	}

	return response.SuccessCtx(c, userResponse)
}

// 获取当前登录用户信息
func (h *UserHandler) FindDetail(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.UnauthorizedCtx(c, "未授权")
	}

	return h.FindOne(c)
}

// 创建用户
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req system_payload.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	// 检查用户名是否已存在
	existingUser, err := h.repo.FindByUsername(c.Context(), req.Username)
	if err == nil && existingUser != nil {
		return response.FailWithCodeCtx(c, fiber.StatusConflict, "用户名已存在")
	}

	// 使用默认密码并进行哈希处理
	password := utils.GetEnv("DEFAULT_PASSWORD", "123456")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to hash password")
	}

	user := &system_db.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Avatar:   req.Avatar,
		Status:   "active",
	}

	// 开始事务
	tx, err := h.repo.GetDB().BeginTx(c.Context(), nil)
	if err != nil {
		return response.InternalServerCtx(c, "开始事务失败")
	}
	defer tx.Rollback()

	// 创建用户
	if err := h.repo.Create(c.Context(), user); err != nil {
		return response.InternalServerCtx(c, "创建用户失败")
	}

	// 设置默认部门
	if err := h.repo.SetDefaultDepartment(c.Context(), user.ID, req.PrimaryDepartmentID); err != nil {
		return response.InternalServerCtx(c, "设置默认部门失败")
	}

	// 创建用户角色关联
	if len(req.RoleIDs) > 0 {
		if err := h.repo.CreateUserRoles(c.Context(), user.ID, req.RoleIDs); err != nil {
			return response.InternalServerCtx(c, "创建用户角色失败")
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return response.InternalServerCtx(c, "提交事务失败")
	}

	return response.SuccessMsgCtx(c, "用户创建成功")
}

// 更新用户
func (h *UserHandler) Update(c *fiber.Ctx) error {
	var req system_payload.UpdateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	// 检查用户是否存在
	user, err := h.repo.FindOne(c.Context(), req.ID)
	if err != nil {
		return response.NotFoundCtx(c, "User not found")
	}

	// 更新字段
	user.Email = req.Email
	user.FullName = req.FullName
	user.Avatar = req.Avatar
	user.Status = req.Status

	// 开始事务
	tx, err := h.repo.GetDB().BeginTx(c.Context(), nil)
	if err != nil {
		return response.InternalServerCtx(c, "Failed to start transaction")
	}
	defer tx.Rollback()

	// 更新用户基本信息
	if err := h.repo.Update(c.Context(), user); err != nil {
		return response.InternalServerCtx(c, "Failed to update user")
	}

	// 更新默认部门
	if req.PrimaryDepartmentID != "" {
		if err := h.repo.SetDefaultDepartment(c.Context(), user.ID, req.PrimaryDepartmentID); err != nil {
			return response.InternalServerCtx(c, "Failed to set default department")
		}
	}

	// 更新角色关联
	if req.RoleIDs != nil {
		// 删除旧的角色关联
		if err := h.repo.DeleteUserRoles(c.Context(), user.ID); err != nil {
			return response.InternalServerCtx(c, "Failed to delete old roles")
		}
		// 创建新的角色关联
		if len(req.RoleIDs) > 0 {
			if err := h.repo.CreateUserRoles(c.Context(), user.ID, req.RoleIDs); err != nil {
				return response.InternalServerCtx(c, "Failed to create user roles")
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return response.InternalServerCtx(c, "Failed to commit transaction")
	}

	return response.SuccessMsgCtx(c, "User updated successfully")
}

// 删除用户
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	// 检查用户是否存在
	_, err := h.repo.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "User not found")
	}

	if err := h.repo.Delete(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "Failed to delete user")
	}

	return response.SuccessMsgCtx(c, "User deleted successfully")
}

// 修改当前登录用户的密码
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}

	var req system_payload.ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, "请求参数错误: "+err.Error())
	}

	// 查询用户是否存在
	user, err := h.repo.FindOne(c.Context(), userID.(string))
	if err != nil {
		return response.NotFoundCtx(c, "用户不存在")
	}

	// 验证旧密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return response.BadRequestCtx(c, "旧密码不正确")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return response.InternalServerCtx(c, "密码加密失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := h.repo.Update(c.Context(), user); err != nil {
		return response.InternalServerCtx(c, "密码更新失败")
	}

	return response.SuccessMsgCtx(c, "密码修改成功")
}

// 上传头像到 MinIO
func (h *UserHandler) UploadAvatar(c *fiber.Ctx) error {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequestCtx(c, "文件上传失败: "+err.Error())
	}

	// 验证文件类型（只允许图片）
	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		return response.BadRequestCtx(c, "只支持上传图片格式 (jpg, png, gif, webp)")
	}

	// 验证文件大小（限制为 5MB）
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if file.Size > maxSize {
		return response.BadRequestCtx(c, "头像文件大小不能超过 5MB")
	}

	// 生成唯一的文件路径：avatars/{uuid}/{filename}
	avatarID := uuid.New()
	fileExt := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("avatars/%s%s", avatarID.String(), fileExt)

	// 打开文件
	f, err := file.Open()
	if err != nil {
		return response.InternalServerCtx(c, "无法打开文件: "+err.Error())
	}
	defer f.Close()

	// 上传到 MinIO
	err = h.minioService.UploadFile(c.Context(), objectName, f, file.Size, contentType)
	if err != nil {
		return response.InternalServerCtx(c, "上传头像失败: "+err.Error())
	}

	// 返回文件路径给前端
	return c.JSON(fiber.Map{
		"message":   "头像上传成功",
		"file_path": objectName,
	})
}

func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}

	// 使用默认密码并进行哈希处理
	password := utils.GetEnv("DEFAULT_PASSWORD", "123456")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return response.InternalServerCtx(c, "密码哈希失败")
	}

	user, err := h.repo.FindOne(c.Context(), userID.(string))
	if err != nil {
		return response.NotFoundCtx(c, "用户不存在")
	}

	user.Password = string(hashedPassword)
	if err := h.repo.Update(c.Context(), user); err != nil {
		return response.InternalServerCtx(c, "密码更新失败")
	}

	return response.SuccessMsgCtx(c, "密码重置成功")
}
