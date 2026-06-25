package repository

import (
	"context"
)

type FolderExpander struct {
	folderRepo     FolderRepository
	permissionRepo FolderPermissionRepository
}

// NewFolderExpander 创建 FolderExpander
func NewFolderExpander(folderRepo FolderRepository, permissionRepo FolderPermissionRepository) *FolderExpander {
	return &FolderExpander{
		folderRepo:     folderRepo,
		permissionRepo: permissionRepo,
	}
}

// ExpandAndFilter 获取用户有权限访问的文件夹列表（递归展开子文件夹）
func (e *FolderExpander) ExpandAndFilter(ctx context.Context, folderID string, userID string) ([]string, error) {
	// 1. 调用 GetAllSubFolderIDs 展开文件夹树
	allFolderIDs, err := e.folderRepo.GetAllSubFolderIDs(ctx, folderID)
	if err != nil {
		return nil, err
	}

	// 2. 调用 BatchCheckPermissions 批量检查权限
	permissionMap, err := e.permissionRepo.BatchCheckPermissions(ctx, allFolderIDs, userID)
	if err != nil {
		return nil, err
	}

	// 3. 返回用户有权限的所有文件夹 ID
	var result []string
	for _, fid := range allFolderIDs {
		if permissionMap[fid] {
			result = append(result, fid)
		}
	}

	return result, nil
}