package db

import (
	"time"

	"github.com/uptrace/bun"
)

type IndexBatch struct {
	bun.BaseModel `bun:"table:rag_index_batches,alias:rib"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`
	RootFolderID string     `bun:"root_folder_id,type:varchar(32),notnull" json:"root_folder_id"`
	Status       string     `bun:"status,notnull" json:"status"`
	TotalCount   int        `bun:"total_count,notnull" json:"total_count"`
	ErrorMessage *string    `bun:"error_message" json:"error_message"`
	StartedAt    *time.Time `bun:"started_at" json:"started_at"`
	FinishedAt   *time.Time `bun:"finished_at" json:"finished_at"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
