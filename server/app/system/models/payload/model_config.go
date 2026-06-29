package payload

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

type CreateAIPlatformRequest struct {
	Name           string  `json:"name"`
	ProviderType   string  `json:"provider_type"`
	DefaultBaseURL string  `json:"default_base_url"`
	BaseURL        string  `json:"base_url"`
	APIKey         string  `json:"api_key"`
	ModelListPath  string  `json:"model_list_path"`
	AuthScheme     string  `json:"auth_scheme"`
	DocsURL        *string `json:"docs_url,omitempty"`
	APIKeyURL      *string `json:"api_key_url,omitempty"`
}

type UpdateAIPlatformConfigRequest struct {
	BaseURL     string  `json:"base_url"`
	APIKey      *string `json:"api_key,omitempty"`
	ClearAPIKey bool    `json:"clear_api_key"`
}

type CreateAIModelRequest struct {
	PlatformID   string   `json:"platform_id"`
	ModelName    string   `json:"model_name"`
	DisplayName  string   `json:"display_name"`
	ModelType    string   `json:"model_type"`
	Capabilities []string `json:"capabilities,omitempty"`
	Source       string   `json:"source"`
	IsDefault    bool     `json:"is_default"`
	Temperature  float32  `json:"temperature"`
	TopP         float32  `json:"top_p"`
	MaxTokens    *int64   `json:"max_tokens,omitempty"`
	TopK         *int64   `json:"top_k,omitempty"`
}

type UpdateAIModelRequest struct {
	Status       *string  `json:"status,omitempty"`
	DisplayName  *string  `json:"display_name,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}
