package router

import (
	"github.com/gofiber/fiber/v2"

	ai_handlers "sag-wiki/app/ai/handlers"
	"sag-wiki/infrastructure/database"
)

// 配置模型配置路由
func SetupModelConfigRoutes(router fiber.Router, dbService *database.DatabaseService) {
	modelConfigHandler := ai_handlers.NewModelConfigHandler(dbService)

	modelConfig := router.Group("/ai/model-config")
	{
		// 获取模型配置列表
		modelConfig.Get("/list", modelConfigHandler.FindList)
		// 创建模型配置
		modelConfig.Post("/", modelConfigHandler.Create)
		// 更新模型配置
		modelConfig.Put("/", modelConfigHandler.Update)
		// 删除模型配置
		modelConfig.Delete("/:id", modelConfigHandler.Delete)
		// 设置默认模型配置
		modelConfig.Post("/:id/default", modelConfigHandler.SetDefault)
	}
}
