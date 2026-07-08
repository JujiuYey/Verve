package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	learning_db "verve/app/learning/models/db"
	learning_service "verve/app/learning/service"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
)

type ObjectiveHandler struct {
	db    *database.DatabaseService
	minio *storage.MinIOService
}

func NewObjectiveHandler(db *database.DatabaseService, minio *storage.MinIOService) *ObjectiveHandler {
	return &ObjectiveHandler{db: db, minio: minio}
}

type ensureObjectivesByDocumentRequest struct {
	DocumentID string `json:"document_id"`
}

type ensureObjectivesByDocumentResponse struct {
	DocumentID       string                           `json:"document_id"`
	FirstObjectiveID string                           `json:"first_objective_id"`
	CreatedCount     int                              `json:"created_count"`
	Reused           bool                             `json:"reused"`
	Objectives       []*learning_db.LearningObjective `json:"objectives"`
}

func (h *ObjectiveHandler) List(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	documentID := strings.TrimSpace(c.Query("document_id"))
	folderID := strings.TrimSpace(c.Query("folder_id"))

	var objectives []*learning_db.LearningObjective
	var err error
	if documentID != "" {
		objectives, err = h.db.Objectives.FindByDocument(c.Context(), documentID)
	} else if folderID != "" {
		objectives, err = h.db.Objectives.FindByFolder(c.Context(), folderID)
	} else {
		objectives, err = h.db.Objectives.FindRecentByUser(c.Context(), userID, 50)
	}
	if err != nil {
		return response.InternalServerCtx(c, "读取学习小节失败")
	}

	filtered := make([]*learning_db.LearningObjective, 0, len(objectives))
	for _, objective := range objectives {
		if objective.UserID != userID {
			continue
		}
		filtered = append(filtered, objective)
	}

	return response.SuccessCtx(c, filtered)
}

func (h *ObjectiveHandler) FindOne(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	objective, err := h.db.Objectives.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "学习小节不存在")
	}
	if objective.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	return response.SuccessCtx(c, objective)
}

func (h *ObjectiveHandler) EnsureByDocument(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req ensureObjectivesByDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	documentID := strings.TrimSpace(req.DocumentID)
	if documentID == "" {
		return response.BadRequestCtx(c, "缺少文档 ID")
	}

	doc, err := h.db.Documents.FindOne(c.Context(), documentID)
	if err != nil {
		return response.NotFoundCtx(c, "文档不存在")
	}
	folder, err := h.db.Folders.FindOne(c.Context(), doc.FolderID)
	if err != nil {
		return response.NotFoundCtx(c, "文件夹不存在")
	}
	if folder.UserID != nil && *folder.UserID != "" && *folder.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	existing, err := h.db.Objectives.FindByDocument(c.Context(), documentID)
	if err != nil {
		return response.InternalServerCtx(c, "读取学习小节失败")
	}
	reusable := filterObjectivesForUser(existing, userID)
	if len(reusable) > 0 {
		return response.SuccessCtx(c, buildEnsureObjectivesByDocumentResponse(documentID, reusable, true))
	}

	if h.minio == nil {
		return response.InternalServerCtx(c, "文档存储服务不可用")
	}
	content, err := h.minio.GetFileContent(c.Context(), doc.FilePath)
	if err != nil {
		return response.InternalServerCtx(c, "读取文档内容失败")
	}
	objectives, err := learning_service.NewObjectiveGenerationService(h.db).GenerateFromMarkdown(c.Context(), userID, doc, folder, content)
	if err != nil {
		return response.InternalServerCtx(c, "生成学习小节失败")
	}

	return response.SuccessCtx(c, buildEnsureObjectivesByDocumentResponse(documentID, objectives, false))
}

func filterObjectivesForUser(objectives []*learning_db.LearningObjective, userID string) []*learning_db.LearningObjective {
	filtered := make([]*learning_db.LearningObjective, 0, len(objectives))
	for _, objective := range objectives {
		if objective.UserID != userID {
			continue
		}
		filtered = append(filtered, objective)
	}
	return filtered
}

func buildEnsureObjectivesByDocumentResponse(documentID string, objectives []*learning_db.LearningObjective, reused bool) ensureObjectivesByDocumentResponse {
	out := ensureObjectivesByDocumentResponse{
		DocumentID:   documentID,
		CreatedCount: len(objectives),
		Reused:       reused,
		Objectives:   objectives,
	}
	for _, objective := range objectives {
		if objective.ID != "" {
			out.FirstObjectiveID = objective.ID
			break
		}
	}
	return out
}
