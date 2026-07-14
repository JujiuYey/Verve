package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	file_router "verve/app/file/router"
	learning_router "verve/app/learning/router"
	rag_router "verve/app/rag/router"
	rag_service "verve/app/rag/service"
	system_router "verve/app/system/router"
	wiki_router "verve/app/wiki/router"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
	"verve/infrastructure/vector"
)

// 配置路由
func SetupRouter(
	dbService *database.DatabaseService,
	minioService *storage.MinIOService,
	vectorStore vector.Store,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Verve",
		ErrorHandler: customErrorHandler,
	})

	// 全局中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5200",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept",
		AllowCredentials: false,
	}))

	// 路由组
	api := app.Group("/api")
	embedder := rag_service.NewOpenAICompatibleEmbedder(dbService.ModelConfigs)
	indexer := rag_service.NewIndexer(dbService.RAGChunks, dbService.RAGJobs, dbService.Folders, dbService.Documents, minioService, embedder, vectorStore)
	retriever := rag_service.NewRetriever(dbService.RAGChunks, embedder, vectorStore)

	// 文件访问路由（公开访问，通过后端代理 MinIO）
	file_router.SetupFileRoutes(api.Group("/"), minioService)

	// 系统配置路由
	system_router.SetupPlatformRoutes(api.Group("/"), dbService)
	system_router.SetupModelRoutes(api.Group("/"), dbService)
	system_router.SetupAgentModelConfigRoutes(api.Group("/"), dbService)

	// 知识库路由
	wiki_router.SetupFolderRoutes(api.Group("/"), dbService)
	wiki_router.SetupDocumentRoutes(api.Group("/"), dbService, minioService, indexer)

	// RAG 路由
	rag_router.SetupRAGRoutes(
		api.Group("/"),
		indexer,
		dbService.RAGJobs,
	)

	// 学习平台路由
	learning_router.SetupLearningRoutes(api.Group("/"), dbService, minioService, retriever)

	return app
}

// customErrorHandler 自定义错误处理
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
