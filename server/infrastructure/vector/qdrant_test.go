package vector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQdrantSearchRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/collections/verve_wiki_chunks/points/search" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		filter := body["filter"].(map[string]any)
		must := filter["must"].([]any)
		match := must[0].(map[string]any)["match"].(map[string]any)
		if match["value"] != "root-1" {
			t.Fatalf("root filter = %#v", match["value"])
		}
		_, _ = w.Write([]byte(`{"result":[{"id":"point-1","score":0.9}]}`))
	}))
	defer server.Close()

	store := NewQdrantStoreWithClient(server.URL, server.Client())
	filter := map[string]any{
		"must": []map[string]any{{"key": "root_folder_id", "match": map[string]any{"value": "root-1"}}},
	}
	results, err := store.Search(context.Background(), WikiChunkCollection, []float32{0.1, 0.2}, filter, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].PointID != "point-1" {
		t.Fatalf("results = %#v", results)
	}
}

func TestQdrantDeleteByDocumentRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/collections/verve_wiki_chunks/points/delete" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("wait") != "true" {
			t.Fatalf("wait = %s", r.URL.Query().Get("wait"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		filter := body["filter"].(map[string]any)
		must := filter["must"].([]any)
		match := must[0].(map[string]any)["match"].(map[string]any)
		if match["value"] != "doc-1" {
			t.Fatalf("document filter = %#v", match["value"])
		}
		_, _ = w.Write([]byte(`{"result":{"operation_id":1,"status":"completed"}}`))
	}))
	defer server.Close()

	store := NewQdrantStoreWithClient(server.URL, server.Client())
	if err := store.DeleteByDocument(context.Background(), WikiChunkCollection, "doc-1"); err != nil {
		t.Fatal(err)
	}
}
