package service

import (
	"context"
	"reflect"
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"
)

func TestBuildExerciseMemoryEventAndItemsForPassVerdictWithEvidence(t *testing.T) {
	folderID := "folder-1"
	documentID := "doc-1"
	obj := &learning_db.LearningObjective{
		ID:               "objective-1",
		Title:            "值类型与字面量",
		SourceFolderID:   &folderID,
		SourceDocumentID: &documentID,
	}
	result := &JudgeResult{
		Verdict:               "pass",
		MasteryAfter:          "explained",
		Feedback:              "解释清楚了值必须有类型",
		Evidence:              "能说明类型由编译期检查",
		WeakPoints:            []string{"缺少反例"},
		ImprovementSuggestion: "补一个 int x = \"hello\" 的反例",
		ReviewRequired:        false,
	}

	event := buildExerciseMemoryEvent("user-1", obj, "session-1", result)
	event.ID = "event-1"
	items := buildExerciseMemoryItems("user-1", obj, event.ID, result)

	if event.UserID != "user-1" {
		t.Fatalf("event user id = %q", event.UserID)
	}
	if event.FolderID != &folderID || event.DocumentID != &documentID || event.ObjectiveID == nil || *event.ObjectiveID != obj.ID {
		t.Fatalf("event source ids = folder:%v document:%v objective:%v", event.FolderID, event.DocumentID, event.ObjectiveID)
	}
	if event.SessionID == nil || *event.SessionID != "session-1" {
		t.Fatalf("event session id = %v", event.SessionID)
	}
	if event.SourceType != "exercise" || event.SourceID == nil || *event.SourceID != "session-1" {
		t.Fatalf("event source = %q %v", event.SourceType, event.SourceID)
	}
	if event.EventType != "examiner_judgement" {
		t.Fatalf("event type = %q", event.EventType)
	}
	if event.Content != result.Feedback {
		t.Fatalf("event content = %q", event.Content)
	}

	assertEvidence(t, event.Evidence, "verdict", result.Verdict)
	assertEvidence(t, event.Evidence, "mastery_after", result.MasteryAfter)
	assertEvidence(t, event.Evidence, "evidence", result.Evidence)
	assertEvidence(t, event.Evidence, "improvement_suggestion", result.ImprovementSuggestion)
	assertEvidence(t, event.Evidence, "review_required", result.ReviewRequired)
	if !reflect.DeepEqual(event.Evidence["weak_points"], result.WeakPoints) {
		t.Fatalf("weak_points evidence = %#v", event.Evidence["weak_points"])
	}

	if len(items) != 2 {
		t.Fatalf("item count = %d, items = %#v", len(items), items)
	}
	if items[0].Kind != "mastered_concept" {
		t.Fatalf("first item kind = %q", items[0].Kind)
	}
	if items[0].Statement != "用户已经能解释：值类型与字面量" {
		t.Fatalf("first item statement = %q", items[0].Statement)
	}
	if items[1].Kind != "verification_evidence" || items[1].Statement != result.Evidence {
		t.Fatalf("second item = %#v", items[1])
	}
	for _, item := range items {
		if item.UserID != "user-1" {
			t.Fatalf("item user id = %q", item.UserID)
		}
		if item.FolderID != &folderID || item.DocumentID != &documentID || item.ObjectiveID == nil || *item.ObjectiveID != obj.ID {
			t.Fatalf("item source ids = folder:%v document:%v objective:%v", item.FolderID, item.DocumentID, item.ObjectiveID)
		}
		if item.Confidence != "observed" {
			t.Fatalf("item confidence = %q", item.Confidence)
		}
		if !reflect.DeepEqual(item.EvidenceEventIDs, []string{event.ID}) {
			t.Fatalf("evidence event ids = %#v", item.EvidenceEventIDs)
		}
	}
}

func TestBuildExerciseMemoryItemsDoesNotCreateMasteredConceptForNonPassVerdict(t *testing.T) {
	obj := &learning_db.LearningObjective{ID: "objective-1", Title: "接口基础"}
	result := &JudgeResult{Verdict: "partial", Evidence: "能说出 interface 是方法集合"}

	items := buildExerciseMemoryItems("user-1", obj, "event-1", result)

	if len(items) != 1 {
		t.Fatalf("item count = %d, items = %#v", len(items), items)
	}
	if items[0].Kind != "verification_evidence" {
		t.Fatalf("item kind = %q", items[0].Kind)
	}
}

func TestBuildExerciseMemoryItemsDoesNotPromoteWeakPoints(t *testing.T) {
	obj := &learning_db.LearningObjective{ID: "objective-1", Title: "接口基础"}
	result := &JudgeResult{
		Verdict:    "fail",
		Evidence:   "能说出 interface 是方法集合",
		WeakPoints: []string{"没有解释隐式实现"},
	}

	event := buildExerciseMemoryEvent("user-1", obj, "", result)
	items := buildExerciseMemoryItems("user-1", obj, "event-1", result)

	if !reflect.DeepEqual(event.Evidence["weak_points"], result.WeakPoints) {
		t.Fatalf("weak_points evidence = %#v", event.Evidence["weak_points"])
	}
	for _, item := range items {
		if strings.Contains(item.Statement, "隐式实现") || item.Kind == "weak_point" {
			t.Fatalf("weak point was promoted to item: %#v", item)
		}
	}
}

func TestBuildExerciseMemoryEventContentFallbacks(t *testing.T) {
	obj := &learning_db.LearningObjective{ID: "objective-1", Title: "接口基础"}

	event := buildExerciseMemoryEvent("user-1", obj, "", &JudgeResult{
		Verdict:  "partial",
		Evidence: "能说出 interface 是方法集合",
	})
	if event.Content != "能说出 interface 是方法集合" {
		t.Fatalf("content = %q", event.Content)
	}

	event = buildExerciseMemoryEvent("user-1", obj, "", &JudgeResult{Verdict: "fail"})
	if event.Content != "接口基础" {
		t.Fatalf("content = %q", event.Content)
	}
	if event.SessionID != nil || event.SourceID != nil {
		t.Fatalf("blank session should not set session/source ids: session=%v source=%v", event.SessionID, event.SourceID)
	}
}

func TestRecordExerciseJudgementValidatesInputs(t *testing.T) {
	service := NewMemoryService(nil)
	err := service.RecordExerciseJudgement(context.Background(), "user-1", &learning_db.LearningObjective{}, "", &JudgeResult{})
	if err == nil || !strings.Contains(err.Error(), "memory database repository is not configured") {
		t.Fatalf("expected missing repository error, got %v", err)
	}

	service = NewMemoryService(nil)
	err = service.RecordExerciseJudgement(context.Background(), "user-1", nil, "", &JudgeResult{})
	if err == nil || !strings.Contains(err.Error(), "learning objective is required") {
		t.Fatalf("expected objective error, got %v", err)
	}

	err = service.RecordExerciseJudgement(context.Background(), "user-1", &learning_db.LearningObjective{}, "", nil)
	if err == nil || !strings.Contains(err.Error(), "judge result is required") {
		t.Fatalf("expected result error, got %v", err)
	}
}

func assertEvidence(t *testing.T, evidence map[string]interface{}, key string, want interface{}) {
	t.Helper()
	if got := evidence[key]; got != want {
		t.Fatalf("evidence[%s] = %#v, want %#v", key, got, want)
	}
}
