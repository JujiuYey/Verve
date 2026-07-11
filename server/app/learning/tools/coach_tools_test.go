package tools

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	rag_payload "verve/app/rag/models/payload"
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

type stubCoachKnowledgeSearcher struct {
	called       bool
	rootFolderID string
	results      []rag_payload.SearchResult
}

func (s *stubCoachKnowledgeSearcher) Search(_ context.Context, rootFolderID, _ string, _ int) ([]rag_payload.SearchResult, error) {
	s.called = true
	s.rootFolderID = rootFolderID
	return s.results, nil
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

func TestSearchAccessibleWikiKnowledgeAllowsOwnedAndPublicRoots(t *testing.T) {
	owner := "user-1"
	other := "user-2"
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "owned", UserID: &owner},
		{ID: "public"},
		{ID: "foreign", UserID: &other},
	}}

	for _, rootFolderID := range []string{"owned", "public"} {
		t.Run(rootFolderID, func(t *testing.T) {
			documents := &stubCoachDocumentLister{documents: []*wiki_db.Document{{ID: "doc-visible", FolderID: rootFolderID}}}
			searcher := &stubCoachKnowledgeSearcher{results: []rag_payload.SearchResult{{DocumentID: "doc-visible", DocumentTitle: "Go", Content: "值有具体类型"}}}
			output, err := searchAccessibleWikiKnowledge(context.Background(), folders, documents, searcher, owner, &SearchWikiKnowledgeInput{
				RootFolderID: rootFolderID,
				Query:        "值和类型",
				Limit:        3,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !searcher.called || searcher.rootFolderID != rootFolderID {
				t.Fatalf("search call = called:%v root:%q", searcher.called, searcher.rootFolderID)
			}
			if len(output.Results) != 1 || output.Results[0]["content"] != "值有具体类型" {
				t.Fatalf("output = %#v", output)
			}
		})
	}
}

func TestSearchAccessibleWikiKnowledgeRejectsForeignMissingAndBlankRootsBeforeRetrieval(t *testing.T) {
	owner := "user-1"
	other := "user-2"
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "owned", UserID: &owner},
		{ID: "foreign", UserID: &other},
	}}

	for _, rootFolderID := range []string{"foreign", "missing", ""} {
		t.Run(rootFolderID, func(t *testing.T) {
			documents := &stubCoachDocumentLister{}
			searcher := &stubCoachKnowledgeSearcher{}
			_, err := searchAccessibleWikiKnowledge(context.Background(), folders, documents, searcher, owner, &SearchWikiKnowledgeInput{
				RootFolderID: rootFolderID,
				Query:        "secret",
			})
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("error = %v, want sql.ErrNoRows", err)
			}
			if searcher.called {
				t.Fatal("unauthorized root must be rejected before retrieval")
			}
		})
	}
}

func TestSearchAccessibleWikiKnowledgeFiltersPrivateChildDocumentsUnderPublicRoot(t *testing.T) {
	owner := "user-1"
	other := "user-2"
	publicRootID := "public-root"
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: publicRootID},
		{ID: "public-child", ParentID: &publicRootID},
		{ID: "private-child", ParentID: &publicRootID, UserID: &other},
	}}
	documents := &stubCoachDocumentLister{documents: []*wiki_db.Document{
		{ID: "doc-visible", FolderID: "public-child"},
		{ID: "doc-hidden", FolderID: "private-child"},
	}}
	searcher := &stubCoachKnowledgeSearcher{results: []rag_payload.SearchResult{
		{DocumentID: "doc-hidden", FolderID: "private-child", Content: "private secret"},
		{DocumentID: "doc-visible", FolderID: "public-child", Content: "public lesson"},
	}}

	output, err := searchAccessibleWikiKnowledge(context.Background(), folders, documents, searcher, owner, &SearchWikiKnowledgeInput{
		RootFolderID: publicRootID,
		Query:        "lesson",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(output.Results) != 1 || output.Results[0]["content"] != "public lesson" {
		t.Fatalf("output leaked inaccessible result: %#v", output)
	}
}
