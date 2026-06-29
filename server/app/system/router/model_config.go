package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "sag-wiki/app/system/handlers"
	"sag-wiki/infrastructure/database"
)

func SetupModelConfigRoutes(router fiber.Router, dbService *database.DatabaseService) {
	modelConfigHandler := system_handlers.NewModelConfigHandler(dbService)

	modelConfig := router.Group("/system/model-config")
	{
		modelConfig.Get("/platforms", modelConfigHandler.FindPlatforms)
		modelConfig.Post("/platforms", modelConfigHandler.CreatePlatform)
		modelConfig.Put("/platforms/:id/config", modelConfigHandler.UpdatePlatformConfig)
		modelConfig.Delete("/platforms/:id", modelConfigHandler.DeletePlatform)
		modelConfig.Post("/platforms/:id/sync-models", modelConfigHandler.SyncModels)
		modelConfig.Get("/models", modelConfigHandler.FindModels)
		modelConfig.Post("/models", modelConfigHandler.CreateModel)
		modelConfig.Put("/models/:id", modelConfigHandler.UpdateModel)
		modelConfig.Delete("/models/:id", modelConfigHandler.DeleteModel)
		modelConfig.Get("/list", modelConfigHandler.FindList)
		modelConfig.Post("/", modelConfigHandler.Create)
		modelConfig.Put("/", modelConfigHandler.Update)
		modelConfig.Delete("/:id", modelConfigHandler.Delete)
		modelConfig.Post("/:id/default", modelConfigHandler.SetDefault)
	}
}
