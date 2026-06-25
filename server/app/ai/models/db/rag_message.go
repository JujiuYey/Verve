package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 对话消息模型
type RagMessage struct {
	bun.BaseModel `bun:"table:ai_rag_messages,alias:rm"`

	ID              string                   `bun:"id,pk,type:varchar(32)" json:"id"`
	SessionID       string                   `bun:"session_id,notnull" json:"session_id"`
	Role            string                   `bun:"role,notnull" json:"role"`
	Content         string                   `bun:"content,notnull" json:"content"`
	RetrievedChunks []map[string]interface{} `bun:"retrieved_chunks,type:jsonb" json:"retrieved_chunks,omitempty"`
	RelevanceScore  *float64                 `bun:"relevance_score" json:"relevance_score,omitempty"`
	PromptTokens    *int64                   `bun:"prompt_tokens" json:"prompt_tokens,omitempty"`     // 输入 token 数
	CompletionTokens *int64                  `bun:"completion_tokens" json:"completion_tokens,omitempty"` // 输出 token 数
	TotalTokens     *int64                   `bun:"total_tokens" json:"total_tokens,omitempty"`       // 总 token 数
	CreatedAt       time.Time                `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time                `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// 关联关系
	Session *RagSession `bun:"rel:belongs-to,join:session_id=id" json:"session,omitempty"`
}
