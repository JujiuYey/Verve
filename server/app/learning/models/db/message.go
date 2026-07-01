package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 陪练对话消息(多 agent)
type LearningMessage struct {
	bun.BaseModel `bun:"table:learning_messages,alias:lm"`

	ID               string                 `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	SessionID        string                 `bun:"session_id,notnull" json:"session_id"`                                    // 会话ID
	Role             string                 `bun:"role,notnull" json:"role"`                                                // user / assistant / system
	AgentType        *string                `bun:"agent_type" json:"agent_type,omitempty"`                                  // tutor / examiner / guide
	Content          string                 `bun:"content,notnull" json:"content"`                                          // 消息内容
	ToolUsed         *string                `bun:"tool_used" json:"tool_used,omitempty"`                                    // 使用的工具
	ToolResult       map[string]interface{} `bun:"tool_result,type:jsonb" json:"tool_result,omitempty"`                     // 工具返回结果
	PromptTokens     *int64                 `bun:"prompt_tokens" json:"prompt_tokens,omitempty"`                            // 输入 token 数
	CompletionTokens *int64                 `bun:"completion_tokens" json:"completion_tokens,omitempty"`                    // 输出 token 数
	TotalTokens      *int64                 `bun:"total_tokens" json:"total_tokens,omitempty"`                              // 总 token 数
	CreatedAt        time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt        time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
