package repository

import (
	"context"
	"sort"

	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	wiki_db "verve/app/wiki/models/db"
)

type TimelineRepository struct {
	db *bun.DB
}

func NewTimelineRepository(db *bun.DB) *TimelineRepository {
	return &TimelineRepository{db: db}
}

func (r *TimelineRepository) FindBySession(ctx context.Context, sessionID string) ([]*learning_payload.TimelineItem, error) {
	turns := make([]*learning_db.LearningTurn, 0)
	if err := r.db.NewSelect().Model(&turns).Where("session_id = ?", sessionID).OrderExpr("created_at ASC, id ASC").Scan(ctx); err != nil {
		return nil, err
	}
	return r.loadItems(ctx, turns)
}

func (r *TimelineRepository) FindByTurn(ctx context.Context, turnID string) (*learning_payload.TimelineItem, error) {
	turn := new(learning_db.LearningTurn)
	if err := r.db.NewSelect().Model(turn).Where("id = ?", turnID).Scan(ctx); err != nil {
		return nil, err
	}
	items, err := r.loadItems(ctx, []*learning_db.LearningTurn{turn})
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

func (r *TimelineRepository) loadItems(ctx context.Context, turns []*learning_db.LearningTurn) ([]*learning_payload.TimelineItem, error) {
	if len(turns) == 0 {
		return []*learning_payload.TimelineItem{}, nil
	}
	ids := make([]string, 0, len(turns))
	for _, turn := range turns {
		ids = append(ids, turn.ID)
	}
	messages := make([]*learning_db.LearningMessage, 0)
	if err := r.db.NewSelect().Model(&messages).Where("turn_id IN (?)", bun.In(ids)).OrderExpr("created_at ASC, id ASC").Scan(ctx); err != nil {
		return nil, err
	}
	reviews := make([]*learning_db.LearningExplanationReview, 0)
	if err := r.db.NewSelect().Model(&reviews).Where("turn_id IN (?)", bun.In(ids)).Scan(ctx); err != nil {
		return nil, err
	}
	interventions := make([]*learning_db.LearningTeachingIntervention, 0)
	if err := r.db.NewSelect().Model(&interventions).Where("turn_id IN (?)", bun.In(ids)).Scan(ctx); err != nil {
		return nil, err
	}
	requests := make([]*wiki_db.DocumentChangeRequest, 0)
	if err := r.db.NewSelect().Model(&requests).
		Where("source_type = ?", "learning_turn").
		Where("source_id IN (?)", bun.In(ids)).Scan(ctx); err != nil {
		return nil, err
	}
	return assembleTimeline(turns, messages, reviews, interventions, requests), nil
}

func assembleTimeline(turns []*learning_db.LearningTurn, messages []*learning_db.LearningMessage, reviews []*learning_db.LearningExplanationReview, interventions []*learning_db.LearningTeachingIntervention, requests []*wiki_db.DocumentChangeRequest) []*learning_payload.TimelineItem {
	sortedTurns := append([]*learning_db.LearningTurn(nil), turns...)
	sort.SliceStable(sortedTurns, func(i, j int) bool {
		if sortedTurns[i].CreatedAt.Equal(sortedTurns[j].CreatedAt) {
			return sortedTurns[i].ID < sortedTurns[j].ID
		}
		return sortedTurns[i].CreatedAt.Before(sortedTurns[j].CreatedAt)
	})
	items := make(map[string]*learning_payload.TimelineItem, len(sortedTurns))
	result := make([]*learning_payload.TimelineItem, 0, len(sortedTurns))
	for _, turn := range sortedTurns {
		item := &learning_payload.TimelineItem{Turn: turn}
		items[turn.ID] = item
		result = append(result, item)
	}
	for _, message := range messages {
		item := items[message.TurnID]
		if item == nil {
			continue
		}
		if message.Role == "user" && item.UserMessage == nil {
			item.UserMessage = message
		}
		if message.Role == "assistant" && item.AssistantMessage == nil {
			item.AssistantMessage = message
		}
	}
	for _, review := range reviews {
		if item := items[review.TurnID]; item != nil {
			item.Artifact = &learning_payload.TurnArtifact{Type: learning_payload.ArtifactExplanationReview, Data: review}
		}
	}
	for _, intervention := range interventions {
		if item := items[intervention.TurnID]; item != nil {
			item.Artifact = &learning_payload.TurnArtifact{Type: learning_payload.ArtifactTeachingIntervention, Data: intervention}
		}
	}
	for _, request := range requests {
		if item := items[request.SourceID]; item != nil {
			item.Artifact = &learning_payload.TurnArtifact{Type: learning_payload.ArtifactWikiChangeRequest, Data: request}
		}
	}
	return result
}
