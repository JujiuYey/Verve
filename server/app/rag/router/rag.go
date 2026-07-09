package router

import (
	"github.com/gofiber/fiber/v2"

	rag_handlers "verve/app/rag/handlers"
	rag_repo "verve/app/rag/repository"
	rag_service "verve/app/rag/service"
	wiki_repo "verve/app/wiki/repository"
)

func SetupRAGRoutes(
	router fiber.Router,
	indexer *rag_service.Indexer,
	retriever *rag_service.Retriever,
	folders wiki_repo.FolderRepository,
	docs *wiki_repo.DocumentRepository,
	jobs *rag_repo.IndexJobRepository,
) {
	handler := rag_handlers.NewRAGHandler(indexer, retriever, folders, docs, jobs)

	rag := router.Group("/rag/wiki")
	{
		rag.Post("/documents/:id/index", handler.IndexDocument)
		rag.Delete("/documents/:id/index", handler.DeleteDocumentIndex)
		rag.Post("/folders/:id/index", handler.IndexFolder)
		rag.Get("/jobs", handler.ListJobs)
		rag.Post("/search", handler.Search)
	}
}
