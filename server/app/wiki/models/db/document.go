package db

import (
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
	FilePath        string  `bun:"file_path" json:"file_path"`   // 中的文件路径
	Status          string  `bun:"status,notnull" json:"status"` // pending, processing, completed, failed
	ChunkCount      int     `bun:"chunk_count" json:"chunk_count"`
	ErrorMessage    *string `bun:"error_message" json:"error_message,omitempty"`
}

// 文档状态常量
const (
	DocumentStatusPending    = "pending"
	DocumentStatusProcessing = "processing"
	DocumentStatusCompleted  = "completed"
	DocumentStatusFailed     = "failed"
)
