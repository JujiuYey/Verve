package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 系统模型(具体 LLM)
type SysModel struct {
	bun.BaseModel `bun:"table:sys_models,alias:sm"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	PlatformID   string     `bun:"platform_id,notnull,type:varchar(32)" json:"platform_id"`                 // 所属平台ID
	ModelName    string     `bun:"model_name,notnull" json:"model_name"`                                    // 模型名
	DisplayName  string     `bun:"display_name,notnull" json:"display_name"`                                // 显示名称
	Status       string     `bun:"status,notnull" json:"status"`                                            // 状态,active / inactive
	LastSyncedAt *time.Time `bun:"last_synced_at" json:"last_synced_at,omitempty"`                          // 最近一次同步时间
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
