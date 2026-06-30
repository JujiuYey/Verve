package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
)

const (
	modelStatusActive   = "active"
	modelStatusInactive = "inactive"
)

type ModelConfigRepository interface {
	FindOne(ctx context.Context, id string) (*system_db.ModelConfig, error)
	FindDefault(ctx context.Context) (*system_db.ModelConfig, error)
	FindDefaultByType(ctx context.Context, modelType string) (*system_db.ModelConfig, error)
	FindByModelName(ctx context.Context, modelName string) (*system_db.ModelConfig, error)
	SetDefault(ctx context.Context, id string) error
	FindList(ctx context.Context) ([]*system_db.ModelConfig, error)
	Create(ctx context.Context, config *system_db.ModelConfig) error
	Update(ctx context.Context, config *system_db.ModelConfig) error
	Delete(ctx context.Context, id string) error
	FindPlatforms(ctx context.Context) ([]*system_db.SysModelPlatform, error)
	FindPlatform(ctx context.Context, id string) (*system_db.SysModelPlatform, error)
	CreatePlatform(ctx context.Context, platform *system_db.SysModelPlatform) error
	UpdatePlatformConfig(ctx context.Context, platformID string, baseURL string, apiKey *string, clearAPIKey bool) (*system_db.SysModelPlatform, error)
	UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error
	DeletePlatform(ctx context.Context, id string) error
	FindModels(ctx context.Context) ([]*system_db.SysModel, error)
	ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error)
	CreateModel(ctx context.Context, model *system_db.SysModel) error
	UpdateModel(ctx context.Context, modelID string, update ModelUpdate) (*system_db.SysModel, error)
	DeleteModel(ctx context.Context, id string) error
}

type ModelUpdate struct {
	Status       *string
	DisplayName  *string
	Capabilities []string
}

type modelConfigRepository struct {
	db *bun.DB
}

func NewModelConfigRepository(database *bun.DB) ModelConfigRepository {
	return &modelConfigRepository{db: database}
}

func buildModelConfig(platform *system_db.SysModelPlatform, model *system_db.SysModel) *system_db.ModelConfig {
	baseURL := platform.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = platform.DefaultBaseURL
	}

	return &system_db.ModelConfig{
		ID:          model.ID,
		Vendor:      platform.Name,
		Name:        model.DisplayName,
		APIKey:      platform.APIKeyCiphertext,
		BaseURL:     baseURL,
		ModelType:   model.ModelType,
		Model:       model.ModelName,
		Temperature: model.Temperature,
		TopP:        model.TopP,
		MaxTokens:   model.MaxTokens,
		TopK:        model.TopK,
		IsActive:    model.Status == modelStatusActive,
		IsDefault:   model.IsDefault,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func newID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

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

type modelConfigRow struct {
	ModelID          string    `bun:"model_id"`
	ModelName        string    `bun:"model_name"`
	DisplayName      string    `bun:"display_name"`
	ModelType        string    `bun:"model_type"`
	Status           string    `bun:"status"`
	IsDefault        bool      `bun:"is_default"`
	Temperature      float32   `bun:"temperature"`
	TopP             float32   `bun:"top_p"`
	MaxTokens        *int64    `bun:"max_tokens"`
	TopK             *int64    `bun:"top_k"`
	CreatedAt        time.Time `bun:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at"`
	PlatformID       string    `bun:"platform_id"`
	PlatformName     string    `bun:"platform_name"`
	DefaultBaseURL   string    `bun:"default_base_url"`
	BaseURL          string    `bun:"base_url"`
	APIKeyCiphertext string    `bun:"api_key_ciphertext"`
}

func (r modelConfigRow) toModelConfig() *system_db.ModelConfig {
	return buildModelConfig(
		&system_db.SysModelPlatform{
			ID:               r.PlatformID,
			Name:             r.PlatformName,
			DefaultBaseURL:   r.DefaultBaseURL,
			BaseURL:          r.BaseURL,
			APIKeyCiphertext: r.APIKeyCiphertext,
		},
		&system_db.SysModel{
			ID:          r.ModelID,
			ModelName:   r.ModelName,
			DisplayName: r.DisplayName,
			ModelType:   r.ModelType,
			Status:      r.Status,
			IsDefault:   r.IsDefault,
			Temperature: r.Temperature,
			TopP:        r.TopP,
			MaxTokens:   r.MaxTokens,
			TopK:        r.TopK,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		},
	)
}

func (r *modelConfigRepository) findModelConfig(ctx context.Context, filter func(*bun.SelectQuery) *bun.SelectQuery) (*system_db.ModelConfig, error) {
	var row modelConfigRow

	query := r.db.NewSelect().
		TableExpr("sys_models AS sm").
		ColumnExpr("sm.id AS model_id").
		ColumnExpr("sm.model_name").
		ColumnExpr("sm.display_name").
		ColumnExpr("sm.model_type").
		ColumnExpr("sm.status").
		ColumnExpr("sm.is_default").
		ColumnExpr("sm.temperature").
		ColumnExpr("sm.top_p").
		ColumnExpr("sm.max_tokens").
		ColumnExpr("sm.top_k").
		ColumnExpr("sm.created_at").
		ColumnExpr("sm.updated_at").
		ColumnExpr("smp.id AS platform_id").
		ColumnExpr("smp.name AS platform_name").
		ColumnExpr("smp.default_base_url").
		ColumnExpr("smp.base_url").
		ColumnExpr("smp.api_key_ciphertext").
		Join("JOIN sys_model_platforms AS smp ON smp.id = sm.platform_id")
	query = filter(query)

	if err := query.Scan(ctx, &row); err != nil {
		return nil, err
	}

	return row.toModelConfig(), nil
}

func (r *modelConfigRepository) FindOne(ctx context.Context, id string) (*system_db.ModelConfig, error) {
	config, err := r.findModelConfig(ctx, func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Where("sm.id = ?", id)
	})
	if err != nil {
		return nil, fmt.Errorf("根据 ID 获取模型配置失败: %w", err)
	}
	return config, nil
}

