package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type emptyMemoryRows struct{}

func (emptyMemoryRows) Columns() []string         { return []string{"id"} }
func (emptyMemoryRows) Close() error              { return nil }
func (emptyMemoryRows) Next([]driver.Value) error { return io.EOF }

func TestMemoryRepositoryFindItemsByDocumentScopesUserAndDocument(t *testing.T) {
	recorder := &reviewQueryRecorder{rows: emptyMemoryRows{}}
	sqldb := sql.OpenDB(recorder)
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()
	repo := NewMemoryRepository(db)

	items, err := repo.FindItemsByDocument(context.Background(), "user-1", "doc-1", 12)
	if err != nil {
		t.Fatal(err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("items = %#v", items)
	}
	for _, want := range []string{
		`WHERE (user_id = 'user-1') AND (document_id = 'doc-1')`,
		`ORDER BY "last_seen_at" DESC`, `LIMIT 12`,
	} {
		if !strings.Contains(recorder.selectQuery, want) {
			t.Fatalf("query does not contain %q: %s", want, recorder.selectQuery)
		}
	}
}

func TestMemoryRepositoryFindItemsByFoldersScopesUserAndFolderSet(t *testing.T) {
	recorder := &reviewQueryRecorder{rows: emptyMemoryRows{}}
	sqldb := sql.OpenDB(recorder)
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()
	repo := NewMemoryRepository(db)

	items, err := repo.FindItemsByFolders(context.Background(), "user-1", []string{"root", "child"}, 15)
	if err != nil {
		t.Fatal(err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("items = %#v", items)
	}
	for _, want := range []string{
		`WHERE (user_id = 'user-1') AND (folder_id IN ('root', 'child'))`,
		`ORDER BY "last_seen_at" DESC`, `LIMIT 15`,
	} {
		if !strings.Contains(recorder.selectQuery, want) {
			t.Fatalf("query does not contain %q: %s", want, recorder.selectQuery)
		}
	}
}
