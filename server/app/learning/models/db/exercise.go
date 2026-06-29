package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 练习与验证记录
type LearningExercise struct {
	bun.BaseModel `bun:"table:learning_exercises,alias:le"`

	ID           string    `bun:"id,pk,type:varchar(32)" json:"id"`
	SessionID    string    `bun:"session_id,notnull" json:"session_id"`
	ObjectiveID  string    `bun:"objective_id,notnull" json:"objective_id"`
	UserID       string    `bun:"user_id,notnull" json:"user_id"`
	Type         string    `bun:"type,notnull" json:"type"`     // explain / choice / cloze / paste_output / code_snippet
	Prompt       string    `bun:"prompt,notnull" json:"prompt"` // 题目 / 要求
	UserAnswer   *string   `bun:"user_answer" json:"user_answer,omitempty"`
	Verdict      *string   `bun:"verdict" json:"verdict,omitempty"`             // pass / partial / fail
	MasteryAfter *string   `bun:"mastery_after" json:"mastery_after,omitempty"` // 判定后掌握层级
	Feedback     *string   `bun:"feedback" json:"feedback,omitempty"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
