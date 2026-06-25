package db

import (
	"time"

	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
)

// 对话会话模型
type AgentSession struct {
	bun.BaseModel `bun:"table:ai_agent_sessions,alias:as"`

	ID           string    `bun:"id,pk,type:varchar(32)" json:"id"`
	Title        *string   `bun:"title" json:"title,omitempty"`
	UserID       *string   `bun:"user_id" json:"user_id,omitempty"`
	MessageCount int       `bun:"message_count" json:"message_count"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// 关联关系
	Messages []*AgentMessage `bun:"rel:has-many,join:id=session_id" json:"messages,omitempty"`
	User     *system_db.User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
