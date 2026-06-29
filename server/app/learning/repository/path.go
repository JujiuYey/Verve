package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "sag-wiki/app/learning/models/db"
)

// 学习路线数据访问层
type PathRepository struct {
	db *bun.DB
}

func NewPathRepository(database *bun.DB) *PathRepository {
	return &PathRepository{db: database}
}

func (r *PathRepository) GetDB() *bun.DB { return r.db }

func (r *PathRepository) FindOne(ctx context.Context, id string) (*learning_db.LearningPath, error) {
	path := new(learning_db.LearningPath)
	if err := r.db.NewSelect().Model(path).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return path, nil
}

// 一个目标一条路线
func (r *PathRepository) FindByGoal(ctx context.Context, goalID string) (*learning_db.LearningPath, error) {
	path := new(learning_db.LearningPath)
	if err := r.db.NewSelect().Model(path).Where("goal_id = ?", goalID).Scan(ctx); err != nil {
		return nil, err
	}
	return path, nil
}

func (r *PathRepository) Create(ctx context.Context, path *learning_db.LearningPath) error {
	path.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(path).Exec(ctx)
	return err
}

func (r *PathRepository) Update(ctx context.Context, path *learning_db.LearningPath) error {
	_, err := r.db.NewUpdate().Model(path).Where("id = ?", path.ID).Exec(ctx)
	return err
}
