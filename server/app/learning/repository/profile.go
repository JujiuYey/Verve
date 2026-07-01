package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 学习画像数据访问层
type ProfileRepository struct {
	db *bun.DB
}

func NewProfileRepository(database *bun.DB) *ProfileRepository {
	return &ProfileRepository{db: database}
}

func (r *ProfileRepository) GetDB() *bun.DB { return r.db }

// 一个目标一份画像
func (r *ProfileRepository) FindByGoal(ctx context.Context, goalID string) (*learning_db.LearningProfile, error) {
	profile := new(learning_db.LearningProfile)
	if err := r.db.NewSelect().Model(profile).Where("goal_id = ?", goalID).Scan(ctx); err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *ProfileRepository) Create(ctx context.Context, profile *learning_db.LearningProfile) error {
	profile.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(profile).Exec(ctx)
	return err
}

func (r *ProfileRepository) Update(ctx context.Context, profile *learning_db.LearningProfile) error {
	_, err := r.db.NewUpdate().Model(profile).Where("id = ?", profile.ID).Exec(ctx)
	return err
}
