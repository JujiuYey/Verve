package repository

import (
	"testing"

	ai_db "sag-wiki/app/ai/models/db"
)

func TestBuildModelConfigUsesPlatformAndModelFields(t *testing.T) {
	t.Parallel()

	platform := &ai_db.SysModelPlatform{
		ID:               "platform-1",
		Name:             "DeepSeek",
		DefaultBaseURL:   "https://default.example.com",
		BaseURL:          "https://api.deepseek.com",
		APIKeyCiphertext: "ciphertext-key",
	}
	model := &ai_db.SysModel{
		ID:          "model-1",
		DisplayName: "DeepSeek Chat",
		ModelName:   "deepseek-chat",
		ModelType:   ai_db.ModelTypeChat,
		Status:      "active",
		IsDefault:   true,
		Temperature: 0.3,
		TopP:        0.8,
	}

	config := buildModelConfig(platform, model)

	if config.ID != model.ID {
		t.Fatalf("expected model ID %q, got %q", model.ID, config.ID)
	}
	if config.Vendor != platform.Name {
		t.Fatalf("expected vendor %q, got %q", platform.Name, config.Vendor)
	}
	if config.BaseURL != platform.BaseURL {
		t.Fatalf("expected base URL %q, got %q", platform.BaseURL, config.BaseURL)
	}
	if config.APIKey != platform.APIKeyCiphertext {
		t.Fatalf("expected API key ciphertext to be used internally")
	}
	if !config.IsActive || !config.IsDefault {
		t.Fatalf("expected active default config, got active=%v default=%v", config.IsActive, config.IsDefault)
	}
}

func TestBuildModelConfigFallsBackToDefaultBaseURL(t *testing.T) {
	t.Parallel()

	config := buildModelConfig(
		&ai_db.SysModelPlatform{Name: "OpenAI", DefaultBaseURL: "https://api.openai.com/v1"},
		&ai_db.SysModel{Status: "inactive"},
	)

	if config.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected default base URL fallback, got %q", config.BaseURL)
	}
	if config.IsActive {
		t.Fatalf("expected inactive status to map to IsActive=false")
	}
}
