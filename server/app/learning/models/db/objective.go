package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 学习小节状态(来自 Wiki 文档)
type LearningObjective struct {
	bun.BaseModel `bun:"table:learning_objectives,alias:lo"`

	ID               string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	UserID           string    `bun:"user_id,notnull" json:"user_id"`                                          // 用户ID
	StageTitle       *string   `bun:"stage_title" json:"stage_title,omitempty"`                                // 所属阶段标题
	Title            string    `bun:"title,notnull" json:"title"`                                              // 小目标标题
	Detail           *string   `bun:"detail" json:"detail,omitempty"`                                          // 详细说明
	SourceDocumentID *string   `bun:"source_document_id" json:"source_document_id,omitempty"`                  // 来源文档ID
	SourceFolderID   *string   `bun:"source_folder_id" json:"source_folder_id,omitempty"`                      // 来源文件夹ID
	SourceFolderPath *string   `bun:"source_folder_path" json:"source_folder_path,omitempty"`                  // 来源文件夹路径
	OrderIndex       int       `bun:"order_index,notnull" json:"order_index"`                                  // 排序索引
	Status           string    `bun:"status,notnull" json:"status"`                                            // pending / active / completed / review
	MasteryLevel     string    `bun:"mastery_level,notnull" json:"mastery_level"`                              // none→seen→heard→explained→written→verified
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt        time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
