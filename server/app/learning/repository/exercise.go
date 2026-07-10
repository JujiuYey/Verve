package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 练习与验证数据访问层
type ExerciseRepository struct {
	db *bun.DB
}

func NewExerciseRepository(database *bun.DB) *ExerciseRepository {
	return &ExerciseRepository{db: database}
}

func (r *ExerciseRepository) GetDB() *bun.DB { return r.db }

// 按用户分页列出练习记录
func (r *ExerciseRepository) FindByUser(ctx context.Context, userID string, offset, limit int) ([]*learning_db.LearningExercise, int, error) {
	var exercises []*learning_db.LearningExercise
	query := r.db.NewSelect().Model(&exercises).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, err
	}
	return exercises, total, nil
}

// 按用户和小目标分页列出练习记录
func (r *ExerciseRepository) FindByUserAndObjective(ctx context.Context, userID, objectiveID string, offset, limit int) ([]*learning_db.LearningExercise, int, error) {
	var exercises []*learning_db.LearningExercise
	query := r.db.NewSelect().Model(&exercises).
		Where("user_id = ?", userID).
		Where("objective_id = ?", objectiveID).
		Order("created_at DESC")

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, err
	}
	return exercises, total, nil
}

// 最近验证记录，用于学习调度器判断薄弱点和复习动作。
func (r *ExerciseRepository) FindRecentByUser(ctx context.Context, userID string, limit int) ([]*learning_db.LearningExercise, error) {
	var exercises []*learning_db.LearningExercise
	query := r.db.NewSelect().Model(&exercises).
		Where("user_id = ?", userID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return exercises, nil
}

// 按会话列出练习记录
func (r *ExerciseRepository) FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningExercise, error) {
	var exercises []*learning_db.LearningExercise
	err := r.db.NewSelect().Model(&exercises).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return exercises, nil
}

func (r *ExerciseRepository) Create(ctx context.Context, exercise *learning_db.LearningExercise) error {
	exercise.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(exercise).Exec(ctx)
	return err
}
