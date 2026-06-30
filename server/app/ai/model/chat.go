package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"

	system_db "sag-wiki/app/system/models/db"
)

const (
	PlannerModelName    = "MiniMax-M3"
	PlannerModelBaseURL = "https://api.minimaxi.com/v1"
)

// NewChatModel 根据默认配置创建 ChatModel
func NewChatModel(ctx context.Context, repo interface {
	FindDefault(ctx context.Context) (*system_db.ModelConfig, error)
}) (model.ToolCallingChatModel, error) {
	config, err := repo.FindDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取默认模型配置失败: %w", err)
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      config.APIKey,
		BaseURL:     config.BaseURL,
		Model:       config.Model,
		Temperature: &config.Temperature,
		TopP:        &config.TopP,
	})
}

// NewPlannerChatModel 临时固定 Planner 使用 MiniMax-M3。
func NewPlannerChatModel(ctx context.Context, repo interface {
	FindByModelName(ctx context.Context, modelName string) (*system_db.ModelConfig, error)
}) (model.ToolCallingChatModel, error) {
	config, err := repo.FindByModelName(ctx, PlannerModelName)
	if err != nil {
		return nil, fmt.Errorf("获取 Planner 模型配置失败: %w", err)
	}

	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = PlannerModelBaseURL
	}

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      config.APIKey,
		BaseURL:     baseURL,
		Model:       PlannerModelName,
		Temperature: &config.Temperature,
		TopP:        &config.TopP,
	})
}

// NewChatModelByID 根据指定 ID 创建 ChatModel
func NewChatModelByID(ctx context.Context, repo interface {
	FindOne(ctx context.Context, id string) (*system_db.ModelConfig, error)
}, id string) (model.ToolCallingChatModel, error) {
	config, err := repo.FindOne(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取模型配置失败: %w", err)
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      config.APIKey,
		BaseURL:     config.BaseURL,
		Model:       config.Model,
		Temperature: &config.Temperature,
		TopP:        &config.TopP,
	})
}
