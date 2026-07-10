package service

import (
	"context"
	"errors"
	"strings"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/database"
)

// MemoryService records learning observations into reusable memory.
type MemoryService struct {
	db *database.DatabaseService
}

func NewMemoryService(db *database.DatabaseService) *MemoryService {
	return &MemoryService{db: db}
}

func (s *MemoryService) RecordExerciseJudgement(ctx context.Context, userID string, obj *learning_db.LearningObjective, sessionID string, result *JudgeResult) error {
	if obj == nil {
		return errors.New("learning objective is required")
	}
	if result == nil {
		return errors.New("judge result is required")
	}
	if s == nil || s.db == nil || s.db.Memories == nil {
		return errors.New("memory database repository is not configured")
	}

	event := buildExerciseMemoryEvent(userID, obj, sessionID, result)
	if err := s.db.Memories.CreateEvent(ctx, event); err != nil {
		return err
	}

	for _, item := range buildExerciseMemoryItems(userID, obj, event.ID, result) {
		if err := s.db.Memories.CreateItem(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func buildExerciseMemoryEvent(userID string, obj *learning_db.LearningObjective, sessionID string, result *JudgeResult) *learning_db.LearningMemoryEvent {
	event := &learning_db.LearningMemoryEvent{
		UserID:      userID,
		FolderID:    obj.SourceFolderID,
		DocumentID:  obj.SourceDocumentID,
		ObjectiveID: &obj.ID,
		SourceType:  "exercise",
		EventType:   "examiner_judgement",
		Content:     firstNonBlank(result.Feedback, result.Evidence, obj.Title),
		Evidence: map[string]interface{}{
			"verdict":                result.Verdict,
			"mastery_after":          result.MasteryAfter,
			"evidence":               result.Evidence,
			"weak_points":            result.WeakPoints,
			"improvement_suggestion": result.ImprovementSuggestion,
			"review_required":        result.ReviewRequired,
		},
	}

	if sessionID != "" {
		event.SessionID = &sessionID
		event.SourceID = &sessionID
	}

	return event
}

func buildExerciseMemoryItems(userID string, obj *learning_db.LearningObjective, eventID string, result *JudgeResult) []*learning_db.LearningMemoryItem {
	items := make([]*learning_db.LearningMemoryItem, 0, 2)
	title := strings.TrimSpace(obj.Title)
	if result.Verdict == "pass" && title != "" {
		items = append(items, buildExerciseMemoryItem(userID, obj, eventID, "mastered_concept", "用户已经能解释："+title))
	}

	evidence := strings.TrimSpace(result.Evidence)
	if evidence != "" {
		items = append(items, buildExerciseMemoryItem(userID, obj, eventID, "verification_evidence", evidence))
	}

	return items
}

func buildExerciseMemoryItem(userID string, obj *learning_db.LearningObjective, eventID string, kind string, statement string) *learning_db.LearningMemoryItem {
	return &learning_db.LearningMemoryItem{
		UserID:           userID,
		FolderID:         obj.SourceFolderID,
		DocumentID:       obj.SourceDocumentID,
		ObjectiveID:      &obj.ID,
		Kind:             kind,
		Statement:        statement,
		EvidenceEventIDs: []string{eventID},
		Confidence:       "observed",
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
