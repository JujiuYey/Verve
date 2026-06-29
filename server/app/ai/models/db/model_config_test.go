package db

import (
	"reflect"
	"strings"
	"testing"
)

func TestSystemModelTablesUseSysPrefix(t *testing.T) {
	t.Parallel()

	assertBunTable(t, SysModelPlatform{}, "sys_model_platforms")
	assertBunTable(t, SysModel{}, "sys_models")
}

func assertBunTable(t *testing.T, model any, tableName string) {
	t.Helper()

	field, ok := reflect.TypeOf(model).FieldByName("BaseModel")
	if !ok {
		t.Fatalf("%T should embed bun.BaseModel", model)
	}

	if !strings.Contains(string(field.Tag), "table:"+tableName) {
		t.Fatalf("expected %T to use table %q, got bun tag %q", model, tableName, field.Tag)
	}
}
