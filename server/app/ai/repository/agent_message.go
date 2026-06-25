package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	ai_db "sag-wiki/app/ai/models/db"
)

// 消息仓储
type AgentMessageRepository struct {
	db *bun.DB
}

// 创建 Agent 消息仓储
func NewAgentMessageRepository(database *bun.DB) *AgentMessageRepository {
	return &AgentMessageRepository{db: database}
}

// 创建 Agent 消息
func (r *AgentMessageRepository) Create(ctx context.Context, sessionID, role, content string, toolUsed *string, toolResult map[string]interface{}, promptTokens, completionTokens, totalTokens *int64) (*ai_db.AgentMessage, error) {
	message := &ai_db.AgentMessage{
		SessionID:        sessionID,
		Role:             role,
		Content:          content,
		ToolUsed:         toolUsed,
		ToolResult:       toolResult,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}

	message.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(message).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 Agent 消息失败: %w", err)
	}

	return message, nil
}

// 根据会话 ID 获取消息列表
func (r *AgentMessageRepository) ListBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*ai_db.AgentMessage, error) {
	var messages []*ai_db.AgentMessage

	err := r.db.NewSelect().
		Model(&messages).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("获取 Agent 消息列表失败: %w", err)
	}

	return messages, nil
}

// 删除 Agent 消息
func (r *AgentMessageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*ai_db.AgentMessage)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("删除 Agent 消息失败: %w", err)
	}

	return nil
}
