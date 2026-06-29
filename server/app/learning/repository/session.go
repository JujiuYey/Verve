package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "sag-wiki/app/learning/models/db"
)

// 学习会话数据访问层
type SessionRepository struct {
	db *bun.DB
}

func NewSessionRepository(database *bun.DB) *SessionRepository {
	return &SessionRepository{db: database}
}

func (r *SessionRepository) GetDB() *bun.DB { return r.db }

func (r *SessionRepository) FindOne(ctx context.Context, id string) (*learning_db.LearningSession, error) {
	session := new(learning_db.LearningSession)
	if err := r.db.NewSelect().Model(session).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) Create(ctx context.Context, session *learning_db.LearningSession) error {
	session.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(session).Exec(ctx)
	return err
}

func (r *SessionRepository) Update(ctx context.Context, session *learning_db.LearningSession) error {
	_, err := r.db.NewUpdate().Model(session).Where("id = ?", session.ID).Exec(ctx)
	return err
}
