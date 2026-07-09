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
	PlatformID  string `json:"platform_id"`
	ModelName   string `json:"model_name"`
	DisplayName string `json:"display_name"`
}

type UpdateAIModelRequest struct {
	Status      *string `json:"status,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
}
