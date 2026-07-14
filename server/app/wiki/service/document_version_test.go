package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
)

var _ = errors.Is

type versionFileStoreFake struct {
	calls   *[]string
	content string
	paths   []string
}

func (f *versionFileStoreFake) UploadFile(_ context.Context, path string, _ io.Reader, _ int64, _ string) error {
	*f.calls = append(*f.calls, "storage")
	f.paths = append(f.paths, path)
	return nil
}
func (f *versionFileStoreFake) GetFileContent(_ context.Context, _ string) (string, error) {
	return f.content, nil
}

type versionDocumentFinderFake struct{ document *wiki_db.Document }

func (f versionDocumentFinderFake) FindOne(context.Context, string) (*wiki_db.Document, error) {
	return f.document, nil
}

type versionRevisionWriterFake struct {
	calls   *[]string
	initial *wiki_db.Document
}

func (f *versionRevisionWriterFake) CreateInitial(_ context.Context, document *wiki_db.Document, _ *wiki_db.DocumentRevision, _ *rag_db.IndexJob) error {
	*f.calls = append(*f.calls, "database")
	f.initial = document
	return nil
}
func (f *versionRevisionWriterFake) EnsureLegacyRevision(context.Context, *wiki_repo.LegacyRevisionSnapshot) error {
	return nil
}

type versionPublisherFake struct {
	calls []string
	input wiki_repo.ApplyVersionInput
}

func (f *versionPublisherFake) ApplyDirectEdit(_ context.Context, input wiki_repo.ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	f.calls = append(f.calls, "database")
	f.input = input
	return &wiki_db.DocumentRevision{DocumentID: input.DocumentID, Version: input.ExpectedVersion + 1}, &rag_db.IndexJob{ID: "job-2"}, nil
}
func (f *versionPublisherFake) ApplyChangeRequest(_ context.Context, input wiki_repo.ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	f.calls = append(f.calls, "publisher:change_request")
	f.input = input
	return &wiki_db.DocumentRevision{DocumentID: input.DocumentID, Version: input.ExpectedVersion + 1}, &rag_db.IndexJob{ID: "job-3"}, nil
}
func (f *versionPublisherFake) FindAppliedChangeRequest(context.Context, string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	return nil, nil, errors.New("not used")
}

type versionRequestFinderFake struct {
	request *wiki_db.DocumentChangeRequest
}

func (f versionRequestFinderFake) FindChangeRequest(context.Context, string) (*wiki_db.DocumentChangeRequest, error) {
	return f.request, nil
}

func TestRevisionObjectPathUsesDocumentAndVersion(t *testing.T) {
	t.Parallel()

	got := revisionObjectPath("doc-1", 2, "notes.md")
	if got != "documents/doc-1/revisions/2/notes.md" {
		t.Fatalf("path = %q", got)
	}
}

func TestCreateInitialWritesImmutableObjectBeforeDatabase(t *testing.T) {
	calls := make([]string, 0, 2)
	files := &versionFileStoreFake{calls: &calls}
	revisions := &versionRevisionWriterFake{calls: &calls}
	service := NewDocumentVersionService(revisions, nil, nil, files)

	document, job, err := service.CreateInitial(context.Background(), InitialDocumentInput{
		FolderID: "folder-1", Filename: "notes.md", Content: []byte("# Go\n"), ContentType: "text/markdown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if document.CurrentVersion != 1 || document.ContentHash == nil || *document.ContentHash != "13ad8e753213f8ac6d0e7917677a03eac600790d616194982eb3c3040ab07665" {
		t.Fatalf("document = %#v", document)
	}
	if job.DocumentVersion != 1 || job.ObjectPath != document.FilePath {
		t.Fatalf("job = %#v document=%#v", job, document)
	}
	if len(files.paths) != 1 || !strings.Contains(files.paths[0], "/revisions/1/") {
		t.Fatalf("paths = %#v", files.paths)
	}
	if strings.Join(calls, ",") != "storage,database" {
		t.Fatalf("call order = %#v", calls)
	}
}

func TestApplyChangeRequestAppliesProposedChangeWithoutUserGate(t *testing.T) {
	calls := make([]string, 0)
	files := &versionFileStoreFake{calls: &calls}
	publisher := &versionPublisherFake{}
	service := NewDocumentVersionService(
		&versionRevisionWriterFake{},
		publisher,
		versionDocumentFinderFake{document: &wiki_db.Document{ID: "doc-1", CurrentVersion: 1}},
		files,
		versionRequestFinderFake{request: &wiki_db.DocumentChangeRequest{
			ID: "request-1", Status: wiki_db.ChangeRequestStatusProposed,
		}},
	)

	if _, _, err := service.ApplyChangeRequest(context.Background(), "request-1"); err != nil {
		t.Fatalf("apply = %v", err)
	}
	if publisher.calls == nil {
		t.Fatal("publisher was not invoked")
	}
}

func TestSaveDirectEditPublishesNextImmutableVersion(t *testing.T) {
	calls := make([]string, 0, 2)
	hash := strings.Repeat("a", 64)
	files := &versionFileStoreFake{calls: &calls}
	publisher := &versionPublisherFake{}
	service := NewDocumentVersionService(nil, publisher, versionDocumentFinderFake{document: &wiki_db.Document{
		ID: "doc-1", Filename: "notes.md", CurrentVersion: 1, ContentHash: &hash,
	}}, files)

	revision, _, err := service.SaveDirectEdit(context.Background(), "doc-1", "# Updated\n")
	if err != nil {
		t.Fatal(err)
	}
	if revision.Version != 2 || publisher.input.ExpectedVersion != 1 {
		t.Fatalf("revision=%#v input=%#v", revision, publisher.input)
	}
	if publisher.input.ObjectPath != "documents/doc-1/revisions/2/notes.md" {
		t.Fatalf("object path = %q", publisher.input.ObjectPath)
	}
	if strings.Join(calls, ",") != "storage" || len(files.paths) != 1 {
		t.Fatalf("storage calls = %#v paths=%#v", calls, files.paths)
	}
}
