package db

import (
	"reflect"
	"strings"
	"testing"
)

func TestWikiModelsExposeSortOrder(t *testing.T) {
	t.Parallel()

	assertSortOrderField(t, Folder{})
	assertSortOrderField(t, Document{})
}

func TestDocumentVersionModels(t *testing.T) {
	t.Parallel()

	assertField(t, Document{}, "CurrentVersion", `bun:"current_version,notnull"`)
	assertField(t, DocumentRevision{}, "ObjectPath", `bun:"object_path,notnull"`)
	assertField(t, DocumentRevision{}, "ContentHash", `bun:"content_hash,notnull"`)
	assertField(t, DocumentChangeRequest{}, "BaseVersion", `bun:"base_version,notnull"`)
	assertField(t, DocumentChangeRequest{}, "Status", `bun:"status,notnull"`)

	statuses := map[string]bool{
		ChangeRequestStatusProposed:  true,
		ChangeRequestStatusApplied:   true,
		ChangeRequestStatusFailed:    true,
		ChangeRequestStatusCancelled: true,
		ChangeRequestStatusConflict:  true,
	}
	if len(statuses) != 5 {
		t.Fatalf("change-request statuses must remain distinct: %#v", statuses)
	}
}

func assertSortOrderField(t *testing.T, model any) {
	t.Helper()

	field, ok := reflect.TypeOf(model).FieldByName("SortOrder")
	if !ok {
		t.Fatalf("%T should expose SortOrder", model)
	}

	if !strings.Contains(string(field.Tag), `bun:"sort_order,notnull"`) {
		t.Fatalf("expected %T SortOrder bun tag to map sort_order, got %q", model, field.Tag)
	}

	if !strings.Contains(string(field.Tag), `json:"sort_order"`) {
		t.Fatalf("expected %T SortOrder json tag to be sort_order, got %q", model, field.Tag)
	}
}

func assertField(t *testing.T, model any, name, expectedTag string) {
	t.Helper()

	field, ok := reflect.TypeOf(model).FieldByName(name)
	if !ok {
		t.Fatalf("%T should expose %s", model, name)
	}
	if !strings.Contains(string(field.Tag), expectedTag) {
		t.Fatalf("expected %T %s bun tag to contain %q, got %q", model, name, expectedTag, field.Tag)
	}
}
