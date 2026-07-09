package router

import (
	"github.com/gofiber/fiber/v2"

	rag_handlers "verve/app/rag/handlers"
	rag_queue "verve/app/rag/queue"
	rag_repo "verve/app/rag/repository"
	rag_service "verve/app/rag/service"
)

func SetupRAGRoutes(
	router fiber.Router,
	indexer *rag_service.Indexer,
	retriever *rag_service.Retriever,
	jobs *rag_repo.IndexJobRepository,
	enqueuer *rag_queue.Enqueuer,
) {
	handler := rag_handlers.NewRAGHandler(indexer, retriever, jobs, enqueuer)

	rag := router.Group("/rag/wiki")
	{
		rag.Post("/documents/:id/index", handler.IndexDocument)
		rag.Delete("/documents/:id/index", handler.DeleteDocumentIndex)
		rag.Post("/folders/:id/index", handler.IndexFolder)
		rag.Get("/jobs", handler.ListJobs)
		rag.Post("/search", handler.Search)
	}
}
