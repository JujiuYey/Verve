package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 模型平台(LLM 提供方)
type SysModelPlatform struct {
	bun.BaseModel `bun:"table:sys_model_platforms,alias:smp"`

	ID               string                 `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	Name             string                 `bun:"name,notnull" json:"name"`                                                // 平台名称
	ProviderType     string                 `bun:"provider_type,notnull" json:"provider_type"`                              // 提供方类型,openai_compatible 等
	DefaultBaseURL   string                 `bun:"default_base_url,notnull" json:"default_base_url"`                        // 提供方默认 BaseURL
	BaseURL          string                 `bun:"base_url,notnull" json:"base_url"`                                        // 当前使用 BaseURL
	APIKeyCiphertext string                 `bun:"api_key_ciphertext" json:"-"`                                             // 加密后的 API Key
	APIKeyHint       *string                `bun:"api_key_hint" json:"api_key_hint,omitempty"`                              // API Key 末位提示
	ExtraHeaders     map[string]interface{} `bun:"extra_headers,type:jsonb" json:"extra_headers,omitempty"`                 // 额外请求头
	ModelListPath    string                 `bun:"model_list_path,notnull" json:"model_list_path"`                          // 模型列表接口路径
	AuthScheme       string                 `bun:"auth_scheme,notnull" json:"auth_scheme"`                                  // 鉴权方式,bearer / x-api-key
	DocsURL          *string                `bun:"docs_url" json:"docs_url,omitempty"`                                      // 文档地址
	APIKeyURL        *string                `bun:"api_key_url" json:"api_key_url,omitempty"`                                // API Key 申请地址
	Enabled          bool                   `bun:"enabled,notnull" json:"enabled"`                                          // 是否启用
	SortOrder        int                    `bun:"sort_order,notnull" json:"sort_order"`                                    // 排序权重
	LastModelSyncAt  *time.Time             `bun:"last_model_sync_at" json:"last_model_sync_at,omitempty"`                  // 最近一次同步模型时间
	Metadata         map[string]interface{} `bun:"metadata,type:jsonb" json:"metadata,omitempty"`                           // 扩展元数据
	CreatedAt        time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt        time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
