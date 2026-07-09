package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	rag_db "verve/app/rag/models/db"
	rag_repo "verve/app/rag/repository"
	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
	"verve/infrastructure/vector"
)

type documentFinder interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Document, error)
}

type contentReader interface {
	GetFileContent(ctx context.Context, objectName string) (string, error)
}

type Indexer struct {
	chunks   chunkWriter
	jobs     jobWriter
	docs     documentFinder
	content  contentReader
	chunker  *MarkdownChunker
	resolver *RootResolver
	embedder Embedder
	vectors  vector.Store
}

type chunkWriter interface {
	ReplaceDocumentChunks(ctx context.Context, documentID string, chunks []*rag_db.WikiChunk) error
	DeleteByDocument(ctx context.Context, documentID string) error
}

type jobWriter interface {
	CreatePending(ctx context.Context, documentID string) (*rag_db.IndexJob, error)
	FindOne(ctx context.Context, jobID string) (*rag_db.IndexJob, error)
	MarkRunning(ctx context.Context, jobID string, rootFolderID string) error
	MarkCompleted(ctx context.Context, jobID string, chunkCount int) error
	MarkPendingRetry(ctx context.Context, jobID string, message string) error
	MarkFailed(ctx context.Context, jobID string, message string) error
}

func NewIndexer(
	chunks *rag_repo.ChunkRepository,
	jobs *rag_repo.IndexJobRepository,
	folders wiki_repo.FolderRepository,
	docs *wiki_repo.DocumentRepository,
	content contentReader,
	embedder Embedder,
	vectors vector.Store,
) *Indexer {
	return NewIndexerWithDependencies(chunks, jobs, docs, content, NewRootResolver(folders), NewMarkdownChunker(1800), embedder, vectors)
}

func NewIndexerWithDependencies(
	chunks chunkWriter,
	jobs jobWriter,
	docs documentFinder,
	content contentReader,
	resolver *RootResolver,
	chunker *MarkdownChunker,
	embedder Embedder,
	vectors vector.Store,
) *Indexer {
	return &Indexer{
		chunks: chunks, jobs: jobs, docs: docs, content: content,
		resolver: resolver, chunker: chunker, embedder: embedder, vectors: vectors,
	}
}

func (s *Indexer) IndexDocument(ctx context.Context, documentID string) error {
	if err := s.CheckReady(ctx); err != nil {
		return err
	}
	job, err := s.jobs.CreatePending(ctx, documentID)
	if err != nil {
		return err
	}
	return s.processDocumentWithJob(ctx, job, documentID, true)
}

func (s *Indexer) ProcessJob(ctx context.Context, jobID string) error {
	if err := s.CheckReady(ctx); err != nil {
		return err
	}
	job, err := s.jobs.FindOne(ctx, jobID)
	if err != nil {
		return err
	}
	err = s.processDocumentWithJob(ctx, job, job.DocumentID, false)
	if err == nil {
		return nil
	}
	latest, findErr := s.jobs.FindOne(ctx, jobID)
	if findErr == nil && latest.AttemptCount < latest.MaxAttempts {
		_ = s.jobs.MarkPendingRetry(ctx, jobID, err.Error())
		return err
	}
	_ = s.jobs.MarkFailed(ctx, jobID, err.Error())
	return fmt.Errorf("%w: %v", ErrJobAttemptsExhausted, err)
}

func (s *Indexer) processDocumentWithJob(ctx context.Context, job *rag_db.IndexJob, documentID string, markFailedImmediately bool) error {
	fail := func(err error) error {
		if markFailedImmediately {
			_ = s.jobs.MarkFailed(ctx, job.ID, err.Error())
		}
		return err
	}

	doc, err := s.docs.FindOne(ctx, documentID)
	if err != nil {
		return fail(err)
	}
	scope, err := s.resolver.Resolve(ctx, doc.FolderID)
	if err != nil {
		return fail(err)
	}
	if err := s.jobs.MarkRunning(ctx, job.ID, scope.RootFolderID); err != nil {
		return fail(err)
	}
	markdown, err := s.content.GetFileContent(ctx, doc.FilePath)
	if err != nil {
		return fail(err)
	}
	markdownChunks := s.chunker.Chunk(markdown)
	if len(markdownChunks) == 0 {
		if err := s.vectors.DeleteByDocument(ctx, vector.WikiChunkCollection, documentID); err != nil {
			return fail(err)
		}
		if err := s.chunks.ReplaceDocumentChunks(ctx, documentID, nil); err != nil {
			return fail(err)
		}
		return s.jobs.MarkCompleted(ctx, job.ID, 0)
	}
	texts := make([]string, 0, len(markdownChunks))
	for _, chunk := range markdownChunks {
		texts = append(texts, chunk.Content)
	}
	embedded, err := s.embedder.EmbedTexts(ctx, texts)
	if err != nil {
		return fail(err)
	}
	if len(embedded.Embeddings) != len(markdownChunks) {
		return fail(fmt.Errorf("embedding count mismatch: got %d want %d", len(embedded.Embeddings), len(markdownChunks)))
	}
	if err := s.vectors.EnsureCollection(ctx, vector.WikiChunkCollection, embedded.Dimension); err != nil {
		return fail(err)
	}

	now := time.Now()
	points := make([]vector.Point, 0, len(markdownChunks))
	records := make([]*rag_db.WikiChunk, 0, len(markdownChunks))
	for i, chunk := range markdownChunks {
		pointID := compactPointID()
		points = append(points, vector.Point{
			ID:     pointID,
			Vector: embedded.Embeddings[i],
			Payload: map[string]any{
				"root_folder_id": scope.RootFolderID,
				"folder_id":      doc.FolderID,
				"document_id":    doc.ID,
				"heading_path":   chunk.HeadingPath,
			},
		})
		records = append(records, &rag_db.WikiChunk{
			ID:             compactPointID(),
			RootFolderID:   scope.RootFolderID,
			FolderID:       doc.FolderID,
			DocumentID:     doc.ID,
			DocumentTitle:  doc.Filename,
			FolderPath:     scope.FolderPath,
			HeadingPath:    chunk.HeadingPath,
			ChunkIndex:     i,
			BlockType:      chunk.BlockType,
			Content:        chunk.Content,
			ContentHash:    chunk.ContentHash,
			TokenCount:     chunk.TokenCount,
			VectorPointID:  pointID,
			EmbeddingModel: embedded.Model,
			IndexedAt:      now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	if err := s.vectors.DeleteByDocument(ctx, vector.WikiChunkCollection, documentID); err != nil {
		return fail(err)
	}
	if err := s.vectors.Upsert(ctx, vector.WikiChunkCollection, points); err != nil {
		return fail(err)
	}
	if err := s.chunks.ReplaceDocumentChunks(ctx, documentID, records); err != nil {
		return fail(err)
	}
	return s.jobs.MarkCompleted(ctx, job.ID, len(records))
}

var ErrJobAttemptsExhausted = errors.New("job attempts exhausted")

func (s *Indexer) CheckReady(ctx context.Context) error {
	if checker, ok := s.embedder.(ReadyChecker); ok {
		return checker.CheckReady(ctx)
	}
	return nil
}

func (s *Indexer) DeleteDocumentIndex(ctx context.Context, documentID string) error {
	if strings.TrimSpace(documentID) == "" {
		return nil
	}
	if err := s.vectors.DeleteByDocument(ctx, vector.WikiChunkCollection, documentID); err != nil {
		return err
	}
	return s.chunks.DeleteByDocument(ctx, documentID)
}

func compactPointID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:32]
}
