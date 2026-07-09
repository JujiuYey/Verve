package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "verve/app/system/handlers"
	"verve/infrastructure/database"
)

func SetupAgentModelConfigRoutes(router fiber.Router, dbService *database.DatabaseService) {
	handler := system_handlers.NewAgentModelConfigHandler(dbService)

	configs := router.Group("/system/agent-model-configs")
	{
		configs.Get("/", handler.FindConfigs)
		configs.Put("/:agentKey/:sceneKey", handler.UpsertConfig)
	}
}
