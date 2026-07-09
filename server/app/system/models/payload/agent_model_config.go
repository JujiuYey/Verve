package payload

type UpsertAgentModelConfigRequest struct {
	ModelID string                 `json:"model_id"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Enabled *bool                  `json:"enabled,omitempty"`
}
