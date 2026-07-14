package repository

import (
	"context"

	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

type ReviewRepository struct {
	db *bun.DB
}

func NewReviewRepository(database *bun.DB) *ReviewRepository {
	return &ReviewRepository{db: database}
}

func (r *ReviewRepository) FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningExplanationReview, error) {
	reviews := make([]*learning_db.LearningExplanationReview, 0)
	err := reviewSelect(r.db, &reviews).
		Where("lt.session_id = ?", sessionID).
		OrderExpr("ler.created_at ASC, ler.id ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) FindByTurn(ctx context.Context, turnID string) (*learning_db.LearningExplanationReview, error) {
	review := new(learning_db.LearningExplanationReview)
	if err := reviewSelect(r.db, review).Where("ler.turn_id = ?", turnID).Scan(ctx); err != nil {
		return nil, err
	}
	return review, nil
}

func reviewSelect(db bun.IDB, model interface{}) *bun.SelectQuery {
	return db.NewSelect().Model(model).
		ColumnExpr("ler.*").
		Join("JOIN learning_turns AS lt ON lt.id = ler.turn_id")
}
