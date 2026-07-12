package repository

import (
	"context"

	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

type InterventionRepository struct {
	db *bun.DB
}

func NewInterventionRepository(db *bun.DB) *InterventionRepository {
	return &InterventionRepository{db: db}
}

func (r *InterventionRepository) FindByTurn(ctx context.Context, turnID string) (*learning_db.LearningTeachingIntervention, error) {
	intervention := new(learning_db.LearningTeachingIntervention)
	if err := r.db.NewSelect().Model(intervention).Where("turn_id = ?", turnID).Scan(ctx); err != nil {
		return nil, err
	}
	return intervention, nil
}
