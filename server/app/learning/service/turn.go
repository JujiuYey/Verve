package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	learning_repo "verve/app/learning/repository"
	wiki_db "verve/app/wiki/models/db"
)

var (
	ErrTurnProcessing     = errors.New("learning turn is processing")
	ErrTurnSessionCompleted = errors.New("learning session is completed")
	ErrUnsupportedAgent   = errors.New("unsupported learning agent")
	ErrInvalidTurnRequest = errors.New("invalid learning turn request")
)

type TurnRequest struct {
	RequestID               string  `json:"request_id"`
	AgentType               string  `json:"agent_type"`
	Content                 string  `json:"content"`
	ReplacesChangeRequestID *string `json:"replaces_change_request_id,omitempty"`
}

type AgentInput struct {
	Session                 *learning_db.LearningSession
	Turn                    *learning_db.LearningTurn
	Content                 string
	RequestID               string
	ReplacesChangeRequestID *string
}

type AgentOutput struct {
	AssistantContent string
	Review           *learning_db.LearningExplanationReview
	Intervention     *learning_db.LearningTeachingIntervention
	ChangeRequest    *wiki_db.DocumentChangeRequest
}

type AgentProcessor interface {
	Process(context.Context, AgentInput) (*AgentOutput, error)
}

type AgentProcessorFunc func(context.Context, AgentInput) (*AgentOutput, error)

func (f AgentProcessorFunc) Process(ctx context.Context, input AgentInput) (*AgentOutput, error) {
	return f(ctx, input)
}

type turnSessionReader interface {
	FindOne(context.Context, string) (*learning_db.LearningSession, error)
}

type turnLifecycleStore interface {
	BeginTurn(context.Context, learning_repo.BeginTurnInput) (*learning_repo.BeginTurnResult, error)
	RetryFailedTurn(context.Context, string) error
	FailTurn(context.Context, string, string, string) error
	CompleteListenerTurn(context.Context, string, string, string, *learning_db.LearningExplanationReview) error
	CompleteTeacherTurn(context.Context, string, string, string, *learning_db.LearningTeachingIntervention) error
	CompleteTurn(context.Context, string, string, string) error
}

type turnTimelineReader interface {
	FindByTurn(context.Context, string) (*learning_payload.TimelineItem, error)
}

type turnMemoryRecorder interface {
	RecordExplanationReview(context.Context, *learning_db.LearningSession, *learning_db.LearningExplanationReview) error
	RecordTeachingIntervention(context.Context, *learning_db.LearningSession, *learning_db.LearningTeachingIntervention) error
}

type TurnService struct {
	sessions   turnSessionReader
	turns      turnLifecycleStore
	timeline   turnTimelineReader
	processors map[string]AgentProcessor
	memory     turnMemoryRecorder
}

func NewTurnService(sessions turnSessionReader, turns turnLifecycleStore, timeline turnTimelineReader, processors map[string]AgentProcessor, memory turnMemoryRecorder) *TurnService {
	return &TurnService{sessions: sessions, turns: turns, timeline: timeline, processors: processors, memory: memory}
}

func (s *TurnService) processorFor(agentType string) (AgentProcessor, error) {
	agentType = strings.TrimSpace(agentType)
	processor := s.processors[agentType]
	if processor == nil {
		return nil, fmt.Errorf("%w %q", ErrUnsupportedAgent, agentType)
	}
	return processor, nil
}

