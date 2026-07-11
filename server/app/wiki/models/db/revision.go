package db

import (
	"time"

	"github.com/uptrace/bun"
)

// DocumentRevision 文档不可变修订模型
type DocumentRevision struct {
	bun.BaseModel `bun:"table:wiki_document_revisions,alias:wdr"`

	ID              string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	DocumentID      string    `bun:"document_id,type:varchar(32),notnull" json:"document_id"`                 // 文档ID
	Version         int64     `bun:"version,notnull" json:"version"`                                          // 文档版本号
	ObjectPath      string    `bun:"object_path,notnull" json:"object_path"`                                  // 不可变对象路径
	ContentHash     string    `bun:"content_hash,notnull" json:"content_hash"`                                // 内容哈希
	FileSize        int64     `bun:"file_size,notnull" json:"file_size"`                                      // 文件大小(字节)
	ChangeRequestID *string   `bun:"change_request_id,type:varchar(32)" json:"change_request_id"`             // 变更申请ID
	ChangedBy       string    `bun:"changed_by,type:varchar(32),notnull" json:"changed_by"`                   // 修改用户ID
	ChangeSummary   string    `bun:"change_summary,notnull" json:"change_summary"`                            // 修改摘要
	CreatedAt       time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
}
