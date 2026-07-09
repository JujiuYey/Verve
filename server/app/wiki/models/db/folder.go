package db

import (
	"time"

	"verve/app/system/models/db"

	"github.com/uptrace/bun"
)

// 文件夹模型
type Folder struct {
	bun.BaseModel `bun:"table:wiki_folders,alias:f"`

	ID            string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	Name          string    `bun:"name,notnull" json:"name"`                                                // 文件夹名称
	Description   *string   `bun:"description" json:"description"`                                          // 描述
	ParentID      *string   `bun:"parent_id,type:varchar(32)" json:"parent_id"`                             // 父文件夹ID
	UserID        *string   `bun:"user_id,type:varchar(32)" json:"user_id"`                                 // 所属用户ID
	SortOrder     int       `bun:"sort_order,notnull" json:"sort_order"`                                    // 排序权重
	CreatedBy     *string   `bun:"created_by,type:varchar(32)" json:"created_by"`                           // 创建人ID
	UpdatedBy     *string   `bun:"updated_by,type:varchar(32)" json:"updated_by"`                           // 更新人ID
	CreatedAt     time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt     time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
	CreatedByUser *db.User  `bun:"rel:has-one,join:created_by=id" json:"created_by_user"`                   // 关联查询：创建人信息
	UpdatedByUser *db.User  `bun:"rel:has-one,join:updated_by=id" json:"updated_by_user"`                   // 关联查询：更新人信息
}
