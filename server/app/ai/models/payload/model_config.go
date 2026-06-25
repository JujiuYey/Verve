package payload

// 创建模型配置请求载体
type CreateModelConfigRequest struct {
	Name        string  `bun:"name,notnull" json:"name"`
	Vendor      string  `bun:"vendor,notnull" json:"vendor"`
	APIKey      string  `bun:"api_key,notnull" json:"api_key"`
	BaseURL     string  `bun:"base_url,notnull" json:"base_url"`
	ModelType   string  `bun:"model_type,notnull" json:"model_type"`
	Model       string  `bun:"model,notnull" json:"model"`
	Temperature float32 `bun:"temperature,notnull" json:"temperature"`
	TopP        float32 `bun:"top_p,notnull" json:"top_p"`
	MaxTokens   *int64  `bun:"max_tokens" json:"max_tokens,omitempty"`
	TopK        *int64  `bun:"top_k" json:"top_k,omitempty"`
	IsActive    bool    `bun:"is_active,notnull" json:"is_active"`
	IsDefault   bool    `bun:"is_default,notnull" json:"is_default"`
}

// 更新模型配置请求载体
type UpdateModelConfigRequest struct {
	ID          string  `bun:"id,notnull" json:"id"`
	Vendor      string  `bun:"vendor,notnull" json:"vendor"`
	Name        string  `bun:"name,notnull" json:"name"`
	APIKey      string  `bun:"api_key,notnull" json:"api_key"`
	BaseURL     string  `bun:"base_url,notnull" json:"base_url"`
	ModelType   string  `bun:"model_type,notnull" json:"model_type"`
	Model       string  `bun:"model,notnull" json:"model"`
	Temperature float32 `bun:"temperature,notnull" json:"temperature"`
	TopP        float32 `bun:"top_p,notnull" json:"top_p"`
	MaxTokens   *int64  `bun:"max_tokens" json:"max_tokens,omitempty"`
	TopK        *int64  `bun:"top_k" json:"top_k,omitempty"`
	IsActive    bool    `bun:"is_active,notnull" json:"is_active"`
	IsDefault   bool    `bun:"is_default,notnull" json:"is_default"`
}
