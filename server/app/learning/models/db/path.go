package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习路线(一个目标一条)
type LearningPath struct {
	bun.BaseModel `bun:"table:learning_paths,alias:lp"`

	ID                 string                   `bun:"id,pk,type:varchar(32)" json:"id"`
	GoalID             string                   `bun:"goal_id,notnull" json:"goal_id"`
	UserID             string                   `bun:"user_id,notnull" json:"user_id"`
	Overview           []map[string]interface{} `bun:"overview,type:jsonb" json:"overview,omitempty"` // 阶段大纲概要
	CurrentObjectiveID *string                  `bun:"current_objective_id" json:"current_objective_id,omitempty"`
	Status             string                   `bun:"status,notnull" json:"status"` // active / completed
	CreatedAt          time.Time                `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time                `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
