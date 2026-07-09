package db

import (
	"time"

	"github.com/uptrace/bun"
)

// AgentModelConfig 记录 agent 在具体场景下使用哪个模型。
type AgentModelConfig struct {
	bun.BaseModel `bun:"table:sys_agent_model_configs,alias:samc"`

	ID        string                 `bun:"id,pk,type:varchar(32)" json:"id"`
	AgentKey  string                 `bun:"agent_key,notnull" json:"agent_key"`
	SceneKey  string                 `bun:"scene_key,notnull" json:"scene_key"`
	ModelID   string                 `bun:"model_id,notnull,type:varchar(32)" json:"model_id"`
	Params    map[string]interface{} `bun:"params,type:jsonb" json:"params,omitempty"`
	Enabled   bool                   `bun:"enabled,notnull" json:"enabled"`
	CreatedAt time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	Model *SysModel `bun:"rel:belongs-to,join:model_id=id" json:"model,omitempty"`
}
