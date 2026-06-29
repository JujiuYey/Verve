package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	file_router "sag-wiki/app/file/router"
	learning_router "sag-wiki/app/learning/router"
	system_router "sag-wiki/app/system/router"
	"sag-wiki/infrastructure/database"
	qdrantdao "sag-wiki/infrastructure/qdrant"
	"sag-wiki/infrastructure/storage"
	"sag-wiki/middleware"
)

// 配置路由
func SetupRouter(
	dbService *database.DatabaseService,
	minioService *storage.MinIOService,
	chunkDAO *qdrantdao.ChunkDAO,
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
		// 用户管理路由
		system_router.SetupUserRoutes(protected.Group("/"), dbService, minioService)

		// 模型配置路由
		system_router.SetupModelConfigRoutes(protected.Group("/"), dbService)

		// 学习平台路由
		learning_router.SetupLearningRoutes(protected.Group("/"), dbService)
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
