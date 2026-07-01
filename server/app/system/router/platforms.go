package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "verve/app/system/handlers"
	"verve/infrastructure/database"
)

// SetupPlatformRoutes 配置模型平台路由
func SetupPlatformRoutes(router fiber.Router, dbService *database.DatabaseService) {
	platformHandler := system_handlers.NewPlatformHandler(dbService)

	platforms := router.Group("/system/platforms")
	{
		platforms.Get("/", platformHandler.FindPlatforms)
		platforms.Post("/", platformHandler.CreatePlatform)
		platforms.Put("/:id/config", platformHandler.UpdatePlatformConfig)
		platforms.Delete("/:id", platformHandler.DeletePlatform)
		platforms.Post("/:id/sync-models", platformHandler.SyncModels)
	}
}
