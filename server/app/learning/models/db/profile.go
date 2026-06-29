package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习画像(一个目标一份)
type LearningProfile struct {
	bun.BaseModel `bun:"table:learning_profiles,alias:lpr"`

	ID                 string    `bun:"id,pk,type:varchar(32)" json:"id"`
	UserID             string    `bun:"user_id,notnull" json:"user_id"`
	GoalID             string    `bun:"goal_id,notnull" json:"goal_id"`
	CurrentLevel       *string   `bun:"current_level" json:"current_level,omitempty"`
	CompletedTopics    []string  `bun:"completed_topics,type:jsonb" json:"completed_topics,omitempty"`
	WeakPoints         []string  `bun:"weak_points,type:jsonb" json:"weak_points,omitempty"`
	VerificationHabits *string   `bun:"verification_habits" json:"verification_habits,omitempty"`
	NextGoal           *string   `bun:"next_goal" json:"next_goal,omitempty"`
	CreatedAt          time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
