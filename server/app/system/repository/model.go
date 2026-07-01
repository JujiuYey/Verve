package repository

import (
	"context"
	"strings"

	system_db "sag-wiki/app/system/models/db"
)

const (
	ModelTypeChat      = "chat"
	ModelTypeEmbedding = "embedding"
)

// FindModels 查询所有模型,按 created_at 降序排列
func (r *modelConfigRepository) FindModels(ctx context.Context) ([]*system_db.SysModel, error) {
	var models []*system_db.SysModel
	err := r.db.NewSelect().Model(&models).Order("created_at DESC").Scan(ctx)
	return models, err
}

// ModelExistsByPlatformAndName 判断指定平台下是否存在同名模型
func (r *modelConfigRepository) ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error) {
	count, err := r.db.NewSelect().Model((*system_db.SysModel)(nil)).
		Where("platform_id = ?", platformID).
		Where("model_name = ?", modelName).
		Count(ctx)
	return count > 0, err
}

// CreateModel 创建模型,自动填充 ID 与默认状态
func (r *modelConfigRepository) CreateModel(ctx context.Context, model *system_db.SysModel) error {
	model.ID = newID()
	if model.Status == "" {
		model.Status = modelStatusActive
	}
	_, err := r.db.NewInsert().Model(model).Exec(ctx)
	return err
}

// UpdateModel 按 ModelUpdate 选择性更新模型状态、显示名称、能力标签
func (r *modelConfigRepository) UpdateModel(ctx context.Context, modelID string, update ModelUpdate) (*system_db.SysModel, error) {
	var model system_db.SysModel
	if err := r.db.NewSelect().Model(&model).Where("id = ?", modelID).Scan(ctx); err != nil {
		return nil, err
	}
	if update.Status != nil {
		model.Status = *update.Status
	}
	if update.DisplayName != nil {
		model.DisplayName = strings.TrimSpace(*update.DisplayName)
	}
	if update.Capabilities != nil {
		model.Capabilities = update.Capabilities
	}
	if _, err := r.db.NewUpdate().Model(&model).WherePK().Exec(ctx); err != nil {
		return nil, err
	}
	return &model, nil
}

// DeleteModel 删除模型
func (r *modelConfigRepository) DeleteModel(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*system_db.SysModel)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
