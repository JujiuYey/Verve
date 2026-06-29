package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "sag-wiki/app/learning/models/db"
)

// 练习与验证数据访问层
type ExerciseRepository struct {
	db *bun.DB
}

func NewExerciseRepository(database *bun.DB) *ExerciseRepository {
	return &ExerciseRepository{db: database}
}

func (r *ExerciseRepository) GetDB() *bun.DB { return r.db }

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
