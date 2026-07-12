package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

var (
	ErrTurnNotProcessing   = errors.New("learning turn is not processing")
	ErrTurnRequestConflict = errors.New("learning turn request payload conflicts with existing request")
)

type BeginTurnInput struct {
	SessionID string
	RequestID string
	AgentType string
	Content   string
}

type BeginTurnResult struct {
	Turn    *learning_db.LearningTurn
	Created bool
}

// 学习轮次数据访问层
type TurnRepository struct {
	db *bun.DB
}

func NewTurnRepository(database *bun.DB) *TurnRepository {
	return &TurnRepository{db: database}
}

func (r *TurnRepository) BeginListenerTurn(ctx context.Context, sessionID, requestID, explanation string) (*BeginTurnResult, error) {
	return r.BeginTurn(ctx, BeginTurnInput{
		SessionID: sessionID, RequestID: requestID,
		AgentType: learning_db.LearningAgentListener, Content: explanation,
	})
}

func (r *TurnRepository) BeginTurn(ctx context.Context, input BeginTurnInput) (*BeginTurnResult, error) {
	now := time.Now()
	turn := &learning_db.LearningTurn{
		ID: newLearningID(), SessionID: input.SessionID, RequestID: input.RequestID,
		AgentType: input.AgentType, Status: learning_db.LearningTurnProcessing,
		StartedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	result := &BeginTurnResult{Turn: turn}
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		insertResult, err := tx.NewInsert().Model(turn).
			On("CONFLICT (session_id, request_id) DO NOTHING").
			Returning("").
			Exec(ctx)
		if err != nil {
			return err
		}
		affected, err := insertResult.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			existing := new(learning_db.LearningTurn)
			if err := tx.NewSelect().Model(existing).
				Where("session_id = ?", input.SessionID).
				Where("request_id = ?", input.RequestID).
				Scan(ctx); err != nil {
				return err
			}
			var content string
			if err := tx.NewSelect().Table("learning_messages").
				Column("content").
				Where("turn_id = ?", existing.ID).
				Where("role = ?", "user").
				Limit(1).
				Scan(ctx, &content); err != nil {
				return err
			}
			if existing.AgentType != input.AgentType || content != input.Content {
				return ErrTurnRequestConflict
			}
			result.Turn = existing
			return nil
		}

		message := &learning_db.LearningMessage{
			ID: newLearningID(), SessionID: input.SessionID, TurnID: turn.ID,
			Role: "user", Content: input.Content, CreatedAt: now, UpdatedAt: now,
		}
		if _, err := tx.NewInsert().Model(message).Returning("").Exec(ctx); err != nil {
			return err
		}
		result.Created = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TurnRepository) CompleteListenerTurn(ctx context.Context, sessionID, turnID, assistantContent string, review *learning_db.LearningExplanationReview) error {
	if review == nil {
		return errors.New("learning explanation review is required")
	}
	now := time.Now()
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		result, err := tx.NewUpdate().Model((*learning_db.LearningTurn)(nil)).
			Set("status = ?", learning_db.LearningTurnCompleted).
			Set("completed_at = ?", now).
			Set("updated_at = ?", now).
			Set("error_code = NULL").
			Set("error_message = NULL").
			Where("id = ?", turnID).
			Where("session_id = ?", sessionID).
			Where("status = ?", learning_db.LearningTurnProcessing).
			Exec(ctx)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected != 1 {
			return ErrTurnNotProcessing
		}

		message := &learning_db.LearningMessage{
			ID: newLearningID(), SessionID: sessionID, TurnID: turnID,
			Role: "assistant", Content: assistantContent, CreatedAt: now, UpdatedAt: now,
		}
		if _, err := tx.NewInsert().Model(message).Returning("").Exec(ctx); err != nil {
			return err
		}
		review.ID = newLearningID()
		review.TurnID = turnID
		review.CreatedAt = now
		_, err = tx.NewInsert().Model(review).Returning("").Exec(ctx)
		return err
	})
}

func (r *TurnRepository) CompleteTeacherTurn(ctx context.Context, sessionID, turnID, assistantContent string, intervention *learning_db.LearningTeachingIntervention) error {
	if intervention == nil {
		return errors.New("learning teaching intervention is required")
	}
	return r.completeTurn(ctx, sessionID, turnID, assistantContent, func(ctx context.Context, tx bun.Tx, now time.Time) error {
		intervention.ID = newLearningID()
		intervention.TurnID = turnID
		intervention.CreatedAt = now
		_, err := tx.NewInsert().Model(intervention).Returning("").Exec(ctx)
		return err
	})
}

func (r *TurnRepository) CompleteTurn(ctx context.Context, sessionID, turnID, assistantContent string) error {
	return r.completeTurn(ctx, sessionID, turnID, assistantContent, nil)
}

func (r *TurnRepository) completeTurn(ctx context.Context, sessionID, turnID, assistantContent string, writeArtifact func(context.Context, bun.Tx, time.Time) error) error {
	now := time.Now()
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		result, err := tx.NewUpdate().Model((*learning_db.LearningTurn)(nil)).
			Set("status = ?", learning_db.LearningTurnCompleted).
			Set("completed_at = ?", now).
			Set("updated_at = ?", now).
			Set("error_code = NULL").
			Set("error_message = NULL").
			Where("id = ?", turnID).
			Where("session_id = ?", sessionID).
			Where("status = ?", learning_db.LearningTurnProcessing).
			Exec(ctx)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected != 1 {
			return ErrTurnNotProcessing
		}
		message := &learning_db.LearningMessage{
			ID: newLearningID(), SessionID: sessionID, TurnID: turnID,
			Role: "assistant", Content: assistantContent, CreatedAt: now, UpdatedAt: now,
		}
		if _, err := tx.NewInsert().Model(message).Returning("").Exec(ctx); err != nil {
			return err
		}
		if writeArtifact != nil {
			return writeArtifact(ctx, tx, now)
		}
		return nil
	})
}

func (r *TurnRepository) FailTurn(ctx context.Context, turnID, errorCode, errorMessage string) error {
	now := time.Now()
	errorMessage = strings.TrimSpace(errorMessage)
	if len(errorMessage) > 2000 {
		errorMessage = errorMessage[:2000]
	}
	result, err := r.db.NewUpdate().Model((*learning_db.LearningTurn)(nil)).
		Set("status = ?", learning_db.LearningTurnFailed).
		Set("error_code = ?", errorCode).
		Set("error_message = ?", errorMessage).
		Set("completed_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", turnID).
		Where("status = ?", learning_db.LearningTurnProcessing).
		Exec(ctx)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrTurnNotProcessing
	}
	return nil
}

func (r *TurnRepository) RetryFailedTurn(ctx context.Context, turnID string) error {
	now := time.Now()
	result, err := r.db.NewUpdate().Model((*learning_db.LearningTurn)(nil)).
		Set("status = ?", learning_db.LearningTurnProcessing).
		Set("error_code = NULL").
		Set("error_message = NULL").
		Set("completed_at = NULL").
		Set("started_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", turnID).
		Where("status = ?", learning_db.LearningTurnFailed).
		Exec(ctx)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrTurnNotProcessing
	}
	return nil
}

func newLearningID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
