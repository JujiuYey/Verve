package db

import (
	"time"

	"sag-wiki/app/system/models/db"

	"github.com/uptrace/bun"
)

// 文件夹模型
type Folder struct {
	bun.BaseModel `bun:"table:wiki_folders,alias:f"`

	ID            string    `bun:"id,pk,type:varchar(32)" json:"id"`
	Name          string    `bun:"name,notnull" json:"name"`
	Description   *string   `bun:"description" json:"description"`
	ParentID      *string   `bun:"parent_id,type:varchar(32)" json:"parent_id"`
	UserID        *string   `bun:"user_id,type:varchar(32)" json:"user_id"`
	CreatedBy     *string   `bun:"created_by,type:varchar(32)" json:"created_by"`
	UpdatedBy     *string   `bun:"updated_by,type:varchar(32)" json:"updated_by"`
	CreatedAt     time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// 关联查询：创建人信息
	CreatedByUser *db.User `bun:"rel:has-one,join:created_by=id" json:"created_by_user"`
	// 关联查询：更新人信息
	UpdatedByUser *db.User `bun:"rel:has-one,join:updated_by=id" json:"updated_by_user"`
}
