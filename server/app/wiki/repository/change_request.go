package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	wiki_db "verve/app/wiki/models/db"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var (
	ErrVersionConflict          = errors.New("文档版本冲突")
	ErrChangeRequestNotProposed = errors.New("变更申请不是待应用状态")
	ErrChangeRequestForbidden   = errors.New("无权操作变更申请")
)

// ChangeRequestRepository 管理文档变更申请。
type ChangeRequestRepository struct {
	db *bun.DB
}

func NewChangeRequestRepository(db *bun.DB) *ChangeRequestRepository {
	return &ChangeRequestRepository{db: db}
}

func (r *ChangeRequestRepository) CreateProposal(ctx context.Context, request *wiki_db.DocumentChangeRequest) error {
	if request.ID == "" {
		request.ID = newCompactID()
	}
	if request.Status == "" {
		request.Status = wiki_db.ChangeRequestStatusProposed
	}
	result, err := r.db.NewInsert().Model(request).
		On("CONFLICT (document_id, request_id) DO NOTHING").
		Returning("").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建文档变更申请失败: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("读取文档变更申请写入结果失败: %w", err)
	}
	if affected == 0 {
		existing := new(wiki_db.DocumentChangeRequest)
		if err := r.db.NewSelect().Model(existing).
			Where("document_id = ?", request.DocumentID).
			Where("request_id = ?", request.RequestID).
			Scan(ctx); err != nil {
			return fmt.Errorf("读取已有文档变更申请失败: %w", err)
		}
		if existing.SourceType != request.SourceType || existing.SourceID != request.SourceID || existing.Instruction != request.Instruction {
			return ErrVersionConflict
		}
		*request = *existing
	}
	return nil
}

func (r *ChangeRequestRepository) FindChangeRequest(ctx context.Context, id string) (*wiki_db.DocumentChangeRequest, error) {
	request := new(wiki_db.DocumentChangeRequest)
	if err := r.db.NewSelect().Model(request).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, fmt.Errorf("获取文档变更申请失败: %w", err)
	}
	return request, nil
}

// CancelChangeRequest 只允许申请人取消待应用申请，重复取消保持成功。
func (r *ChangeRequestRepository) CancelChangeRequest(ctx context.Context, id, userID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		request := new(wiki_db.DocumentChangeRequest)
		if err := tx.NewSelect().Model(request).Where("id = ?", id).For("UPDATE").Scan(ctx); err != nil {
			if err == sql.ErrNoRows {
				return err
			}
			return fmt.Errorf("锁定文档变更申请失败: %w", err)
		}
		if request.RequestedBy != userID {
			return ErrChangeRequestForbidden
		}
		if request.Status == wiki_db.ChangeRequestStatusCancelled {
			return nil
		}
		if request.Status != wiki_db.ChangeRequestStatusProposed {
			return ErrChangeRequestNotProposed
		}
		if _, err := tx.NewUpdate().Model((*wiki_db.DocumentChangeRequest)(nil)).
			Set("status = ?", wiki_db.ChangeRequestStatusCancelled).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", id).
			Exec(ctx); err != nil {
			return fmt.Errorf("取消文档变更申请失败: %w", err)
		}
		return nil
	})
}

func newCompactID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
