package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 解释审阅
type LearningExplanationReview struct {
	bun.BaseModel `bun:"table:learning_explanation_reviews,alias:ler"`

	ID                 string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 审阅ID
	SessionID          string    `bun:"session_id,notnull,type:varchar(32)" json:"session_id"`                   // 会话ID
	DocumentID         string    `bun:"document_id,notnull,type:varchar(32)" json:"document_id"`                 // Wiki文档ID
	UserID             string    `bun:"user_id,notnull,type:varchar(32)" json:"user_id"`                         // 用户ID
	Explanation        string    `bun:"explanation,notnull" json:"explanation"`                                  // 学习者解释
	HeardSummary       string    `bun:"heard_summary,notnull" json:"heard_summary"`                              // 倾听摘要
	ClearPoints        []string  `bun:"clear_points,type:jsonb,notnull" json:"clear_points"`                     // 已讲清内容
	ConfusingPoints    []string  `bun:"confusing_points,type:jsonb,notnull" json:"confusing_points"`             // 含混内容
	Misconceptions     []string  `bun:"misconceptions,type:jsonb,notnull" json:"misconceptions"`                 // 误解内容
	FollowUpQuestion   string    `bun:"follow_up_question,notnull" json:"follow_up_question"`                    // 后续追问
	ExplanationSummary string    `bun:"explanation_summary,notnull" json:"explanation_summary"`                  // 解释摘要
	ReadyToWrapUp      bool      `bun:"ready_to_wrap_up,notnull" json:"ready_to_wrap_up"`                        // 是否可以结束练习
	ContextSufficient  bool      `bun:"context_sufficient,notnull" json:"context_sufficient"`                    // 文档依据是否充分
	CreatedAt          time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
}
