package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	ai_db "sag-wiki/app/ai/models/db"
)

const (
	modelStatusActive   = "active"
	modelStatusInactive = "inactive"
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

	FindPlatforms(ctx context.Context) ([]*ai_db.SysModelPlatform, error)
	FindPlatform(ctx context.Context, id string) (*ai_db.SysModelPlatform, error)
	CreatePlatform(ctx context.Context, platform *ai_db.SysModelPlatform) error
	UpdatePlatformConfig(ctx context.Context, platformID string, baseURL string, apiKey *string, clearAPIKey bool) (*ai_db.SysModelPlatform, error)
	UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error
	DeletePlatform(ctx context.Context, id string) error

	FindModels(ctx context.Context) ([]*ai_db.SysModel, error)
	ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error)
	CreateModel(ctx context.Context, model *ai_db.SysModel) error
	UpdateModel(ctx context.Context, modelID string, update ModelUpdate) (*ai_db.SysModel, error)
	DeleteModel(ctx context.Context, id string) error
}

type ModelUpdate struct {
	Status       *string
	DisplayName  *string
	Capabilities []string
}

// 模型配置仓储
type modelConfigRepository struct {
	db *bun.DB
}

func NewModelConfigRepository(database *bun.DB) ModelConfigRepository {
	return &modelConfigRepository{db: database}
}

func buildModelConfig(platform *ai_db.SysModelPlatform, model *ai_db.SysModel) *ai_db.ModelConfig {
	baseURL := platform.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = platform.DefaultBaseURL
	}

	return &ai_db.ModelConfig{
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

func (r *modelConfigRepository) findModelConfig(ctx context.Context, filter func(*bun.SelectQuery) *bun.SelectQuery) (*ai_db.ModelConfig, error) {
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

func (r modelConfigRow) toModelConfig() *ai_db.ModelConfig {
	return buildModelConfig(
		&ai_db.SysModelPlatform{
			ID:               r.PlatformID,
			Name:             r.PlatformName,
			DefaultBaseURL:   r.DefaultBaseURL,
			BaseURL:          r.BaseURL,
			APIKeyCiphertext: r.APIKeyCiphertext,
		},
		&ai_db.SysModel{
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

// 根据 ID 获取模型配置
func (r *modelConfigRepository) FindOne(ctx context.Context, id string) (*ai_db.ModelConfig, error) {
	config, err := r.findModelConfig(ctx, func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Where("sm.id = ?", id)
	})
	if err != nil {
		return nil, fmt.Errorf("根据 ID 获取模型配置失败: %w", err)
	}
	return config, nil
}

// 获取默认模型配置
func (r *modelConfigRepository) FindDefault(ctx context.Context) (*ai_db.ModelConfig, error) {
	return r.FindDefaultByType(ctx, ai_db.ModelTypeChat)
}

// FindDefaultByType 获取指定类型的默认模型配置
func (r *modelConfigRepository) FindDefaultByType(ctx context.Context, modelType string) (*ai_db.ModelConfig, error) {
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

// 设置默认模型
func (r *modelConfigRepository) SetDefault(ctx context.Context, id string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var model ai_db.SysModel
		if err := tx.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx); err != nil {
			return fmt.Errorf("获取模型配置失败: %w", err)
		}

		if _, err := tx.NewUpdate().
			Model((*ai_db.SysModel)(nil)).
			Set("is_default = ?", false).
			Where("model_type = ?", model.ModelType).
			Where("id != ?", id).
			Exec(ctx); err != nil {
			return fmt.Errorf("设置默认模型失败: %w", err)
		}

		if _, err := tx.NewUpdate().
			Model((*ai_db.SysModel)(nil)).
			Set("is_default = ?", true).
			Where("id = ?", id).
			Exec(ctx); err != nil {
			return fmt.Errorf("设置默认模型失败: %w", err)
		}

		return nil
	})
}

// 获取模型配置列表
func (r *modelConfigRepository) FindList(ctx context.Context) ([]*ai_db.ModelConfig, error) {
	var models []*ai_db.SysModel
	if err := r.db.NewSelect().
		Model(&models).
		Order("created_at DESC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("获取模型配置列表失败: %w", err)
	}

	platformIDs := make([]string, 0, len(models))
	for _, model := range models {
		platformIDs = append(platformIDs, model.PlatformID)
	}

	platforms := make([]*ai_db.SysModelPlatform, 0)
	if len(platformIDs) > 0 {
		if err := r.db.NewSelect().
			Model(&platforms).
			Where("id IN (?)", bun.In(platformIDs)).
			Scan(ctx); err != nil {
			return nil, fmt.Errorf("获取模型平台失败: %w", err)
		}
	}

	platformByID := make(map[string]*ai_db.SysModelPlatform, len(platforms))
	for _, platform := range platforms {
		platformByID[platform.ID] = platform
	}

	configs := make([]*ai_db.ModelConfig, 0, len(models))
	for _, model := range models {
		if platform := platformByID[model.PlatformID]; platform != nil {
			configs = append(configs, buildModelConfig(platform, model))
		}
	}

	return configs, nil
}

// 创建模型配置
func (r *modelConfigRepository) Create(ctx context.Context, config *ai_db.ModelConfig) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		platform := &ai_db.SysModelPlatform{
			ID:               newID(),
			Name:             config.Vendor,
			ProviderType:     "openai_compatible",
			DefaultBaseURL:   config.BaseURL,
			BaseURL:          config.BaseURL,
			APIKeyCiphertext: config.APIKey,
			APIKeyHint:       apiKeyHint(config.APIKey),
			ModelListPath:    "/models",
			AuthScheme:       "bearer",
			Enabled:          true,
			SortOrder:        0,
		}
		if strings.TrimSpace(platform.Name) == "" {
			platform.Name = config.Name
		}

		if _, err := tx.NewInsert().Model(platform).Exec(ctx); err != nil {
			return err
		}

		model := &ai_db.SysModel{
			ID:          newID(),
			PlatformID:  platform.ID,
			ModelName:   config.Model,
			DisplayName: config.Name,
			ModelType:   config.ModelType,
			Status:      statusFromActive(config.IsActive),
			Source:      "manual",
			IsDefault:   config.IsDefault,
			Temperature: config.Temperature,
			TopP:        config.TopP,
			MaxTokens:   config.MaxTokens,
			TopK:        config.TopK,
		}
		if _, err := tx.NewInsert().Model(model).Exec(ctx); err != nil {
			return err
		}

		config.ID = model.ID
		return nil
	})
}

