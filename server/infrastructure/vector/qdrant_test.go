package vector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestQdrantSearchOmitsEmptyFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, exists := body["filter"]; exists {
			t.Fatalf("request contains filter: %#v", body["filter"])
		}
		_, _ = w.Write([]byte(`{"result":[]}`))
	}))
	defer server.Close()

	store := NewQdrantStoreWithClient(server.URL, server.Client())
	if _, err := store.Search(context.Background(), WikiChunkCollection, []float32{0.1}, nil, 8); err != nil {
		t.Fatal(err)
	}
}

func TestQdrantEnsureCollectionAllowsExistingMatchingCollection(t *testing.T) {
	requests := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "PUT /collections/verve_wiki_chunks":
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"status":{"error":"collection already exists"}}`))
		case "GET /collections/verve_wiki_chunks":
			_, _ = w.Write([]byte(`{"result":{"config":{"params":{"vectors":{"size":1024}}}}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	store := NewQdrantStoreWithClient(server.URL, server.Client())
	if err := store.EnsureCollection(context.Background(), WikiChunkCollection, 1024); err != nil {
		t.Fatal(err)
	}
	wantRequests := []string{"PUT /collections/verve_wiki_chunks", "GET /collections/verve_wiki_chunks"}
	if !reflect.DeepEqual(requests, wantRequests) {
		t.Fatalf("requests = %#v, want %#v", requests, wantRequests)
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
