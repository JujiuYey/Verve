package db

import (
	"time"

	"github.com/uptrace/bun"
)

const (
	LearningAgentListener = "listener"
	LearningAgentTeacher  = "teacher"
	LearningAgentCurator  = "curator"

	LearningTurnProcessing = "processing"
	LearningTurnCompleted  = "completed"
	LearningTurnFailed     = "failed"
)

// 学习处理轮次
type LearningTurn struct {
	bun.BaseModel `bun:"table:learning_turns,alias:lt"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 轮次ID
	SessionID    string     `bun:"session_id,notnull,type:varchar(32)" json:"session_id"`                   // 学习会话ID
	RequestID    string     `bun:"request_id,notnull,type:varchar(64)" json:"request_id"`                   // 请求幂等标识
	AgentType    string     `bun:"agent_type,notnull,type:varchar(32)" json:"agent_type"`                   // 处理Agent类型
	Status       string     `bun:"status,notnull,type:varchar(20)" json:"status"`                           // 处理状态
	ErrorCode    *string    `bun:"error_code,type:varchar(64)" json:"error_code,omitempty"`                 // 失败错误码
	ErrorMessage *string    `bun:"error_message" json:"error_message,omitempty"`                            // 失败原因
	StartedAt    time.Time  `bun:"started_at,nullzero,notnull,default:current_timestamp" json:"started_at"` // 开始时间
	CompletedAt  *time.Time `bun:"completed_at" json:"completed_at,omitempty"`                              // 完成时间
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