// 更新模型配置
func (r *modelConfigRepository) Update(ctx context.Context, config *ai_db.ModelConfig) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var model ai_db.SysModel
		if err := tx.NewSelect().Model(&model).Where("id = ?", config.ID).Scan(ctx); err != nil {
			return err
		}

		if _, err := tx.NewUpdate().
			Model((*ai_db.SysModelPlatform)(nil)).
			Set("name = ?", config.Vendor).
			Set("base_url = ?", config.BaseURL).
			Set("api_key_ciphertext = ?", config.APIKey).
			Set("api_key_hint = ?", apiKeyHint(config.APIKey)).
			Where("id = ?", model.PlatformID).
			Exec(ctx); err != nil {
			return err
		}

		model.ModelName = config.Model
		model.DisplayName = config.Name
		model.ModelType = config.ModelType
		model.Status = statusFromActive(config.IsActive)
		model.IsDefault = config.IsDefault
		model.Temperature = config.Temperature
		model.TopP = config.TopP
		model.MaxTokens = config.MaxTokens
		model.TopK = config.TopK

		_, err := tx.NewUpdate().Model(&model).Where("id = ?", model.ID).Exec(ctx)
		return err
	})
}

// 删除模型配置
func (r *modelConfigRepository) Delete(ctx context.Context, id string) error {
	return r.DeleteModel(ctx, id)
}

