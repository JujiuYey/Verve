package service

import (
	"context"
	"testing"

	rag_db "verve/app/rag/models/db"
	"verve/infrastructure/vector"
)

type fakeEmbedder struct {
	result EmbeddingResult
}

func (f fakeEmbedder) EmbedTexts(ctx context.Context, texts []string) (EmbeddingResult, error) {
	return f.result, nil
}

type fakeChunkFinder struct {
	chunks   []*rag_db.WikiChunk
	pointIDs []string
}

func (f *fakeChunkFinder) FindByPointIDs(ctx context.Context, pointIDs []string) ([]*rag_db.WikiChunk, error) {
	f.pointIDs = pointIDs
	return f.chunks, nil
}

type fakeVectorStore struct {
	searchResults []vector.ScoredPoint
	points        []vector.Point
	filter        map[string]any
	limit         int
}

func (f *fakeVectorStore) EnsureCollection(ctx context.Context, collection string, dimension int) error {
	return nil
}

func (f *fakeVectorStore) Upsert(ctx context.Context, collection string, points []vector.Point) error {
	f.points = points
	return nil
}

func (f *fakeVectorStore) Search(ctx context.Context, collection string, v []float32, filter map[string]any, limit int) ([]vector.ScoredPoint, error) {
	f.filter = filter
	f.limit = limit
	return f.searchResults, nil
}

func (f *fakeVectorStore) DeleteByDocument(ctx context.Context, collection string, documentID string) error {
	return nil
}

func TestRetrieverPreservesVectorScoreOrder(t *testing.T) {
	store := &fakeVectorStore{searchResults: []vector.ScoredPoint{
		{PointID: "p2", Score: 0.9},
		{PointID: "p1", Score: 0.8},
	}}
	retriever := NewRetriever(
		&fakeChunkFinder{chunks: []*rag_db.WikiChunk{
			{ID: "c1", VectorPointID: "p1", RootFolderID: "root", DocumentTitle: "one.md"},
			{ID: "c2", VectorPointID: "p2", RootFolderID: "root", DocumentTitle: "two.md"},
		}},
		fakeEmbedder{result: EmbeddingResult{Model: "embed", Dimension: 2, Embeddings: [][]float32{{0.1, 0.2}}}},
		store,
	)

	results, err := retriever.Search(context.Background(), "root", "channel", 20)
	if err != nil {
		t.Fatal(err)
	}
	if store.limit != 12 {
		t.Fatalf("limit = %d", store.limit)
	}
	if len(results) != 2 || results[0].ChunkID != "c2" || results[1].ChunkID != "c1" {
		t.Fatalf("results = %#v", results)
	}
}

func TestRetrieverMatchesQdrantNormalizedUUIDPointID(t *testing.T) {
	store := &fakeVectorStore{searchResults: []vector.ScoredPoint{
		{PointID: "001928ba-3d04-4986-b359-36ccd62380d1", Score: 0.9},
	}}
	finder := &fakeChunkFinder{chunks: []*rag_db.WikiChunk{
		{
			ID:            "c1",
			VectorPointID: "001928ba3d044986b35936ccd62380d1",
			RootFolderID:  "root",
			DocumentTitle: "one.md",
		},
	}}
	retriever := NewRetriever(
		finder,
		fakeEmbedder{result: EmbeddingResult{Model: "embed", Dimension: 2, Embeddings: [][]float32{{0.1, 0.2}}}},
		store,
	)

	results, err := retriever.Search(context.Background(), "root", "channel", 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ChunkID != "c1" {
		t.Fatalf("results = %#v", results)
	}
	if len(finder.pointIDs) != 2 ||
		finder.pointIDs[0] != "001928ba-3d04-4986-b359-36ccd62380d1" ||
		finder.pointIDs[1] != "001928ba3d044986b35936ccd62380d1" {
		t.Fatalf("point id candidates = %#v", finder.pointIDs)
	}
}
