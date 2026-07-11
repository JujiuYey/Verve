package router

import (
	"github.com/gofiber/fiber/v2"

	learning_handlers "verve/app/learning/handlers"
	rag_service "verve/app/rag/service"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
)

// 配置学习平台路由(/api/learning/*)
func SetupLearningRoutes(router fiber.Router, dbService *database.DatabaseService, minioService *storage.MinIOService, retriever *rag_service.Retriever) {
	sessionHandler := learning_handlers.NewSessionHandler(dbService, minioService, retriever)
	coachHandler := learning_handlers.NewCoachHandler(dbService, retriever)
	memoryHandler := learning_handlers.NewMemoryHandler(dbService)

	learning := router.Group("/learning")
	{
		// 学习会话
		session := learning.Group("/session")
		{
			session.Post("/", sessionHandler.Create)
			session.Get("/:id", sessionHandler.FindOne)
			session.Post("/:id/review", sessionHandler.Review)
			session.Post("/:id/complete", sessionHandler.Complete) // 结束本节
		}

		coach := learning.Group("/coach")
		{
			coach.Post("/chat", coachHandler.Chat)
		}

		learning.Get("/memory", memoryHandler.Get)
	}
}
