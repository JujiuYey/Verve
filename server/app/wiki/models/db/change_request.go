package db

import (
	"time"

	"github.com/uptrace/bun"
)

const (
	ChangeRequestStatusProposed  = "proposed"
	ChangeRequestStatusApplied   = "applied"
	ChangeRequestStatusFailed    = "failed"
	ChangeRequestStatusCancelled = "cancelled"
	ChangeRequestStatusConflict  = "conflict"
)

// DocumentChangeRequest 文档变更申请模型
type DocumentChangeRequest struct {
	bun.BaseModel `bun:"table:wiki_document_change_requests,alias:wdcr"`

	ID                      string     `bun:"id,pk,type:varchar(32)" json:"id"`                                              // 主键ID
	DocumentID              string     `bun:"document_id,type:varchar(32),notnull" json:"document_id"`                       // 文档ID
	SourceType              string     `bun:"source_type,notnull" json:"source_type"`                                        // 来源类型
	SourceID                string     `bun:"source_id,notnull" json:"source_id"`                                            // 来源记录ID
	RequestID               string     `bun:"request_id,notnull" json:"request_id"`                                          // 幂等请求ID
	ReplacesChangeRequestID *string    `bun:"replaces_change_request_id,type:varchar(32)" json:"replaces_change_request_id"` // 替代的变更申请ID
	BaseVersion             int64      `bun:"base_version,notnull" json:"base_version"`                                      // 基准文档版本号
	Instruction             string     `bun:"instruction,notnull" json:"instruction"`                                        // 用户修改要求
	ChangeSummary           string     `bun:"change_summary,notnull" json:"change_summary"`                                  // 修改摘要
	ProposedContent         string     `bun:"proposed_content,notnull" json:"proposed_content"`                              // 建议完整内容
	ProposedDiff            string     `bun:"proposed_diff,notnull" json:"proposed_diff"`                                    // 建议内容差异
	Status                  string     `bun:"status,notnull" json:"status"`                                                  // 申请状态
	ErrorMessage            *string    `bun:"error_message" json:"error_message"`                                            // 失败原因
	AppliedVersion          *int64     `bun:"applied_version" json:"applied_version"`                                        // 已应用版本号
	CreatedAt               time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`       // 创建时间
	UpdatedAt               time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`       // 更新时间
	AppliedAt               *time.Time `bun:"applied_at" json:"applied_at"`                                                  // 应用时间
}
