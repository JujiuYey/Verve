package handlers

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"verve/common/response"
	"verve/infrastructure/database"
)

// 学习记忆处理器
type MemoryHandler struct {
	db *database.DatabaseService
}

func NewMemoryHandler(db *database.DatabaseService) *MemoryHandler {
	return &MemoryHandler{db: db}
}

type memoryResponse struct {
	Summary    string       `json:"summary"`
	Highlights []string     `json:"highlights"`
	Items      []memoryItem `json:"items"`
}

type memoryItem struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Statement  string    `json:"statement"`
	Confidence string    `json:"confidence"`
	FolderID   *string   `json:"folder_id"`
	DocumentID *string   `json:"document_id"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// 学习记忆读取(全局或按文件夹)
func (h *MemoryHandler) Get(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	folderID := c.Query("folder_id")

	if h.db == nil || h.db.Memories == nil {
		return response.SuccessCtx(c, memoryResponse{
			Highlights: []string{},
			Items:      []memoryItem{},
		})
	}

	resp := memoryResponse{
		Highlights: []string{},
		Items:      []memoryItem{},
	}

	summary, err := h.db.Memories.FindSummaryByFolder(c.Context(), userID, folderID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return response.InternalServerCtx(c, "获取学习记忆摘要失败")
	}
	if summary != nil {
		resp.Summary = summary.Summary
		if summary.Highlights != nil {
			resp.Highlights = summary.Highlights
		}
	}

	items, err := h.db.Memories.FindItemsByUser(c.Context(), userID, folderID, 50)
	if err != nil {
		return response.InternalServerCtx(c, "获取学习记忆条目失败")
	}

	for _, item := range items {
		if item == nil {
			continue
		}
		resp.Items = append(resp.Items, memoryItem{
			ID:         item.ID,
			Kind:       item.Kind,
			Statement:  item.Statement,
			Confidence: item.Confidence,
			FolderID:   item.FolderID,
			DocumentID: item.DocumentID,
			LastSeenAt: item.LastSeenAt,
		})
	}

	return response.SuccessCtx(c, resp)
}
