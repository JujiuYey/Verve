package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 陪练对话消息(多 agent)
type LearningMessage struct {
	bun.BaseModel `bun:"table:learning_messages,alias:lm"`

	ID               string                 `bun:"id,pk,type:varchar(32)" json:"id"`
	SessionID        string                 `bun:"session_id,notnull" json:"session_id"`
	Role             string                 `bun:"role,notnull" json:"role"`            // user / assistant / system
	AgentType        *string                `bun:"agent_type" json:"agent_type,omitempty"` // planner / tutor / examiner / orchestrator
	Content          string                 `bun:"content,notnull" json:"content"`
	ToolUsed         *string                `bun:"tool_used" json:"tool_used,omitempty"`
	ToolResult       map[string]interface{} `bun:"tool_result,type:jsonb" json:"tool_result,omitempty"`
	PromptTokens     *int64                 `bun:"prompt_tokens" json:"prompt_tokens,omitempty"`
	CompletionTokens *int64                 `bun:"completion_tokens" json:"completion_tokens,omitempty"`
	TotalTokens      *int64                 `bun:"total_tokens" json:"total_tokens,omitempty"`
	CreatedAt        time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
