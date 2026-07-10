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
	sessionHandler := learning_handlers.NewSessionHandler(dbService)
	journalHandler := learning_handlers.NewJournalHandler(dbService)
	exerciseHandler := learning_handlers.NewExerciseHandler(dbService)
	objectiveHandler := learning_handlers.NewObjectiveHandler(dbService, minioService)
	coachHandler := learning_handlers.NewCoachHandler(dbService, minioService, retriever)
	memoryHandler := learning_handlers.NewMemoryHandler(dbService)

	learning := router.Group("/learning")
	{
		// 学习会话
		session := learning.Group("/session")
		{
			session.Post("/", sessionHandler.Create)
			session.Get("/:id", sessionHandler.FindOne)
			session.Post("/:id/chat", sessionHandler.Chat)         // 陪练对话(SSE)
			session.Post("/:id/exercise", sessionHandler.Exercise) // 提交练习验证
			session.Post("/:id/complete", sessionHandler.Complete) // 结束本节
		}

		// 学习日志
		journal := learning.Group("/journal")
		{
			journal.Get("/page", journalHandler.FindPage)
		}

		// 费曼练习记录
		exercise := learning.Group("/exercise")
		{
			exercise.Get("/page", exerciseHandler.FindPage)
		}

		objective := learning.Group("/objective")
		{
			objective.Get("/", objectiveHandler.List)
			objective.Post("/ensure-by-document", objectiveHandler.EnsureByDocument)
			objective.Get("/:id", objectiveHandler.FindOne)
		}

		coach := learning.Group("/coach")
		{
			coach.Post("/chat", coachHandler.Chat)
		}

		learning.Get("/memory", memoryHandler.Get)
	}
}
