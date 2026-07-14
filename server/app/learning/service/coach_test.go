package service

import (
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
)

func TestBuildCoachQueryIncludesRuntimeContext(t *testing.T) {
	folderID := "folder-go"
	docID := "doc-interface"

	ctx := CoachRuntimeContext{
		Folders: []*wiki_db.Folder{
			{ID: folderID, Name: "Go 基础"},
		},
		Documents: []*wiki_db.Document{
			{ID: docID, FolderID: folderID, Filename: "interface.md"},
		},
		MemoryItems: []*learning_db.LearningMemoryItem{
			{Kind: "explanation_evidence", Statement: "能解释接口的隐式实现"},
		},
	}

	query := BuildCoachQuery(ctx, "继续学习")

	for _, want := range []string{
		"学习者说:继续学习",
		"Go 基础",
		"interface.md",
		"能解释接口的隐式实现",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %q:\n%s", want, query)
		}
	}
	for _, forbidden := range []string{"objective", "学习小节", "mastery", "掌握度", "create_learning_objectives", "first_objective_id"} {
		if strings.Contains(query, forbidden) {
			t.Fatalf("query contains removed concept %q:\n%s", forbidden, query)
		}
	}
}

func TestBuildCoachQueryNavigatesDirectlyToDocuments(t *testing.T) {
	ctx := CoachRuntimeContext{
		Folders: []*wiki_db.Folder{
			{ID: "folder-go", Name: "Go 基础"},
		},
		Documents: []*wiki_db.Document{
			{ID: "doc-hello", FolderID: "folder-go", Filename: "hello.md"},
		},
	}

	query := BuildCoachQuery(ctx, "继续学习")

	for _, want := range []string{"doc-hello", `"document_id":"..."`} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %q:\n%s", want, query)
		}
	}
}

func TestBuildCoachQueryIncludesExplicitEmptyStates(t *testing.T) {
	query := BuildCoachQuery(CoachRuntimeContext{}, "继续学习")

	for _, want := range []string{
		"- 暂无文件夹",
		"- 暂无文档",
		"- 暂无学习记忆",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %q:\n%s", want, query)
		}
	}
}

func TestParseCoachActionFindsNavigateAction(t *testing.T) {
	content := "我们继续接口基础。\n\n<ACTION>{\"type\":\"navigate_to_practice\",\"document_id\":\"doc-interface\",\"label\":\"开始费曼练习\"}</ACTION>"

	action := ParseCoachAction(content)

	if action == nil {
		t.Fatalf("expected action")
	}
	if action.Type != "navigate_to_practice" {
		t.Fatalf("type = %q", action.Type)
	}
	if action.DocumentID != "doc-interface" {
		t.Fatalf("document id = %q", action.DocumentID)
	}
	if action.Label != "开始费曼练习" {
		t.Fatalf("label = %q", action.Label)
	}
}

func TestParseCoachActionRejectsNavigateActionWithoutDocument(t *testing.T) {
	for _, content := range []string{
		`<ACTION>{"type":"navigate_to_practice","label":"开始费曼练习"}</ACTION>`,
		`<ACTION>{"type":"navigate_to_practice","objective_id":"obj-legacy","label":"进入练习"}</ACTION>`,
	} {
		if action := ParseCoachAction(content); action != nil {
			t.Fatalf("expected invalid navigation action to be rejected: %#v", action)
		}
	}
}

func TestCoachFolderAccessibilityUsesProvidedList(t *testing.T) {
	folders := []*wiki_db.Folder{
		{ID: "owned"},
		{ID: "shared"},
	}
	if !CoachFolderIsAccessible(folders, "owned") || !CoachFolderIsAccessible(folders, "shared") {
		t.Fatal("folders provided to the agent should all be reachable")
	}
	if CoachFolderIsAccessible(folders, "missing") || CoachFolderIsAccessible(folders, "") {
		t.Fatal("folders not in the supplied list should be unreachable")
	}
}

func TestParseCoachActionForDocumentsRequiresNavigateTypeAndScopedDocument(t *testing.T) {
	documents := []*wiki_db.Document{{ID: "doc-visible"}}

	valid := `<ACTION>{"type":"navigate_to_practice","document_id":"doc-visible","label":"开始费曼练习"}</ACTION>`
	if action := ParseCoachActionForDocuments(valid, documents); action == nil || action.DocumentID != "doc-visible" {
		t.Fatalf("valid action = %#v", action)
	}

	invalid := []string{
		`<ACTION>{"type":"navigate_to_practice","document_id":"doc-foreign"}</ACTION>`,
		`<ACTION>{"type":"open_wiki","document_id":"doc-visible"}</ACTION>`,
		`<ACTION>{"type":"navigate_to_practice"}</ACTION>`,
	}
	for _, content := range invalid {
		if action := ParseCoachActionForDocuments(content, documents); action != nil {
			t.Fatalf("out-of-scope action should be rejected: %#v", action)
		}
	}
}
