package repository

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"
)

func TestApplyVersionInputExposesImmutableRevisionData(t *testing.T) {
	t.Parallel()

	typeOfInput := reflect.TypeOf(ApplyVersionInput{})
	for _, name := range []string{
		"ChangeRequestID", "DocumentID", "ExpectedVersion", "ObjectPath", "ContentHash", "FileSize", "ChangedBy", "ChangeSummary",
	} {
		if _, ok := typeOfInput.FieldByName(name); !ok {
			t.Fatalf("ApplyVersionInput.%s is missing", name)
		}
	}
}

func TestVersionRepositoryLocksDocumentAndChangeRequest(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("version.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`Where("id = ?", *input.ChangeRequestID).For("UPDATE")`,
		`Where("id = ?", input.DocumentID).For("UPDATE")`,
	} {
		if !strings.Contains(string(source), want) {
			t.Fatalf("version transaction must retain row lock %q", want)
		}
	}
}

func TestVersionRepositoryExposesConflictAndChangeRequestErrors(t *testing.T) {
	t.Parallel()

	for _, err := range []error{ErrVersionConflict, ErrChangeRequestNotProposed, ErrChangeRequestForbidden} {
		if err == nil {
			t.Fatal("version repository sentinel errors must be non-nil")
		}
	}
}

func TestCaptureVersionConflictAllowsTransactionToCommit(t *testing.T) {
	conflicted := false
	if err := captureVersionConflict(ErrVersionConflict, &conflicted); err != nil {
		t.Fatalf("transaction callback error = %v", err)
	}
	if !conflicted {
		t.Fatal("version conflict was not captured for return after commit")
	}
	other := errors.New("database unavailable")
	if err := captureVersionConflict(other, &conflicted); !errors.Is(err, other) {
		t.Fatalf("non-conflict error = %v", err)
	}
}

func TestVersionRepositoryContracts(t *testing.T) {
	t.Parallel()

	var revisions *RevisionRepository
	var requests *ChangeRequestRepository
	var versions *VersionRepository

	var _ func(context.Context, *wiki_db.Document, *wiki_db.DocumentRevision, *rag_db.IndexJob) error = revisions.CreateInitial
	var _ func(context.Context, *LegacyRevisionSnapshot) error = revisions.EnsureLegacyRevision
	var _ func(context.Context, *wiki_db.DocumentChangeRequest) error = requests.CreateProposal
	var _ func(context.Context, string) (*wiki_db.DocumentChangeRequest, error) = requests.FindChangeRequest
	var _ func(context.Context, string) ([]string, error) = revisions.ListRevisionObjectPaths
	var _ func(context.Context, ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) = versions.ApplyChangeRequest
	var _ func(context.Context, ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) = versions.ApplyDirectEdit
	var _ func(context.Context, string, string) error = requests.CancelChangeRequest

	if errors.Is(ErrVersionConflict, ErrChangeRequestNotProposed) {
		t.Fatal("version errors must remain distinguishable")
	}
}
