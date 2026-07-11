package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习会话(一节课)
type LearningSession struct {
	bun.BaseModel `bun:"table:learning_sessions,alias:ls"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	UserID       string     `bun:"user_id,notnull" json:"user_id"`                                          // 用户ID
	DocumentID   string     `bun:"document_id,notnull" json:"document_id"`                                  // Wiki 文档ID
	ObjectiveID  string     `bun:"-" json:"objective_id,omitempty"`                                         // 旧流程临时编译字段
	Status       string     `bun:"status,notnull" json:"status"`                                            // active / completed / abandoned
	Summary      *string    `bun:"summary" json:"summary,omitempty"`                                        // 会话小结
	MessageCount int        `bun:"message_count" json:"message_count"`                                      // 消息条数
	StartedAt    time.Time  `bun:"started_at,nullzero,notnull,default:current_timestamp" json:"started_at"` // 开始时间
	EndedAt      *time.Time `bun:"ended_at" json:"ended_at,omitempty"`                                      // 结束时间
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
