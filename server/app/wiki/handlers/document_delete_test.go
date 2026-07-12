package handlers

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"

	wiki_db "verve/app/wiki/models/db"
)

type fakeDocumentRepository struct {
	doc          *wiki_db.Document
	deleteErr    error
	deleteCalled bool
	calls        *[]string
}

func (f *fakeDocumentRepository) Page(ctx context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error) {
	return nil, 0, nil
}
func (f *fakeDocumentRepository) List(ctx context.Context, name, folderID string) ([]*wiki_db.Document, error) {
	return nil, nil
}
func (f *fakeDocumentRepository) FindOne(ctx context.Context, id string) (*wiki_db.Document, error) {
	return f.doc, nil
}
func (f *fakeDocumentRepository) Create(ctx context.Context, folderID string, filename string, fileSize int64, filePath string) (*wiki_db.Document, error) {
	return nil, nil
}
func (f *fakeDocumentRepository) UpdateFileSize(ctx context.Context, docID string, fileSize int64) error {
	return nil
}
func (f *fakeDocumentRepository) Delete(ctx context.Context, id string) error {
	return f.DeleteWithChunks(ctx, id)
}
func (f *fakeDocumentRepository) DeleteWithChunks(ctx context.Context, id string) error {
	f.deleteCalled = true
	*f.calls = append(*f.calls, "db")
	return f.deleteErr
}

type fakeDocumentFileStore struct {
	deleteErr    error
	deleteCalled bool
	calls        *[]string
}

func (f *fakeDocumentFileStore) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	return nil
}
func (f *fakeDocumentFileStore) GetPresignedURL(ctx context.Context, objectName string) (string, error) {
	return "", nil
}
func (f *fakeDocumentFileStore) GetFileContent(ctx context.Context, objectName string) (string, error) {
	return "", nil
}
func (f *fakeDocumentFileStore) PutFileContent(ctx context.Context, objectName string, content string) error {
	return nil
}
func (f *fakeDocumentFileStore) DeleteFile(ctx context.Context, objectName string) error {
	f.deleteCalled = true
	*f.calls = append(*f.calls, "minio")
	return f.deleteErr
}

type fakeDocumentIndexer struct {
	deleteErr    error
	deleteCalled bool
	calls        *[]string
}

func (f *fakeDocumentIndexer) ProcessJob(ctx context.Context, jobID string) error {
	return nil
}
func (f *fakeDocumentIndexer) DeleteDocumentVectors(ctx context.Context, documentID string) error {
	f.deleteCalled = true
	*f.calls = append(*f.calls, "qdrant")
	return f.deleteErr
}

func TestDocumentDeleteRemovesIndexFileThenDatabase(t *testing.T) {
	calls := make([]string, 0, 3)
	repo := &fakeDocumentRepository{
		doc:   &wiki_db.Document{ID: "doc-1", FilePath: "documents/doc-1.md"},
		calls: &calls,
	}
	files := &fakeDocumentFileStore{calls: &calls}
	indexer := &fakeDocumentIndexer{calls: &calls}
	handler := NewDocumentHandlerWithDependencies(repo, files, indexer)
	app := fiber.New()
	app.Delete("/documents/:id", handler.Delete)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/documents/doc-1", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	wantCalls := []string{"qdrant", "minio", "db"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}
}

func TestDocumentDeleteStopsWhenIndexDeleteFails(t *testing.T) {
	calls := make([]string, 0, 3)
	repo := &fakeDocumentRepository{
		doc:   &wiki_db.Document{ID: "doc-1", FilePath: "documents/doc-1.md"},
		calls: &calls,
	}
	files := &fakeDocumentFileStore{calls: &calls}
	indexer := &fakeDocumentIndexer{deleteErr: errors.New("qdrant down"), calls: &calls}
	handler := NewDocumentHandlerWithDependencies(repo, files, indexer)
	app := fiber.New()
	app.Delete("/documents/:id", handler.Delete)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/documents/doc-1", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	wantCalls := []string{"qdrant"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}
	if files.deleteCalled || repo.deleteCalled {
		t.Fatal("file or database delete ran after index delete failed")
	}
}
