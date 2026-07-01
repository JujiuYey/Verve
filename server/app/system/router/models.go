package router

import (
	"github.com/gofiber/fiber/v2"

	system_handlers "sag-wiki/app/system/handlers"
	"sag-wiki/infrastructure/database"
)

// SetupModelRoutes 配置已启用模型路由
func SetupModelRoutes(router fiber.Router, dbService *database.DatabaseService) {
	modelHandler := system_handlers.NewModelHandler(dbService)

	models := router.Group("/system/models")
	{
		models.Get("/", modelHandler.FindModels)
		models.Post("/", modelHandler.CreateModel)
		models.Put("/:id", modelHandler.UpdateModel)
		models.Delete("/:id", modelHandler.DeleteModel)
	}
}