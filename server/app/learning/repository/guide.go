package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 导学结果缓存数据访问层
type GuideRepository struct {
	db *bun.DB
}

func NewGuideRepository(database *bun.DB) *GuideRepository {
	return &GuideRepository{db: database}
}

func (r *GuideRepository) FindByObjectiveAndHash(ctx context.Context, objectiveID, contentHash string) (*learning_db.LearningGuide, error) {
	guide := new(learning_db.LearningGuide)
	if err := r.db.NewSelect().Model(guide).
		Where("objective_id = ?", objectiveID).
		Where("content_hash = ?", contentHash).
		Scan(ctx); err != nil {
		return nil, err
	}
	return guide, nil
}

func (r *GuideRepository) Upsert(ctx context.Context, guide *learning_db.LearningGuide) error {
	if guide.ID == "" {
		guide.ID = newID()
	}
	guide.UpdatedAt = time.Now()
	_, err := r.db.NewInsert().Model(guide).
		On("CONFLICT (objective_id, content_hash) DO UPDATE").
		Set("result = EXCLUDED.result").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func newID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
