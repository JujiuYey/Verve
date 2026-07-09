package db

import (
	"time"

	"github.com/uptrace/bun"
)

type IndexJob struct {
	bun.BaseModel `bun:"table:rag_index_jobs,alias:rij"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`
	DocumentID   string     `bun:"document_id,type:varchar(32),notnull" json:"document_id"`
	RootFolderID *string    `bun:"root_folder_id,type:varchar(32)" json:"root_folder_id"`
	Status       string     `bun:"status,notnull" json:"status"`
	ErrorMessage *string    `bun:"error_message" json:"error_message"`
	ChunkCount   int        `bun:"chunk_count,notnull" json:"chunk_count"`
	StartedAt    *time.Time `bun:"started_at" json:"started_at"`
	FinishedAt   *time.Time `bun:"finished_at" json:"finished_at"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
