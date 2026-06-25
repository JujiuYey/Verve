package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	wiki_db "sag-wiki/app/wiki/models/db"
)

// 文件夹权限仓储接口
type FolderPermissionRepository interface {
	FindByFolderID(ctx context.Context, folderID string) ([]*wiki_db.FolderPermission, error)
	FindByUserID(ctx context.Context, userID string) ([]*wiki_db.FolderPermission, error)
	FindOne(ctx context.Context, folderID, userID string) (*wiki_db.FolderPermission, error)
	Create(ctx context.Context, permission *wiki_db.FolderPermission) error
	Delete(ctx context.Context, folderID, userID string) error
	DeleteByFolderID(ctx context.Context, folderID string) error
	CheckPermission(ctx context.Context, folderID, userID string) (bool, error)
	BatchCreate(ctx context.Context, permissions []*wiki_db.FolderPermission) error
	BatchCheckPermissions(ctx context.Context, folderIDs []string, userID string) (map[string]bool, error)
}

type folderPermissionRepository struct {
	db *bun.DB
}

// 创建文件夹权限仓储实例
func NewFolderPermissionRepository(db *bun.DB) FolderPermissionRepository {
	return &folderPermissionRepository{db: db}
}

// 根据文件夹ID获取所有权限
func (r *folderPermissionRepository) FindByFolderID(ctx context.Context, folderID string) ([]*wiki_db.FolderPermission, error) {
	var permissions []*wiki_db.FolderPermission
	err := r.db.NewSelect().
		Model(&permissions).
		Where("folder_id = ?", folderID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// 根据用户ID获取所有权限
func (r *folderPermissionRepository) FindByUserID(ctx context.Context, userID string) ([]*wiki_db.FolderPermission, error) {
	var permissions []*wiki_db.FolderPermission
	err := r.db.NewSelect().
		Model(&permissions).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// 根据文件夹ID和用户ID获取权限
func (r *folderPermissionRepository) FindOne(ctx context.Context, folderID, userID string) (*wiki_db.FolderPermission, error) {
	permission := new(wiki_db.FolderPermission)
	err := r.db.NewSelect().
		Model(permission).
		Where("folder_id = ?", folderID).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

// 创建权限
func (r *folderPermissionRepository) Create(ctx context.Context, permission *wiki_db.FolderPermission) error {
	permission.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := r.db.NewInsert().Model(permission).Exec(ctx)
	return err
}

// 删除权限
func (r *folderPermissionRepository) Delete(ctx context.Context, folderID, userID string) error {
	_, err := r.db.NewDelete().
		Model((*wiki_db.FolderPermission)(nil)).
		Where("folder_id = ?", folderID).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

// 删除文件夹的所有权限
func (r *folderPermissionRepository) DeleteByFolderID(ctx context.Context, folderID string) error {
	_, err := r.db.NewDelete().
		Model((*wiki_db.FolderPermission)(nil)).
		Where("folder_id = ?", folderID).
		Exec(ctx)
	return err
}

// 检查用户是否有权限访问文件夹
func (r *folderPermissionRepository) CheckPermission(ctx context.Context, folderID, userID string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*wiki_db.FolderPermission)(nil)).
		Where("folder_id = ?", folderID).
		Where("user_id = ?", userID).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 批量创建权限
func (r *folderPermissionRepository) BatchCreate(ctx context.Context, permissions []*wiki_db.FolderPermission) error {
	if len(permissions) == 0 {
		return nil
	}

	for _, permission := range permissions {
		permission.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	_, err := r.db.NewInsert().Model(&permissions).Exec(ctx)
	return err
}

// 批量检查用户对多个文件夹的权限
func (r *folderPermissionRepository) BatchCheckPermissions(ctx context.Context, folderIDs []string, userID string) (map[string]bool, error) {
	result := make(map[string]bool)

	if len(folderIDs) == 0 {
		return result, nil
	}

	// 查询用户在所有指定文件夹的权限记录
	var permissions []*wiki_db.FolderPermission
	err := r.db.NewSelect().
		Model(&permissions).
		Column("folder_id").
		Where("user_id = ?", userID).
		Where("folder_id IN (?)", bun.In(folderIDs)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 初始化所有文件夹为无权限
	for _, folderID := range folderIDs {
		result[folderID] = false
	}

	// 标记有权限的文件夹
	for _, p := range permissions {
		result[p.FolderID] = true
	}

	return result, nil
}
