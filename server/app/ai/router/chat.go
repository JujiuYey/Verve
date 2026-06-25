package router

import (
	"github.com/gofiber/fiber/v2"

	ai_handlers "sag-wiki/app/ai/handlers"
	ai_service "sag-wiki/app/ai/service"
	"sag-wiki/infrastructure/database"
)

// 配置聊天路由
func SetupChatRoutes(router fiber.Router, dbService *database.DatabaseService, retrievalService *ai_service.RetrievalService) {
	chatHandler := ai_handlers.NewChatHandler(dbService, dbService.ModelConfigs, retrievalService)

	chat := router.Group("/ai/chat")
	{
		// 发送消息
		chat.Post("/", chatHandler.Chat)
	}
}
