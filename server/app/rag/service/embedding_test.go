package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	system_db "verve/app/system/models/db"
)

type fakeDefaultEmbeddingModelRepo struct {
	model     *system_db.SysModel
	platform  *system_db.SysModelPlatform
	agentKey  string
	sceneKey  string
	modelType string
	err       error
}

func (r *fakeDefaultEmbeddingModelRepo) FindAgentModelWithPlatform(ctx context.Context, agentKey string, sceneKey string, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
	r.agentKey = agentKey
	r.sceneKey = sceneKey
	r.modelType = modelType
	if r.err != nil {
		return nil, nil, r.err
	}
	return r.model, r.platform, nil
}

func TestOpenAICompatibleEmbedderRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/embeddings" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "text-embedding-3-small" {
			t.Fatalf("model = %q", body.Model)
		}
		if len(body.Input) != 2 || body.Input[0] != "chunk one" || body.Input[1] != "chunk two" {
			t.Fatalf("input = %#v", body.Input)
		}
		_, _ = w.Write([]byte(`{"model":"text-embedding-3-small","data":[{"embedding":[0.1,0.2]},{"embedding":[0.3,0.4]}]}`))
	}))
	defer server.Close()

	repo := &fakeDefaultEmbeddingModelRepo{
		model: &system_db.SysModel{
			ModelName: "text-embedding-3-small",
		},
		platform: &system_db.SysModelPlatform{
			BaseURL:          server.URL,
			APIKeyCiphertext: "test-key",
		},
	}
	embedder := NewOpenAICompatibleEmbedderWithClient(repo, server.Client())

	result, err := embedder.EmbedTexts(context.Background(), []string{"chunk one", "chunk two"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.agentKey != "wiki_rag" || repo.sceneKey != "embedding" || repo.modelType != "embedding" {
		t.Fatalf("model usage = %s/%s/%s", repo.agentKey, repo.sceneKey, repo.modelType)
	}
	if result.Model != "text-embedding-3-small" {
		t.Fatalf("result model = %q", result.Model)
	}
	if result.Dimension != 2 {
		t.Fatalf("dimension = %d", result.Dimension)
	}
}

func TestOpenAICompatibleEmbedderCheckReadyHidesNoRowsDetail(t *testing.T) {
	embedder := NewOpenAICompatibleEmbedderWithClient(&fakeDefaultEmbeddingModelRepo{
		err: sql.ErrNoRows,
	}, nil)

	err := embedder.CheckReady(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "未配置知识库 embedding 模型，请先在 Agent 配置中设置 wiki_rag.embedding" {
		t.Fatalf("error = %q", err.Error())
	}
}
