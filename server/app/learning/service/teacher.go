package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/llm"
	"verve/infrastructure/llm/prompts"
)

var ErrIndexNotReady = errors.New("index_not_ready")

type TeachingResult struct {
	Response           string                                 `json:"response"`
	QuestionSummary    string                                 `json:"question_summary"`
	KnowledgeGaps      []string                               `json:"knowledge_gaps"`
	ExplanationSummary string                                 `json:"explanation_summary"`
	KeyPoints          []string                               `json:"key_points"`
	Examples           []string                               `json:"examples"`
	Evidence           []learning_db.LearningTeachingEvidence `json:"evidence"`
}

type TeachingRequest struct {
	DocumentID  string
	Question    string
	PriorTurns  []*learning_db.LearningExplanationReview
	MemoryItems []*learning_db.LearningMemoryItem
}

type agentTextRunner func(context.Context, string) (string, error)

type TeacherService struct {
	contextBuilder *FeynmanContextBuilder
	run            agentTextRunner
}

func NewTeacherService(source FeynmanDocumentSource) *TeacherService {
	return newTeacherService(source, runLearningTeacher)
}

func newTeacherService(source FeynmanDocumentSource, run agentTextRunner) *TeacherService {
	return &TeacherService{contextBuilder: NewFeynmanContextBuilder(source), run: run}
}

func (s *TeacherService) Teach(ctx context.Context, request TeachingRequest) (*TeachingResult, error) {
	question := strings.TrimSpace(request.Question)
	if question == "" {
		return nil, errors.New("question is required")
	}
	documentContext, err := s.contextBuilder.Build(ctx, request.DocumentID, question)
	if err != nil {
		return nil, err
	}
	if documentContext.Mode == FeynmanContextModeRAG && !documentContext.ContextSufficient {
		return nil, ErrIndexNotReady
	}
	evidence := make([]prompts.LearningTeacherEvidence, 0, len(documentContext.Evidence))
	for _, item := range documentContext.Evidence {
		evidence = append(evidence, prompts.LearningTeacherEvidence{
			ChunkID: item.ChunkID, DocumentVersion: item.DocumentVersion, ChunkIndex: item.ChunkIndex,
			HeadingPath: item.HeadingPath, Content: item.Content,
		})
	}
	query := prompts.LearningTeacherQueryPrompt(prompts.LearningTeacherQueryInput{
		DocumentTitle: documentContext.Title, DocumentVersion: documentContext.DocumentVersion,
		Mode: documentContext.Mode, FullText: documentContext.FullText, Evidence: evidence,
		PriorSummary: summarizeReviews(request.PriorTurns), MemorySummary: summarizeMemory(request.MemoryItems), Question: question,
	})
	text, err := s.run(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("run LearningTeacher: %w", err)
	}
	result, err := parseTeachingResult(text)
	if err != nil {
		return nil, err
	}
	for _, item := range result.Evidence {
		if item.DocumentVersion != documentContext.DocumentVersion {
			return nil, errors.New("LearningTeacher evidence does not match current document version")
		}
	}
	return result, nil
}

func runLearningTeacher(ctx context.Context, query string) (string, error) {
	agent, err := llm.NewLearningTeacherAgent(ctx)
	if err != nil {
		return "", err
	}
	return collectText(adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent}).Query(ctx, query))
}

func parseTeachingResult(text string) (*TeachingResult, error) {
	for _, candidate := range jsonObjectCandidates(strings.TrimSpace(text)) {
		var result TeachingResult
		if err := json.Unmarshal([]byte(candidate), &result); err != nil {
			continue
		}
		normalizeTeachingResult(&result)
		if result.Response == "" || result.QuestionSummary == "" || result.ExplanationSummary == "" {
			continue
		}
		return &result, nil
	}
	return nil, errors.New("LearningTeacher output is invalid")
}

func normalizeTeachingResult(result *TeachingResult) {
	result.Response = strings.TrimSpace(result.Response)
	result.QuestionSummary = strings.TrimSpace(result.QuestionSummary)
	result.ExplanationSummary = strings.TrimSpace(result.ExplanationSummary)
	if result.KnowledgeGaps == nil {
		result.KnowledgeGaps = []string{}
	}
	if result.KeyPoints == nil {
		result.KeyPoints = []string{}
	}
	if result.Examples == nil {
		result.Examples = []string{}
	}
	if result.Evidence == nil {
		result.Evidence = []learning_db.LearningTeachingEvidence{}
	}
}

func summarizeReviews(items []*learning_db.LearningExplanationReview) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && strings.TrimSpace(item.ExplanationSummary) != "" {
			parts = append(parts, strings.TrimSpace(item.ExplanationSummary))
		}
	}
	return truncateRunes(strings.Join(parts, "\n"), 4000)
}

func summarizeMemory(items []*learning_db.LearningMemoryItem) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && strings.TrimSpace(item.Statement) != "" {
			parts = append(parts, strings.TrimSpace(item.Statement))
		}
	}
	return truncateRunes(strings.Join(parts, "\n"), 4000)
}
