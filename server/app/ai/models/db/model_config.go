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
	ID          string    `json:"id"`
	Vendor      string    `json:"vendor"`
	Name        string    `json:"name"`
	APIKey      string    `json:"-"`
	BaseURL     string    `json:"base_url"`
	ModelType   string    `json:"model_type"`
	Model       string    `json:"model"`
	Temperature float32   `json:"temperature"`
	TopP        float32   `json:"top_p"`
	MaxTokens   *int64    `json:"max_tokens,omitempty"`
	TopK        *int64    `json:"top_k,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SysModelPlatform struct {
	bun.BaseModel `bun:"table:sys_model_platforms,alias:smp"`

	ID               string                 `bun:"id,pk,type:varchar(32)" json:"id"`
	Name             string                 `bun:"name,notnull" json:"name"`
	ProviderType     string                 `bun:"provider_type,notnull" json:"provider_type"`
	DefaultBaseURL   string                 `bun:"default_base_url,notnull" json:"default_base_url"`
	BaseURL          string                 `bun:"base_url,notnull" json:"base_url"`
	APIKeyCiphertext string                 `bun:"api_key_ciphertext" json:"-"`
	APIKeyHint       *string                `bun:"api_key_hint" json:"api_key_hint,omitempty"`
	ExtraHeaders     map[string]interface{} `bun:"extra_headers,type:jsonb" json:"extra_headers,omitempty"`
	ModelListPath    string                 `bun:"model_list_path,notnull" json:"model_list_path"`
	AuthScheme       string                 `bun:"auth_scheme,notnull" json:"auth_scheme"`
	DocsURL          *string                `bun:"docs_url" json:"docs_url,omitempty"`
	APIKeyURL        *string                `bun:"api_key_url" json:"api_key_url,omitempty"`
	Enabled          bool                   `bun:"enabled,notnull" json:"enabled"`
	SortOrder        int                    `bun:"sort_order,notnull" json:"sort_order"`
	LastModelSyncAt  *time.Time             `bun:"last_model_sync_at" json:"last_model_sync_at,omitempty"`
	Metadata         map[string]interface{} `bun:"metadata,type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type SysModel struct {
	bun.BaseModel `bun:"table:sys_models,alias:sm"`

	ID           string                 `bun:"id,pk,type:varchar(32)" json:"id"`
	PlatformID   string                 `bun:"platform_id,notnull,type:varchar(32)" json:"platform_id"`
	ModelName    string                 `bun:"model_name,notnull" json:"model_name"`
	DisplayName  string                 `bun:"display_name,notnull" json:"display_name"`
	ModelType    string                 `bun:"model_type,notnull" json:"model_type"`
	Capabilities []string               `bun:"capabilities,type:jsonb" json:"capabilities,omitempty"`
	Source       string                 `bun:"source,notnull" json:"source"`
	Status       string                 `bun:"status,notnull" json:"status"`
	IsDefault    bool                   `bun:"is_default,notnull" json:"is_default"`
	Temperature  float32                `bun:"temperature,notnull" json:"temperature"`
	TopP         float32                `bun:"top_p,notnull" json:"top_p"`
	MaxTokens    *int64                 `bun:"max_tokens" json:"max_tokens,omitempty"`
	TopK         *int64                 `bun:"top_k" json:"top_k,omitempty"`
	LastSyncedAt *time.Time             `bun:"last_synced_at" json:"last_synced_at,omitempty"`
	Metadata     map[string]interface{} `bun:"metadata,type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
