package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习路线(一个目标一条)
type LearningPath struct {
	bun.BaseModel `bun:"table:learning_paths,alias:lp"`

	ID                 string                   `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	GoalID             string                   `bun:"goal_id,notnull" json:"goal_id"`                                          // 目标ID
	UserID             string                   `bun:"user_id,notnull" json:"user_id"`                                          // 用户ID
	Overview           []map[string]interface{} `bun:"overview,type:jsonb" json:"overview,omitempty"`                           // 阶段大纲概要
	CurrentObjectiveID *string                  `bun:"current_objective_id" json:"current_objective_id,omitempty"`              // 当前进行中的小目标ID
	Status             string                   `bun:"status,notnull" json:"status"`                                            // active / completed
	CreatedAt          time.Time                `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt          time.Time                `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
