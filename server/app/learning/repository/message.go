package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 陪练消息数据访问层
type MessageRepository struct {
	db *bun.DB
}

func NewMessageRepository(database *bun.DB) *MessageRepository {
	return &MessageRepository{db: database}
}

func (r *MessageRepository) GetDB() *bun.DB { return r.db }

// 按会话列出消息(时间顺序)
func (r *MessageRepository) FindBySession(ctx context.Context, sessionID string) ([]*learning_db.LearningMessage, error) {
	var messages []*learning_db.LearningMessage
	err := r.db.NewSelect().Model(&messages).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *MessageRepository) Create(ctx context.Context, message *learning_db.LearningMessage) error {
	message.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(message).Exec(ctx)
	return err
}
