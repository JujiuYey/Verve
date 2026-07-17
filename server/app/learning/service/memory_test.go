package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
)

type fakeMemoryWriter struct {
	event           *learning_db.LearningMemoryEvent
	items           []*learning_db.LearningMemoryItem
	readItems       []*learning_db.LearningMemoryItem
	readDocumentID  string
	readDocumentIDs []string
}

func (f *fakeMemoryWriter) FindItemsByDocument(_ context.Context, documentID string, _ int) ([]*learning_db.LearningMemoryItem, error) {
	f.readDocumentID = documentID
	return f.readItems, nil
}

func (f *fakeMemoryWriter) FindItemsByDocuments(_ context.Context, documentIDs []string, _ int) ([]*learning_db.LearningMemoryItem, error) {
	f.readDocumentIDs = append([]string(nil), documentIDs...)
	return f.readItems, nil
}

func (f *fakeMemoryWriter) CreateEvent(_ context.Context, event *learning_db.LearningMemoryEvent) error {
	event.ID = "event-1"
	f.event = event
	return nil
}

func (f *fakeMemoryWriter) CreateItem(_ context.Context, item *learning_db.LearningMemoryItem) error {
	f.items = append(f.items, item)
	return nil
}

type fakeMemoryDocuments struct {
	documents map[string]*wiki_db.Document
}

func (f fakeMemoryDocuments) FindOne(_ context.Context, id string) (*wiki_db.Document, error) {
	if document := f.documents[id]; document != nil {
		return document, nil
	}
	return nil, sql.ErrNoRows
}

type fakeMemoryFolders struct {
	folders []*wiki_db.Folder
}

func (f fakeMemoryFolders) FindOne(_ context.Context, id string) (*wiki_db.Folder, error) {
	for _, folder := range f.folders {
		if folder != nil && folder.ID == id {
			return folder, nil
		}
	}
	return nil, sql.ErrNoRows
}

func TestRecordExplanationReviewStoresEvidenceAndMisconceptionsWithoutMastery(t *testing.T) {
	writer := &fakeMemoryWriter{}
	documents := fakeMemoryDocuments{documents: map[string]*wiki_db.Document{"doc-1": {ID: "doc-1", FolderID: "folder-child"}}}
	folders := fakeMemoryFolders{folders: []*wiki_db.Folder{{ID: "folder-child"}}}
	service := newMemoryService(writer, documents, folders)
	session := &learning_db.LearningSession{ID: "session-1", DocumentID: "doc-1"}
	review := &learning_db.LearningExplanationReview{
		ID:                 "review-1",
		HeardSummary:       "解释了值与类型",
		ClearPoints:        []string{"值有具体类型", "类型决定可用操作"},
		Misconceptions:     []string{"把动态类型说成变量声明类型"},
		FollowUpQuestion:   "接口值的动态类型是什么？",
		ExplanationSummary: "值和类型存在约束关系",
	}

	if err := service.RecordExplanationReview(context.Background(), session, review); err != nil {
		t.Fatal(err)
	}
	if writer.event == nil || writer.event.SourceType != "explanation_review" || writer.event.EventType != "explanation_review" || writer.event.DocumentID == nil || *writer.event.DocumentID != "doc-1" {
		t.Fatalf("event = %#v", writer.event)
	}
	if writer.event.FolderID == nil || *writer.event.FolderID != "folder-child" {
		t.Fatalf("event folder = %#v", writer.event.FolderID)
	}
	if writer.event.SessionID == nil || *writer.event.SessionID != "session-1" || writer.event.SourceID == nil || *writer.event.SourceID != "review-1" {
		t.Fatalf("event source = %#v", writer.event)
	}
	if len(writer.items) != 3 {
		t.Fatalf("items = %#v", writer.items)
	}
	wantKinds := []string{"explanation_evidence", "explanation_evidence", "misconception"}
	for i, item := range writer.items {
		if item.Kind != wantKinds[i] {
			t.Fatalf("item %d kind = %q", i, item.Kind)
		}
		if item.Kind == "mastered_concept" {
			t.Fatal("single explanation created mastered_concept")
		}
		if item.DocumentID == nil || *item.DocumentID != "doc-1" {
			t.Fatalf("item document = %#v", item.DocumentID)
		}
		if item.FolderID == nil || *item.FolderID != "folder-child" {
			t.Fatalf("item folder = %#v", item.FolderID)
		}
	}
	if writer.event.Evidence["follow_up_question"] != review.FollowUpQuestion {
		t.Fatalf("follow-up evidence = %#v", writer.event.Evidence["follow_up_question"])
	}
}

func TestRecordExplanationReviewValidatesInputs(t *testing.T) {
	service := newMemoryService(nil, nil, nil)
	if err := service.RecordExplanationReview(context.Background(), &learning_db.LearningSession{}, &learning_db.LearningExplanationReview{}); err == nil || !strings.Contains(err.Error(), "repository") {
		t.Fatalf("missing repository error = %v", err)
	}
	service = newMemoryService(&fakeMemoryWriter{}, fakeMemoryDocuments{}, fakeMemoryFolders{})
	if err := service.RecordExplanationReview(context.Background(), nil, &learning_db.LearningExplanationReview{}); err == nil || !strings.Contains(err.Error(), "session") {
		t.Fatalf("missing session error = %v", err)
	}
	if err := service.RecordExplanationReview(context.Background(), &learning_db.LearningSession{}, nil); err == nil || !strings.Contains(err.Error(), "review") {
		t.Fatalf("missing review error = %v", err)
	}
}

func TestMemoryServiceFindDocumentItemsForwardsDocumentScope(t *testing.T) {
	writer := &fakeMemoryWriter{readItems: []*learning_db.LearningMemoryItem{{ID: "memory-1"}}}
	service := newMemoryService(writer, nil, nil)
	items, err := service.FindDocumentItems(context.Background(), "doc-1", 20)
	if err != nil {
		t.Fatal(err)
	}
	if writer.readDocumentID != "doc-1" {
		t.Fatalf("scope = document:%q", writer.readDocumentID)
	}
	if len(items) != 1 || items[0].ID != "memory-1" {
		t.Fatalf("items = %#v", items)
	}
}

func TestMemoryServiceFindDocumentItemsBatchForwardsDocumentScope(t *testing.T) {
	writer := &fakeMemoryWriter{readItems: []*learning_db.LearningMemoryItem{{ID: "memory-1"}}}
	service := newMemoryService(writer, nil, nil)
	items, err := service.FindDocumentItemsBatch(context.Background(), []string{"doc-1", "doc-2"}, 20)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(writer.readDocumentIDs, ",") != "doc-1,doc-2" {
		t.Fatalf("document scope = %v", writer.readDocumentIDs)
	}
	if len(items) != 1 || items[0].ID != "memory-1" {
		t.Fatalf("items = %#v", items)
	}
}
