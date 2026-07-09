package db

import (
	"time"

	"github.com/uptrace/bun"
)

// WikiChunk RAG 文档切块模型
type WikiChunk struct {
	bun.BaseModel `bun:"table:rag_wiki_chunks,alias:rwc"`

	ID             string    `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 主键ID
	RootFolderID   string    `bun:"root_folder_id,type:varchar(32),notnull" json:"root_folder_id"`           // 知识库根文件夹ID
	FolderID       string    `bun:"folder_id,type:varchar(32),notnull" json:"folder_id"`                     // 所属文件夹ID
	DocumentID     string    `bun:"document_id,type:varchar(32),notnull" json:"document_id"`                 // 文档ID
	DocumentTitle  string    `bun:"document_title,notnull" json:"document_title"`                            // 文档标题
	FolderPath     string    `bun:"folder_path,notnull" json:"folder_path"`                                  // 文件夹路径
	HeadingPath    string    `bun:"heading_path,notnull" json:"heading_path"`                                // 标题路径
	ChunkIndex     int       `bun:"chunk_index,notnull" json:"chunk_index"`                                  // 切块序号
	BlockType      string    `bun:"block_type,notnull" json:"block_type"`                                    // 内容块类型
	Content        string    `bun:"content,notnull" json:"content"`                                          // 切块内容
	ContentHash    string    `bun:"content_hash,notnull" json:"content_hash"`                                // 内容哈希
	TokenCount     int       `bun:"token_count,notnull" json:"token_count"`                                  // 估算 token 数
	VectorPointID  string    `bun:"vector_point_id,notnull" json:"vector_point_id"`                          // 向量库 point ID
	EmbeddingModel string    `bun:"embedding_model,notnull" json:"embedding_model"`                          // Embedding 模型
	IndexedAt      time.Time `bun:"indexed_at,nullzero,notnull,default:current_timestamp" json:"indexed_at"` // 索引时间
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
	UpdatedAt      time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"` // 更新时间
}
