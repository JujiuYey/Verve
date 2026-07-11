package db

import (
	"reflect"
	"strings"
	"testing"
)

func TestLearningTurnModelCarriesAgentRequestLifecycle(t *testing.T) {
	modelType := reflect.TypeOf(LearningTurn{})
	wantFields := map[string]string{
		"SessionID": "session_id",
		"RequestID": "request_id",
		"AgentType": "agent_type",
		"Status":    "status",
		"ErrorCode": "error_code",
	}
	for name, column := range wantFields {
		field, ok := modelType.FieldByName(name)
		if !ok {
			t.Fatalf("LearningTurn is missing field %s", name)
		}
		if !strings.Contains(field.Tag.Get("bun"), column) {
			t.Fatalf("LearningTurn.%s bun tag = %q, want column %q", name, field.Tag.Get("bun"), column)
		}
	}
	if LearningAgentListener != "listener" || LearningAgentTeacher != "teacher" || LearningAgentCurator != "curator" {
		t.Fatalf("agent constants = %q, %q, %q", LearningAgentListener, LearningAgentTeacher, LearningAgentCurator)
	}
}

func TestLearningTeachingInterventionBelongsToOneTurn(t *testing.T) {
	modelType := reflect.TypeOf(LearningTeachingIntervention{})
	for _, name := range []string{"TurnID", "QuestionSummary", "KnowledgeGaps", "ExplanationSummary", "KeyPoints", "Examples", "Evidence"} {
		if _, ok := modelType.FieldByName(name); !ok {
			t.Fatalf("LearningTeachingIntervention is missing field %s", name)
		}
	}
}
