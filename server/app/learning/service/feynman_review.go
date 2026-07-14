package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/eino/adk"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/llm"
	"verve/infrastructure/llm/prompts"
)

const (
	FeynmanPriorHistoryCharacterBudget = 12000
	FeynmanRecentTurnLimit             = 4
	feynmanPriorSummaryCharacterBudget = 4000
	feynmanRecentTurnCharacterBudget   = (FeynmanPriorHistoryCharacterBudget - feynmanPriorSummaryCharacterBudget) / FeynmanRecentTurnLimit
	feynmanRecentExplanationBudget     = 1250
	feynmanRecentReviewBudget          = feynmanRecentTurnCharacterBudget - feynmanRecentExplanationBudget
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
	Review(ctx context.Context, request FeynmanReviewRequest) (*FeynmanReview, error)
}

type FeynmanReviewRequest struct {
	DocumentID  string
	Explanation string
	PriorTurns  []*learning_db.LearningExplanationReview
	MemoryItems []*learning_db.LearningMemoryItem
}

type FeynmanReviewService struct {
	contextBuilder *FeynmanContextBuilder
}

func NewFeynmanReviewService(source FeynmanDocumentSource) *FeynmanReviewService {
	return &FeynmanReviewService{contextBuilder: NewFeynmanContextBuilder(source)}
}

func (s *FeynmanReviewService) Review(ctx context.Context, request FeynmanReviewRequest) (*FeynmanReview, error) {
	if s == nil || s.contextBuilder == nil {
		return nil, errors.New("Feynman reviewer is not configured")
	}
	explanation := strings.TrimSpace(request.Explanation)
	if explanation == "" {
		return nil, errors.New("explanation is required")
	}
	documentContext, err := s.contextBuilder.Build(ctx, request.DocumentID, explanation)
	if err != nil {
		return nil, err
	}

	agent, err := llm.NewFeynmanReviewerAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Feynman reviewer: %w", err)
	}
	request.Explanation = explanation
	query := prompts.FeynmanReviewerQueryPrompt(feynmanReviewerQueryInput(documentContext, request))
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	text, err := collectText(runner.Query(ctx, query))
	if err != nil {
		return nil, fmt.Errorf("run Feynman reviewer: %w", err)
	}
	return parseFeynmanReviewOutput(text, documentContext.ContextSufficient)
}

func feynmanReviewerQueryInput(documentContext *FeynmanDocumentContext, request FeynmanReviewRequest) prompts.FeynmanReviewerQueryInput {
	evidence := make([]prompts.FeynmanReviewerEvidence, 0, len(documentContext.Evidence))
	for _, item := range documentContext.Evidence {
		evidence = append(evidence, prompts.FeynmanReviewerEvidence{
			ChunkIndex: item.ChunkIndex, HeadingPath: item.HeadingPath, Content: item.Content,
		})
	}
	prior := nonNilReviews(request.PriorTurns)
	recentStart := len(prior) - FeynmanRecentTurnLimit
	if recentStart < 0 {
		recentStart = 0
	}
	turns := make([]prompts.FeynmanReviewerTurn, 0, len(prior)-recentStart)
	for _, item := range prior[recentStart:] {
		// ExplanationReview no longer carries the original learner explanation
		// (the read-only column was dropped along with user_id). The summary fields
		// below still carry enough context for the LLM to reason about prior turns.
		turns = append(turns, prompts.FeynmanReviewerTurn{
			Explanation: "",
			Review:      compactPriorReview(item, feynmanRecentReviewBudget),
		})
	}
	memoryItems := make([]prompts.FeynmanReviewerMemoryItem, 0, len(request.MemoryItems))
	for _, item := range request.MemoryItems {
		if item == nil || strings.TrimSpace(item.Statement) == "" {
			continue
		}
		memoryItems = append(memoryItems, prompts.FeynmanReviewerMemoryItem{
			Kind: strings.TrimSpace(item.Kind), Statement: strings.TrimSpace(item.Statement), Confidence: strings.TrimSpace(item.Confidence),
		})
	}
	return prompts.FeynmanReviewerQueryInput{
		DocumentTitle: documentContext.Title, Outline: documentContext.Outline,
		Mode: documentContext.Mode, FullText: documentContext.FullText, Evidence: evidence,
		ContextSufficient:          documentContext.ContextSufficient,
		ContextInsufficiencyReason: documentContext.ContextInsufficiencyReason,
		PriorSummary:               compactOlderReviewSummaries(prior[:recentStart], feynmanPriorSummaryCharacterBudget),
		PriorTurns:                 turns, MemoryItems: memoryItems, NewExplanation: request.Explanation,
	}
}

func nonNilReviews(items []*learning_db.LearningExplanationReview) []*learning_db.LearningExplanationReview {
	result := make([]*learning_db.LearningExplanationReview, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, item)
		}
	}
	return result
}

func compactPriorReview(item *learning_db.LearningExplanationReview, budget int) string {
	compact := struct {
		ExplanationSummary string `json:"explanation_summary"`
		FollowUpQuestion   string `json:"follow_up_question,omitempty"`
		Misconceptions     string `json:"misconceptions,omitempty"`
	}{
		ExplanationSummary: truncateRunes(strings.TrimSpace(item.ExplanationSummary), 250),
		FollowUpQuestion:   truncateRunes(strings.TrimSpace(item.FollowUpQuestion), 180),
		Misconceptions:     truncateRunes(strings.Join(uniqueNonEmptyStrings(item.Misconceptions), "; "), 180),
	}
	data, _ := json.Marshal(compact)
	if utf8.RuneCount(data) <= budget {
		return string(data)
	}
	compact.FollowUpQuestion = ""
	compact.Misconceptions = ""
	compact.ExplanationSummary = truncateRunes(compact.ExplanationSummary, budget/2)
	data, _ = json.Marshal(compact)
	if utf8.RuneCount(data) <= budget {
		return string(data)
	}
	return "{}"
}

func compactOlderReviewSummaries(items []*learning_db.LearningExplanationReview, budget int) string {
	lines := make([]string, 0, len(items))
	used := 0
	for i := len(items) - 1; i >= 0 && used < budget; i-- {
		summary := truncateRunes(strings.TrimSpace(items[i].ExplanationSummary), 500)
		if summary == "" {
			continue
		}
		line := fmt.Sprintf("turn %d: %s", i+1, summary)
		remaining := budget - used
		line = truncateRunes(line, remaining)
		lines = append(lines, line)
		used += utf8.RuneCountInString(line)
		if used < budget {
			used++ // newline
		}
	}
	for left, right := 0, len(lines)-1; left < right; left, right = left+1, right-1 {
		lines[left], lines[right] = lines[right], lines[left]
	}
	return truncateRunes(strings.Join(lines, "\n"), budget)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
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
