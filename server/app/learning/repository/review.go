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
		ColumnExpr("lt.session_id AS session_id").
		ColumnExpr("ls.document_id AS document_id").
		ColumnExpr("ls.user_id AS user_id").
		ColumnExpr("user_message.content AS explanation").
		Join("JOIN learning_turns AS lt ON lt.id = ler.turn_id").
		Join("JOIN learning_sessions AS ls ON ls.id = lt.session_id").
		Join("JOIN learning_messages AS user_message ON user_message.turn_id = lt.id AND user_message.role = 'user'")
}
