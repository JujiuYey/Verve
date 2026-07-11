package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习记忆事件(原始观察)
type LearningMemoryEvent struct {
	bun.BaseModel `bun:"table:learning_memory_events,alias:lme"`

	ID          string                 `bun:"id,pk,type:varchar(32)" json:"id"`                                          // 主键ID
	UserID      string                 `bun:"user_id,notnull" json:"user_id"`                                            // 用户ID
	FolderID    *string                `bun:"folder_id,type:varchar(32)" json:"folder_id,omitempty"`                     // Wiki 文件夹ID
	DocumentID  *string                `bun:"document_id,type:varchar(32)" json:"document_id,omitempty"`                 // Wiki 文档ID
	ObjectiveID *string                `bun:"-" json:"objective_id,omitempty"`                                           // 旧流程临时编译字段
	SessionID   *string                `bun:"session_id,type:varchar(32)" json:"session_id,omitempty"`                   // 学习会话ID
	SourceType  string                 `bun:"source_type,notnull" json:"source_type"`                                    // 来源类型
	SourceID    *string                `bun:"source_id,type:varchar(32)" json:"source_id,omitempty"`                     // 来源ID
	EventType   string                 `bun:"event_type,notnull" json:"event_type"`                                      // 事件类型
	Content     string                 `bun:"content,notnull" json:"content"`                                            // 事件内容
	Evidence    map[string]interface{} `bun:"evidence,type:jsonb,notnull" json:"evidence,omitempty"`                     // 证据元数据
	OccurredAt  time.Time              `bun:"occurred_at,nullzero,notnull,default:current_timestamp" json:"occurred_at"` // 发生时间
	CreatedAt   time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`   // 创建时间
}

// 学习记忆条目(可复用事实)
type LearningMemoryItem struct {
	bun.BaseModel `bun:"table:learning_memory_items,alias:lmi"`

	ID               string    `bun:"id,pk,type:varchar(32)" json:"id"`                                            // 主键ID
	UserID           string    `bun:"user_id,notnull" json:"user_id"`                                              // 用户ID
	FolderID         *string   `bun:"folder_id,type:varchar(32)" json:"folder_id,omitempty"`                       // Wiki 文件夹ID
	DocumentID       *string   `bun:"document_id,type:varchar(32)" json:"document_id,omitempty"`                   // Wiki 文档ID
	ObjectiveID      *string   `bun:"-" json:"objective_id,omitempty"`                                             // 旧流程临时编译字段
	Kind             string    `bun:"kind,notnull" json:"kind"`                                                    // 记忆类型
	Statement        string    `bun:"statement,notnull" json:"statement"`                                          // 记忆内容
	EvidenceEventIDs []string  `bun:"evidence_event_ids,type:jsonb,notnull" json:"evidence_event_ids"`             // 证据事件ID列表
	Confidence       string    `bun:"confidence,notnull,default:'observed'" json:"confidence"`                     // 置信度
	LastSeenAt       time.Time `bun:"last_seen_at,nullzero,notnull,default:current_timestamp" json:"last_seen_at"` // 最近观察时间
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`     // 创建时间
	UpdatedAt        time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`     // 更新时间
}

// 学习记忆汇总(按 Wiki 文件夹)
type LearningMemorySummary struct {
	bun.BaseModel `bun:"table:learning_memory_summaries,alias:lms"`

	ID                   string    `bun:"id,pk,type:varchar(32)" json:"id"`                                                  // 主键ID
	UserID               string    `bun:"user_id,notnull" json:"user_id"`                                                    // 用户ID
	FolderID             *string   `bun:"folder_id,type:varchar(32)" json:"folder_id,omitempty"`                             // Wiki 文件夹ID
	Summary              string    `bun:"summary,notnull" json:"summary"`                                                    // 汇总内容
	Highlights           []string  `bun:"highlights,type:jsonb,notnull" json:"highlights"`                                   // 重点摘要
	GeneratedFromEventID *string   `bun:"generated_from_event_id,type:varchar(32)" json:"generated_from_event_id,omitempty"` // 生成起点事件ID
	GeneratedAt          time.Time `bun:"generated_at,nullzero,notnull,default:current_timestamp" json:"generated_at"`       // 生成时间
	CreatedAt            time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`           // 创建时间
	UpdatedAt            time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`           // 更新时间
}
