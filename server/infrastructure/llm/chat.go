package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"

	system_db "verve/app/system/models/db"
)

// AgentModelResolver 暴露查找 agent 场景模型配置所需的最小接口。
// system_repository.ModelConfigRepository 已经满足该接口。
type AgentModelResolver interface {
	FindAgentModelWithPlatform(ctx context.Context, agentKey, sceneKey string) (*system_db.SysModel, *system_db.SysModelPlatform, error)
}

const (
	defaultTemperature    = float32(0.7)
	structuredTemperature = float32(0.2)
	defaultTopP           = float32(0.9)
)

func resolveAgentModel(ctx context.Context, resolver AgentModelResolver, agentKey, sceneKey string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
	if resolver == nil {
		return nil, nil, fmt.Errorf("agent model resolver is required for %s/%s", agentKey, sceneKey)
	}
	return resolver.FindAgentModelWithPlatform(ctx, agentKey, sceneKey)
}

// NewChatModel 根据 agent_key / scene_key 在系统配置中查找模型和平台,
// 并基于 OpenAI 兼容接口构造对话模型。未配置时返回错误,避免静默回退。
func NewChatModel(ctx context.Context, resolver AgentModelResolver, agentKey, sceneKey string) (model.ToolCallingChatModel, error) {
	modelCfg, platform, err := resolveAgentModel(ctx, resolver, agentKey, sceneKey)
	if err != nil {
		return nil, fmt.Errorf("agent %s/%s model not configured: %w", agentKey, sceneKey, err)
	}
	temperature := defaultTemperature
	topP := defaultTopP
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      strings.TrimSpace(platform.APIKeyCiphertext),
		BaseURL:     strings.TrimRight(strings.TrimSpace(platform.BaseURL), "/"),
		Model:       strings.TrimSpace(modelCfg.ModelName),
		Temperature: &temperature,
		TopP:        &topP,
	})
}

// NewStructuredChatModel 用于只需要 JSON 的任务,在 OpenAI 兼容接口上请求 JSON object。
func NewStructuredChatModel(ctx context.Context, resolver AgentModelResolver, agentKey, sceneKey string) (model.ToolCallingChatModel, error) {
	modelCfg, platform, err := resolveAgentModel(ctx, resolver, agentKey, sceneKey)
	if err != nil {
		return nil, fmt.Errorf("agent %s/%s model not configured: %w", agentKey, sceneKey, err)
	}
	temperature := structuredTemperature
	topP := defaultTopP
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      strings.TrimSpace(platform.APIKeyCiphertext),
		BaseURL:     strings.TrimRight(strings.TrimSpace(platform.BaseURL), "/"),
		Model:       strings.TrimSpace(modelCfg.ModelName),
		Temperature: &temperature,
		TopP:        &topP,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
}
