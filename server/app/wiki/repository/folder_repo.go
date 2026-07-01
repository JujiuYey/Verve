package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	wiki_db "verve/app/wiki/models/db"
)

// 文件夹仓储接口
type FolderRepository interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Folder, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*wiki_db.Folder, error)
	GetAll(ctx context.Context) ([]*wiki_db.Folder, error)
	Page(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*wiki_db.Folder, int, error)
	Create(ctx context.Context, folder *wiki_db.Folder) error
	Update(ctx context.Context, folder *wiki_db.Folder) error
	Delete(ctx context.Context, id string) error
	GetAllSubFolderIDs(ctx context.Context, parentID string) ([]string, error)
	GetDB() *bun.DB
}

type folderRepository struct {
	db *bun.DB
}

// 创建文件夹仓储实例
func NewFolderRepository(db *bun.DB) FolderRepository {
	return &folderRepository{db: db}
}

// 获取数据库连接
func (r *folderRepository) GetDB() *bun.DB {
	return r.db
}

// 根据ID获取文件夹
func (r *folderRepository) FindOne(ctx context.Context, id string) (*wiki_db.Folder, error) {
	folder := new(wiki_db.Folder)
	err := r.db.NewSelect().
		Model(folder).
		Where("f.id = ?", id).
		Relation("CreatedByUser").
		Relation("UpdatedByUser").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return folder, nil
}

// 获取所有文件夹列表（不分页）
func (r *folderRepository) List(ctx context.Context, filters map[string]interface{}) ([]*wiki_db.Folder, error) {
	var folders []*wiki_db.Folder

	query := r.db.NewSelect().
		Model(&folders).
		Relation("CreatedByUser").
		Relation("UpdatedByUser")

	if parentID, ok := filters["parent_id"].(string); ok && parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}
	if _, ok := filters["root"]; ok {
		query = query.Where("parent_id IS NULL")
	}
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	query = query.Order("created_at ASC")

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return folders, nil
}

// 获取所有文件夹（用于构建树形结构）
func (r *folderRepository) GetAll(ctx context.Context) ([]*wiki_db.Folder, error) {
	var folders []*wiki_db.Folder

	query := r.db.NewSelect().
		Model(&folders).
		Relation("CreatedByUser").
		Relation("UpdatedByUser").
		Order("created_at ASC")

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return folders, nil
}

// 获取文件夹列表（分页）
func (r *folderRepository) Page(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*wiki_db.Folder, int, error) {
	var folders []*wiki_db.Folder

	query := r.db.NewSelect().
		Model(&folders).
		Relation("CreatedByUser").
		Relation("UpdatedByUser")

	if parentID, ok := filters["parent_id"].(string); ok && parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}
	if _, ok := filters["root"]; ok {
		query = query.Where("parent_id IS NULL")
	}
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	query = query.Order("created_at ASC")

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return folders, total, nil
}

// 创建文件夹
func (r *folderRepository) Create(ctx context.Context, folder *wiki_db.Folder) error {
	folder.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(folder).Exec(ctx)
	return err
}

// 更新文件夹
func (r *folderRepository) Update(ctx context.Context, folder *wiki_db.Folder) error {
	_, err := r.db.NewUpdate().
		Model(folder).
		Where("id = ?", folder.ID).
		Exec(ctx)
	return err
}

// 删除文件夹
func (r *folderRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*wiki_db.Folder)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// 递归获取所有子文件夹ID（包括自身）
func (r *folderRepository) GetAllSubFolderIDs(ctx context.Context, parentID string) ([]string, error) {
	query := `
		WITH RECURSIVE folder_tree AS (
			SELECT id FROM wiki_folders WHERE id = ?
			UNION ALL
			SELECT f.id FROM wiki_folders f
			INNER JOIN folder_tree ft ON f.parent_id = ft.id
		)
		SELECT id FROM folder_tree
	`

	var folderIDs []string
	err := r.db.NewRaw(query, parentID).Scan(ctx, &folderIDs)
	if err != nil {
		return nil, err
	}

	return folderIDs, nil
}