func (r *modelConfigRepository) FindDefault(ctx context.Context) (*system_db.ModelConfig, error) {
	return r.FindDefaultByType(ctx, system_db.ModelTypeChat)
}

func (r *modelConfigRepository) FindDefaultByType(ctx context.Context, modelType string) (*system_db.ModelConfig, error) {
	config, err := r.findModelConfig(ctx, func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.
			Where("sm.is_default = ?", true).
			Where("sm.model_type = ?", modelType).
			Where("sm.status = ?", modelStatusActive).
			Where("smp.enabled = ?", true)
	})
	if err != nil {
		return nil, fmt.Errorf("获取默认模型配置失败: %w", err)
	}
	return config, nil
}

func (r *modelConfigRepository) FindByModelName(ctx context.Context, modelName string) (*system_db.ModelConfig, error) {
	config, err := r.findModelConfig(ctx, func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.
			Where("(sm.model_name = ? OR sm.display_name = ?)", modelName, modelName).
			Where("sm.model_type = ?", system_db.ModelTypeChat).
			Where("sm.status = ?", modelStatusActive).
			Where("smp.enabled = ?", true)
	})
	if err != nil {
		return nil, fmt.Errorf("根据模型名称获取模型配置失败: %w", err)
	}
	return config, nil
}

func (r *modelConfigRepository) SetDefault(ctx context.Context, id string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var model system_db.SysModel
		if err := tx.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx); err != nil {
			return fmt.Errorf("获取模型配置失败: %w", err)
		}

		if _, err := tx.NewUpdate().
			Model((*system_db.SysModel)(nil)).
			Set("is_default = ?", false).
			Where("model_type = ?", model.ModelType).
			Where("id != ?", id).
			Exec(ctx); err != nil {
			return fmt.Errorf("设置默认模型失败: %w", err)
		}

		if _, err := tx.NewUpdate().
			Model((*system_db.SysModel)(nil)).
			Set("is_default = ?", true).
			Where("id = ?", id).
			Exec(ctx); err != nil {
			return fmt.Errorf("设置默认模型失败: %w", err)
		}

		return nil
	})
}

func (r *modelConfigRepository) FindList(ctx context.Context) ([]*system_db.ModelConfig, error) {
	var models []*system_db.SysModel
	if err := r.db.NewSelect().Model(&models).Order("created_at DESC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("获取模型配置列表失败: %w", err)
	}

	platformIDs := make([]string, 0, len(models))
	for _, model := range models {
		platformIDs = append(platformIDs, model.PlatformID)
	}

	platforms := make([]*system_db.SysModelPlatform, 0)
	if len(platformIDs) > 0 {
		if err := r.db.NewSelect().Model(&platforms).Where("id IN (?)", bun.In(platformIDs)).Scan(ctx); err != nil {
			return nil, fmt.Errorf("获取模型平台失败: %w", err)
		}
	}

	platformByID := make(map[string]*system_db.SysModelPlatform, len(platforms))
	for _, platform := range platforms {
		platformByID[platform.ID] = platform
	}

	configs := make([]*system_db.ModelConfig, 0, len(models))
	for _, model := range models {
		if platform := platformByID[model.PlatformID]; platform != nil {
			configs = append(configs, buildModelConfig(platform, model))
		}
	}
	return configs, nil
}

