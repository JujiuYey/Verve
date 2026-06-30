package handlers

import (
	"testing"

	learning_db "sag-wiki/app/learning/models/db"
	wiki_db "sag-wiki/app/wiki/models/db"
)

func TestBuildStagesFromFolderStructureUsesTopLevelFolders(t *testing.T) {
	rootID := "root"
	stageOneID := "stage-one"
	stageTwoID := "stage-two"
	nestedID := "nested"

	root := &wiki_db.Folder{ID: rootID, Name: "golang"}
	folders := []*wiki_db.Folder{
		root,
		{ID: stageTwoID, Name: "02-类型状态与控制流", ParentID: &rootID},
		{ID: stageOneID, Name: "01-工具链与HelloWorld", ParentID: &rootID},
		{ID: nestedID, Name: "深入补充", ParentID: &stageTwoID},
	}
	docs := []*wiki_db.Document{
		{ID: "doc-root", Filename: "00-课程说明.md", FolderID: rootID},
		{ID: "doc-2", Filename: "02-变量声明赋值零值与作用域.md", FolderID: nestedID},
		{ID: "doc-1", Filename: "01-工具链与HelloWorld.md", FolderID: stageOneID},
	}
	goal := &learning_db.LearningGoal{UserID: "user-1"}

	stages := buildStagesFromFolderStructure(goal, root, folders, docs)
	if len(stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(stages))
	}

	if stages[0].Title != "基础资料" {
		t.Fatalf("expected root docs to become 基础资料 stage, got %q", stages[0].Title)
	}
	if got := stages[0].Objectives[0].Title; got != "课程说明" {
		t.Fatalf("expected cleaned root doc title, got %q", got)
	}
	if got := stages[0].Objectives[0].SourceDocumentID; got == nil || *got != "doc-root" {
		t.Fatalf("expected root objective source document id, got %v", got)
	}
	if got := stages[0].Objectives[0].SourceFolderID; got == nil || *got != rootID {
		t.Fatalf("expected root objective source folder id, got %v", got)
	}
	if got := stages[0].Objectives[0].SourceFolderPath; got == nil || *got != "golang" {
		t.Fatalf("expected root objective source folder path, got %v", got)
	}

	if stages[1].Title != "01-工具链与HelloWorld" {
		t.Fatalf("expected first numbered folder stage, got %q", stages[1].Title)
	}
	if got := stages[1].Objectives[0].Title; got != "工具链与HelloWorld" {
		t.Fatalf("expected cleaned stage doc title, got %q", got)
	}
	if got := stages[1].Objectives[0].SourceDocumentID; got == nil || *got != "doc-1" {
		t.Fatalf("expected stage objective source document id, got %v", got)
	}
	if got := stages[1].Objectives[0].SourceFolderID; got == nil || *got != stageOneID {
		t.Fatalf("expected stage objective source folder id, got %v", got)
	}
	if got := stages[1].Objectives[0].SourceFolderPath; got == nil || *got != "golang / 01-工具链与HelloWorld" {
		t.Fatalf("expected stage objective source folder path, got %v", got)
	}

	if stages[2].Title != "02-类型状态与控制流" {
		t.Fatalf("expected nested doc to map to top-level stage, got %q", stages[2].Title)
	}
	if got := *stages[2].Objectives[0].Detail; got != "来源目录: golang / 02-类型状态与控制流 / 深入补充\n来源文档: 02-变量声明赋值零值与作用域.md" {
		t.Fatalf("unexpected nested doc detail: %q", got)
	}
	if got := stages[2].Objectives[0].SourceDocumentID; got == nil || *got != "doc-2" {
		t.Fatalf("expected nested objective source document id, got %v", got)
	}
	if got := stages[2].Objectives[0].SourceFolderID; got == nil || *got != nestedID {
		t.Fatalf("expected nested objective source folder id, got %v", got)
	}
	if got := stages[2].Objectives[0].SourceFolderPath; got == nil || *got != "golang / 02-类型状态与控制流 / 深入补充" {
		t.Fatalf("expected nested objective source folder path, got %v", got)
	}
}
