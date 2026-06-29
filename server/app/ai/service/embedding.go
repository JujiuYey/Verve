package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	system_db "sag-wiki/app/system/models/db"
	system_repository "sag-wiki/app/system/repository"
)

type EmbeddingService struct {
	repo       system_repository.ModelConfigRepository
	httpClient *http.Client
}

func NewEmbeddingService(repo system_repository.ModelConfigRepository) *EmbeddingService {
	return &EmbeddingService{
		repo:       repo,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// GetEmbedding 调用 embedding API 获取向量
func (s *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	config, err := s.repo.FindDefaultByType(ctx, system_db.ModelTypeEmbedding)
	if err != nil {
		return nil, fmt.Errorf("获取 embedding 配置失败: %w", err)
	}

	return s.CallEmbeddingAPI(ctx, config, text)
}

// CallEmbeddingAPI 调用 embedding API
func (s *EmbeddingService) CallEmbeddingAPI(ctx context.Context, config *system_db.ModelConfig, text string) ([]float32, error) {
	url := fmt.Sprintf("%s/embeddings", config.BaseURL)

	reqBody := map[string]interface{}{
		"input": text,
		"model": config.Model,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 embedding API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API 返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("embedding 返回为空")
	}

	return result.Data[0].Embedding, nil
}

// BatchEmbeddings 批量获取 embeddings
func (s *EmbeddingService) BatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, 0, len(texts))
	for _, text := range texts {
		emb, err := s.GetEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		results = append(results, emb)
	}
	return results, nil
}
