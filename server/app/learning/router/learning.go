package router

import (
	"github.com/gofiber/fiber/v2"

	learning_handlers "sag-wiki/app/learning/handlers"
	"sag-wiki/infrastructure/database"
)

// 配置学习平台路由(/api/learning/*)
func SetupLearningRoutes(router fiber.Router, dbService *database.DatabaseService) {
	goalHandler := learning_handlers.NewGoalHandler(dbService)
	sessionHandler := learning_handlers.NewSessionHandler(dbService)
	profileHandler := learning_handlers.NewProfileHandler(dbService)
	journalHandler := learning_handlers.NewJournalHandler(dbService)

	learning := router.Group("/learning")
	{
		// 继续上次 / 今日推荐
		learning.Get("/continue", goalHandler.Continue)

		// 学习目标
		goal := learning.Group("/goal")
		{
			goal.Get("/page", goalHandler.FindPage) // 注意:/page 在 /:id 之前注册
			goal.Post("/from-folder", goalHandler.CreateFromFolder)
			goal.Post("/", goalHandler.Create)
			goal.Put("/", goalHandler.Update)
			goal.Get("/:id", goalHandler.FindOne)
			goal.Delete("/:id", goalHandler.Delete)
			goal.Get("/:id/profile", profileHandler.Get)
		}

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
	}
}
