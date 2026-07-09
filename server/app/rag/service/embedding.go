package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	system_db "verve/app/system/models/db"
	system_repo "verve/app/system/repository"
)

type Embedder interface {
	EmbedTexts(ctx context.Context, texts []string) (EmbeddingResult, error)
}

type ReadyChecker interface {
	CheckReady(ctx context.Context) error
}

type EmbeddingResult struct {
	Model      string
	Dimension  int
	Embeddings [][]float32
}

const (
	AgentKeyWikiRAG      = "wiki_rag"
	SceneKeyRAGEmbedding = "embedding"
)

type AgentModelRepository interface {
	FindAgentModelWithPlatform(ctx context.Context, agentKey string, sceneKey string, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error)
}

type OpenAICompatibleEmbedder struct {
	models AgentModelRepository
	client *http.Client
}

func NewOpenAICompatibleEmbedder(models AgentModelRepository) *OpenAICompatibleEmbedder {
	return &OpenAICompatibleEmbedder{
		models: models,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func NewOpenAICompatibleEmbedderWithClient(models AgentModelRepository, client *http.Client) *OpenAICompatibleEmbedder {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &OpenAICompatibleEmbedder{models: models, client: client}
}

func (e *OpenAICompatibleEmbedder) CheckReady(ctx context.Context) error {
	_, platform, err := e.models.FindAgentModelWithPlatform(ctx, AgentKeyWikiRAG, SceneKeyRAGEmbedding, system_repo.ModelTypeEmbedding)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("未配置知识库 embedding 模型，请先在 Agent 配置中设置 wiki_rag.embedding")
		}
		return fmt.Errorf("未配置知识库 embedding 模型，请先在 Agent 配置中设置 wiki_rag.embedding: %w", err)
	}
	if strings.TrimSpace(platform.APIKeyCiphertext) == "" {
		return errors.New("知识库 embedding 模型平台未配置 API Key")
	}
	return nil
}

func (e *OpenAICompatibleEmbedder) EmbedTexts(ctx context.Context, texts []string) (EmbeddingResult, error) {
	if len(texts) == 0 {
		return EmbeddingResult{}, nil
	}
	model, platform, err := e.models.FindAgentModelWithPlatform(ctx, AgentKeyWikiRAG, SceneKeyRAGEmbedding, system_repo.ModelTypeEmbedding)
	if err != nil {
		return EmbeddingResult{}, err
	}
	apiKey := strings.TrimSpace(platform.APIKeyCiphertext)
	if apiKey == "" {
		return EmbeddingResult{}, errors.New("embedding model platform api key is not configured")
	}

	body, err := json.Marshal(map[string]any{
		"model": model.ModelName,
		"input": texts,
	})
	if err != nil {
		return EmbeddingResult{}, err
	}
	endpoint := strings.TrimRight(platform.BaseURL, "/") + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return EmbeddingResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return EmbeddingResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return EmbeddingResult{}, fmt.Errorf("embedding request failed: status %d", resp.StatusCode)
	}

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Model string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return EmbeddingResult{}, err
	}
	if len(parsed.Data) != len(texts) {
		return EmbeddingResult{}, fmt.Errorf("embedding count mismatch: got %d want %d", len(parsed.Data), len(texts))
	}
	result := EmbeddingResult{
		Model:      parsed.Model,
		Embeddings: make([][]float32, 0, len(parsed.Data)),
	}
	if result.Model == "" {
		result.Model = model.ModelName
	}
	for _, item := range parsed.Data {
		if result.Dimension == 0 {
			result.Dimension = len(item.Embedding)
		}
		result.Embeddings = append(result.Embeddings, item.Embedding)
	}
	return result, nil
}
