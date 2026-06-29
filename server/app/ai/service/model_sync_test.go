package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ai_db "sag-wiki/app/ai/models/db"
)

func TestModelSyncServiceFetchModelsUsesPlatformAuthAndHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("expected /v1/models, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", got)
		}
		if got := r.Header.Get("X-Test-Tenant"); got != "tenant-1" {
			t.Fatalf("expected extra header, got %q", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "MiniMax-M2"},
			},
		})
	}))
	defer server.Close()

	svc := NewModelSyncService(nil)
	platform := &ai_db.SysModelPlatform{
		BaseURL:       server.URL + "/v1/",
		ModelListPath: "models",
		AuthScheme:    "bearer",
		ExtraHeaders: map[string]interface{}{
			"X-Test-Tenant": "tenant-1",
		},
	}

	result, err := svc.fetchModels(context.Background(), platform, "test-key")
	if err != nil {
		t.Fatalf("fetch models: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].ID != "MiniMax-M2" {
		t.Fatalf("unexpected models response: %+v", result.Data)
	}
}
