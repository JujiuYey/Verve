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

func (s *stubCoachDocumentLister) List(context.Context, string, string) ([]*wiki_db.Document, error) {
	s.called = true
	return s.documents, nil
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

func TestLoadAccessibleCoachDocumentsReturnsAllDocumentsInKnownFolders(t *testing.T) {
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "owned"},
		{ID: "shared"},
	}}
	documents := &stubCoachDocumentLister{documents: []*wiki_db.Document{
		{ID: "doc-owned", FolderID: "owned"},
		{ID: "doc-shared", FolderID: "shared"},
		{ID: "doc-orphan", FolderID: "missing"},
	}}

	got, err := loadAccessibleCoachDocuments(context.Background(), folders, documents, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "doc-owned" || got[1].ID != "doc-shared" {
		t.Fatalf("documents = %#v", got)
	}
}

func TestLoadAccessibleCoachDocumentsRejectsUnknownFolderScope(t *testing.T) {
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{{ID: "known"}}}
	documents := &stubCoachDocumentLister{}

	_, err := loadAccessibleCoachDocuments(context.Background(), folders, documents, "foreign")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
	if documents.called {
		t.Fatal("unknown folder scope must be rejected before loading documents")
	}
}

func TestSearchAccessibleWikiKnowledgeReturnsResultsForKnownRoots(t *testing.T) {
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "owned"},
		{ID: "public"},
	}}

	for _, rootFolderID := range []string{"owned", "public"} {
		t.Run(rootFolderID, func(t *testing.T) {
			documents := &stubCoachDocumentLister{documents: []*wiki_db.Document{{ID: "doc-visible", FolderID: rootFolderID}}}
			searcher := &stubCoachKnowledgeSearcher{results: []rag_payload.SearchResult{{DocumentID: "doc-visible", DocumentTitle: "Go", Content: "值有具体类型"}}}
			output, err := searchAccessibleWikiKnowledge(context.Background(), folders, documents, searcher, &SearchWikiKnowledgeInput{
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

func TestSearchAccessibleWikiKnowledgeRejectsUnknownRootsBeforeRetrieval(t *testing.T) {
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "owned"},
	}}

	for _, rootFolderID := range []string{"foreign", "missing", ""} {
		t.Run(rootFolderID, func(t *testing.T) {
			documents := &stubCoachDocumentLister{}
			searcher := &stubCoachKnowledgeSearcher{}
			_, err := searchAccessibleWikiKnowledge(context.Background(), folders, documents, searcher, &SearchWikiKnowledgeInput{
				RootFolderID: rootFolderID,
				Query:        "secret",
			})
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("error = %v, want sql.ErrNoRows", err)
			}
			if searcher.called {
				t.Fatal("unknown root must be rejected before retrieval")
			}
		})
	}
}

func TestSearchAccessibleWikiKnowledgeReturnsAllIndexedMatchesInSingleTenant(t *testing.T) {
	// In single-tenant mode every Wiki folder and document is reachable, so the
	// tool returns every search hit whose document ID is known to the indexer.
	folders := stubCoachFolderLister{folders: []*wiki_db.Folder{
		{ID: "root-a"},
		{ID: "root-b"},
	}}
	documents := &stubCoachDocumentLister{documents: []*wiki_db.Document{
		{ID: "doc-a", FolderID: "root-a"},
		{ID: "doc-b", FolderID: "root-b"},
	}}
	searcher := &stubCoachKnowledgeSearcher{results: []rag_payload.SearchResult{
		{DocumentID: "doc-b", Content: "lesson in B"},
		{DocumentID: "doc-a", Content: "lesson in A"},
	}}

	output, err := searchAccessibleWikiKnowledge(context.Background(), folders, documents, searcher, &SearchWikiKnowledgeInput{
		RootFolderID: "root-a",
		Query:        "lesson",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(output.Results) != 2 {
		t.Fatalf("expected both matches to be returned in single-tenant: %#v", output)
	}
}
