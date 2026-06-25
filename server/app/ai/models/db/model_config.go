package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 模型类型枚举
const (
	ModelTypeChat      = "chat"      // 对话模型
	ModelTypeEmbedding = "embedding" // 向量模型
)

// 模型配置模型
type ModelConfig struct {
	bun.BaseModel `bun:"table:ai_model_config,alias:mc"`

	ID          string    `bun:"id,pk,type:varchar(32)" json:"id"`
	Vendor      string    `bun:"vendor,notnull" json:"vendor"`
	Name        string    `bun:"name,notnull" json:"name"`
	APIKey      string    `bun:"api_key,notnull" json:"api_key"`
	BaseURL     string    `bun:"base_url,notnull" json:"base_url"`
	ModelType   string    `bun:"model_type,notnull" json:"model_type"`
	Model       string    `bun:"model,notnull" json:"model"`
	Temperature float32   `bun:"temperature,notnull" json:"temperature"`
	TopP        float32   `bun:"top_p,notnull" json:"top_p"`
	MaxTokens   *int64    `bun:"max_tokens" json:"max_tokens,omitempty"`
	TopK        *int64    `bun:"top_k" json:"top_k,omitempty"`
	IsActive    bool      `bun:"is_active,notnull" json:"is_active"`
	IsDefault   bool      `bun:"is_default,notnull" json:"is_default"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
