package repository

import (
	"context"
	"strings"
	"time"

	system_db "verve/app/system/models/db"
)

// apiKeyHint 根据 API Key 生成脱敏提示,空值返回 nil
func apiKeyHint(apiKey string) *string {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil
	}
	if len(key) <= 8 {
		hint := "已保存"
		return &hint
	}
	hint := key[:2] + "****" + key[len(key)-4:]
	return &hint
}

// FindPlatforms 查询所有平台,按 sort_order 升序、created_at 降序排列
func (r *modelConfigRepository) FindPlatforms(ctx context.Context) ([]*system_db.SysModelPlatform, error) {
	var platforms []*system_db.SysModelPlatform
	if err := r.db.NewSelect().Model(&platforms).Order("sort_order ASC, created_at DESC").Scan(ctx); err != nil {
		return nil, err
	}
	for _, platform := range platforms {
		platform.APIKeyHint = apiKeyHint(platform.APIKeyCiphertext)
	}
	return platforms, nil
}

// FindPlatform 根据 ID 查询单个平台
func (r *modelConfigRepository) FindPlatform(ctx context.Context, id string) (*system_db.SysModelPlatform, error) {
	platform := new(system_db.SysModelPlatform)
	if err := r.db.NewSelect().Model(platform).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	platform.APIKeyHint = apiKeyHint(platform.APIKeyCiphertext)
	return platform, nil
}

// CreatePlatform 创建平台,自动填充 ID、默认 ModelListPath、AuthScheme、Enabled
func (r *modelConfigRepository) CreatePlatform(ctx context.Context, platform *system_db.SysModelPlatform) error {
	platform.ID = newID()
	platform.APIKeyHint = apiKeyHint(platform.APIKeyCiphertext)
	if platform.ModelListPath == "" {
		platform.ModelListPath = "/models"
	}
	if platform.AuthScheme == "" {
		platform.AuthScheme = "bearer"
	}
	platform.Enabled = true
	_, err := r.db.NewInsert().Model(platform).Exec(ctx)
	return err
}

// UpdatePlatformConfig 更新平台 BaseURL 与 API Key,clearAPIKey=true 时清空密钥
func (r *modelConfigRepository) UpdatePlatformConfig(ctx context.Context, platformID string, baseURL string, apiKey *string, clearAPIKey bool) (*system_db.SysModelPlatform, error) {
	platform, err := r.FindPlatform(ctx, platformID)
	if err != nil {
		return nil, err
	}
	platform.BaseURL = baseURL
	if clearAPIKey {
		platform.APIKeyCiphertext = ""
	} else if apiKey != nil && strings.TrimSpace(*apiKey) != "" {
		platform.APIKeyCiphertext = strings.TrimSpace(*apiKey)
	}
	platform.APIKeyHint = apiKeyHint(platform.APIKeyCiphertext)
	if _, err := r.db.NewUpdate().Model(platform).WherePK().Exec(ctx); err != nil {
		return nil, err
	}
	return platform, nil
}

// UpdatePlatformLastModelSyncAt 更新平台最近一次同步模型时间
func (r *modelConfigRepository) UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*system_db.SysModelPlatform)(nil)).
		Set("last_model_sync_at = ?", syncedAt).
		Where("id = ?", platformID).
		Exec(ctx)
	return err
}

// DeletePlatform 删除平台
func (r *modelConfigRepository) DeletePlatform(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*system_db.SysModelPlatform)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
