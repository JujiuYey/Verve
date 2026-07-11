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
