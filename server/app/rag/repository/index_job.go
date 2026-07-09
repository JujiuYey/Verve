package repository

import (
	"context"
	"strings"
	"time"

	rag_db "verve/app/rag/models/db"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type IndexJobRepository struct {
	db *bun.DB
}

func NewIndexJobRepository(db *bun.DB) *IndexJobRepository {
	return &IndexJobRepository{db: db}
}

func (r *IndexJobRepository) CreatePending(ctx context.Context, documentID string) (*rag_db.IndexJob, error) {
	job := &rag_db.IndexJob{
		ID:          compactUUID(),
		DocumentID:  documentID,
		Status:      "pending",
		MaxAttempts: 3,
	}
	_, err := r.db.NewInsert().Model(job).Exec(ctx)
	return job, err
}

func (r *IndexJobRepository) FindOne(ctx context.Context, jobID string) (*rag_db.IndexJob, error) {
	job := new(rag_db.IndexJob)
	err := r.db.NewSelect().Model(job).Where("id = ?", jobID).Scan(ctx)
	return job, err
}

func (r *IndexJobRepository) MarkRunning(ctx context.Context, jobID string, rootFolderID string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "running").
		Set("root_folder_id = ?", rootFolderID).
		Set("attempt_count = attempt_count + 1").
		Set("started_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func (r *IndexJobRepository) MarkCompleted(ctx context.Context, jobID string, chunkCount int) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "completed").
		Set("chunk_count = ?", chunkCount).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func (r *IndexJobRepository) MarkPendingRetry(ctx context.Context, jobID string, message string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "pending").
		Set("error_message = ?", message).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func (r *IndexJobRepository) MarkFailed(ctx context.Context, jobID string, message string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "failed").
		Set("error_message = ?", message).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func (r *IndexJobRepository) ListRecent(ctx context.Context, rootFolderID string, limit int) ([]*rag_db.IndexJob, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var jobs []*rag_db.IndexJob
	query := r.db.NewSelect().
		Model(&jobs).
		Order("created_at DESC").
		Limit(limit)
	if strings.TrimSpace(rootFolderID) != "" {
		query = query.Where("root_folder_id = ?", rootFolderID)
	}
	err := query.Scan(ctx)
	return jobs, err
}

func compactUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:32]
}
