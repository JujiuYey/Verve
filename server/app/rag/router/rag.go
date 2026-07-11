package router

import (
	"github.com/gofiber/fiber/v2"

	rag_handlers "verve/app/rag/handlers"
	rag_repo "verve/app/rag/repository"
	rag_service "verve/app/rag/service"
)

func SetupRAGRoutes(
	router fiber.Router,
	indexer *rag_service.Indexer,
	jobs *rag_repo.IndexJobRepository,
) {
	handler := rag_handlers.NewRAGHandler(indexer, jobs)

	rag := router.Group("/rag/wiki")
	{
		rag.Post("/documents/:id/index", handler.IndexDocument)
		rag.Delete("/documents/:id/index", handler.DeleteDocumentIndex)
		rag.Get("/jobs", handler.ListJobs)
	}
}
