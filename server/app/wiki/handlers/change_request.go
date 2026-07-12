package handlers

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
	wiki_service "verve/app/wiki/service"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
)

type changeRequestVersionService interface {
	ApplyChangeRequest(ctx context.Context, userID, requestID string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error)
	CancelChangeRequest(ctx context.Context, userID, requestID string) error
}

// ChangeRequestHandler 处理文档修改申请的确认与取消。
type ChangeRequestHandler struct {
	versions changeRequestVersionService
	requests *wiki_repo.ChangeRequestRepository
}

func NewChangeRequestHandler(dbService *database.DatabaseService, minio *storage.MinIOService) *ChangeRequestHandler {
	return &ChangeRequestHandler{
		versions: wiki_service.NewDocumentVersionService(
			dbService.Revisions, dbService.Versions, dbService.Documents, minio, dbService.ChangeRequests,
		),
		requests: dbService.ChangeRequests,
	}
}

func (h *ChangeRequestHandler) Apply(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	requestID := c.Params("id")
	request, revision, job, err := h.apply(c.Context(), userID, requestID)
	if err != nil {
		return h.writeError(c, err)
	}
	return response.SuccessCtx(c, fiber.Map{"change_request": request, "revision": revision, "index_job": job})
}

func (h *ChangeRequestHandler) Cancel(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	if err := h.versions.CancelChangeRequest(c.Context(), userID, c.Params("id")); err != nil {
		return h.writeError(c, err)
	}
	return response.SuccessMsgCtx(c, "文档变更申请已取消")
}

func (h *ChangeRequestHandler) apply(ctx context.Context, userID, requestID string) (*wiki_db.DocumentChangeRequest, *wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	revision, job, err := h.versions.ApplyChangeRequest(ctx, userID, requestID)
	if err != nil {
		return nil, nil, nil, err
	}
	request, err := h.requests.FindChangeRequest(ctx, requestID)
	if err != nil {
		return nil, nil, nil, err
	}
	return request, revision, job, nil
}

func (h *ChangeRequestHandler) writeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, wiki_repo.ErrVersionConflict):
		return response.FailWithCodeCtx(c, fiber.StatusConflict, "文档版本已变化，请重新生成修改建议")
	case errors.Is(err, wiki_repo.ErrChangeRequestForbidden):
		return response.ForbiddenCtx(c, "无权操作该文档变更申请")
	case errors.Is(err, wiki_repo.ErrChangeRequestNotProposed):
		return response.BadRequestCtx(c, "文档变更申请当前不可操作")
	default:
		return response.InternalServerCtx(c, "操作文档变更申请失败")
	}
}
