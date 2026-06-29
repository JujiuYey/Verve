package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"

	system_db "sag-wiki/app/system/models/db"
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
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
		Model:   config.Model,
		Temperature: &config.Temperature,
		TopP: &config.TopP,
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
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
		Model:   config.Model,
		Temperature: &config.Temperature,
		TopP: &config.TopP,
	})
}
