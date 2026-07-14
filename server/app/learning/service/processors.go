package service

import (
	"context"
	"encoding/json"
	"errors"

	learning_db "verve/app/learning/models/db"
)

type processorReviewReader interface {
	FindBySession(context.Context, string) ([]*learning_db.LearningExplanationReview, error)
}

type processorMemoryReader interface {
	FindDocumentItems(context.Context, string, int) ([]*learning_db.LearningMemoryItem, error)
}

type ListenerProcessor struct {
	reviewer FeynmanReviewer
	reviews  processorReviewReader
	memory   processorMemoryReader
}

func NewListenerProcessor(reviewer FeynmanReviewer, reviews processorReviewReader, memory processorMemoryReader) *ListenerProcessor {
	return &ListenerProcessor{reviewer: reviewer, reviews: reviews, memory: memory}
}

func (p *ListenerProcessor) Process(ctx context.Context, input AgentInput) (*AgentOutput, error) {
	prior, err := p.reviews.FindBySession(ctx, input.Session.ID)
	if err != nil {
		return nil, err
	}
	memory := make([]*learning_db.LearningMemoryItem, 0)
	if p.memory != nil {
		if loaded, loadErr := p.memory.FindDocumentItems(ctx, input.Session.DocumentID, 20); loadErr == nil && loaded != nil {
			memory = loaded
		}
	}
	result, err := p.reviewer.Review(ctx, FeynmanReviewRequest{
		DocumentID: input.Session.DocumentID, Explanation: input.Content,
		PriorTurns: prior, MemoryItems: memory,
	})
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &AgentOutput{AssistantContent: string(data), Review: &learning_db.LearningExplanationReview{
		HeardSummary: result.HeardSummary, ClearPoints: result.ClearPoints, ConfusingPoints: result.ConfusingPoints,
		Misconceptions: result.Misconceptions, FollowUpQuestion: result.FollowUpQuestion,
		ExplanationSummary: result.ExplanationSummary, ReadyToWrapUp: result.ReadyToWrapUp, ContextSufficient: result.ContextSufficient,
	}}, nil
}

type TeacherProcessor struct {
	teacher *TeacherService
	reviews processorReviewReader
	memory  processorMemoryReader
}

func NewTeacherProcessor(teacher *TeacherService, reviews processorReviewReader, memory processorMemoryReader) *TeacherProcessor {
	return &TeacherProcessor{teacher: teacher, reviews: reviews, memory: memory}
}

func (p *TeacherProcessor) Process(ctx context.Context, input AgentInput) (*AgentOutput, error) {
	prior, err := p.reviews.FindBySession(ctx, input.Session.ID)
	if err != nil {
		return nil, err
	}
	memory := make([]*learning_db.LearningMemoryItem, 0)
	if p.memory != nil {
		if loaded, loadErr := p.memory.FindDocumentItems(ctx, input.Session.DocumentID, 20); loadErr == nil && loaded != nil {
			memory = loaded
		}
	}
	result, err := p.teacher.Teach(ctx, TeachingRequest{
		DocumentID: input.Session.DocumentID, Question: input.Content,
		PriorTurns: prior, MemoryItems: memory,
	})
	if err != nil {
		return nil, err
	}
	return &AgentOutput{AssistantContent: result.Response, Intervention: &learning_db.LearningTeachingIntervention{
		QuestionSummary: result.QuestionSummary, KnowledgeGaps: result.KnowledgeGaps,
		ExplanationSummary: result.ExplanationSummary, KeyPoints: result.KeyPoints, Examples: result.Examples, Evidence: result.Evidence,
	}}, nil
}

type CuratorProcessor struct{ curator *CuratorService }

func NewCuratorProcessor(curator *CuratorService) *CuratorProcessor {
	return &CuratorProcessor{curator: curator}
}

func (p *CuratorProcessor) Process(ctx context.Context, input AgentInput) (*AgentOutput, error) {
	if p.curator == nil {
		return nil, errors.New("curator is not configured")
	}
	request, err := p.curator.Propose(ctx, CuratorRequest{
		DocumentID: input.Session.DocumentID, TurnID: input.Turn.ID,
		RequestID: input.RequestID, Instruction: input.Content, ReplacesChangeRequestID: input.ReplacesChangeRequestID,
	})
	if err != nil {
		return nil, err
	}
	return &AgentOutput{AssistantContent: request.ChangeSummary, ChangeRequest: request}, nil
}
