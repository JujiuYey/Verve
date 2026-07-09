package router

import (
	"github.com/gofiber/fiber/v2"

	wiki_handlers "verve/app/wiki/handlers"
	"verve/infrastructure/database"
)

func SetupAgentInstanceRoutes(router fiber.Router, dbService *database.DatabaseService) {
	handler := wiki_handlers.NewAgentInstanceHandler(dbService)

	agents := router.Group("/wiki/agent-instances")
	{
		agents.Get("/", handler.FindByRoot)
		agents.Post("/ensure", handler.Ensure)
	}
}
