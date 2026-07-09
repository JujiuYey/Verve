package db

import (
	"time"

	"github.com/uptrace/bun"
)

type WikiChunk struct {
	bun.BaseModel `bun:"table:rag_wiki_chunks,alias:rwc"`

	ID             string    `bun:"id,pk,type:varchar(32)" json:"id"`
	RootFolderID   string    `bun:"root_folder_id,type:varchar(32),notnull" json:"root_folder_id"`
	FolderID       string    `bun:"folder_id,type:varchar(32),notnull" json:"folder_id"`
	DocumentID     string    `bun:"document_id,type:varchar(32),notnull" json:"document_id"`
	DocumentTitle  string    `bun:"document_title,notnull" json:"document_title"`
	FolderPath     string    `bun:"folder_path,notnull" json:"folder_path"`
	HeadingPath    string    `bun:"heading_path,notnull" json:"heading_path"`
	ChunkIndex     int       `bun:"chunk_index,notnull" json:"chunk_index"`
	BlockType      string    `bun:"block_type,notnull" json:"block_type"`
	Content        string    `bun:"content,notnull" json:"content"`
	ContentHash    string    `bun:"content_hash,notnull" json:"content_hash"`
	TokenCount     int       `bun:"token_count,notnull" json:"token_count"`
	VectorPointID  string    `bun:"vector_point_id,notnull" json:"vector_point_id"`
	EmbeddingModel string    `bun:"embedding_model,notnull" json:"embedding_model"`
	IndexedAt      time.Time `bun:"indexed_at,nullzero,notnull,default:current_timestamp" json:"indexed_at"`
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
