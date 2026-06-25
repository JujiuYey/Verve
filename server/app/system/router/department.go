package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "sag-wiki/app/system/handlers"
	"sag-wiki/infrastructure/database"
)

// 配置部门管理路由
func SetupDepartmentRoutes(router fiber.Router, dbService *database.DatabaseService) {
	deptHandler := system_handlers.NewDepartmentHandler(dbService)

	department := router.Group("/system/department")
	{
		// 获取部门树
		department.Get("/tree", deptHandler.FindTree)
		// 获取部门分页列表
		department.Get("/page", deptHandler.FindPage)
		// 获取部门选项
		department.Get("/options", deptHandler.FindOptions)
		// 搜索用户和部门
		department.Get("/search", deptHandler.Search)
		// 获取部门详情
		department.Get("/:id", deptHandler.FindOne)
		// 获取部门下的用户列表
		department.Get("/:id/users", deptHandler.FindUsers)
		// 创建部门
		department.Post("/", deptHandler.Create)
		// 更新部门
		department.Put("/", deptHandler.Update)
		// 删除部门
		department.Delete("/:id", deptHandler.Delete)
	}
}
