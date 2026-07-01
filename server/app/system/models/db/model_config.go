package db

import "time"

const (
	ModelTypeChat      = "chat"
	ModelTypeEmbedding = "embedding"
)

// 模型运行时配置(组合平台与模型后的最终配置)
type ModelConfig struct {
	ID          string    `json:"id"`                   // 模型ID
	Vendor      string    `json:"vendor"`               // 提供方
	Name        string    `json:"name"`                 // 模型名
	APIKey      string    `json:"-"`                    // API Key
	BaseURL     string    `json:"base_url"`             // BaseURL
	ModelType   string    `json:"model_type"`           // 模型类型,chat / embedding
	Model       string    `json:"model"`                // 模型标识
	Temperature float32   `json:"temperature"`          // 温度参数
	TopP        float32   `json:"top_p"`                // top_p 参数
	MaxTokens   *int64    `json:"max_tokens,omitempty"` // 最大输出 token 数
	TopK        *int64    `json:"top_k,omitempty"`      // top_k 参数
	IsActive    bool      `json:"is_active"`            // 是否启用
	IsDefault   bool      `json:"is_default"`           // 是否默认
	CreatedAt   time.Time `json:"created_at"`           // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`           // 更新时间
}
