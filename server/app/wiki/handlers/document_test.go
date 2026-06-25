package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	wiki_db "sag-wiki/app/wiki/models/db"
)

type stubDocumentRepository struct {
	pageFn func(ctx context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error)
	listFn func(ctx context.Context, name, folderID string) ([]*wiki_db.Document, error)
}

func (s *stubDocumentRepository) Page(ctx context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error) {
	if s.pageFn != nil {
		return s.pageFn(ctx, pageSize, offset, name, folderID)
	}
	return nil, 0, nil
}

func (s *stubDocumentRepository) List(ctx context.Context, name, folderID string) ([]*wiki_db.Document, error) {
	if s.listFn != nil {
		return s.listFn(ctx, name, folderID)
	}
	return nil, nil
}

func (*stubDocumentRepository) FindOne(context.Context, string) (*wiki_db.Document, error) {
	return nil, errors.New("not implemented")
}

func (*stubDocumentRepository) Create(context.Context, string, string, int64, string) (*wiki_db.Document, error) {
	return nil, errors.New("not implemented")
}

func (*stubDocumentRepository) UpdateStatus(context.Context, string, string, int, *string) error {
	return errors.New("not implemented")
}

func (*stubDocumentRepository) UpdateFileSize(context.Context, string, int64) error {
	return errors.New("not implemented")
}

func (*stubDocumentRepository) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func TestDocumentHandlerFindPagePassesFolderIDAndNameFilters(t *testing.T) {
	t.Parallel()

	repo := &stubDocumentRepository{
		pageFn: func(_ context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error) {
			if pageSize != 20 {
				t.Fatalf("expected pageSize 20, got %d", pageSize)
			}
			if offset != 20 {
				t.Fatalf("expected offset 20, got %d", offset)
			}
			if folderID != "folder-1" {
				t.Fatalf("expected folderID folder-1, got %q", folderID)
			}
			if name != "spec" {
				t.Fatalf("expected name spec, got %q", name)
			}
			return []*wiki_db.Document{}, 0, nil
		},
	}

	app := fiber.New()
	app.Get("/wiki/documents/page", (&DocumentHandler{documentRepository: repo}).FindPage)

	req := httptest.NewRequest(http.MethodGet, "/wiki/documents/page?page=2&page_size=20&folder_id=folder-1&name=spec", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("expected request to succeed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
}

func TestDocumentHandlerFindListPassesFolderIDAndNameFilters(t *testing.T) {
	t.Parallel()

	repo := &stubDocumentRepository{
		listFn: func(_ context.Context, name, folderID string) ([]*wiki_db.Document, error) {
			if folderID != "folder-2" {
				t.Fatalf("expected folderID folder-2, got %q", folderID)
			}
			if name != "guide" {
				t.Fatalf("expected name guide, got %q", name)
			}
			return []*wiki_db.Document{
				{
					ID:          "doc-1",
					Filename:    "guide.md",
					FileSize:    1024,
					ContentType: "text/markdown",
					FolderID:    folderID,
					FilePath:    "documents/doc-1/guide.md",
					Status:      wiki_db.DocumentStatusPending,
				},
			}, nil
		},
	}

	app := fiber.New()
	app.Get("/wiki/documents/list", (&DocumentHandler{documentRepository: repo}).FindList)

	req := httptest.NewRequest(http.MethodGet, "/wiki/documents/list?folder_id=folder-2&name=guide", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("expected request to succeed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
}
