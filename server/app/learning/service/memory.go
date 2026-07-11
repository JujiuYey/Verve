package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
	"verve/infrastructure/database"
)

type memoryWriter interface {
	CreateEvent(ctx context.Context, event *learning_db.LearningMemoryEvent) error
	CreateItem(ctx context.Context, item *learning_db.LearningMemoryItem) error
}

type memoryReader interface {
	FindItemsByDocument(ctx context.Context, userID, documentID string, limit int) ([]*learning_db.LearningMemoryItem, error)
	FindItemsByUser(ctx context.Context, userID, folderID string, limit int) ([]*learning_db.LearningMemoryItem, error)
	FindItemsByFolders(ctx context.Context, userID string, folderIDs []string, limit int) ([]*learning_db.LearningMemoryItem, error)
}

type memoryDocumentFinder interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Document, error)
}

type memoryFolderScope interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Folder, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*wiki_db.Folder, error)
	GetAllSubFolderIDs(ctx context.Context, parentID string) ([]string, error)
}

type MemoryService struct {
	repository memoryWriter
	reader     memoryReader
	documents  memoryDocumentFinder
	folders    memoryFolderScope
}

func NewMemoryService(db *database.DatabaseService) *MemoryService {
	if db == nil {
		return newMemoryService(nil, nil, nil)
	}
	return newMemoryService(db.Memories, db.Documents, db.Folders)
}

func newMemoryService(repository memoryWriter, documents memoryDocumentFinder, folders memoryFolderScope) *MemoryService {
	service := &MemoryService{repository: repository, documents: documents, folders: folders}
	if reader, ok := repository.(memoryReader); ok {
		service.reader = reader
	}
	return service
}

func (s *MemoryService) FindCoachItems(ctx context.Context, userID, rootFolderID string, limit int) ([]*learning_db.LearningMemoryItem, error) {
	if s == nil || s.reader == nil {
		return nil, errors.New("memory repository is not configured")
	}
	rootFolderID = strings.TrimSpace(rootFolderID)
	if rootFolderID == "" {
		return s.reader.FindItemsByUser(ctx, userID, "", limit)
	}
	if s.folders == nil {
		return nil, errors.New("memory folder scope is not configured")
	}
	folders, err := s.folders.List(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	accessibleFolders := FilterCoachFoldersForUser(folders, userID)
	if !CoachFolderIsAccessible(accessibleFolders, rootFolderID) {
		return nil, sql.ErrNoRows
	}
	descendantIDs, err := s.folders.GetAllSubFolderIDs(ctx, rootFolderID)
	if err != nil {
		return nil, err
	}
	accessibleIDs := make(map[string]struct{}, len(accessibleFolders))
	for _, folder := range accessibleFolders {
		accessibleIDs[folder.ID] = struct{}{}
	}
	allowedIDs := make([]string, 0, len(descendantIDs))
	seen := make(map[string]struct{}, len(descendantIDs))
	for _, folderID := range descendantIDs {
		if _, ok := accessibleIDs[folderID]; !ok {
			continue
		}
		if _, ok := seen[folderID]; ok {
			continue
		}
		seen[folderID] = struct{}{}
		allowedIDs = append(allowedIDs, folderID)
	}
	return s.reader.FindItemsByFolders(ctx, userID, allowedIDs, limit)
}

func (s *MemoryService) FindDocumentItems(ctx context.Context, userID, documentID string, limit int) ([]*learning_db.LearningMemoryItem, error) {
	if s == nil || s.reader == nil {
		return nil, errors.New("memory repository is not configured")
	}
	return s.reader.FindItemsByDocument(ctx, userID, documentID, limit)
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
	if err := validateExplanationMemoryScope(userID, session, review); err != nil {
		return err
	}
	if s.documents == nil || s.folders == nil {
		return errors.New("memory document scope is not configured")
	}
	document, err := s.documents.FindOne(ctx, session.DocumentID)
	if err != nil {
		return err
	}
	folder, err := s.folders.FindOne(ctx, document.FolderID)
	if err != nil {
		return err
	}
	if len(FilterCoachFoldersForUser([]*wiki_db.Folder{folder}, userID)) == 0 {
		return sql.ErrNoRows
	}

	event := buildExplanationMemoryEvent(userID, document.FolderID, session, review)
	if err := s.repository.CreateEvent(ctx, event); err != nil {
		return err
	}
	for _, item := range buildExplanationMemoryItems(userID, document.FolderID, session, event.ID, review) {
		if err := s.repository.CreateItem(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func validateExplanationMemoryScope(userID string, session *learning_db.LearningSession, review *learning_db.LearningExplanationReview) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(session.DocumentID) == "" {
		return errors.New("memory user and document scope are required")
	}
	if session.UserID != "" && session.UserID != userID {
		return errors.New("learning session user does not match memory user")
	}
	if review.UserID != "" && review.UserID != userID {
		return errors.New("explanation review user does not match memory user")
	}
	if review.SessionID != "" && review.SessionID != session.ID {
		return errors.New("explanation review session does not match learning session")
	}
	if review.DocumentID != "" && review.DocumentID != session.DocumentID {
		return errors.New("explanation review document does not match learning session")
	}
	return nil
}

func buildExplanationMemoryEvent(userID, folderID string, session *learning_db.LearningSession, review *learning_db.LearningExplanationReview) *learning_db.LearningMemoryEvent {
	folderIDCopy := folderID
	documentID := session.DocumentID
	sessionID := session.ID
	sourceID := review.ID
	return &learning_db.LearningMemoryEvent{
		UserID: userID, FolderID: &folderIDCopy, DocumentID: &documentID, SessionID: &sessionID,
		SourceType: "explanation_review", SourceID: &sourceID, EventType: "explanation_review",
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

func buildExplanationMemoryItems(userID, folderID string, session *learning_db.LearningSession, eventID string, review *learning_db.LearningExplanationReview) []*learning_db.LearningMemoryItem {
	items := make([]*learning_db.LearningMemoryItem, 0, len(review.ClearPoints)+len(review.Misconceptions))
	for _, point := range review.ClearPoints {
		if point = strings.TrimSpace(point); point != "" {
			items = append(items, buildExplanationMemoryItem(userID, folderID, session.DocumentID, eventID, "explanation_evidence", point))
		}
	}
	for _, misconception := range review.Misconceptions {
		if misconception = strings.TrimSpace(misconception); misconception != "" {
			items = append(items, buildExplanationMemoryItem(userID, folderID, session.DocumentID, eventID, "misconception", misconception))
		}
	}
	return items
}

func buildExplanationMemoryItem(userID, folderID, documentID, eventID, kind, statement string) *learning_db.LearningMemoryItem {
	folderIDCopy := folderID
	return &learning_db.LearningMemoryItem{
		UserID: userID, FolderID: &folderIDCopy, DocumentID: &documentID, Kind: kind, Statement: statement,
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
