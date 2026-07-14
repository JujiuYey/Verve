package service

import (
	"context"
	"strings"
	"testing"

	wiki_db "verve/app/wiki/models/db"
)

type curatorChangeRequestStoreFake struct {
	existing *wiki_db.DocumentChangeRequest
}

func (f curatorChangeRequestStoreFake) CreateProposal(context.Context, *wiki_db.DocumentChangeRequest) error {
	return nil
}
func (f curatorChangeRequestStoreFake) FindChangeRequest(context.Context, string) (*wiki_db.DocumentChangeRequest, error) {
	return f.existing, nil
}

func TestCuratorRejectsDocumentsOverCodePointLimit(t *testing.T) {
	source := &fakeFeynmanDocumentSource{
		document: &wiki_db.Document{ID: "doc-1", Filename: "长文.md", CurrentVersion: 1},
		markdown: strings.Repeat("界", CuratorDocumentCharacterLimit+1),
	}
	service := newCuratorService(source, nil, func(context.Context, string) (string, error) {
		t.Fatal("agent must not run for an oversized document")
		return "", nil
	})

	if _, err := service.Propose(context.Background(), CuratorRequest{DocumentID: "doc-1", Instruction: "整理结构"}); err == nil {
		t.Fatal("oversized document must be rejected")
	}
}

func TestCuratorBuildsDeterministicUnifiedDiff(t *testing.T) {
	diff := buildUnifiedDiff("# 标题\n\n旧内容\n", "# 标题\n\n新内容\n")
	for _, want := range []string{"--- current.md", "+++ proposed.md", "-旧内容", "+新内容"} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff missing %q:\n%s", want, diff)
		}
	}
}

func TestCuratorRejectsReplacementFromAnotherDocument(t *testing.T) {
	replaces := "change-other"
	source := &fakeFeynmanDocumentSource{document: &wiki_db.Document{ID: "doc-1", Filename: "文档.md", CurrentVersion: 2}, markdown: "# 文档"}
	service := newCuratorService(source, curatorChangeRequestStoreFake{existing: &wiki_db.DocumentChangeRequest{ID: replaces, DocumentID: "doc-other"}}, func(context.Context, string) (string, error) {
		t.Fatal("agent must not run for an invalid replacement")
		return "", nil
	})
	_, err := service.Propose(context.Background(), CuratorRequest{DocumentID: "doc-1", ReplacesChangeRequestID: &replaces, Instruction: "重写"})
	if err == nil || !strings.Contains(err.Error(), "not replaceable") {
		t.Fatalf("error = %v", err)
	}
}
