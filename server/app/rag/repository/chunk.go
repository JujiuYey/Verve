package repository

import (
	"context"

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
		Where("vector_point_id IN (?)", bun.In(pointIDs)).
		Scan(ctx)
	return chunks, err
}

func (r *ChunkRepository) DeleteByDocument(ctx context.Context, documentID string) error {
	_, err := r.db.NewDelete().Model((*rag_db.WikiChunk)(nil)).Where("document_id = ?", documentID).Exec(ctx)
	return err
}
