package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 文件夹权限模型
type FolderPermission struct {
	bun.BaseModel `bun:"table:wiki_folder_permissions,alias:fp"`

	ID             string    `bun:"id,pk,type:varchar(32)" json:"id"`
	FolderID       string    `bun:"folder_id,notnull,type:varchar(32)" json:"folder_id"`
	UserID         *string   `bun:"user_id,type:varchar(32)" json:"user_id"`
	DepartmentID   *string   `bun:"department_id,type:varchar(32)" json:"department_id"`
	PermissionType string    `bun:"permission_type,notnull,default:'read'" json:"permission_type"`
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// 权限类型常量
const (
	PermissionManage = "manage"
	PermissionEdit   = "edit"
	PermissionRead   = "read"
)
