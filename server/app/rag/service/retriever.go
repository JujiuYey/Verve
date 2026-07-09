package service

import (
	"context"
	"fmt"
	"strings"

	rag_db "verve/app/rag/models/db"
	rag_payload "verve/app/rag/models/payload"
	"verve/infrastructure/vector"
)

type chunkFinder interface {
	FindByPointIDs(ctx context.Context, pointIDs []string) ([]*rag_db.WikiChunk, error)
}

type Retriever struct {
	chunks  chunkFinder
	embed   Embedder
	vectors vector.Store
}

func NewRetriever(chunks chunkFinder, embed Embedder, vectors vector.Store) *Retriever {
	return &Retriever{chunks: chunks, embed: embed, vectors: vectors}
}

func (r *Retriever) Search(ctx context.Context, rootFolderID, query string, limit int) ([]rag_payload.SearchResult, error) {
	rootFolderID = strings.TrimSpace(rootFolderID)
	query = strings.TrimSpace(query)
	if rootFolderID == "" {
		return nil, fmt.Errorf("root_folder_id is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit = normalizeSearchLimit(limit)
	embedding, err := r.embed.EmbedTexts(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embedding.Embeddings) == 0 {
		return []rag_payload.SearchResult{}, nil
	}
	filter := map[string]any{
		"must": []map[string]any{
			{"key": "root_folder_id", "match": map[string]any{"value": rootFolderID}},
		},
	}
	points, err := r.vectors.Search(ctx, vector.WikiChunkCollection, embedding.Embeddings[0], filter, limit)
	if err != nil {
		return nil, err
	}
	pointIDs := make([]string, 0, len(points))
	for _, point := range points {
		pointIDs = append(pointIDs, point.PointID)
	}
	chunks, err := r.chunks.FindByPointIDs(ctx, pointIDs)
	if err != nil {
		return nil, err
	}
	byPointID := make(map[string]*rag_db.WikiChunk, len(chunks))
	for _, chunk := range chunks {
		byPointID[chunk.VectorPointID] = chunk
	}
	results := make([]rag_payload.SearchResult, 0, len(points))
	for _, point := range points {
		chunk, ok := byPointID[point.PointID]
		if !ok {
			continue
		}
		results = append(results, rag_payload.SearchResult{
			ChunkID:       chunk.ID,
			Score:         point.Score,
			RootFolderID:  chunk.RootFolderID,
			FolderID:      chunk.FolderID,
			DocumentID:    chunk.DocumentID,
			DocumentTitle: chunk.DocumentTitle,
			FolderPath:    chunk.FolderPath,
			HeadingPath:   chunk.HeadingPath,
			Content:       chunk.Content,
		})
	}
	return results, nil
}

func normalizeSearchLimit(limit int) int {
	if limit <= 0 {
		return 6
	}
	if limit > 12 {
		return 12
	}
	return limit
}
