package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 文档模型
type Document struct {
	bun.BaseModel `bun:"table:wiki_documents,alias:d"`

	ID          string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	Filename    string    `bun:"filename,notnull" json:"filename"`                                        // 文件名
	FileSize    int64     `bun:"file_size,notnull" json:"file_size"`                                      // 文件大小(字节)
	ContentType string    `bun:"content_type" json:"content_type"`                                        // MIME 类型
	FolderID    string    `bun:"folder_id,type:varchar(32)" json:"folder_id"`                             // 所属文件夹ID
	FilePath    string    `bun:"file_path" json:"file_path"`                                              // 存储路径
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
