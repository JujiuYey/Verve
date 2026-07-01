package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 每日学习日志
type LearningJournal struct {
	bun.BaseModel `bun:"table:learning_journals,alias:lj"`

	ID         string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	UserID     string    `bun:"user_id,notnull" json:"user_id"`                                          // 用户ID
	GoalID     string    `bun:"goal_id,notnull" json:"goal_id"`                                          // 目标ID
	Date       time.Time `bun:"date,notnull,type:date" json:"date"`                                      // 日期
	Learned    *string   `bun:"learned" json:"learned,omitempty"`                                        // 今日所学
	Evidence   *string   `bun:"evidence" json:"evidence,omitempty"`                                      // 证据 / 产物
	WeakPoints *string   `bun:"weak_points" json:"weak_points,omitempty"`                                // 薄弱点
	NextStep   *string   `bun:"next_step" json:"next_step,omitempty"`                                    // 下一步计划
	CreatedAt  time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt  time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
