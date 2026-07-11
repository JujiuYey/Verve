package service

import (
	"context"
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"
)

type fakeMemoryWriter struct {
	event *learning_db.LearningMemoryEvent
	items []*learning_db.LearningMemoryItem
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

func TestRecordExplanationReviewStoresEvidenceAndMisconceptionsWithoutMastery(t *testing.T) {
	writer := &fakeMemoryWriter{}
	service := newMemoryService(writer)
	session := &learning_db.LearningSession{ID: "session-1", DocumentID: "doc-1"}
	review := &learning_db.LearningExplanationReview{
		ID: "review-1", Explanation: "值有类型,类型约束操作", HeardSummary: "解释了值与类型",
		ClearPoints:      []string{"值有具体类型", "类型决定可用操作"},
		Misconceptions:   []string{"把动态类型说成变量声明类型"},
		FollowUpQuestion: "接口值的动态类型是什么？", ExplanationSummary: "值和类型存在约束关系",
	}

	if err := service.RecordExplanationReview(context.Background(), "user-1", session, review); err != nil {
		t.Fatal(err)
	}
	if writer.event == nil || writer.event.EventType != "explanation_review" || writer.event.DocumentID == nil || *writer.event.DocumentID != "doc-1" {
		t.Fatalf("event = %#v", writer.event)
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
	}
	if writer.event.Evidence["follow_up_question"] != review.FollowUpQuestion {
		t.Fatalf("follow-up evidence = %#v", writer.event.Evidence["follow_up_question"])
	}
}

func TestRecordExplanationReviewValidatesInputs(t *testing.T) {
	service := newMemoryService(nil)
	if err := service.RecordExplanationReview(context.Background(), "user-1", &learning_db.LearningSession{}, &learning_db.LearningExplanationReview{}); err == nil || !strings.Contains(err.Error(), "repository") {
		t.Fatalf("missing repository error = %v", err)
	}
	service = newMemoryService(&fakeMemoryWriter{})
	if err := service.RecordExplanationReview(context.Background(), "user-1", nil, &learning_db.LearningExplanationReview{}); err == nil || !strings.Contains(err.Error(), "session") {
		t.Fatalf("missing session error = %v", err)
	}
	if err := service.RecordExplanationReview(context.Background(), "user-1", &learning_db.LearningSession{}, nil); err == nil || !strings.Contains(err.Error(), "review") {
		t.Fatalf("missing review error = %v", err)
	}
}
