package db

import (
	"time"

	"github.com/uptrace/bun"
)

// LearningExplanationReview records one structured review turn in a document-bound session.
type LearningExplanationReview struct {
	bun.BaseModel `bun:"table:learning_explanation_reviews,alias:ler"`

	ID                 string    `bun:"id,pk,type:varchar(32)" json:"id"`
	SessionID          string    `bun:"session_id,notnull,type:varchar(32)" json:"session_id"`
	DocumentID         string    `bun:"document_id,notnull,type:varchar(32)" json:"document_id"`
	UserID             string    `bun:"user_id,notnull,type:varchar(32)" json:"user_id"`
	Explanation        string    `bun:"explanation,notnull" json:"explanation"`
	HeardSummary       string    `bun:"heard_summary,notnull" json:"heard_summary"`
	ClearPoints        []string  `bun:"clear_points,type:jsonb,notnull" json:"clear_points"`
	ConfusingPoints    []string  `bun:"confusing_points,type:jsonb,notnull" json:"confusing_points"`
	Misconceptions     []string  `bun:"misconceptions,type:jsonb,notnull" json:"misconceptions"`
	FollowUpQuestion   string    `bun:"follow_up_question,notnull" json:"follow_up_question"`
	ExplanationSummary string    `bun:"explanation_summary,notnull" json:"explanation_summary"`
	ReadyToWrapUp      bool      `bun:"ready_to_wrap_up,notnull" json:"ready_to_wrap_up"`
	ContextSufficient  bool      `bun:"context_sufficient,notnull" json:"context_sufficient"`
	CreatedAt          time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}
