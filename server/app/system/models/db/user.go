package db


import (
	"time"

	"github.com/uptrace/bun"
)

// 用户模型
type User struct {
	bun.BaseModel `bun:"table:sys_users,alias:su"`

	ID        string     `bun:"id,pk,type:varchar(32)" json:"id"`
	Username  string     `bun:"username,notnull" json:"username"`
	Email     string     `bun:"email,notnull" json:"email"`
	Password  string     `bun:"password,notnull" json:"-"`
	FullName  *string    `bun:"full_name" json:"full_name"`
	Avatar    *string    `bun:"avatar" json:"avatar"`
	Status    string     `bun:"status,notnull,default:'active'" json:"status"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
