package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习会话(一节课)
type LearningSession struct {
	bun.BaseModel `bun:"table:learning_sessions,alias:ls"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`
	UserID       string     `bun:"user_id,notnull" json:"user_id"`
	GoalID       string     `bun:"goal_id,notnull" json:"goal_id"`
	ObjectiveID  string     `bun:"objective_id,notnull" json:"objective_id"`
	Status       string     `bun:"status,notnull" json:"status"` // active / completed / abandoned
	Summary      *string    `bun:"summary" json:"summary,omitempty"`
	MessageCount int        `bun:"message_count" json:"message_count"`
	StartedAt    time.Time  `bun:"started_at,nullzero,notnull,default:current_timestamp" json:"started_at"`
	EndedAt      *time.Time `bun:"ended_at" json:"ended_at,omitempty"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
