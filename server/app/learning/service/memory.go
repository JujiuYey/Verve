package service

import (
	"context"
	"errors"
	"strings"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/database"
)

type memoryWriter interface {
	CreateEvent(ctx context.Context, event *learning_db.LearningMemoryEvent) error
	CreateItem(ctx context.Context, item *learning_db.LearningMemoryItem) error
}

type MemoryService struct {
	repository memoryWriter
}

func NewMemoryService(db *database.DatabaseService) *MemoryService {
	if db == nil {
		return newMemoryService(nil)
	}
	return newMemoryService(db.Memories)
}

func newMemoryService(repository memoryWriter) *MemoryService {
	return &MemoryService{repository: repository}
}

func (s *MemoryService) RecordExplanationReview(ctx context.Context, userID string, session *learning_db.LearningSession, review *learning_db.LearningExplanationReview) error {
	if session == nil {
		return errors.New("learning session is required")
	}
	if review == nil {
		return errors.New("explanation review is required")
	}
	if s == nil || s.repository == nil {
		return errors.New("memory repository is not configured")
	}

	event := buildExplanationMemoryEvent(userID, session, review)
	if err := s.repository.CreateEvent(ctx, event); err != nil {
		return err
	}
	for _, item := range buildExplanationMemoryItems(userID, session, event.ID, review) {
		if err := s.repository.CreateItem(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func buildExplanationMemoryEvent(userID string, session *learning_db.LearningSession, review *learning_db.LearningExplanationReview) *learning_db.LearningMemoryEvent {
	documentID := session.DocumentID
	sessionID := session.ID
	sourceID := review.ID
	return &learning_db.LearningMemoryEvent{
		UserID: userID, DocumentID: &documentID, SessionID: &sessionID,
		SourceType: "feynman_review", SourceID: &sourceID, EventType: "explanation_review",
		Content: firstNonBlank(review.ExplanationSummary, review.HeardSummary, review.Explanation),
		Evidence: map[string]interface{}{
			"clear_points":       review.ClearPoints,
			"confusing_points":   review.ConfusingPoints,
			"misconceptions":     review.Misconceptions,
			"follow_up_question": review.FollowUpQuestion,
			"ready_to_wrap_up":   review.ReadyToWrapUp,
			"context_sufficient": review.ContextSufficient,
		},
	}
}

func buildExplanationMemoryItems(userID string, session *learning_db.LearningSession, eventID string, review *learning_db.LearningExplanationReview) []*learning_db.LearningMemoryItem {
	items := make([]*learning_db.LearningMemoryItem, 0, len(review.ClearPoints)+len(review.Misconceptions))
	for _, point := range review.ClearPoints {
		if point = strings.TrimSpace(point); point != "" {
			items = append(items, buildExplanationMemoryItem(userID, session.DocumentID, eventID, "explanation_evidence", point))
		}
	}
	for _, misconception := range review.Misconceptions {
		if misconception = strings.TrimSpace(misconception); misconception != "" {
			items = append(items, buildExplanationMemoryItem(userID, session.DocumentID, eventID, "misconception", misconception))
		}
	}
	return items
}

func buildExplanationMemoryItem(userID, documentID, eventID, kind, statement string) *learning_db.LearningMemoryItem {
	return &learning_db.LearningMemoryItem{
		UserID: userID, DocumentID: &documentID, Kind: kind, Statement: statement,
		EvidenceEventIDs: []string{eventID}, Confidence: "observed",
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
