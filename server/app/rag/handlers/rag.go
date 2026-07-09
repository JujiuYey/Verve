package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	rag_payload "verve/app/rag/models/payload"
	rag_queue "verve/app/rag/queue"
	rag_repo "verve/app/rag/repository"
	rag_service "verve/app/rag/service"
	"verve/common/response"
)

type RAGHandler struct {
	indexer   *rag_service.Indexer
	retriever *rag_service.Retriever
	jobs      *rag_repo.IndexJobRepository
	enqueuer  *rag_queue.Enqueuer
}

func NewRAGHandler(
	indexer *rag_service.Indexer,
	retriever *rag_service.Retriever,
	jobs *rag_repo.IndexJobRepository,
	enqueuer *rag_queue.Enqueuer,
) *RAGHandler {
	return &RAGHandler{indexer: indexer, retriever: retriever, jobs: jobs, enqueuer: enqueuer}
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

func (h *RAGHandler) Search(c *fiber.Ctx) error {
	var req rag_payload.SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, "参数错误: "+err.Error())
	}
	results, err := h.retriever.Search(c.Context(), req.RootFolderID, req.Query, req.Limit)
	if err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	return response.SuccessCtx(c, results)
}

func (h *RAGHandler) IndexFolder(c *fiber.Ctx) error {
	folderID := strings.TrimSpace(c.Params("id"))
	if folderID == "" {
		return response.BadRequestCtx(c, "缺少文件夹ID")
	}
	if err := h.indexer.CheckReady(c.Context()); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	batch, count, err := h.enqueuer.EnqueueFolder(c.Context(), folderID)
	if err != nil {
		return response.InternalServerCtx(c, "启动解析队列失败: "+err.Error())
	}
	return response.SuccessCtx(c, rag_payload.IndexFolderResponse{
		BatchID:       batch.ID,
		RootFolderID:  folderID,
		DocumentCount: count,
		StartedAt:     batch.CreatedAt.Format(time.RFC3339),
	})
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
			ID:           job.ID,
			DocumentID:   job.DocumentID,
			RootFolderID: job.RootFolderID,
			Status:       job.Status,
			ErrorMessage: job.ErrorMessage,
			ChunkCount:   job.ChunkCount,
			CreatedAt:    job.CreatedAt.Format(time.RFC3339),
			StartedAt:    formatOptionalTime(job.StartedAt),
			FinishedAt:   formatOptionalTime(job.FinishedAt),
		})
	}
	return response.SuccessCtx(c, result)
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}
