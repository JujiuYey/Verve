package tools

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	wiki_db "verve/app/wiki/models/db"
)

type stubCoachFolderLister struct {
	folders []*wiki_db.Folder
}

func (s stubCoachFolderLister) List(context.Context, map[string]interface{}) ([]*wiki_db.Folder, error) {
	return s.folders, nil
}

type stubCoachDocumentLister struct {
	documents []*wiki_db.Document
	called    bool
}

func (s *stubCoachDocumentLister) List(context.Context, string, string) ([]*wiki_db.Document, error) {
	s.called = true
	return s.documents, nil
}

func TestLoadAccessibleCoachDocumentsFiltersForeignFolders(t *testing.T) {
	owner := "user-1"
	other := "user-2"
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "owned", UserID: &owner},
		{ID: "foreign", UserID: &other},
		{ID: "public"},
	}}
	documents := &stubCoachDocumentLister{documents: []*wiki_db.Document{
		{ID: "doc-owned", FolderID: "owned"},
		{ID: "doc-foreign", FolderID: "foreign"},
		{ID: "doc-public", FolderID: "public"},
	}}

	got, err := loadAccessibleCoachDocuments(context.Background(), folders, documents, owner, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "doc-owned" || got[1].ID != "doc-public" {
		t.Fatalf("documents = %#v", got)
	}
}

func TestLoadAccessibleCoachDocumentsRejectsForeignFolderScope(t *testing.T) {
	owner := "user-1"
	other := "user-2"
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{{ID: "foreign", UserID: &other}}}
	documents := &stubCoachDocumentLister{}

	_, err := loadAccessibleCoachDocuments(context.Background(), folders, documents, owner, "foreign")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
	if documents.called {
		t.Fatal("foreign folder scope must be rejected before loading documents")
	}
}
