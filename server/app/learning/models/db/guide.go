package db

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// 导学 Agent 结果缓存
type LearningGuide struct {
	bun.BaseModel `bun:"table:learning_guides,alias:lgd"`

	ID          string          `bun:"id,pk,type:varchar(32)" json:"id"`
	ObjectiveID string          `bun:"objective_id,notnull" json:"objective_id"`
	UserID      string          `bun:"user_id,notnull" json:"user_id"`
	ContentHash string          `bun:"content_hash,notnull" json:"content_hash"`
	Result      json.RawMessage `bun:"result,type:jsonb,notnull" json:"result"`
	CreatedAt   time.Time       `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
