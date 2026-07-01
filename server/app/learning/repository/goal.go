package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "sag-wiki/app/learning/models/db"
)

// 学习目标数据访问层
type GoalRepository struct {
	db *bun.DB
}

func NewGoalRepository(database *bun.DB) *GoalRepository {
	return &GoalRepository{db: database}
}

func (r *GoalRepository) GetDB() *bun.DB { return r.db }

func (r *GoalRepository) FindOne(ctx context.Context, id string) (*learning_db.LearningGoal, error) {
	goal := new(learning_db.LearningGoal)
	if err := r.db.NewSelect().Model(goal).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return goal, nil
}

// 按用户分页(只返回本人目标)
func (r *GoalRepository) FindByUser(ctx context.Context, userID string, offset, limit int) ([]*learning_db.LearningGoal, int, error) {
	var goals []*learning_db.LearningGoal
	query := r.db.NewSelect().Model(&goals).Where("user_id = ?", userID).Order("created_at DESC")

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, err
	}
	return goals, total, nil
}

// 按用户列出进行中的目标，用于学习调度入口生成继续选项。
func (r *GoalRepository) FindActiveByUser(ctx context.Context, userID string, limit int) ([]*learning_db.LearningGoal, error) {
	var goals []*learning_db.LearningGoal
	query := r.db.NewSelect().Model(&goals).
		Where("user_id = ?", userID).
		Where("status = ?", "active").
		Order("updated_at DESC", "created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return goals, nil
}

func (r *GoalRepository) Create(ctx context.Context, goal *learning_db.LearningGoal) error {
	goal.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(goal).Exec(ctx)
	return err
}

func (r *GoalRepository) Update(ctx context.Context, goal *learning_db.LearningGoal) error {
	_, err := r.db.NewUpdate().Model(goal).Where("id = ?", goal.ID).Exec(ctx)
	return err
}

func (r *GoalRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*learning_db.LearningGoal)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
