package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"verve/config"
)

const WikiChunkCollection = "verve_wiki_chunks"

type ScoredPoint struct {
	PointID string
	Score   float64
}

type Point struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

type Store interface {
	EnsureCollection(ctx context.Context, collection string, dimension int) error
	Upsert(ctx context.Context, collection string, points []Point) error
	Search(ctx context.Context, collection string, vector []float32, filter map[string]any, limit int) ([]ScoredPoint, error)
	DeleteByDocument(ctx context.Context, collection string, documentID string) error
}

type QdrantStore struct {
	baseURL string
	client  *http.Client
}

func NewQdrantStore(cfg *config.QdrantConfig) *QdrantStore {
	return NewQdrantStoreWithClient(cfg.URL, &http.Client{Timeout: 30 * time.Second})
}

func NewQdrantStoreWithClient(baseURL string, client *http.Client) *QdrantStore {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &QdrantStore{baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

func (s *QdrantStore) EnsureCollection(ctx context.Context, collection string, dimension int) error {
	body := map[string]any{
		"vectors": map[string]any{
			"size":     dimension,
			"distance": "Cosine",
		},
	}
	return s.do(ctx, http.MethodPut, fmt.Sprintf("/collections/%s", collection), body, nil)
}

func (s *QdrantStore) Upsert(ctx context.Context, collection string, points []Point) error {
	if len(points) == 0 {
		return nil
	}
	body := map[string]any{"points": points}
	return s.do(ctx, http.MethodPut, fmt.Sprintf("/collections/%s/points?wait=true", collection), body, nil)
}

func (s *QdrantStore) Search(ctx context.Context, collection string, vector []float32, filter map[string]any, limit int) ([]ScoredPoint, error) {
	body := map[string]any{
		"vector":       vector,
		"filter":       filter,
		"limit":        limit,
		"with_payload": false,
		"with_vector":  false,
	}
	var parsed struct {
		Result []struct {
			ID    string  `json:"id"`
			Score float64 `json:"score"`
		} `json:"result"`
	}
	if err := s.do(ctx, http.MethodPost, fmt.Sprintf("/collections/%s/points/search", collection), body, &parsed); err != nil {
		return nil, err
	}
	points := make([]ScoredPoint, 0, len(parsed.Result))
	for _, item := range parsed.Result {
		points = append(points, ScoredPoint{PointID: item.ID, Score: item.Score})
	}
	return points, nil
}

func (s *QdrantStore) DeleteByDocument(ctx context.Context, collection string, documentID string) error {
	body := map[string]any{
		"filter": map[string]any{
			"must": []map[string]any{
				{"key": "document_id", "match": map[string]any{"value": documentID}},
			},
		},
	}
	return s.do(ctx, http.MethodPost, fmt.Sprintf("/collections/%s/points/delete?wait=true", collection), body, nil)
}

func (s *QdrantStore) do(ctx context.Context, method, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("qdrant request failed: %s %s status %d", method, path, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
