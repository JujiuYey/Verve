package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	system_db "verve/app/system/models/db"
)

type fakeDefaultEmbeddingModelRepo struct {
	model    *system_db.SysModel
	platform *system_db.SysModelPlatform
	agentKey string
	sceneKey string
	err      error
}

func (r *fakeDefaultEmbeddingModelRepo) FindAgentModelWithPlatform(ctx context.Context, agentKey string, sceneKey string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
	r.agentKey = agentKey
	r.sceneKey = sceneKey
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
	if repo.agentKey != "wiki_rag" || repo.sceneKey != "embedding" {
		t.Fatalf("model usage = %s/%s", repo.agentKey, repo.sceneKey)
	}
	if result.Model != "text-embedding-3-small" {
		t.Fatalf("result model = %q", result.Model)
	}
	if result.Dimension != 2 {
		t.Fatalf("dimension = %d", result.Dimension)
	}
}

func TestOpenAICompatibleEmbedderBatchesLargeInputs(t *testing.T) {
	batchSizes := make([]int, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		batchSizes = append(batchSizes, len(body.Input))
		type item struct {
			Embedding []float32 `json:"embedding"`
		}
		resp := struct {
			Model string `json:"model"`
			Data  []item `json:"data"`
		}{Model: body.Model, Data: make([]item, 0, len(body.Input))}
		for _, text := range body.Input {
			resp.Data = append(resp.Data, item{Embedding: []float32{float32(len(text))}})
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	repo := &fakeDefaultEmbeddingModelRepo{
		model: &system_db.SysModel{ModelName: "text-embedding-v4"},
		platform: &system_db.SysModelPlatform{
			BaseURL:          server.URL,
			APIKeyCiphertext: "test-key",
		},
	}
	embedder := NewOpenAICompatibleEmbedderWithClient(repo, server.Client())
	texts := make([]string, 25)
	for i := range texts {
		texts[i] = strings.Repeat("x", i+1)
	}

	result, err := embedder.EmbedTexts(context.Background(), texts)
	if err != nil {
		t.Fatal(err)
	}
	wantBatchSizes := []int{10, 10, 5}
	if !reflect.DeepEqual(batchSizes, wantBatchSizes) {
		t.Fatalf("batch sizes = %#v, want %#v", batchSizes, wantBatchSizes)
	}
	if len(result.Embeddings) != len(texts) {
		t.Fatalf("embeddings = %d, want %d", len(result.Embeddings), len(texts))
	}
	for i, embedding := range result.Embeddings {
		if len(embedding) != 1 || embedding[0] != float32(i+1) {
			t.Fatalf("embedding[%d] = %#v", i, embedding)
		}
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
