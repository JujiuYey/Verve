package router

import (
	"github.com/gofiber/fiber/v2"

	wiki_handlers "sag-wiki/app/wiki/handlers"
	"sag-wiki/infrastructure/database"
	qdrantdao "sag-wiki/infrastructure/qdrant"
	"sag-wiki/infrastructure/queue"
	"sag-wiki/infrastructure/storage"
)

// 配置文档管理路由
func SetupDocumentRoutes(
	router fiber.Router,
	dbService *database.DatabaseService,
	minioService *storage.MinIOService,
	taskQueue *queue.TaskQueue,
	chunkDAO *qdrantdao.ChunkDAO,
) {
	docHandler := wiki_handlers.NewDocumentHandler(dbService, minioService, taskQueue, chunkDAO)

	docs := router.Group("/wiki/documents")
	{
		// 获取文档列表（不分页）
		docs.Get("/list", docHandler.FindList)
		// 获取文档列表
		docs.Get("/page", docHandler.FindPage)
		// 获取文档详情
		docs.Get("/:id", docHandler.FindOne)
		// 上传文档
		docs.Post("/upload", docHandler.Upload)
		// 重新处理文档
		docs.Post("/:id/reprocess", docHandler.Reprocess)
		// 删除文档
		docs.Delete("/:id", docHandler.Delete)
		// 下载文档
		docs.Get("/:id/download", docHandler.Download)
		// 获取文档内容
		docs.Get("/:id/content", docHandler.GetContent)
		// 更新文档内容
		docs.Put("/:id/content", docHandler.UpdateContent)
		// 获取文档 chunks
		docs.Get("/:id/chunks", docHandler.GetChunks)
		// 删除文档 chunks
		docs.Delete("/:id/chunks", docHandler.DeleteChunks)
	}
}
