package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 每日学习日志
type LearningJournal struct {
	bun.BaseModel `bun:"table:learning_journals,alias:lj"`

	ID         string    `bun:"id,pk,type:varchar(32)" json:"id"`
	UserID     string    `bun:"user_id,notnull" json:"user_id"`
	GoalID     string    `bun:"goal_id,notnull" json:"goal_id"`
	Date       time.Time `bun:"date,notnull,type:date" json:"date"`
	Learned    *string   `bun:"learned" json:"learned,omitempty"`
	Evidence   *string   `bun:"evidence" json:"evidence,omitempty"`
	WeakPoints *string   `bun:"weak_points" json:"weak_points,omitempty"`
	NextStep   *string   `bun:"next_step" json:"next_step,omitempty"`
	CreatedAt  time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