func (s *TurnService) Submit(ctx context.Context, sessionID string, request TurnRequest) (*learning_payload.TimelineItem, error) {
	request.RequestID = strings.TrimSpace(request.RequestID)
	request.AgentType = strings.TrimSpace(request.AgentType)
	request.Content = strings.TrimSpace(request.Content)
	if request.RequestID == "" || request.Content == "" {
		return nil, fmt.Errorf("%w: request_id and content are required", ErrInvalidTurnRequest)
	}
	processor, err := s.processorFor(request.AgentType)
	if err != nil {
		return nil, err
	}
	session, err := s.sessions.FindOne(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != "active" {
		return nil, ErrTurnSessionCompleted
	}
	started, err := s.turns.BeginTurn(ctx, learning_repo.BeginTurnInput{
		SessionID: sessionID, RequestID: request.RequestID, AgentType: request.AgentType, Content: request.Content,
	})
	if err != nil {
		return nil, err
	}
	if started == nil || started.Turn == nil {
		return nil, errors.New("turn repository returned no turn")
	}
	turn := started.Turn
	if !started.Created {
		switch turn.Status {
		case learning_db.LearningTurnCompleted:
			return s.timeline.FindByTurn(ctx, turn.ID)
		case learning_db.LearningTurnProcessing:
			return nil, ErrTurnProcessing
		case learning_db.LearningTurnFailed:
			if err := s.turns.RetryFailedTurn(ctx, turn.ID); err != nil {
				return nil, err
			}
			turn.Status = learning_db.LearningTurnProcessing
		default:
			return nil, fmt.Errorf("invalid turn status %q", turn.Status)
		}
	}
	output, err := processor.Process(ctx, AgentInput{
		Session: session, Turn: turn, Content: request.Content, RequestID: request.RequestID,
		ReplacesChangeRequestID: request.ReplacesChangeRequestID,
	})
	if err != nil {
		s.fail(ctx, turn.ID, request.AgentType+"_failed", err)
		return nil, err
	}
	if output == nil || strings.TrimSpace(output.AssistantContent) == "" {
		err = errors.New("agent returned no assistant content")
		s.fail(ctx, turn.ID, request.AgentType+"_empty_result", err)
		return nil, err
	}
	switch request.AgentType {
	case learning_db.LearningAgentListener:
		if output.Review == nil {
			err = errors.New("listener returned no review")
			s.fail(ctx, turn.ID, "listener_empty_artifact", err)
			return nil, err
		}
		err = s.turns.CompleteListenerTurn(ctx, sessionID, turn.ID, output.AssistantContent, output.Review)
	case learning_db.LearningAgentTeacher:
		if output.Intervention == nil {
			err = errors.New("teacher returned no intervention")
			s.fail(ctx, turn.ID, "teacher_empty_artifact", err)
			return nil, err
		}
		err = s.turns.CompleteTeacherTurn(ctx, sessionID, turn.ID, output.AssistantContent, output.Intervention)
	case learning_db.LearningAgentCurator:
		if output.ChangeRequest == nil {
			err = errors.New("curator returned no change request")
			s.fail(ctx, turn.ID, "curator_empty_artifact", err)
			return nil, err
		}
		err = s.turns.CompleteTurn(ctx, sessionID, turn.ID, output.AssistantContent)
	}
	if err != nil {
		s.fail(ctx, turn.ID, request.AgentType+"_persistence_failed", err)
		return nil, err
	}
	s.recordMemory(ctx, session, output)
	return s.timeline.FindByTurn(ctx, turn.ID)
}

func (s *TurnService) fail(ctx context.Context, turnID, code string, cause error) {
	if err := s.turns.FailTurn(ctx, turnID, code, cause.Error()); err != nil && !errors.Is(err, learning_repo.ErrTurnNotProcessing) {
		log.Printf("mark learning turn failed: turn_id=%s err=%v", turnID, err)
	}
}

func (s *TurnService) recordMemory(ctx context.Context, session *learning_db.LearningSession, output *AgentOutput) {
	if s.memory == nil {
		return
	}
	var err error
	if output.Review != nil {
		err = s.memory.RecordExplanationReview(ctx, session, output.Review)
	} else if output.Intervention != nil {
		err = s.memory.RecordTeachingIntervention(ctx, session, output.Intervention)
	}
	if err != nil {
		log.Printf("record learning turn memory failed: session_id=%s err=%v", session.ID, err)
	}
}
