package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	ai_router "sag-wiki/app/ai/router"
	ai_service "sag-wiki/app/ai/service"
	file_router "sag-wiki/app/file/router"
	system_router "sag-wiki/app/system/router"
	wiki_router "sag-wiki/app/wiki/router"
	"sag-wiki/infrastructure/database"
	qdrantdao "sag-wiki/infrastructure/qdrant"
	"sag-wiki/infrastructure/queue"
	"sag-wiki/infrastructure/storage"
	"sag-wiki/middleware"
)

// 配置路由
func SetupRouter(
	dbService *database.DatabaseService,
	minioService *storage.MinIOService,
	taskQueue *queue.TaskQueue,
	chunkDAO *qdrantdao.ChunkDAO,
	retrievalService *ai_service.RetrievalService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "SAG-WIKI",
		ErrorHandler: customErrorHandler,
	})

	// 全局中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5200",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// 路由组
	api := app.Group("/api")

	// 文件访问路由（公开访问，通过后端代理 MinIO）
	file_router.SetupFileRoutes(api.Group("/"), minioService)

	// 认证路由（包含公开和受保护的路由）
	system_router.SetupAuthRoutes(api.Group("/"), dbService)

	// 需要认证的路由组
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// 文档管理路由
		wiki_router.SetupDocumentRoutes(protected.Group("/"), dbService, minioService, taskQueue, chunkDAO)

		// 队列监控路由
		system_router.SetupQueueRoutes(protected.Group("/"), taskQueue)

		// 部门管理路由
		system_router.SetupDepartmentRoutes(protected.Group("/"), dbService)

		// 角色管理路由
		system_router.SetupRoleRoutes(protected.Group("/"), dbService)

		// 用户管理路由
		system_router.SetupUserRoutes(protected.Group("/"), dbService, minioService)

		// 文件夹管理路由
		wiki_router.SetupFolderRoutes(protected.Group("/"), dbService)

		// 模型配置路由
		ai_router.SetupModelConfigRoutes(protected.Group("/"), dbService)

		// Collection 路由
		ai_router.SetupCollectionRoutes(protected.Group("/"))

		// 聊天路由
		ai_router.SetupChatRoutes(protected.Group("/"), dbService, retrievalService)
	}

	// 管理员路由（需要 admin 角色）
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
	{
		// 这里可以添加管理员专用的路由
		// 例如：用户管理、系统配置等
	}

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
