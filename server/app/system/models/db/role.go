package db


import (
	"time"

	"github.com/uptrace/bun"
)

// 角色模型
type Role struct {
	bun.BaseModel `bun:"table:sys_roles,alias:sr"`

	ID          string     `bun:"id,pk,type:varchar(32)" json:"id"`
	Name        string     `bun:"name,notnull,unique" json:"name"`
	Description *string    `bun:"description" json:"description"`
	CreatedAt   time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// 用户角色关联
type UserRole struct {
	bun.BaseModel `bun:"table:sys_user_roles,alias:sur"`

	ID        string     `bun:"id,pk,type:varchar(32)" json:"id"`
	UserID    string     `bun:"user_id,notnull,type:varchar(32)" json:"user_id"`
	RoleID    string     `bun:"role_id,notnull,type:varchar(32)" json:"role_id"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
