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

type FeynmanReview struct {
	HeardSummary       string   `json:"heard_summary"`
	ClearPoints        []string `json:"clear_points"`
	ConfusingPoints    []string `json:"confusing_points"`
	Misconceptions     []string `json:"misconceptions"`
	FollowUpQuestion   string   `json:"follow_up_question"`
	ExplanationSummary string   `json:"explanation_summary"`
	ReadyToWrapUp      bool     `json:"ready_to_wrap_up"`
	ContextSufficient  bool     `json:"context_sufficient"`
}

type FeynmanReviewer interface {
	Review(ctx context.Context, documentID, explanation string, prior []*learning_db.LearningExplanationReview) (*FeynmanReview, error)
}

type FeynmanReviewService struct {
	contextBuilder *FeynmanContextBuilder
}

func NewFeynmanReviewService(source FeynmanDocumentSource) *FeynmanReviewService {
	return &FeynmanReviewService{contextBuilder: NewFeynmanContextBuilder(source)}
}

func (s *FeynmanReviewService) Review(ctx context.Context, documentID, explanation string, prior []*learning_db.LearningExplanationReview) (*FeynmanReview, error) {
	if s == nil || s.contextBuilder == nil {
		return nil, errors.New("Feynman reviewer is not configured")
	}
	explanation = strings.TrimSpace(explanation)
	if explanation == "" {
		return nil, errors.New("explanation is required")
	}
	documentContext, err := s.contextBuilder.Build(ctx, documentID, explanation)
	if err != nil {
		return nil, err
	}

	agent, err := llm.NewFeynmanReviewerAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Feynman reviewer: %w", err)
	}
	query := prompts.FeynmanReviewerQueryPrompt(feynmanReviewerQueryInput(documentContext, prior, explanation))
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	text, err := collectText(runner.Query(ctx, query))
	if err != nil {
		return nil, fmt.Errorf("run Feynman reviewer: %w", err)
	}
	return parseFeynmanReviewOutput(text, documentContext.ContextSufficient)
}

func feynmanReviewerQueryInput(documentContext *FeynmanDocumentContext, prior []*learning_db.LearningExplanationReview, explanation string) prompts.FeynmanReviewerQueryInput {
	evidence := make([]prompts.FeynmanReviewerEvidence, 0, len(documentContext.Evidence))
	for _, item := range documentContext.Evidence {
		evidence = append(evidence, prompts.FeynmanReviewerEvidence{
			ChunkIndex: item.ChunkIndex, HeadingPath: item.HeadingPath, Content: item.Content,
		})
	}
	turns := make([]prompts.FeynmanReviewerTurn, 0, len(prior))
	for _, item := range prior {
		if item == nil {
			continue
		}
		reviewJSON, _ := json.Marshal(FeynmanReview{
			HeardSummary: item.HeardSummary, ClearPoints: item.ClearPoints,
			ConfusingPoints: item.ConfusingPoints, Misconceptions: item.Misconceptions,
			FollowUpQuestion: item.FollowUpQuestion, ExplanationSummary: item.ExplanationSummary,
			ReadyToWrapUp: item.ReadyToWrapUp, ContextSufficient: item.ContextSufficient,
		})
		turns = append(turns, prompts.FeynmanReviewerTurn{Explanation: item.Explanation, Review: string(reviewJSON)})
	}
	return prompts.FeynmanReviewerQueryInput{
		DocumentTitle: documentContext.Title, Outline: documentContext.Outline,
		Mode: documentContext.Mode, FullText: documentContext.FullText, Evidence: evidence,
		ContextSufficient:          documentContext.ContextSufficient,
		ContextInsufficiencyReason: documentContext.ContextInsufficiencyReason,
		PriorTurns:                 turns, NewExplanation: explanation,
	}
}

var feynmanReviewRequiredKeys = []string{
	"heard_summary",
	"clear_points",
	"confusing_points",
	"misconceptions",
	"follow_up_question",
	"explanation_summary",
	"ready_to_wrap_up",
	"context_sufficient",
}

func parseFeynmanReviewOutput(text string, contextSufficient bool) (*FeynmanReview, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("Feynman reviewer output is empty")
	}

	var lastErr error
	for _, candidate := range jsonObjectCandidates(text) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(candidate), &fields); err != nil {
			lastErr = err
			continue
		}
		var fieldErr error
		for _, key := range feynmanReviewRequiredKeys {
			value, present := fields[key]
			if !present {
				fieldErr = fmt.Errorf("FeynmanReview is missing required key %q", key)
				break
			}
			if string(value) == "null" && key != "clear_points" && key != "confusing_points" && key != "misconceptions" {
				fieldErr = fmt.Errorf("FeynmanReview key %q cannot be null", key)
				break
			}
		}
		if fieldErr != nil {
			lastErr = fieldErr
			continue
		}

		var review FeynmanReview
		if err := json.Unmarshal([]byte(candidate), &review); err != nil {
			lastErr = err
			continue
		}
		normalizeFeynmanReview(&review)
		if review.HeardSummary == "" || review.ExplanationSummary == "" {
			lastErr = errors.New("JSON object is not a complete FeynmanReview")
			continue
		}
		if !review.ReadyToWrapUp && review.FollowUpQuestion == "" {
			lastErr = errors.New("FeynmanReview requires one follow_up_question while ready_to_wrap_up is false")
			continue
		}
		if review.ReadyToWrapUp && review.FollowUpQuestion != "" {
			lastErr = errors.New("FeynmanReview follow_up_question must be blank while ready_to_wrap_up is true")
			continue
		}
		review.ContextSufficient = contextSufficient
		return &review, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no JSON object found in Feynman reviewer output")
}

func normalizeFeynmanReview(review *FeynmanReview) {
	review.HeardSummary = strings.TrimSpace(review.HeardSummary)
	review.FollowUpQuestion = strings.TrimSpace(review.FollowUpQuestion)
	review.ExplanationSummary = strings.TrimSpace(review.ExplanationSummary)
	review.ClearPoints = uniqueNonEmptyStrings(review.ClearPoints)
	review.ConfusingPoints = uniqueNonEmptyStrings(review.ConfusingPoints)
	review.Misconceptions = uniqueNonEmptyStrings(review.Misconceptions)
}

func uniqueNonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
