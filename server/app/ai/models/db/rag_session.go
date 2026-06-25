package db

import (
	"time"

	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
	wiki_db "sag-wiki/app/wiki/models/db"
)

// 对话会话模型
type RagSession struct {
	bun.BaseModel `bun:"table:ai_rag_sessions,alias:rs"`

	ID           string    `bun:"id,pk,type:varchar(32)" json:"id"`
	Title        *string   `bun:"title" json:"title,omitempty"`
	UserID       *string   `bun:"user_id" json:"user_id,omitempty"`
	FolderID     string    `bun:"folder_id,notnull" json:"folder_id"`
	DocumentID   *string   `bun:"document_id" json:"document_id,omitempty"`
	MessageCount int       `bun:"message_count" json:"message_count"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// 关联关系
	Messages []*RagMessage     `bun:"rel:has-many,join:id=session_id" json:"messages,omitempty"`
	User     *system_db.User   `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Folder   *wiki_db.Folder   `bun:"rel:belongs-to,join:folder_id=id" json:"folder,omitempty"`
	Document *wiki_db.Document `bun:"rel:belongs-to,join:document_id=id" json:"document,omitempty"`
}
