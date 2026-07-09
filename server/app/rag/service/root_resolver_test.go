package service

import (
	"context"
	"fmt"
	"testing"

	wiki_db "verve/app/wiki/models/db"
)

type fakeFolderFinder map[string]*wiki_db.Folder

func (f fakeFolderFinder) FindOne(ctx context.Context, id string) (*wiki_db.Folder, error) {
	folder, ok := f[id]
	if !ok {
		return nil, fmt.Errorf("missing folder %s", id)
	}
	return folder, nil
}

func TestRootResolverBuildsRootAndPath(t *testing.T) {
	rootID := "root"
	childID := "child"
	resolver := NewRootResolver(fakeFolderFinder{
		rootID:  {ID: rootID, Name: "Go Tutorial"},
		childID: {ID: childID, Name: "Channel", ParentID: &rootID},
	})

	scope, err := resolver.Resolve(context.Background(), childID)
	if err != nil {
		t.Fatal(err)
	}
	if scope.RootFolderID != rootID {
		t.Fatalf("root id = %q", scope.RootFolderID)
	}
	if scope.FolderPath != "Go Tutorial/Channel" {
		t.Fatalf("folder path = %q", scope.FolderPath)
	}
}

func TestRootResolverDetectsCycle(t *testing.T) {
	a := "a"
	b := "b"
	resolver := NewRootResolver(fakeFolderFinder{
		a: {ID: a, Name: "A", ParentID: &b},
		b: {ID: b, Name: "B", ParentID: &a},
	})

	_, err := resolver.Resolve(context.Background(), a)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}
