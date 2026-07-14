package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 文件夹模型
type Folder struct {
	bun.BaseModel `bun:"table:wiki_folders,alias:f"`

	ID          string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	Name        string    `bun:"name,notnull" json:"name"`                                                // 文件夹名称
	Description *string   `bun:"description" json:"description"`                                          // 描述
	ParentID    *string   `bun:"parent_id,type:varchar(32)" json:"parent_id"`                             // 父文件夹ID
	SortOrder   int       `bun:"sort_order,notnull" json:"sort_order"`                                    // 排序权重
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
