package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	learning_db "verve/app/learning/models/db"
	"verve/common/response"
	"verve/infrastructure/database"
)

type ObjectiveHandler struct {
	db *database.DatabaseService
}

func NewObjectiveHandler(db *database.DatabaseService) *ObjectiveHandler {
	return &ObjectiveHandler{db: db}
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
