package service

import (
	"context"
	"testing"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"
)

type fakeDocs struct{ doc *wiki_db.Document }

func (f fakeDocs) FindOne(ctx context.Context, id string) (*wiki_db.Document, error) {
	return f.doc, nil
}

type fakeContent struct{ text string }

func (f fakeContent) GetFileContent(ctx context.Context, objectName string) (string, error) {
	return f.text, nil
}

type fakeChunks struct{ saved []*rag_db.WikiChunk }

func (f *fakeChunks) ReplaceDocumentChunks(ctx context.Context, documentID string, chunks []*rag_db.WikiChunk) error {
	f.saved = chunks
	return nil
}
func (f *fakeChunks) DeleteByDocument(ctx context.Context, documentID string) error { return nil }

type fakeJobs struct{}

func (f fakeJobs) CreatePending(ctx context.Context, documentID string) (*rag_db.IndexJob, error) {
	return &rag_db.IndexJob{ID: "job", DocumentID: documentID}, nil
}
func (f fakeJobs) FindOne(ctx context.Context, jobID string) (*rag_db.IndexJob, error) {
	return &rag_db.IndexJob{ID: jobID, DocumentID: "doc", AttemptCount: 1, MaxAttempts: 3}, nil
}
func (f fakeJobs) MarkRunning(ctx context.Context, jobID string, rootFolderID string) error {
	return nil
}
func (f fakeJobs) MarkCompleted(ctx context.Context, jobID string, chunkCount int) error { return nil }
func (f fakeJobs) MarkPendingRetry(ctx context.Context, jobID string, message string) error {
	return nil
}
func (f fakeJobs) MarkFailed(ctx context.Context, jobID string, message string) error { return nil }

type countingJobs struct {
	created int
}

func (f *countingJobs) CreatePending(ctx context.Context, documentID string) (*rag_db.IndexJob, error) {
	f.created++
	return &rag_db.IndexJob{ID: "job", DocumentID: documentID}, nil
}
func (f *countingJobs) FindOne(ctx context.Context, jobID string) (*rag_db.IndexJob, error) {
	return &rag_db.IndexJob{ID: jobID, DocumentID: "doc", AttemptCount: 1, MaxAttempts: 3}, nil
}
func (f *countingJobs) MarkRunning(ctx context.Context, jobID string, rootFolderID string) error {
	return nil
}
func (f *countingJobs) MarkCompleted(ctx context.Context, jobID string, chunkCount int) error {
	return nil
}
func (f *countingJobs) MarkPendingRetry(ctx context.Context, jobID string, message string) error {
	return nil
}
func (f *countingJobs) MarkFailed(ctx context.Context, jobID string, message string) error {
	return nil
}

type notReadyEmbedder struct{}

func (f notReadyEmbedder) EmbedTexts(ctx context.Context, texts []string) (EmbeddingResult, error) {
	return EmbeddingResult{}, nil
}

func (f notReadyEmbedder) CheckReady(ctx context.Context) error {
	return errNotReadyForTest{}
}

type errNotReadyForTest struct{}

func (errNotReadyForTest) Error() string {
	return "embedding not configured"
}

func TestIndexerStoresWikiScopedChunkMetadata(t *testing.T) {
	rootID := "root"
	folderID := "folder"
	chunks := &fakeChunks{}
	store := &fakeVectorStore{}
	indexer := NewIndexerWithDependencies(
		chunks,
		fakeJobs{},
		fakeDocs{doc: &wiki_db.Document{ID: "doc", Filename: "channel.md", FolderID: folderID, FilePath: "objects/channel.md"}},
		fakeContent{text: "# Go\n\n## Channel\n\nChannel is a typed pipe."},
		NewRootResolver(fakeFolderFinder{
			rootID:   {ID: rootID, Name: "Go"},
			folderID: {ID: folderID, Name: "Concurrency", ParentID: &rootID},
		}),
		NewMarkdownChunker(1800),
		fakeEmbedder{result: EmbeddingResult{Model: "embed", Dimension: 2, Embeddings: [][]float32{{0.1, 0.2}}}},
		store,
	)

	if err := indexer.IndexDocument(context.Background(), "doc"); err != nil {
		t.Fatal(err)
	}
	if len(chunks.saved) != 1 {
		t.Fatalf("saved chunks = %d", len(chunks.saved))
	}
	chunk := chunks.saved[0]
	if chunk.RootFolderID != rootID || chunk.FolderPath != "Go/Concurrency" || chunk.HeadingPath != "Go > Channel" {
		t.Fatalf("chunk metadata = %#v", chunk)
	}
	if len(store.points) != 1 {
		t.Fatalf("vector points = %d", len(store.points))
	}
}

func TestIndexerDoesNotCreateJobWhenEmbeddingIsNotReady(t *testing.T) {
	jobs := &countingJobs{}
	indexer := NewIndexerWithDependencies(
		&fakeChunks{},
		jobs,
		fakeDocs{doc: &wiki_db.Document{ID: "doc", Filename: "channel.md", FolderID: "folder", FilePath: "objects/channel.md"}},
		fakeContent{text: "# Go"},
		NewRootResolver(fakeFolderFinder{"folder": {ID: "folder", Name: "Go"}}),
		NewMarkdownChunker(1800),
		notReadyEmbedder{},
		&fakeVectorStore{},
	)

	if err := indexer.IndexDocument(context.Background(), "doc"); err == nil {
		t.Fatal("expected readiness error")
	}
	if jobs.created != 0 {
		t.Fatalf("created jobs = %d", jobs.created)
	}
}
