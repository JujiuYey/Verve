package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
)

const modelStatusActive = "active"

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