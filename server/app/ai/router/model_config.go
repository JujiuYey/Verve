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
		// 获取模型平台列表
		modelConfig.Get("/platforms", modelConfigHandler.FindPlatforms)
		// 创建模型平台
		modelConfig.Post("/platforms", modelConfigHandler.CreatePlatform)
		// 更新模型平台配置
		modelConfig.Put("/platforms/:id/config", modelConfigHandler.UpdatePlatformConfig)
		// 删除模型平台
		modelConfig.Delete("/platforms/:id", modelConfigHandler.DeletePlatform)
		// 同步模型列表
		modelConfig.Post("/platforms/:id/sync-models", modelConfigHandler.SyncModels)
		// 获取模型列表
		modelConfig.Get("/models", modelConfigHandler.FindModels)
		// 创建模型
		modelConfig.Post("/models", modelConfigHandler.CreateModel)
		// 更新模型
		modelConfig.Put("/models/:id", modelConfigHandler.UpdateModel)
		// 删除模型
		modelConfig.Delete("/models/:id", modelConfigHandler.DeleteModel)
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
