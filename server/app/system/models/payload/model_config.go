package payload

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
