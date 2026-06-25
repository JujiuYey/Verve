package router

import (
	"sag-wiki/app/ai/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupCollectionRoutes 设置 collection 路由
func SetupCollectionRoutes(api fiber.Router) {
	collectionHandler, err := handlers.NewCollectionHandler()
	if err != nil {
		panic("初始化 CollectionHandler 失败: " + err.Error())
	}

	collections := api.Group("/ai/collections")
	{
		collections.Get("/", collectionHandler.List)
		collections.Get("/:name", collectionHandler.Get)
		collections.Post("/", collectionHandler.Create)
		collections.Delete("/:name", collectionHandler.Delete)
		collections.Get("/:name/points", collectionHandler.GetPoints)
		collections.Get("/:name/stats", collectionHandler.GetStats)
	}
}
