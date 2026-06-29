package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习目标
type LearningGoal struct {
	bun.BaseModel `bun:"table:learning_goals,alias:lg"`

	ID          string    `bun:"id,pk,type:varchar(32)" json:"id"`
	UserID      string    `bun:"user_id,notnull" json:"user_id"`
	Title       string    `bun:"title,notnull" json:"title"`                 // 用户的一句话目标
	Description *string   `bun:"description" json:"description,omitempty"`
	Source      string    `bun:"source,notnull" json:"source"`               // text / documents(MVP 固定 text)
	Status      string    `bun:"status,notnull" json:"status"`               // active / archived / completed
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
