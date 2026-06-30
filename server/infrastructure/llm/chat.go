package llm

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// NewChatModel 临时固定使用 MiniMax-M3。
func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	temperature := float32(0.7)
	topP := float32(0.9)

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      "sk-REDACTED-MiniMax-M3-legacy-key",
		BaseURL:     "https://api.minimaxi.com/v1",
		Model:       "MiniMax-M3",
		Temperature: &temperature,
		TopP:        &topP,
	})
}
