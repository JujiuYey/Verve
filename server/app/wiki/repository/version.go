package repository

import (
	"context"
	"fmt"
	"time"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"

	"github.com/uptrace/bun"
)

// ApplyVersionInput 是发布一个不可变文档版本所需的已写入对象元数据。
type ApplyVersionInput struct {
	ChangeRequestID *string
	DocumentID      string
	ExpectedVersion int64
	ObjectPath      string
	ContentHash     string
	FileSize        int64
	ChangedBy       string
	ChangeSummary   string
}

// VersionRepository 以事务方式发布文档版本。
type VersionRepository struct {
	db *bun.DB
}

func NewVersionRepository(db *bun.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) ApplyChangeRequest(ctx context.Context, input ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	if input.ChangeRequestID == nil {
		return nil, nil, ErrChangeRequestNotProposed
	}

	var revision *wiki_db.DocumentRevision
	var job *rag_db.IndexJob
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		request := new(wiki_db.DocumentChangeRequest)
		if err := tx.NewSelect().Model(request).Where("id = ?", *input.ChangeRequestID).For("UPDATE").Scan(ctx); err != nil {
			return fmt.Errorf("锁定文档变更申请失败: %w", err)
		}
		if request.Status == wiki_db.ChangeRequestStatusApplied {
			var err error
			revision, job, err = findAppliedVersion(ctx, tx, request.ID)
			return err
		}
		if request.Status != wiki_db.ChangeRequestStatusProposed {
			return ErrChangeRequestNotProposed
		}
		if request.DocumentID != input.DocumentID || request.BaseVersion != input.ExpectedVersion {
			return ErrVersionConflict
		}

		var err error
		revision, job, err = applyLockedVersion(ctx, tx, input, request)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return revision, job, nil
}

func (r *VersionRepository) ApplyDirectEdit(ctx context.Context, input ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	var revision *wiki_db.DocumentRevision
	var job *rag_db.IndexJob
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var err error
		revision, job, err = applyLockedVersion(ctx, tx, input, nil)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return revision, job, nil
}

// FindAppliedChangeRequest 返回已应用申请对应的持久化修订和索引任务。
func (r *VersionRepository) FindAppliedChangeRequest(ctx context.Context, changeRequestID string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	return findAppliedVersion(ctx, r.db, changeRequestID)
}

func applyLockedVersion(ctx context.Context, tx bun.Tx, input ApplyVersionInput, request *wiki_db.DocumentChangeRequest) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	document := new(wiki_db.Document)
	if err := tx.NewSelect().Model(document).Where("id = ?", input.DocumentID).For("UPDATE").Scan(ctx); err != nil {
		return nil, nil, fmt.Errorf("锁定文档失败: %w", err)
	}
	if document.CurrentVersion != input.ExpectedVersion {
		if request != nil {
			if _, err := tx.NewUpdate().Model((*wiki_db.DocumentChangeRequest)(nil)).
				Set("status = ?", wiki_db.ChangeRequestStatusConflict).
				Set("updated_at = ?", time.Now()).
				Where("id = ?", request.ID).
				Exec(ctx); err != nil {
				return nil, nil, fmt.Errorf("标记文档变更冲突失败: %w", err)
			}
		}
		return nil, nil, ErrVersionConflict
	}

	nextVersion := document.CurrentVersion + 1
	revision := &wiki_db.DocumentRevision{
		ID:              newCompactID(),
		DocumentID:      document.ID,
		Version:         nextVersion,
		ObjectPath:      input.ObjectPath,
		ContentHash:     input.ContentHash,
		FileSize:        input.FileSize,
		ChangeRequestID: input.ChangeRequestID,
		ChangedBy:       input.ChangedBy,
		ChangeSummary:   input.ChangeSummary,
	}
	job := &rag_db.IndexJob{
		ID:              newCompactID(),
		DocumentID:      document.ID,
		DocumentVersion: nextVersion,
		ObjectPath:      input.ObjectPath,
		Status:          "pending",
		MaxAttempts:     3,
	}
	if _, err := tx.NewInsert().Model(revision).Exec(ctx); err != nil {
		return nil, nil, fmt.Errorf("创建文档修订失败: %w", err)
	}
	if _, err := tx.NewUpdate().Model((*wiki_db.Document)(nil)).
		Set("current_version = ?", nextVersion).
		Set("file_path = ?", input.ObjectPath).
		Set("content_hash = ?", input.ContentHash).
		Set("file_size = ?", input.FileSize).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", document.ID).
		Exec(ctx); err != nil {
		return nil, nil, fmt.Errorf("更新当前文档版本失败: %w", err)
	}
	if request != nil {
		now := time.Now()
		if _, err := tx.NewUpdate().Model((*wiki_db.DocumentChangeRequest)(nil)).
			Set("status = ?", wiki_db.ChangeRequestStatusApplied).
			Set("applied_version = ?", nextVersion).
			Set("applied_at = ?", now).
			Set("updated_at = ?", now).
			Where("id = ?", request.ID).
			Exec(ctx); err != nil {
			return nil, nil, fmt.Errorf("应用文档变更申请失败: %w", err)
		}
	}
	if _, err := tx.NewInsert().Model(job).Exec(ctx); err != nil {
		return nil, nil, fmt.Errorf("创建文档索引任务失败: %w", err)
	}
	return revision, job, nil
}

type versionQuery interface {
	NewSelect() *bun.SelectQuery
}

func findAppliedVersion(ctx context.Context, db versionQuery, changeRequestID string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	revision := new(wiki_db.DocumentRevision)
	if err := db.NewSelect().Model(revision).Where("change_request_id = ?", changeRequestID).Scan(ctx); err != nil {
		return nil, nil, fmt.Errorf("获取已应用文档修订失败: %w", err)
	}
	job := new(rag_db.IndexJob)
	if err := db.NewSelect().Model(job).
		Where("document_id = ?", revision.DocumentID).
		Where("document_version = ?", revision.Version).
		Scan(ctx); err != nil {
		return nil, nil, fmt.Errorf("获取已应用索引任务失败: %w", err)
	}
	return revision, job, nil
}
