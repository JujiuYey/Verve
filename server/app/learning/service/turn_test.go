package service

import (
	"context"
	"errors"
	"testing"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	learning_repo "verve/app/learning/repository"
)

type recordingProcessor struct {
	calls int
	input AgentInput
}

func (p *recordingProcessor) Process(_ context.Context, input AgentInput) (*AgentOutput, error) {
	p.calls++
	p.input = input
	return &AgentOutput{AssistantContent: "回答"}, nil
}

func TestTurnServiceDispatchesOnlyExplicitSelectedAgent(t *testing.T) {
	listener := &recordingProcessor{}
	teacher := &recordingProcessor{}
	curator := &recordingProcessor{}
	service := &TurnService{processors: map[string]AgentProcessor{
		learning_db.LearningAgentListener: listener,
		learning_db.LearningAgentTeacher:  teacher,
		learning_db.LearningAgentCurator:  curator,
	}}

	processor, err := service.processorFor(learning_db.LearningAgentTeacher)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := processor.Process(context.Background(), AgentInput{Content: "解释 channel"}); err != nil {
		t.Fatal(err)
	}
	if teacher.calls != 1 || listener.calls != 0 || curator.calls != 0 {
		t.Fatalf("calls = listener:%d teacher:%d curator:%d", listener.calls, teacher.calls, curator.calls)
	}
	if _, err := service.processorFor("supervisor"); err == nil {
		t.Fatal("unknown or supervisor agent must be rejected")
	}
}

type turnSessionStoreFake struct{ session *learning_db.LearningSession }

func (f turnSessionStoreFake) FindOne(context.Context, string) (*learning_db.LearningSession, error) {
	return f.session, nil
}

type turnLifecycleStoreFake struct {
	begin        *learning_repo.BeginTurnResult
	failed       bool
	retried      bool
	completed    string
	intervention *learning_db.LearningTeachingIntervention
}

func (f *turnLifecycleStoreFake) BeginTurn(context.Context, learning_repo.BeginTurnInput) (*learning_repo.BeginTurnResult, error) {
	return f.begin, nil
}
func (f *turnLifecycleStoreFake) RetryFailedTurn(context.Context, string) error {
	f.retried = true
	return nil
}
func (f *turnLifecycleStoreFake) FailTurn(context.Context, string, string, string) error {
	f.failed = true
	return nil
}
func (f *turnLifecycleStoreFake) CompleteListenerTurn(context.Context, string, string, string, *learning_db.LearningExplanationReview) error {
	f.completed = learning_db.LearningAgentListener
	return nil
}
func (f *turnLifecycleStoreFake) CompleteTeacherTurn(_ context.Context, _, _, _ string, intervention *learning_db.LearningTeachingIntervention) error {
	f.completed = learning_db.LearningAgentTeacher
	f.intervention = intervention
	return nil
}
func (f *turnLifecycleStoreFake) CompleteTurn(context.Context, string, string, string) error {
	f.completed = learning_db.LearningAgentCurator
	return nil
}

type turnTimelineFake struct {
	item *learning_payload.TimelineItem
}

func (f turnTimelineFake) FindByTurn(context.Context, string) (*learning_payload.TimelineItem, error) {
	return f.item, nil
}

func TestTurnServiceSubmitsTeacherAndReturnsPersistedTimelineItem(t *testing.T) {
	turn := &learning_db.LearningTurn{ID: "turn-1", SessionID: "session-1", AgentType: learning_db.LearningAgentTeacher, Status: learning_db.LearningTurnProcessing}
	lifecycle := &turnLifecycleStoreFake{begin: &learning_repo.BeginTurnResult{Turn: turn, Created: true}}
	processor := &recordingProcessor{}
	processorOutput := &learning_db.LearningTeachingIntervention{QuestionSummary: "问题", ExplanationSummary: "讲解"}
	service := NewTurnService(
		turnSessionStoreFake{session: &learning_db.LearningSession{ID: "session-1", UserID: "user-1", DocumentID: "doc-1", Status: "active"}},
		lifecycle,
		turnTimelineFake{item: &learning_payload.TimelineItem{Turn: &learning_db.LearningTurn{ID: "turn-1", Status: learning_db.LearningTurnCompleted}}},
		map[string]AgentProcessor{learning_db.LearningAgentTeacher: AgentProcessorFunc(func(_ context.Context, input AgentInput) (*AgentOutput, error) {
			processor.calls++
			processor.input = input
			return &AgentOutput{AssistantContent: "回答", Intervention: processorOutput}, nil
		})},
		nil,
	)

	item, err := service.Submit(context.Background(), "user-1", "session-1", TurnRequest{RequestID: "request-1", AgentType: "teacher", Content: "解释 channel"})
	if err != nil {
		t.Fatal(err)
	}
	if item.Turn.Status != learning_db.LearningTurnCompleted || lifecycle.completed != learning_db.LearningAgentTeacher || lifecycle.intervention != processorOutput {
		t.Fatalf("item/lifecycle = %#v/%#v", item, lifecycle)
	}
	if processor.calls != 1 || processor.input.Session.DocumentID != "doc-1" {
		t.Fatalf("processor = %#v", processor)
	}
}

func TestTurnServiceMarksFailedProcessorAndRetriesFailedTurn(t *testing.T) {
	turn := &learning_db.LearningTurn{ID: "turn-1", SessionID: "session-1", AgentType: learning_db.LearningAgentTeacher, Status: learning_db.LearningTurnFailed}
	lifecycle := &turnLifecycleStoreFake{begin: &learning_repo.BeginTurnResult{Turn: turn}}
	service := NewTurnService(
		turnSessionStoreFake{session: &learning_db.LearningSession{ID: "session-1", UserID: "user-1", Status: "active"}},
		lifecycle, turnTimelineFake{},
		map[string]AgentProcessor{learning_db.LearningAgentTeacher: AgentProcessorFunc(func(context.Context, AgentInput) (*AgentOutput, error) {
			return nil, errors.New("model unavailable")
		})}, nil,
	)

	_, err := service.Submit(context.Background(), "user-1", "session-1", TurnRequest{RequestID: "request-1", AgentType: "teacher", Content: "解释"})
	if err == nil || !lifecycle.retried || !lifecycle.failed {
		t.Fatalf("error/retried/failed = %v/%v/%v", err, lifecycle.retried, lifecycle.failed)
	}
}
