package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 学习记忆数据访问层
type MemoryRepository struct {
	db *bun.DB
}

func NewMemoryRepository(database *bun.DB) *MemoryRepository {
	return &MemoryRepository{db: database}
}

func (r *MemoryRepository) CreateEvent(ctx context.Context, event *learning_db.LearningMemoryEvent) error {
	event.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(event).
		On("CONFLICT (source_type, source_id, event_type) WHERE source_id IS NOT NULL DO UPDATE SET id = learning_memory_events.id").
		Returning("id").
		Exec(ctx)
	return err
}

func (r *MemoryRepository) CreateItem(ctx context.Context, item *learning_db.LearningMemoryItem) error {
	item.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	return err
}

func (r *MemoryRepository) FindItemsByUser(ctx context.Context, folderID string, limit int) ([]*learning_db.LearningMemoryItem, error) {
	if limit <= 0 {
		limit = 20
	}

	var items []*learning_db.LearningMemoryItem
	query := r.db.NewSelect().
		Model(&items).
		Order("last_seen_at DESC").
		Limit(limit)
	if folderID != "" {
		query.Where("folder_id = ?", folderID)
	} else {
		query.Where("folder_id IS NULL")
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MemoryRepository) FindItemsByDocument(ctx context.Context, documentID string, limit int) ([]*learning_db.LearningMemoryItem, error) {
	if limit <= 0 {
		limit = 20
	}
	items := make([]*learning_db.LearningMemoryItem, 0)
	err := r.db.NewSelect().
		Model(&items).
		Where("document_id = ?", documentID).
		Order("last_seen_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MemoryRepository) FindItemsByFolders(ctx context.Context, folderIDs []string, limit int) ([]*learning_db.LearningMemoryItem, error) {
	if limit <= 0 {
		limit = 20
	}
	items := make([]*learning_db.LearningMemoryItem, 0)
	if len(folderIDs) == 0 {
		return items, nil
	}
	err := r.db.NewSelect().
		Model(&items).
		Where("folder_id IN (?)", bun.In(folderIDs)).
		Order("last_seen_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MemoryRepository) FindSummaryByFolder(ctx context.Context, folderID string) (*learning_db.LearningMemorySummary, error) {
	summary := new(learning_db.LearningMemorySummary)
	query := r.db.NewSelect().Model(summary)
	if folderID == "" {
		query.Where("folder_id IS NULL")
	} else {
		query.Where("folder_id = ?", folderID)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return summary, nil
}

func (r *MemoryRepository) UpsertSummary(ctx context.Context, summary *learning_db.LearningMemorySummary) error {
	if summary.ID == "" {
		summary.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	summary.UpdatedAt = time.Now()

	if summary.FolderID == nil || *summary.FolderID == "" {
		summary.FolderID = nil
		_, err := r.db.NewInsert().
			Model(summary).
			On("CONFLICT ((1)) WHERE folder_id IS NULL DO UPDATE").
			Set("summary = EXCLUDED.summary").
			Set("highlights = EXCLUDED.highlights").
			Set("generated_from_event_id = EXCLUDED.generated_from_event_id").
			Set("generated_at = EXCLUDED.generated_at").
			Set("updated_at = EXCLUDED.updated_at").
			Exec(ctx)
		return err
	}

	_, err := r.db.NewInsert().
		Model(summary).
		On("CONFLICT (folder_id) WHERE folder_id IS NOT NULL DO UPDATE").
		Set("summary = EXCLUDED.summary").
		Set("highlights = EXCLUDED.highlights").
		Set("generated_from_event_id = EXCLUDED.generated_from_event_id").
		Set("generated_at = EXCLUDED.generated_at").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}
