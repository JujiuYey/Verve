package db

import (
	"time"

	"github.com/uptrace/bun"
)

// AgentModelConfig 记录 agent 在具体场景下使用哪个模型。
type AgentModelConfig struct {
	bun.BaseModel `bun:"table:sys_agent_model_configs,alias:samc"`

	ID        string                 `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	AgentKey  string                 `bun:"agent_key,notnull" json:"agent_key"`                                      // Agent 标识,如 coach / wiki_rag
	SceneKey  string                 `bun:"scene_key,notnull" json:"scene_key"`                                      // 场景标识,如 default / embedding
	ModelID   string                 `bun:"model_id,notnull,type:varchar(32)" json:"model_id"`                       // 绑定的模型 ID
	Params    map[string]interface{} `bun:"params,type:jsonb" json:"params,omitempty"`                               // 场景级模型参数
	Enabled   bool                   `bun:"enabled,notnull" json:"enabled"`                                          // 是否启用该绑定
	CreatedAt time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt time.Time              `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间

	Model *SysModel `bun:"rel:belongs-to,join:model_id=id" json:"model,omitempty"` // 绑定的模型
}
