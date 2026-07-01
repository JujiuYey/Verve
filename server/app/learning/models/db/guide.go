package db

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// 导学 Agent 结果缓存
type LearningGuide struct {
	bun.BaseModel `bun:"table:learning_guides,alias:lgd"`

	ID          string          `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	ObjectiveID string          `bun:"objective_id,notnull" json:"objective_id"`                                // 小目标ID
	UserID      string          `bun:"user_id,notnull" json:"user_id"`                                          // 用户ID
	ContentHash string          `bun:"content_hash,notnull" json:"content_hash"`                                // 内容哈希(用于去重 / 缓存命中)
	Result      json.RawMessage `bun:"result,type:jsonb,notnull" json:"result"`                                 // Agent 输出结果(JSON)
	CreatedAt   time.Time       `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt   time.Time       `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
