package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 部门模型
type Department struct {
	bun.BaseModel `bun:"table:sys_departments,alias:sd"`

	ID          string     `bun:"id,pk,type:varchar(32)" json:"id"`
	Name        string     `bun:"name,notnull" json:"name"`
	Description *string    `bun:"description" json:"description"`
	ParentID    *string    `bun:"parent_id,type:varchar(32)" json:"parent_id"`
	CreatedAt   time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	Children    []*Department `bun:"-" json:"children,omitempty"`
}

// 用户部门关联
type UserDepartment struct {
	bun.BaseModel `bun:"table:sys_user_departments,alias:ud"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`
	UserID       string     `bun:"user_id,notnull,type:varchar(32)" json:"user_id"`
	DepartmentID string     `bun:"department_id,notnull,type:varchar(32)" json:"department_id"`
	IsPrimary    bool       `bun:"is_primary,default:false" json:"is_primary"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
