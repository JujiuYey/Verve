package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 系统模型(具体 LLM)
type SysModel struct {
	bun.BaseModel `bun:"table:sys_models,alias:sm"`

	ID           string                 `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	PlatformID   string                 `bun:"platform_id,notnull,type:varchar(32)" json:"platform_id"`                 // 所属平台ID
	ModelName    string                 `bun:"model_name,notnull" json:"model_name"`                                    // 模型名
	DisplayName  string                 `bun:"display_name,notnull" json:"display_name"`                                // 显示名称
	ModelType    string                 `bun:"model_type,notnull" json:"model_type"`                                    // 模型类型,chat / embedding
	Capabilities []string               `bun:"capabilities,type:jsonb" json:"capabilities,omitempty"`                   // 能力标签,tool / vision 等
	Source       string                 `bun:"source,notnull" json:"source"`                                            // 来源,builtin / remote / custom
	Status       string                 `bun:"status,notnull" json:"status"`                                            // 状态,active / inactive
	IsDefault    bool                   `bun:"is_default,notnull" json:"is_default"`                                    // 是否默认模型
	Temperature  float32                `bun:"temperature,notnull" json:"temperature"`                                  // 温度参数
	TopP         float32                `bun:"top_p,notnull" json:"top_p"`                                              // top_p 参数
	MaxTokens    *int64                 `bun:"max_tokens" json:"max_tokens,omitempty"`                                  // 最大输出 token 数
	TopK         *int64                 `bun:"top_k" json:"top_k,omitempty"`                                            // top_k 参数
	LastSyncedAt *time.Time             `bun:"last_synced_at" json:"last_synced_at,omitempty"`                          // 最近一次同步时间
	Metadata     map[string]interface{} `bun:"metadata,type:jsonb" json:"metadata,omitempty"`                           // 扩展元数据
	CreatedAt    time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt    time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
