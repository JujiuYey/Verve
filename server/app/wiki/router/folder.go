package router

import (
	"github.com/gofiber/fiber/v2"

	wiki_handlers "verve/app/wiki/handlers"
	"verve/infrastructure/database"
)

// 配置文件夹管理路由
func SetupFolderRoutes(router fiber.Router, dbService *database.DatabaseService) {
	folderHandler := wiki_handlers.NewFolderHandler(dbService)

	folders := router.Group("/wiki/folders")
	{
		folders.Get("/page", folderHandler.FindPage)
		folders.Get("/list", folderHandler.FindList)
		folders.Get("/tree", folderHandler.GetTree)
		folders.Get("/:id", folderHandler.FindOne)
		folders.Post("/", folderHandler.Create)
		folders.Put("/", folderHandler.Update)
		folders.Delete("/:id", folderHandler.Delete)
	}
}
