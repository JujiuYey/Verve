package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 用户模型
type User struct {
	bun.BaseModel `bun:"table:sys_users,alias:su"`

	ID        string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	Username  string    `bun:"username,notnull" json:"username"`                                        // 用户名
	Email     string    `bun:"email,notnull" json:"email"`                                              // 邮箱
	Password  string    `bun:"password,notnull" json:"-"`                                               // 密码
	FullName  *string   `bun:"full_name" json:"full_name"`                                              // 全名
	Avatar    *string   `bun:"avatar" json:"avatar"`                                                    // 头像
	Status    string    `bun:"status,notnull,default:'active'" json:"status"`                           // 状态
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
