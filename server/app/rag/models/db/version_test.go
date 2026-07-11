package db

import (
	"reflect"
	"strings"
	"testing"
)

func TestVersionedRAGModels(t *testing.T) {
	t.Parallel()

	assertVersionField(t, IndexJob{}, "DocumentVersion", `bun:"document_version,notnull"`)
	assertVersionField(t, IndexJob{}, "ObjectPath", `bun:"object_path,notnull"`)
	assertVersionField(t, WikiChunk{}, "DocumentVersion", `bun:"document_version,notnull"`)
}

func assertVersionField(t *testing.T, model any, name, expectedTag string) {
	t.Helper()

	field, ok := reflect.TypeOf(model).FieldByName(name)
	if !ok {
		t.Fatalf("%T should expose %s", model, name)
	}
	if !strings.Contains(string(field.Tag), expectedTag) {
		t.Fatalf("expected %T %s bun tag to contain %q, got %q", model, name, expectedTag, field.Tag)
	}
}
