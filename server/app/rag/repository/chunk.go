package repository

import (
	"context"
	"sort"

	rag_db "verve/app/rag/models/db"

	"github.com/uptrace/bun"
)

type ChunkRepository struct {
	db *bun.DB
}

func NewChunkRepository(db *bun.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) ReplaceDocumentChunks(ctx context.Context, documentID string, chunks []*rag_db.WikiChunk) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().Model((*rag_db.WikiChunk)(nil)).Where("document_id = ?", documentID).Exec(ctx); err != nil {
			return err
		}
		if len(chunks) == 0 {
			return nil
		}
		_, err := tx.NewInsert().Model(&chunks).Exec(ctx)
		return err
	})
}

func (r *ChunkRepository) FindByPointIDs(ctx context.Context, pointIDs []string) ([]*rag_db.WikiChunk, error) {
	if len(pointIDs) == 0 {
		return []*rag_db.WikiChunk{}, nil
	}
	var chunks []*rag_db.WikiChunk
	err := r.db.NewSelect().
		Model(&chunks).
		Join("JOIN wiki_documents AS d ON d.id = rwc.document_id").
		Where("rwc.vector_point_id IN (?)", bun.In(pointIDs)).
		Where("rwc.document_version = d.current_version").
		Scan(ctx)
	return chunks, err
}

func (r *ChunkRepository) FindNeighbors(ctx context.Context, documentID string, indexes []int, radius int) ([]*rag_db.WikiChunk, error) {
	if len(indexes) == 0 {
		return []*rag_db.WikiChunk{}, nil
	}
	if radius < 0 {
		radius = 0
	}

	indexSet := make(map[int]struct{}, len(indexes)*(radius*2+1))
	for _, index := range indexes {
		for offset := -radius; offset <= radius; offset++ {
			candidate := index + offset
			if candidate >= 0 {
				indexSet[candidate] = struct{}{}
			}
		}
	}
	expandedIndexes := make([]int, 0, len(indexSet))
	for index := range indexSet {
		expandedIndexes = append(expandedIndexes, index)
	}
	sort.Ints(expandedIndexes)

	var chunks []*rag_db.WikiChunk
	err := r.db.NewSelect().
		Model(&chunks).
		Where("rwc.document_id = ?", documentID).
		Where("rwc.chunk_index IN (?)", bun.In(expandedIndexes)).
		OrderExpr("rwc.chunk_index ASC").
		Scan(ctx)
	return chunks, err
}

func (r *ChunkRepository) DeleteByDocument(ctx context.Context, documentID string) error {
	_, err := r.db.NewDelete().Model((*rag_db.WikiChunk)(nil)).Where("document_id = ?", documentID).Exec(ctx)
	return err
}
