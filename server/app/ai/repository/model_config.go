package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	ai_db "sag-wiki/app/ai/models/db"
)

// 模型配置仓储接口
type ModelConfigRepository interface {
	FindOne(ctx context.Context, id string) (*ai_db.ModelConfig, error)
	FindDefault(ctx context.Context) (*ai_db.ModelConfig, error)
	FindDefaultByType(ctx context.Context, modelType string) (*ai_db.ModelConfig, error)
	SetDefault(ctx context.Context, id string) error
	FindList(ctx context.Context) ([]*ai_db.ModelConfig, error)
	Create(ctx context.Context, config *ai_db.ModelConfig) error
	Update(ctx context.Context, config *ai_db.ModelConfig) error
	Delete(ctx context.Context, id string) error
}

// 模型配置仓储
type modelConfigRepository struct {
	db *bun.DB
}

func NewModelConfigRepository(database *bun.DB) ModelConfigRepository {
	return &modelConfigRepository{db: database}
}

// 根据 ID 获取模型配置
func (r *modelConfigRepository) FindOne(ctx context.Context, id string) (*ai_db.ModelConfig, error) {
	var config ai_db.ModelConfig

	err := r.db.NewSelect().
		Model(&config).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("根据 ID 获取模型配置失败: %w", err)
	}

	return &config, nil
}

// 获取默认模型配置
func (r *modelConfigRepository) FindDefault(ctx context.Context) (*ai_db.ModelConfig, error) {
	var config ai_db.ModelConfig

	err := r.db.NewSelect().
		Model(&config).
		Where("is_default = ?", true).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("获取默认模型配置失败: %w", err)
	}

	return &config, nil
}

// FindDefaultByType 获取指定类型的默认模型配置
func (r *modelConfigRepository) FindDefaultByType(ctx context.Context, modelType string) (*ai_db.ModelConfig, error) {
	var config ai_db.ModelConfig

	err := r.db.NewSelect().
		Model(&config).
		Where("is_default = ?", true).
		Where("model_type = ?", modelType).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("获取默认模型配置失败: %w", err)
	}

	return &config, nil
}

// 设置默认模型
func (r *modelConfigRepository) SetDefault(ctx context.Context, id string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 1. 先获取要设置的模型类型
		var config ai_db.ModelConfig
		err := tx.NewSelect().Model(&config).Where("id = ?", id).Scan(ctx)
		if err != nil {
			return fmt.Errorf("获取模型配置失败: %w", err)
		}

		// 2. 只把同类型的其他模型 is_default 设为 false
		_, err = tx.NewUpdate().
			Model((*ai_db.ModelConfig)(nil)).
			Set("is_default = ?", false).
			Where("model_type = ?", config.ModelType).
			Where("id != ?", id).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("设置默认模型失败: %w", err)
		}

		// 3. 把指定模型的 is_default 设为 true
		_, err = tx.NewUpdate().
			Model((*ai_db.ModelConfig)(nil)).
			Set("is_default = ?", true).
			Where("id = ?", id).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("设置默认模型失败: %w", err)
		}

		return nil
	})
}

// 获取模型配置列表
func (r *modelConfigRepository) FindList(ctx context.Context) ([]*ai_db.ModelConfig, error) {
	var configs []*ai_db.ModelConfig

	err := r.db.NewSelect().
		Model(&configs).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("获取模型配置列表失败: %w", err)
	}

	return configs, nil
}

// 创建模型配置
func (r *modelConfigRepository) Create(ctx context.Context, config *ai_db.ModelConfig) error {
	config.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(config).Exec(ctx)

	return err
}

// 更新模型配置
func (r *modelConfigRepository) Update(ctx context.Context, config *ai_db.ModelConfig) error {
	_, err := r.db.NewUpdate().
		Model(config).
		Where("id = ?", config.ID).
		Exec(ctx)

	return err
}

// 删除模型配置
func (r *modelConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*ai_db.ModelConfig)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}
