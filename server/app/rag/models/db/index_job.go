package db

import (
	"time"

	"github.com/uptrace/bun"
)

// IndexJob 单篇文档索引状态模型
type IndexJob struct {
	bun.BaseModel `bun:"table:rag_index_jobs,alias:rij"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	DocumentID   string     `bun:"document_id,type:varchar(32),notnull" json:"document_id"`                 // 文档ID
	RootFolderID *string    `bun:"root_folder_id,type:varchar(32)" json:"root_folder_id"`                   // 知识库根文件夹ID
	Status       string     `bun:"status,notnull" json:"status"`                                            // 索引状态
	ErrorMessage *string    `bun:"error_message" json:"error_message"`                                      // 错误信息
	ChunkCount   int        `bun:"chunk_count,notnull" json:"chunk_count"`                                  // 切块数量
	AttemptCount int        `bun:"attempt_count,notnull" json:"attempt_count"`                              // 尝试次数
	MaxAttempts  int        `bun:"max_attempts,notnull" json:"max_attempts"`                                // 最大尝试次数
	StartedAt    *time.Time `bun:"started_at" json:"started_at"`                                            // 开始时间
	FinishedAt   *time.Time `bun:"finished_at" json:"finished_at"`                                          // 结束时间
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
