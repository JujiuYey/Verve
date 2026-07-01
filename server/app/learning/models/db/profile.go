package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习画像(一个目标一份)
type LearningProfile struct {
	bun.BaseModel `bun:"table:learning_profiles,alias:lpr"`

	ID                 string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	UserID             string    `bun:"user_id,notnull" json:"user_id"`                                          // 用户ID
	GoalID             string    `bun:"goal_id,notnull" json:"goal_id"`                                          // 目标ID
	CurrentLevel       *string   `bun:"current_level" json:"current_level,omitempty"`                            // 当前水平
	CompletedTopics    []string  `bun:"completed_topics,type:jsonb" json:"completed_topics,omitempty"`           // 已完成主题列表
	WeakPoints         []string  `bun:"weak_points,type:jsonb" json:"weak_points,omitempty"`                     // 薄弱点列表
	VerificationHabits *string   `bun:"verification_habits" json:"verification_habits,omitempty"`                // 验证习惯
	NextGoal           *string   `bun:"next_goal" json:"next_goal,omitempty"`                                    // 下一步目标
	CreatedAt          time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt          time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
