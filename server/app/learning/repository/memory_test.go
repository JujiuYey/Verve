package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type emptyMemoryRows struct{}

func (emptyMemoryRows) Columns() []string         { return []string{"id"} }
func (emptyMemoryRows) Close() error              { return nil }
func (emptyMemoryRows) Next([]driver.Value) error { return io.EOF }

type memoryEventRows struct {
	id   string
	done bool
}

func (r *memoryEventRows) Columns() []string { return []string{"id"} }
func (r *memoryEventRows) Close() error      { return nil }
func (r *memoryEventRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	dest[0] = r.id
	r.done = true
	return nil
}

func TestMemoryRepositoryCreateEventReusesIdempotentSourceEventID(t *testing.T) {
	recorder := &reviewQueryRecorder{rows: &memoryEventRows{id: "event-existing"}}
	sqldb := sql.OpenDB(recorder)
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()
	repo := NewMemoryRepository(db)
	sourceID := "review-1"
	event := &learning_db.LearningMemoryEvent{
		UserID: "user-1", SourceType: "explanation_review", SourceID: &sourceID,
		EventType: "explanation_review", Content: "understood types", Evidence: map[string]interface{}{},
	}

	if err := repo.CreateEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if event.ID != "event-existing" {
		t.Fatalf("event ID = %q", event.ID)
	}
	for _, want := range []string{
		"ON CONFLICT (source_type, source_id, event_type)",
		"DO UPDATE SET id = learning_memory_events.id",
		"RETURNING id",
	} {
		if !strings.Contains(recorder.selectQuery, want) {
			t.Fatalf("event insert missing %q: %s", want, recorder.selectQuery)
		}
	}
}

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
