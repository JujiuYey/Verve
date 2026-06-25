package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "sag-wiki/app/system/handlers"
	"sag-wiki/infrastructure/database"
	"sag-wiki/infrastructure/storage"
)

// 配置用户管理路由
func SetupUserRoutes(router fiber.Router, dbService *database.DatabaseService, minioService *storage.MinIOService) {
	userHandler := system_handlers.NewUserHandler(dbService, minioService)

	user := router.Group("/system/user")
	{
		// 获取用户列表
		user.Get("/page", userHandler.FindPage)
		// 获取当前登录用户信息
		user.Get("/me", userHandler.FindDetail)
		// 获取用户详情
		user.Get("/:id", userHandler.FindOne)
		// 创建用户
		user.Post("/", userHandler.Create)
		// 更新用户
		user.Put("/", userHandler.Update)
		// 删除用户
		user.Delete("/:id", userHandler.Delete)
		// 修改密码
		user.Put("/change-password", userHandler.ChangePassword)
		// 重置密码
		user.Post("/reset-password", userHandler.ResetPassword)
		// 上传头像
		user.Post("/upload-avatar", userHandler.UploadAvatar)
	}
}
