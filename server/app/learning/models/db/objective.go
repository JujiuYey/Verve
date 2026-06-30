package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 小目标(进度与掌握度最小单位)
type LearningObjective struct {
	bun.BaseModel `bun:"table:learning_objectives,alias:lo"`

	ID               string    `bun:"id,pk,type:varchar(32)" json:"id"`
	PathID           string    `bun:"path_id,notnull" json:"path_id"`
	UserID           string    `bun:"user_id,notnull" json:"user_id"`
	StageTitle       *string   `bun:"stage_title" json:"stage_title,omitempty"`
	Title            string    `bun:"title,notnull" json:"title"`
	Detail           *string   `bun:"detail" json:"detail,omitempty"`
	SourceDocumentID *string   `bun:"source_document_id" json:"source_document_id,omitempty"`
	SourceFolderID   *string   `bun:"source_folder_id" json:"source_folder_id,omitempty"`
	SourceFolderPath *string   `bun:"source_folder_path" json:"source_folder_path,omitempty"`
	OrderIndex       int       `bun:"order_index,notnull" json:"order_index"`
	Status           string    `bun:"status,notnull" json:"status"`               // pending / active / completed / review
	MasteryLevel     string    `bun:"mastery_level,notnull" json:"mastery_level"` // none→seen→heard→explained→written→verified
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
