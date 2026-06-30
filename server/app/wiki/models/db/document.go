package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 文档模型
type Document struct {
	bun.BaseModel `bun:"table:wiki_documents,alias:d"`

	ID              string  `bun:"id,pk,type:varchar(32)" json:"id"`
	Filename        string  `bun:"filename,notnull" json:"filename"`
	FileSize        int64   `bun:"file_size,notnull" json:"file_size"`
	ContentType     string  `bun:"content_type" json:"content_type"`
	FolderID        string  `bun:"folder_id,type:varchar(32)" json:"folder_id"`
	FilePath        string  `bun:"file_path" json:"file_path"` // 中的文件路径
	CreatedAt       time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
