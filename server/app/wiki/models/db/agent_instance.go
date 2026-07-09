package db

import (
	"time"

	"github.com/uptrace/bun"
)

// AgentInstance 记录用户绑定到某个 Wiki 根目录的学习 Agent。
type AgentInstance struct {
	bun.BaseModel `bun:"table:wiki_agent_instances,alias:wai"`

	ID           string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	UserID       string    `bun:"user_id,type:varchar(32),notnull" json:"user_id"`                         // 用户ID
	RootFolderID string    `bun:"root_folder_id,type:varchar(32),notnull" json:"root_folder_id"`           // 根文件夹ID
	AgentKey     string    `bun:"agent_key,notnull" json:"agent_key"`                                      // Agent 类型
	Name         string    `bun:"name,notnull" json:"name"`                                                // 展示名称
	Description  *string   `bun:"description" json:"description,omitempty"`                                // 描述
	Status       string    `bun:"status,notnull,default:'active'" json:"status"`                           // 状态
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间

	RootFolder *Folder `bun:"rel:belongs-to,join:root_folder_id=id" json:"root_folder,omitempty"` // 绑定根目录
}
