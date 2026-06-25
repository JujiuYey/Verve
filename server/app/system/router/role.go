package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "sag-wiki/app/system/handlers"
	"sag-wiki/infrastructure/database"
)

// 配置角色管理路由
func SetupRoleRoutes(router fiber.Router, dbService *database.DatabaseService) {
	roleHandler := system_handlers.NewRoleHandler(dbService)

	role := router.Group("/system/role")
	{
		// 获取角色列表
		role.Get("/page", roleHandler.FindPage)
		// 获取角色详情
		role.Get("/:id", roleHandler.FindOne)
		// 创建角色
		role.Post("/", roleHandler.Create)
		// 更新角色
		role.Put("/", roleHandler.Update)
		// 删除角色
		role.Delete("/:id", roleHandler.Delete)
	}
}
