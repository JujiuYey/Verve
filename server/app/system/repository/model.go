package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	system_db "verve/app/system/models/db"
)

const (
	ModelTypeChat      = "chat"
	ModelTypeEmbedding = "embedding"
	ModelTypeRerank    = "rerank"

	modelStatusActive = "active"
)

// ModelConfigRepository 模型平台与已启用模型的数据访问接口
type ModelConfigRepository interface {
	// SysModelPlatform 平台 CRUD
	FindPlatforms(ctx context.Context) ([]*system_db.SysModelPlatform, error)
	FindPlatform(ctx context.Context, id string) (*system_db.SysModelPlatform, error)
	CreatePlatform(ctx context.Context, platform *system_db.SysModelPlatform) error
	UpdatePlatformConfig(ctx context.Context, platformID string, baseURL string, apiKey *string, clearAPIKey bool) (*system_db.SysModelPlatform, error)
	UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error
	DeletePlatform(ctx context.Context, id string) error

	// SysModel 模型 CRUD
	FindModels(ctx context.Context) ([]*system_db.SysModel, error)
	FindDefaultModelWithPlatform(ctx context.Context, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error)
	FindAgentModelConfigs(ctx context.Context) ([]*system_db.AgentModelConfig, error)
	FindAgentModelWithPlatform(ctx context.Context, agentKey, sceneKey, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error)
	UpsertAgentModelConfig(ctx context.Context, config *system_db.AgentModelConfig) (*system_db.AgentModelConfig, error)
	ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error)
	CreateModel(ctx context.Context, model *system_db.SysModel) error
	UpdateModel(ctx context.Context, modelID string, update ModelUpdate) (*system_db.SysModel, error)
	DeleteModel(ctx context.Context, id string) error
}

// ModelUpdate 模型可更新字段集合
type ModelUpdate struct {
	Status       *string
	DisplayName  *string
	Capabilities []string
}

// modelConfigRepository ModelConfigRepository 的实现
type modelConfigRepository struct {
	db *bun.DB
}

// NewModelConfigRepository 创建 ModelConfigRepository
func NewModelConfigRepository(database *bun.DB) ModelConfigRepository {
	return &modelConfigRepository{db: database}
}

// newID 生成 32 位无中划线 UUID
func newID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// FindModels 查询所有模型,按 created_at 降序排列
func (r *modelConfigRepository) FindModels(ctx context.Context) ([]*system_db.SysModel, error) {
	var models []*system_db.SysModel
	err := r.db.NewSelect().Model(&models).Order("created_at DESC").Scan(ctx)
	return models, err
}

func (r *modelConfigRepository) FindDefaultModelWithPlatform(ctx context.Context, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
	model := new(system_db.SysModel)
	err := r.db.NewSelect().
		Model(model).
		Where("model_type = ?", modelType).
		Where("status = ?", modelStatusActive).
		Where("is_default = ?", true).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, nil, err
	}
	platform, err := r.FindPlatform(ctx, model.PlatformID)
	if err != nil {
		return nil, nil, err
	}
	if !platform.Enabled {
		return nil, nil, fmt.Errorf("default %s model platform is disabled", modelType)
	}
	return model, platform, nil
}

func (r *modelConfigRepository) FindAgentModelConfigs(ctx context.Context) ([]*system_db.AgentModelConfig, error) {
	var configs []*system_db.AgentModelConfig
	err := r.db.NewSelect().
		Model(&configs).
		Relation("Model").
		Order("agent_key ASC", "scene_key ASC").
		Scan(ctx)
	return configs, err
}

func (r *modelConfigRepository) FindAgentModelWithPlatform(ctx context.Context, agentKey, sceneKey, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
	config := new(system_db.AgentModelConfig)
	err := r.db.NewSelect().
		Model(config).
		Relation("Model").
		Where("samc.agent_key = ?", strings.TrimSpace(agentKey)).
		Where("samc.scene_key = ?", strings.TrimSpace(sceneKey)).
		Where("samc.enabled = ?", true).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, nil, err
	}
	if config.Model == nil {
		return nil, nil, fmt.Errorf("agent model config %s.%s has no model", agentKey, sceneKey)
	}
	if config.Model.ModelType != modelType {
		return nil, nil, fmt.Errorf("agent model config %s.%s requires %s model, got %s", agentKey, sceneKey, modelType, config.Model.ModelType)
	}
	if config.Model.Status != modelStatusActive {
		return nil, nil, fmt.Errorf("agent model config %s.%s model is inactive", agentKey, sceneKey)
	}
	platform, err := r.FindPlatform(ctx, config.Model.PlatformID)
	if err != nil {
		return nil, nil, err
	}
	if !platform.Enabled {
		return nil, nil, fmt.Errorf("agent model config %s.%s platform is disabled", agentKey, sceneKey)
	}
	return config.Model, platform, nil
}

func (r *modelConfigRepository) UpsertAgentModelConfig(ctx context.Context, config *system_db.AgentModelConfig) (*system_db.AgentModelConfig, error) {
	config.AgentKey = strings.TrimSpace(config.AgentKey)
	config.SceneKey = strings.TrimSpace(config.SceneKey)
	config.ModelID = strings.TrimSpace(config.ModelID)
	if config.ID == "" {
		config.ID = newID()
	}
	if config.Params == nil {
		config.Params = map[string]interface{}{}
	}
	_, err := r.db.NewInsert().
		Model(config).
		On("CONFLICT (agent_key, scene_key) DO UPDATE").
		Set("model_id = EXCLUDED.model_id").
		Set("params = EXCLUDED.params").
		Set("enabled = EXCLUDED.enabled").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	updated := new(system_db.AgentModelConfig)
	err = r.db.NewSelect().
		Model(updated).
		Relation("Model").
		Where("agent_key = ?", config.AgentKey).
		Where("scene_key = ?", config.SceneKey).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return updated, nil
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