func (r *modelConfigRepository) Create(ctx context.Context, config *system_db.ModelConfig) error {
	model := &system_db.SysModel{
		ID:          newID(),
		ModelName:   config.Model,
		DisplayName: config.Name,
		ModelType:   config.ModelType,
		Status:      modelStatusInactive,
		IsDefault:   config.IsDefault,
		Temperature: config.Temperature,
		TopP:        config.TopP,
		MaxTokens:   config.MaxTokens,
		TopK:        config.TopK,
	}
	platform := &system_db.SysModelPlatform{
		ID:               newID(),
		Name:             config.Vendor,
		DefaultBaseURL:   config.BaseURL,
		BaseURL:          config.BaseURL,
		APIKeyCiphertext: config.APIKey,
		Enabled:          config.IsActive,
		ProviderType:     "openai_compatible",
		ModelListPath:    "/models",
		AuthScheme:       "bearer",
	}
	model.PlatformID = platform.ID
	if config.IsActive {
		model.Status = modelStatusActive
	}

	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(platform).Exec(ctx); err != nil {
			return fmt.Errorf("创建模型平台失败: %w", err)
		}
		if config.IsDefault {
			if _, err := tx.NewUpdate().Model((*system_db.SysModel)(nil)).
				Set("is_default = ?", false).
				Where("model_type = ?", config.ModelType).
				Exec(ctx); err != nil {
				return fmt.Errorf("重置默认模型失败: %w", err)
			}
		}
		if _, err := tx.NewInsert().Model(model).Exec(ctx); err != nil {
			return fmt.Errorf("创建模型失败: %w", err)
		}
		config.ID = model.ID
		return nil
	})
}

func (r *modelConfigRepository) Update(ctx context.Context, config *system_db.ModelConfig) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var model system_db.SysModel
		if err := tx.NewSelect().Model(&model).Where("id = ?", config.ID).Scan(ctx); err != nil {
			return fmt.Errorf("获取模型失败: %w", err)
		}
		var platform system_db.SysModelPlatform
		if err := tx.NewSelect().Model(&platform).Where("id = ?", model.PlatformID).Scan(ctx); err != nil {
			return fmt.Errorf("获取模型平台失败: %w", err)
		}

		model.ModelName = config.Model
		model.DisplayName = config.Name
		model.ModelType = config.ModelType
		model.Temperature = config.Temperature
		model.TopP = config.TopP
		model.MaxTokens = config.MaxTokens
		model.TopK = config.TopK
		model.IsDefault = config.IsDefault
		if config.IsActive {
			model.Status = modelStatusActive
		} else {
			model.Status = modelStatusInactive
		}

		platform.Name = config.Vendor
		platform.BaseURL = config.BaseURL
		platform.DefaultBaseURL = config.BaseURL
		platform.Enabled = config.IsActive
		if strings.TrimSpace(config.APIKey) != "" {
			platform.APIKeyCiphertext = config.APIKey
		}

		if config.IsDefault {
			if _, err := tx.NewUpdate().Model((*system_db.SysModel)(nil)).
				Set("is_default = ?", false).
				Where("model_type = ?", config.ModelType).
				Where("id != ?", config.ID).
				Exec(ctx); err != nil {
				return fmt.Errorf("重置默认模型失败: %w", err)
			}
		}
		if _, err := tx.NewUpdate().Model(&platform).WherePK().Exec(ctx); err != nil {
			return fmt.Errorf("更新模型平台失败: %w", err)
		}
		if _, err := tx.NewUpdate().Model(&model).WherePK().Exec(ctx); err != nil {
			return fmt.Errorf("更新模型失败: %w", err)
		}
		return nil
	})
}

func (r *modelConfigRepository) Delete(ctx context.Context, id string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var model system_db.SysModel
		if err := tx.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx); err != nil {
			return fmt.Errorf("获取模型失败: %w", err)
		}
		if _, err := tx.NewDelete().Model((*system_db.SysModel)(nil)).Where("id = ?", id).Exec(ctx); err != nil {
			return fmt.Errorf("删除模型失败: %w", err)
		}
		if _, err := tx.NewDelete().Model((*system_db.SysModelPlatform)(nil)).Where("id = ?", model.PlatformID).Exec(ctx); err != nil {
			return fmt.Errorf("删除模型平台失败: %w", err)
		}
		return nil
	})
}

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

func (r *modelConfigRepository) FindPlatform(ctx context.Context, id string) (*system_db.SysModelPlatform, error) {
	platform := new(system_db.SysModelPlatform)
	if err := r.db.NewSelect().Model(platform).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	platform.APIKeyHint = apiKeyHint(platform.APIKeyCiphertext)
	return platform, nil
}

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

func (r *modelConfigRepository) UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*system_db.SysModelPlatform)(nil)).
		Set("last_model_sync_at = ?", syncedAt).
		Where("id = ?", platformID).
		Exec(ctx)
	return err
}

func (r *modelConfigRepository) DeletePlatform(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*system_db.SysModelPlatform)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *modelConfigRepository) FindModels(ctx context.Context) ([]*system_db.SysModel, error) {
	var models []*system_db.SysModel
	err := r.db.NewSelect().Model(&models).Order("created_at DESC").Scan(ctx)
	return models, err
}

func (r *modelConfigRepository) ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error) {
	count, err := r.db.NewSelect().Model((*system_db.SysModel)(nil)).
		Where("platform_id = ?", platformID).
		Where("model_name = ?", modelName).
		Count(ctx)
	return count > 0, err
}

func (r *modelConfigRepository) CreateModel(ctx context.Context, model *system_db.SysModel) error {
	model.ID = newID()
	if model.Status == "" {
		model.Status = modelStatusActive
	}
	_, err := r.db.NewInsert().Model(model).Exec(ctx)
	return err
}

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

func (r *modelConfigRepository) DeleteModel(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*system_db.SysModel)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
