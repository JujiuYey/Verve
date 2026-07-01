package router

import (
	"github.com/gofiber/fiber/v2"

	file_handlers "verve/app/file/handlers"
	"verve/infrastructure/storage"
)

// 配置文件访问路由
func SetupFileRoutes(router fiber.Router, minioService *storage.MinIOService) {
	fileHandler := file_handlers.NewFileHandler(minioService)

	// 文件下载路由（公开访问）
	router.Get("/files/*filepath", fileHandler.GetFile)

	// 文件预览路由
	router.Get("/preview/*filepath", fileHandler.PreviewFile)
}