func statusFromActive(active bool) string {
	if active {
		return modelStatusActive
	}
	return modelStatusInactive
}

func (r *modelConfigRepository) FindPlatforms(ctx context.Context) ([]*ai_db.SysModelPlatform, error) {
	var platforms []*ai_db.SysModelPlatform
	if err := r.db.NewSelect().
		Model(&platforms).
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return platforms, nil
}

func (r *modelConfigRepository) FindPlatform(ctx context.Context, id string) (*ai_db.SysModelPlatform, error) {
	var platform ai_db.SysModelPlatform
	if err := r.db.NewSelect().Model(&platform).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return &platform, nil
}

func (r *modelConfigRepository) CreatePlatform(ctx context.Context, platform *ai_db.SysModelPlatform) error {
	platform.ID = newID()
	platform.APIKeyHint = apiKeyHint(platform.APIKeyCiphertext)
	if platform.ProviderType == "" {
		platform.ProviderType = "openai_compatible"
	}
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

func (r *modelConfigRepository) UpdatePlatformConfig(ctx context.Context, platformID string, baseURL string, apiKey *string, clearAPIKey bool) (*ai_db.SysModelPlatform, error) {
	var platform ai_db.SysModelPlatform
	if err := r.db.NewSelect().Model(&platform).Where("id = ?", platformID).Scan(ctx); err != nil {
		return nil, err
	}

	platform.BaseURL = baseURL
	if clearAPIKey {
		platform.APIKeyCiphertext = ""
		platform.APIKeyHint = nil
	} else if apiKey != nil {
		platform.APIKeyCiphertext = *apiKey
		platform.APIKeyHint = apiKeyHint(*apiKey)
	}

	if _, err := r.db.NewUpdate().Model(&platform).Where("id = ?", platform.ID).Exec(ctx); err != nil {
		return nil, err
	}
	return &platform, nil
}

func (r *modelConfigRepository) UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*ai_db.SysModelPlatform)(nil)).
		Set("last_model_sync_at = ?", syncedAt).
		Where("id = ?", platformID).
		Exec(ctx)
	return err
}

func (r *modelConfigRepository) DeletePlatform(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*ai_db.SysModelPlatform)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *modelConfigRepository) FindModels(ctx context.Context) ([]*ai_db.SysModel, error) {
	var models []*ai_db.SysModel
	if err := r.db.NewSelect().
		Model(&models).
		Order("created_at DESC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return models, nil
}

func (r *modelConfigRepository) ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error) {
	return r.db.NewSelect().
		Model((*ai_db.SysModel)(nil)).
		Where("platform_id = ?", platformID).
		Where("model_name = ?", modelName).
		Exists(ctx)
}

func (r *modelConfigRepository) CreateModel(ctx context.Context, model *ai_db.SysModel) error {
	model.ID = newID()
	if model.DisplayName == "" {
		model.DisplayName = model.ModelName
	}
	if model.Status == "" {
		model.Status = modelStatusActive
	}
	if model.Source == "" {
		model.Source = "manual"
	}
	if model.Temperature == 0 {
		model.Temperature = 0.7
	}
	if model.TopP == 0 {
		model.TopP = 0.9
	}

	_, err := r.db.NewInsert().Model(model).Exec(ctx)
	return err
}

func (r *modelConfigRepository) UpdateModel(ctx context.Context, modelID string, update ModelUpdate) (*ai_db.SysModel, error) {
	var model ai_db.SysModel
	if err := r.db.NewSelect().Model(&model).Where("id = ?", modelID).Scan(ctx); err != nil {
		return nil, err
	}

	if update.Status != nil {
		model.Status = *update.Status
	}
	if update.DisplayName != nil {
		model.DisplayName = *update.DisplayName
	}
	if update.Capabilities != nil {
		model.Capabilities = update.Capabilities
	}

	if _, err := r.db.NewUpdate().Model(&model).Where("id = ?", model.ID).Exec(ctx); err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *modelConfigRepository) DeleteModel(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*ai_db.SysModel)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
