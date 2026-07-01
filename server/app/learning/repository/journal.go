package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 学习日志数据访问层
type JournalRepository struct {
	db *bun.DB
}

func NewJournalRepository(database *bun.DB) *JournalRepository {
	return &JournalRepository{db: database}
}

func (r *JournalRepository) GetDB() *bun.DB { return r.db }

// 按用户分页(只返回本人日志,按日期倒序)
func (r *JournalRepository) FindByUser(ctx context.Context, userID string, offset, limit int) ([]*learning_db.LearningJournal, int, error) {
	var journals []*learning_db.LearningJournal
	query := r.db.NewSelect().Model(&journals).Where("user_id = ?", userID).Order("date DESC")

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, err
	}
	return journals, total, nil
}

func (r *JournalRepository) Create(ctx context.Context, journal *learning_db.LearningJournal) error {
	journal.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(journal).Exec(ctx)
	return err
}
