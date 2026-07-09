package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	system_db "verve/app/system/models/db"
)

var ErrModelPlatformNotFound = errors.New("model platform not found")

type SyncModelsResult struct {
	Synced  int      `json:"synced"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors"`
}

type ModelSyncRepository interface {
	FindPlatform(ctx context.Context, id string) (*system_db.SysModelPlatform, error)
	ModelExistsByPlatformAndName(ctx context.Context, platformID, modelName string) (bool, error)
	CreateModel(ctx context.Context, model *system_db.SysModel) error
	UpdatePlatformLastModelSyncAt(ctx context.Context, platformID string, syncedAt time.Time) error
}

type ModelSyncService struct {
	repo       ModelSyncRepository
	httpClient *http.Client
}

type openAIModelListResponse struct {
	Data []openAIModelListItem `json:"data"`
}

type openAIModelListItem struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object,omitempty"`
	Created   int64                  `json:"created,omitempty"`
	OwnedBy   string                 `json:"owned_by,omitempty"`
	RawFields map[string]interface{} `json:"-"`
}

func (m *openAIModelListResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.Data = make([]openAIModelListItem, 0, len(raw.Data))
	for _, itemBytes := range raw.Data {
		var item openAIModelListItem
		if err := json.Unmarshal(itemBytes, &item); err != nil {
			return err
		}
		_ = json.Unmarshal(itemBytes, &item.RawFields)
		m.Data = append(m.Data, item)
	}
	return nil
}

func NewModelSyncService(repo ModelSyncRepository) *ModelSyncService {
	return &ModelSyncService{
		repo:       repo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ModelSyncService) SyncModels(ctx context.Context, platformID string) (*SyncModelsResult, error) {
	platform, err := s.repo.FindPlatform(ctx, platformID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrModelPlatformNotFound, err)
	}

	apiKey := strings.TrimSpace(platform.APIKeyCiphertext)
	if apiKey == "" {
		return nil, errors.New("平台 API Key 未配置")
	}

	models, err := s.fetchModels(ctx, platform, apiKey)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := &SyncModelsResult{Errors: []string{}}
	for _, remoteModel := range models.Data {
		modelName := strings.TrimSpace(remoteModel.ID)
		if modelName == "" {
			result.Skipped++
			continue
		}

		exists, err := s.repo.ModelExistsByPlatformAndName(ctx, platform.ID, modelName)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("检查模型 %s 失败: %v", modelName, err))
			continue
		}
		if exists {
			result.Skipped++
			continue
		}

		if err := s.repo.CreateModel(ctx, &system_db.SysModel{
			PlatformID:   platform.ID,
			ModelName:    modelName,
			DisplayName:  modelName,
			Status:       "active",
			LastSyncedAt: &now,
		}); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("创建模型 %s 失败: %v", modelName, err))
			continue
		}
		result.Synced++
	}

	if err := s.repo.UpdatePlatformLastModelSyncAt(ctx, platform.ID, now); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("更新平台同步时间失败: %v", err))
	}

	return result, nil
}

func (s *ModelSyncService) fetchModels(ctx context.Context, platform *system_db.SysModelPlatform, apiKey string) (*openAIModelListResponse, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(platform.BaseURL), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(platform.DefaultBaseURL), "/")
	}
	if baseURL == "" {
		return nil, errors.New("平台 API 地址未配置")
	}

	modelListPath := strings.TrimSpace(platform.ModelListPath)
	if modelListPath == "" {
		modelListPath = "/models"
	}
	if !strings.HasPrefix(modelListPath, "/") {
		modelListPath = "/" + modelListPath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+modelListPath, nil)
	if err != nil {
		return nil, fmt.Errorf("创建模型同步请求失败: %w", err)
	}

	switch platform.AuthScheme {
	case "x_api_key":
		req.Header.Set("x-api-key", apiKey)
	case "both":
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("x-api-key", apiKey)
	default:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	for key, value := range platform.ExtraHeaders {
		if strings.TrimSpace(key) == "" {
			continue
		}
		switch v := value.(type) {
		case string:
			req.Header.Set(key, v)
		default:
			req.Header.Set(key, fmt.Sprint(v))
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求模型列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取模型列表响应失败: %w", err)
	}
	log.Printf("🤖 模型同步响应: platform_id=%s platform=%q url=%s status=%d body_bytes=%d body_preview=%q",
		platform.ID,
		platform.Name,
		baseURL+modelListPath,
		resp.StatusCode,
		len(body),
		truncateForModelSyncLog(string(body), 4000),
	)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("模型提供方返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result openAIModelListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}
	log.Printf("🤖 模型同步解析: platform_id=%s model_count=%d sample=%s",
		platform.ID,
		len(result.Data),
		formatModelSyncSample(result.Data, 8),
	)
	return &result, nil
}

func truncateForModelSyncLog(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return fmt.Sprintf("%s...(truncated %d bytes)", text[:limit], len(text)-limit)
}

func formatModelSyncSample(models []openAIModelListItem, limit int) string {
	if len(models) == 0 {
		return "[]"
	}
	if limit > len(models) {
		limit = len(models)
	}
	sample := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		sample = append(sample, models[i].RawFields)
	}
	data, err := json.Marshal(sample)
	if err != nil {
		return fmt.Sprintf("%+v", sample)
	}
	if len(models) > limit {
		return fmt.Sprintf("%s...(and %d more)", string(data), len(models)-limit)
	}
	return string(data)
}
