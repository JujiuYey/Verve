package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
