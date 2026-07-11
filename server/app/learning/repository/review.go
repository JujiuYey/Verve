package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

type ReviewRepository struct {
	db *bun.DB
}

func NewReviewRepository(database *bun.DB) *ReviewRepository {
	return &ReviewRepository{db: database}
}

func (r *ReviewRepository) Create(ctx context.Context, review *learning_db.LearningExplanationReview) error {
	review.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(review).Exec(ctx)
	return err
}

func (r *ReviewRepository) FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningExplanationReview, error) {
	reviews := make([]*learning_db.LearningExplanationReview, 0)
	err := r.db.NewSelect().
		Model(&reviews).
		Where("session_id = ?", sessionID).
		Order("created_at ASC", "id ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}
