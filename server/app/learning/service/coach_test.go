package service

import (
	"strings"
	"testing"
	"time"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
)

func TestBuildCoachQueryIncludesRuntimeContext(t *testing.T) {
	level := "explained"
	nextGoal := "继续讲接口组合"
	learned := "接口基础"
	nextStep := "复述 interface 的隐式实现"
	folderID := "folder-go"
	docID := "doc-interface"
	detail := "理解接口定义和实现关系"

	ctx := CoachRuntimeContext{
		UserID: "user-1",
		Folders: []*wiki_db.Folder{
			{ID: folderID, Name: "Go 基础"},
		},
		Documents: []*wiki_db.Document{
			{ID: docID, FolderID: folderID, Filename: "interface.md"},
		},
		Objectives: []*learning_db.LearningObjective{
			{
				ID:               "obj-interface",
				Title:            "接口基础",
				Detail:           &detail,
				SourceFolderID:   &folderID,
				SourceDocumentID: &docID,
				Status:           "active",
				MasteryLevel:     "heard",
			},
		},
		Profiles: []*learning_db.LearningProfile{
			{
				FolderID:        folderID,
				CurrentLevel:    &level,
				CompletedTopics: []string{"值类型"},
				WeakPoints:      []string{"缺少反例"},
				NextGoal:        &nextGoal,
			},
		},
		Journals: []*learning_db.LearningJournal{
			{
				FolderID: folderID,
				Date:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
				Learned:  &learned,
				NextStep: &nextStep,
			},
		},
	}

	query := BuildCoachQuery(ctx, "继续学习")

	for _, want := range []string{
		"学习者说:继续学习",
		"Go 基础",
		"interface.md",
		"接口基础",
		"当前水平:explained",
		"缺少反例",
		"复述 interface 的隐式实现",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %q:\n%s", want, query)
		}
	}
}

func TestBuildCoachQueryTellsAgentToGenerateObjectivesWhenDocumentsExist(t *testing.T) {
	ctx := CoachRuntimeContext{
		UserID: "user-1",
		Folders: []*wiki_db.Folder{
			{ID: "folder-go", Name: "Go 基础"},
		},
		Documents: []*wiki_db.Document{
			{ID: "doc-hello", FolderID: "folder-go", Filename: "hello.md"},
		},
	}

	query := BuildCoachQuery(ctx, "继续学习")

	for _, want := range []string{
		"create_learning_objectives",
		"doc-hello",
		"生成学习小节",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %q:\n%s", want, query)
		}
	}
}

func TestParseCoachActionFindsNavigateAction(t *testing.T) {
	content := "我们继续接口基础。\n\n<ACTION>{\"type\":\"navigate_to_practice\",\"objective_id\":\"obj-interface\",\"label\":\"进入练习\"}</ACTION>"

	action := ParseCoachAction(content)

	if action == nil {
		t.Fatalf("expected action")
	}
	if action.Type != "navigate_to_practice" {
		t.Fatalf("type = %q", action.Type)
	}
	if action.ObjectiveID != "obj-interface" {
		t.Fatalf("objective id = %q", action.ObjectiveID)
	}
	if action.Label != "进入练习" {
		t.Fatalf("label = %q", action.Label)
	}
}
