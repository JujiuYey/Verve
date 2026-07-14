package llm

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	system_db "verve/app/system/models/db"
)

type fakeResolver struct {
	model    *system_db.SysModel
	plat     *system_db.SysModelPlatform
	err      error
	gotKey   string
	gotScene string
}

func (f *fakeResolver) FindAgentModelWithPlatform(_ context.Context, agentKey, sceneKey string) (*system_db.SysModel, *system_db.SysModelPlatform, error) {
	f.gotKey = agentKey
	f.gotScene = sceneKey
	if f.err != nil {
		return nil, nil, f.err
	}
	return f.model, f.plat, nil
}

func TestNewChatModelRequiresResolver(t *testing.T) {
	t.Parallel()
	if _, err := NewChatModel(context.Background(), nil, "coach", "default"); err == nil {
		t.Fatal("expected error when resolver is nil")
	}
	if _, err := NewStructuredChatModel(context.Background(), nil, "coach", "default"); err == nil {
		t.Fatal("expected error when resolver is nil")
	}
}

func TestNewChatModelWrapsMissingConfig(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{err: sql.ErrNoRows}
	_, err := NewChatModel(context.Background(), resolver, "coach", "default")
	if err == nil {
		t.Fatal("expected wrapped error")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected wrapped sql.ErrNoRows, got %v", err)
	}
	if !containsAll(err.Error(), []string{"coach", "default"}) {
		t.Fatalf("error should reference agent/scene keys, got %v", err)
	}
}

func TestNewChatModelUsesModelAndPlatformFromResolver(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{
		model: &system_db.SysModel{ID: "m1", ModelName: "gpt-test-1", Status: "active", PlatformID: "p1"},
		plat:  &system_db.SysModelPlatform{ID: "p1", BaseURL: "https://example.com/v1/", APIKeyCiphertext: "  secret-key  "},
	}
	if _, err := NewChatModel(context.Background(), resolver, "coach", "default"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolver.gotKey != "coach" || resolver.gotScene != "default" {
		t.Fatalf("resolver saw key=%q scene=%q", resolver.gotKey, resolver.gotScene)
	}
}

func TestNewStructuredChatModelWrapsMissingConfig(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{err: errors.New("boom")}
	_, err := NewStructuredChatModel(context.Background(), resolver, "learning_teacher", "default")
	if err == nil || !containsAll(err.Error(), []string{"learning_teacher", "default", "boom"}) {
		t.Fatalf("expected wrapped error referencing agent/scene and underlying cause, got %v", err)
	}
}

func TestNewStructuredChatModelUsesModelAndPlatformFromResolver(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{
		model: &system_db.SysModel{ID: "m1", ModelName: "structured-1", Status: "active", PlatformID: "p1"},
		plat:  &system_db.SysModelPlatform{ID: "p1", BaseURL: "https://example.com/v1/", APIKeyCiphertext: "k"},
	}
	if _, err := NewStructuredChatModel(context.Background(), resolver, "wiki_curator", "default"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
