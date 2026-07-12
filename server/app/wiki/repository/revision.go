package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"

	"github.com/uptrace/bun"
)

// RevisionRepository 管理文档不可变修订。
type RevisionRepository struct {
	db *bun.DB
}

// LegacyRevisionSnapshot 是旧文档首次快照的已写入对象元数据。
type LegacyRevisionSnapshot struct {
	DocumentID    string
	ObjectPath    string
	ContentHash   string
	FileSize      int64
	ChangedBy     string
	ChangeSummary string
}

func NewRevisionRepository(db *bun.DB) *RevisionRepository {
	return &RevisionRepository{db: db}
}

// CreateInitial 原子写入新文档、初始修订和对应索引任务。
func (r *RevisionRepository) CreateInitial(ctx context.Context, document *wiki_db.Document, revision *wiki_db.DocumentRevision, job *rag_db.IndexJob) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(document).Exec(ctx); err != nil {
			return fmt.Errorf("创建文档失败: %w", err)
		}
		if _, err := tx.NewInsert().Model(revision).Exec(ctx); err != nil {
			return fmt.Errorf("创建初始修订失败: %w", err)
		}
		if _, err := tx.NewInsert().Model(job).Exec(ctx); err != nil {
			return fmt.Errorf("创建索引任务失败: %w", err)
		}
		return nil
	})
}

// EnsureLegacyRevision 为尚未版本化的文档持久化版本一快照。
func (r *RevisionRepository) EnsureLegacyRevision(ctx context.Context, snapshot *LegacyRevisionSnapshot) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		document := new(wiki_db.Document)
		if err := tx.NewSelect().Model(document).Where("id = ?", snapshot.DocumentID).For("UPDATE").Scan(ctx); err != nil {
			return fmt.Errorf("锁定旧文档失败: %w", err)
		}

		existing := new(wiki_db.DocumentRevision)
		err := tx.NewSelect().Model(existing).
			Where("document_id = ?", snapshot.DocumentID).
			Where("version = ?", 1).
			For("UPDATE").
			Scan(ctx)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("查询旧文档修订失败: %w", err)
		}
		if document.CurrentVersion != 1 {
			return ErrVersionConflict
		}

		revision := &wiki_db.DocumentRevision{
			ID:            newCompactID(),
			DocumentID:    snapshot.DocumentID,
			Version:       1,
			ObjectPath:    snapshot.ObjectPath,
			ContentHash:   snapshot.ContentHash,
			FileSize:      snapshot.FileSize,
			ChangedBy:     snapshot.ChangedBy,
			ChangeSummary: snapshot.ChangeSummary,
		}
		if _, err := tx.NewInsert().Model(revision).Exec(ctx); err != nil {
			return fmt.Errorf("创建旧文档修订失败: %w", err)
		}
		if _, err := tx.NewUpdate().Model((*wiki_db.Document)(nil)).
			Set("file_path = ?", snapshot.ObjectPath).
			Set("content_hash = ?", snapshot.ContentHash).
			Set("file_size = ?", snapshot.FileSize).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", snapshot.DocumentID).
			Exec(ctx); err != nil {
			return fmt.Errorf("更新旧文档快照失败: %w", err)
		}
		return nil
	})
}

func (r *RevisionRepository) ListRevisionObjectPaths(ctx context.Context, documentID string) ([]string, error) {
	paths := make([]string, 0)
	err := r.db.NewSelect().
		Model((*wiki_db.DocumentRevision)(nil)).
		Column("object_path").
		Where("document_id = ?", documentID).
		Order("version ASC").
		Scan(ctx, &paths)
	if err != nil {
		return nil, fmt.Errorf("获取文档修订对象路径失败: %w", err)
	}
	return paths, nil
}
