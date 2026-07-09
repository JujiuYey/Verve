package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	wiki_db "verve/app/wiki/models/db"
)

// 文档仓储
type DocumentRepository struct {
	db *bun.DB
}

// 创建文档仓储
func NewDocumentRepository(db *bun.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// 创建文档记录（使用指定的 UUID）
func (r *DocumentRepository) Create(ctx context.Context, folderID string, filename string, fileSize int64, filePath string) (*wiki_db.Document, error) {
	doc := &wiki_db.Document{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		Filename:    filename,
		FileSize:    fileSize,
		ContentType: "text/markdown",
		FolderID:    folderID,
		FilePath:    filePath,
	}

	_, err := r.db.NewInsert().Model(doc).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建文档记录失败: %w", err)
	}

	return doc, nil
}

// 更新文件大小
func (r *DocumentRepository) UpdateFileSize(ctx context.Context, docID string, fileSize int64) error {
	_, err := r.db.NewUpdate().
		Model((*wiki_db.Document)(nil)).
		Set("file_size = ?", fileSize).
		Where("id = ?", docID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("更新文件大小失败: %w", err)
	}

	return nil
}

// 根据 ID 获取文档
func (r *DocumentRepository) FindOne(ctx context.Context, id string) (*wiki_db.Document, error) {
	doc := new(wiki_db.Document)
	err := r.db.NewSelect().
		Model(doc).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %w", err)
	}

	return doc, nil
}

// 列出所有文档（支持分页和查询）
func (r *DocumentRepository) Page(ctx context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error) {
	var docs []*wiki_db.Document

	query := r.db.NewSelect().
		Model(&docs)

	// 根据名称查询
	if name != "" {
		query = query.Where("filename LIKE ?", "%"+name+"%")
	}

	// 根据知识库 ID 查询
	if folderID != "" {
		query = query.Where("folder_id = ?", folderID)
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取文档总数失败: %w", err)
	}

	// 获取分页数据
	err = query.
		Order("sort_order ASC", "created_at ASC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("列出文档失败: %w", err)
	}

	return docs, total, nil
}

// 列出所有文档（不分页）
func (r *DocumentRepository) List(ctx context.Context, name, folderID string) ([]*wiki_db.Document, error) {
	var docs []*wiki_db.Document

	query := r.db.NewSelect().
		Model(&docs)

	if name != "" {
		query = query.Where("filename LIKE ?", "%"+name+"%")
	}

	if folderID != "" {
		query = query.Where("folder_id = ?", folderID)
	}

	err := query.
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("列出文档失败: %w", err)
	}

	return docs, nil
}

// 删除文档
func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*wiki_db.Document)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}

	return nil
}

// GetDocumentsByFolderIDs 根据文件夹ID列表获取所有文档
func (r *DocumentRepository) GetDocumentsByFolderIDs(ctx context.Context, folderIDs []string) ([]*wiki_db.Document, error) {
	if len(folderIDs) == 0 {
		return []*wiki_db.Document{}, nil
	}

	var docs []*wiki_db.Document
	err := r.db.NewSelect().
		Model(&docs).
		Where("folder_id IN (?)", bun.In(folderIDs)).
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取文档列表失败: %w", err)
	}

	return docs, nil
}
