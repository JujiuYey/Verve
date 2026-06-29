package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	system_payload "sag-wiki/app/system/models/payload"
	system_repo "sag-wiki/app/system/repository"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
	"sag-wiki/utils"
)

// 认证处理器
type AuthHandler struct {
	userRepo *system_repo.UserRepository
}

// 创建认证处理器
func NewAuthHandler(dbService *database.DatabaseService) *AuthHandler {
	return &AuthHandler{
		userRepo: system_repo.NewUserRepository(dbService.GetDB()),
	}
}

// 用户登录
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req system_payload.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	ctx := c.Context()

	user, err := h.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.FailWithCodeCtx(c, 401, "用户不存在")
		}
		return response.InternalServerCtx(c, "数据库错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return response.FailWithCodeCtx(c, 401, "用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return response.FailWithCodeCtx(c, 403, "用户账号已禁用")
	}

	// to C 个人订阅,无角色体系(jwt 保留 roles 字段,传空)
	roles := []string{}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email, roles)
	if err != nil {
		return response.InternalServerCtx(c, "生成访问令牌失败")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return response.InternalServerCtx(c, "生成刷新令牌失败")
	}

	// 返回响应
	return response.SuccessCtx(c, system_payload.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: system_payload.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	})
}

// 刷新访问令牌
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req system_payload.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	userID, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return response.FailWithCodeCtx(c, 401, "刷新令牌无效")
	}

	// 获取用户信息
	ctx := c.Context()
	user, err := h.userRepo.FindOne(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.FailWithCodeCtx(c, 401, "用户不存在")
		}
		return response.InternalServerCtx(c, "数据库错误")
	}

	if user.Status != "active" {
		return response.FailWithCodeCtx(c, 403, "用户账号已禁用")
	}

	roles := []string{}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email, roles)
	if err != nil {
		return response.InternalServerCtx(c, "生成访问令牌失败")
	}

	return response.SuccessCtx(c, system_payload.RefreshTokenResponse{
		AccessToken: accessToken,
	})
}
