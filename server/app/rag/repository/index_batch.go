package repository

import (
	"context"
	"time"

	rag_db "verve/app/rag/models/db"

	"github.com/uptrace/bun"
)

type IndexBatchRepository struct {
	db *bun.DB
}

func NewIndexBatchRepository(db *bun.DB) *IndexBatchRepository {
	return &IndexBatchRepository{db: db}
}

func (r *IndexBatchRepository) Create(ctx context.Context, rootFolderID string, totalCount int) (*rag_db.IndexBatch, error) {
	now := time.Now()
	batch := &rag_db.IndexBatch{
		ID:           compactUUID(),
		RootFolderID: rootFolderID,
		Status:       "pending",
		TotalCount:   totalCount,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := r.db.NewInsert().Model(batch).Exec(ctx)
	return batch, err
}

func (r *IndexBatchRepository) MarkRunning(ctx context.Context, batchID string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexBatch)(nil)).
		Set("status = ?", "running").
		Set("started_at = COALESCE(started_at, ?)", now).
		Set("updated_at = ?", now).
		Where("id = ?", batchID).
		Where("status IN (?)", bun.In([]string{"pending", "running"})).
		Exec(ctx)
	return err
}

func (r *IndexBatchRepository) MarkCompleted(ctx context.Context, batchID string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexBatch)(nil)).
		Set("status = ?", "completed").
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", batchID).
		Exec(ctx)
	return err
}

func (r *IndexBatchRepository) RefreshStatus(ctx context.Context, batchID string) error {
	var counts []struct {
		Status string `bun:"status"`
		Count  int    `bun:"count"`
	}
	if err := r.db.NewSelect().
		Model((*rag_db.IndexJob)(nil)).
		ColumnExpr("status").
		ColumnExpr("COUNT(*) AS count").
		Where("batch_id = ?", batchID).
		Group("status").
		Scan(ctx, &counts); err != nil {
		return err
	}
	total := 0
	completed := 0
	failed := 0
	for _, row := range counts {
		total += row.Count
		switch row.Status {
		case "completed":
			completed = row.Count
		case "failed":
			failed = row.Count
		}
	}
	if total == 0 || completed+failed < total {
		return nil
	}
	status := "completed"
	if failed > 0 {
		status = "failed"
	}
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexBatch)(nil)).
		Set("status = ?", status).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", batchID).
		Exec(ctx)
	return err
}
