package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	rag_payload "verve/app/rag/models/payload"
	rag_repo "verve/app/rag/repository"
	rag_service "verve/app/rag/service"
	"verve/common/response"
)

type RAGHandler struct {
	indexer *rag_service.Indexer
	jobs    *rag_repo.IndexJobRepository
}

func NewRAGHandler(
	indexer *rag_service.Indexer,
	jobs *rag_repo.IndexJobRepository,
) *RAGHandler {
	return &RAGHandler{indexer: indexer, jobs: jobs}
}

func (h *RAGHandler) IndexDocument(c *fiber.Ctx) error {
	documentID := strings.TrimSpace(c.Params("id"))
	if documentID == "" {
		return response.BadRequestCtx(c, "缺少文档ID")
	}
	if err := h.indexer.IndexDocument(c.Context(), documentID); err != nil {
		return response.InternalServerCtx(c, "索引文档失败: "+err.Error())
	}
	return response.SuccessMsgCtx(c, "文档索引完成")
}

func (h *RAGHandler) DeleteDocumentIndex(c *fiber.Ctx) error {
	documentID := strings.TrimSpace(c.Params("id"))
	if documentID == "" {
		return response.BadRequestCtx(c, "缺少文档ID")
	}
	if err := h.indexer.DeleteDocumentIndex(c.Context(), documentID); err != nil {
		return response.InternalServerCtx(c, "删除文档索引失败: "+err.Error())
	}
	return response.SuccessMsgCtx(c, "文档索引已删除")
}

func (h *RAGHandler) ListJobs(c *fiber.Ctx) error {
	rootFolderID := strings.TrimSpace(c.Query("root_folder_id"))
	jobs, err := h.jobs.ListRecent(c.Context(), rootFolderID, 100)
	if err != nil {
		return response.InternalServerCtx(c, "获取解析任务失败: "+err.Error())
	}
	result := make([]rag_payload.IndexJobProgress, 0, len(jobs))
	for _, job := range jobs {
		result = append(result, rag_payload.IndexJobProgress{
			ID:              job.ID,
			DocumentID:      job.DocumentID,
			DocumentVersion: job.DocumentVersion,
			RootFolderID:    job.RootFolderID,
			Status:          job.Status,
			ErrorMessage:    job.ErrorMessage,
			ChunkCount:      job.ChunkCount,
			CreatedAt:       job.CreatedAt.Format(time.RFC3339),
			StartedAt:       formatOptionalTime(job.StartedAt),
			FinishedAt:      formatOptionalTime(job.FinishedAt),
		})
	}
	return response.SuccessCtx(c, result)
}

func (h *RAGHandler) DocumentIndexStatus(c *fiber.Ctx) error {
	job, err := h.jobs.FindCurrentVersion(c.Context(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return response.NotFoundCtx(c, "文档索引任务不存在")
	}
	return response.SuccessCtx(c, rag_payload.IndexJobProgress{
		ID: job.ID, DocumentID: job.DocumentID, DocumentVersion: job.DocumentVersion, RootFolderID: job.RootFolderID,
		Status: job.Status, ErrorMessage: job.ErrorMessage, ChunkCount: job.ChunkCount,
		CreatedAt: job.CreatedAt.Format(time.RFC3339), StartedAt: formatOptionalTime(job.StartedAt), FinishedAt: formatOptionalTime(job.FinishedAt),
	})
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}
