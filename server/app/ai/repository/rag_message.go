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
type RagMessageRepository struct {
	db *bun.DB
}

// 创建 RAG 消息仓储
func NewRagMessageRepository(database *bun.DB) *RagMessageRepository {
	return &RagMessageRepository{db: database}
}

// 创建 RAG 消息
func (r *RagMessageRepository) Create(ctx context.Context, sessionID, role, content string, retrievedChunks []map[string]interface{}, relevanceScore *float64, promptTokens, completionTokens, totalTokens *int64) (*ai_db.RagMessage, error) {
	message := &ai_db.RagMessage{
		SessionID:        sessionID,
		Role:             role,
		Content:          content,
		RetrievedChunks:  retrievedChunks,
		RelevanceScore:   relevanceScore,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}

	message.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(message).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 RAG 消息失败: %w", err)
	}

	return message, nil
}

// 根据会话 ID 获取消息列表
func (r *RagMessageRepository) ListBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*ai_db.RagMessage, error) {
	var messages []*ai_db.RagMessage

	err := r.db.NewSelect().
		Model(&messages).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("获取 RAG 消息列表失败: %w", err)
	}

	return messages, nil
}

// 删除 RAG 消息
func (r *RagMessageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*ai_db.RagMessage)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("删除 RAG 消息失败: %w", err)
	}

	return nil
}

// 获取会话的 token 统计
// 注意：只统计 assistant 消息的 token，因为 prompt_tokens 包含历史消息会重复计算
func (r *RagMessageRepository) GetSessionTokenStats(ctx context.Context, sessionID string) (int64, int64, int64, error) {
	var stats struct {
		TotalPromptTokens     int64 `bun:"total_prompt_tokens"`
		TotalCompletionTokens int64 `bun:"total_completion_tokens"`
		TotalTokens           int64 `bun:"total_tokens"`
	}

	err := r.db.NewSelect().
		Model((*ai_db.RagMessage)(nil)).
		ColumnExpr("COALESCE(SUM(prompt_tokens), 0) AS total_prompt_tokens").
		ColumnExpr("COALESCE(SUM(completion_tokens), 0) AS total_completion_tokens").
		ColumnExpr("COALESCE(SUM(total_tokens), 0) AS total_tokens").
		Where("session_id = ?", sessionID).
		Where("role = ?", "assistant").
		Scan(ctx, &stats)

	if err != nil {
		return 0, 0, 0, fmt.Errorf("获取会话 token 统计失败: %w", err)
	}

	return stats.TotalPromptTokens, stats.TotalCompletionTokens, stats.TotalTokens, nil
}
