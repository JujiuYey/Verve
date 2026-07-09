package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	system_db "verve/app/system/models/db"
)

type fakeDefaultEmbeddingModelRepo struct {
	model    *system_db.SysModel
	platform *system_db.SysModelPlatform
}

func (r fakeDefaultEmbeddingModelRepo) FindDefaultModelWithPlatform(ctx context.Context, modelType string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
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

	embedder := NewOpenAICompatibleEmbedderWithClient(fakeDefaultEmbeddingModelRepo{
		model: &system_db.SysModel{
			ModelName: "text-embedding-3-small",
		},
		platform: &system_db.SysModelPlatform{
			BaseURL:          server.URL,
			APIKeyCiphertext: "test-key",
		},
	}, server.Client())

	result, err := embedder.EmbedTexts(context.Background(), []string{"chunk one", "chunk two"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Model != "text-embedding-3-small" {
		t.Fatalf("result model = %q", result.Model)
	}
	if result.Dimension != 2 {
		t.Fatalf("dimension = %d", result.Dimension)
	}
}
