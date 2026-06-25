package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "sag-wiki/app/system/handlers"
	"sag-wiki/infrastructure/database"
)

// 配置认证路由
func SetupAuthRoutes(router fiber.Router, dbService *database.DatabaseService) {
	authHandler := system_handlers.NewAuthHandler(dbService)

	// 公开路由（无需认证）
	auth := router.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)
		auth.Post("/refresh-token", authHandler.RefreshToken)
	}

	// 需要认证的路由
	protected := router.Group("")
	protected.Use(func(c *fiber.Ctx) error {
		// 这里可以添加认证检查，但实际在 middleware 中处理
		return c.Next()
	})
}
