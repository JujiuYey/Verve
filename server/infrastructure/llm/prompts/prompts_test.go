package prompts

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultAgentPromptsContainCriticalContracts(t *testing.T) {
	tests := []struct {
		name   string
		render func(Input) string
		want   []string
	}{
		{
			name:   "guide",
			render: GuidePrompt,
			want: []string{
				"只输出 JSON",
				`"practice_points"`,
				"顶层第一个字符必须是 {",
			},
		},
		{
			name:   "objective generator",
			render: ObjectiveGeneratorPrompt,
			want: []string{
				"只输出 JSON",
				`"objectives"`,
				"通常生成 3-8 个小节",
			},
		},
		{
			name:   "coach",
			render: CoachPrompt,
			want: []string{
				"学习调度 agent",
				"<ACTION>",
				`"navigate_to_practice"`,
			},
		},
		{
			name:   "tutor",
			render: TutorPrompt,
			want: []string{
				"费曼式编程陪练",
				"每次只推进一个认知点",
				"不连续追问",
			},
		},
		{
			name:   "examiner",
			render: ExaminerPrompt,
			want: []string{
				"只输出 JSON",
				`"verdict":"pass|partial|fail"`,
				"review_required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := tt.render(Input{PresetKey: DefaultPresetKey})
			if strings.TrimSpace(prompt) == "" {
				t.Fatalf("prompt is empty")
			}
			for _, want := range tt.want {
				if !strings.Contains(prompt, want) {
					t.Fatalf("prompt does not contain %q:\n%s", want, prompt)
				}
			}
		})
	}
}

func TestAgentPromptsFallBackToDefaultPreset(t *testing.T) {
	tests := []struct {
		name   string
		render func(Input) string
	}{
		{name: "guide", render: GuidePrompt},
		{name: "objective generator", render: ObjectiveGeneratorPrompt},
		{name: "coach", render: CoachPrompt},
		{name: "tutor", render: TutorPrompt},
		{name: "examiner", render: ExaminerPrompt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultPrompt := tt.render(Input{PresetKey: DefaultPresetKey})
			emptyPresetPrompt := tt.render(Input{})
			unknownPresetPrompt := tt.render(Input{PresetKey: "future-preset"})

			if emptyPresetPrompt != defaultPrompt {
				t.Fatalf("empty preset did not fall back to default")
			}
			if unknownPresetPrompt != defaultPrompt {
				t.Fatalf("unknown preset did not fall back to default")
			}
		})
	}
}

func TestCoachQueryPromptRendersRuntimeContext(t *testing.T) {
	prompt := CoachQueryPrompt(CoachQueryInput{
		Message: "  继续学习  ",
		Folders: []CoachFolder{
			{ID: "folder-go", Name: "Go 基础", Description: "接口和组合"},
		},
		Documents: []CoachDocument{
			{ID: "doc-interface", FolderID: "folder-go", Filename: "interface.md"},
		},
		Objectives: []CoachObjective{
			{
				ID:               "obj-interface",
				Title:            "接口基础",
				Detail:           "理解接口定义和实现关系",
				SourceFolderID:   "folder-go",
				SourceDocumentID: "doc-interface",
				Status:           "active",
				MasteryLevel:     "heard",
			},
		},
		MemoryItems: []CoachMemoryItem{
			{
				FolderID:   "folder-go",
				Kind:       "mastered_concept",
				Statement:  "用户已经能用自己的话解释 Go 接口的隐式实现",
				Confidence: "confirmed",
			},
		},
		Profiles: []CoachProfile{
			{
				FolderID:        "folder-go",
				CurrentLevel:    "explained",
				CompletedTopics: []string{"值类型"},
				WeakPoints:      []string{"缺少反例"},
				NextGoal:        "继续讲接口组合",
			},
		},
		Journals: []CoachJournal{
			{
				FolderID: "folder-go",
				Date:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
				Learned:  "接口基础",
				NextStep: "复述 interface 的隐式实现",
			},
		},
	})

	for _, want := range []string{
		"学习者说:继续学习",
		"Go 基础 (folder-go) - 接口和组合",
		"interface.md (doc-interface), folder=folder-go",
		"接口基础 (obj-interface), status=active, mastery=heard, folder=folder-go, document=doc-interface",
		"要点:理解接口定义和实现关系",
		"## 学习记忆",
		"已掌握",
		"用户已经能用自己的话解释 Go 接口的隐式实现",
		"当前水平:explained",
		"已掌握内容:值类型",
		"上次建议:复述 interface 的隐式实现",
		"<ACTION>",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt does not contain %q:\n%s", want, prompt)
		}
	}
	if strings.Index(prompt, "## 学习记忆") > strings.Index(prompt, "## 学习画像") {
		t.Fatalf("learning memory should appear before profiles:\n%s", prompt)
	}
}

func TestCoachQueryPromptRendersExplicitEmptyStates(t *testing.T) {
	prompt := CoachQueryPrompt(CoachQueryInput{Message: "继续学习"})

	for _, want := range []string{
		"- 暂无文件夹",
		"- 暂无文档",
		"- 暂无学习小节。可以建议用户先去 Wiki 补充资料。",
		"- 暂无学习记忆",
		"- 暂无画像",
		"- 暂无记录",
		"如果不能确定,只问用户一个选择题。",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt does not contain %q:\n%s", want, prompt)
		}
	}
}

func TestCoachQueryPromptSkipsBlankMemoryStatements(t *testing.T) {
	prompt := CoachQueryPrompt(CoachQueryInput{
		Message: "继续学习",
		MemoryItems: []CoachMemoryItem{
			{Kind: "mastered_concept", Statement: "  "},
		},
	})

	if !strings.Contains(prompt, "- 暂无学习记忆") {
		t.Fatalf("prompt should render empty memory state:\n%s", prompt)
	}
}

func TestCoachQueryPromptUsesFallbackMemoryKindLabel(t *testing.T) {
	prompt := CoachQueryPrompt(CoachQueryInput{
		Message: "继续学习",
		MemoryItems: []CoachMemoryItem{
			{Statement: "用户能解释值与类型"},
			{Kind: "custom_fact", Statement: "用户偏好先看例子"},
		},
	})

	for _, want := range []string{
		"记忆: 用户能解释值与类型",
		"custom_fact: 用户偏好先看例子",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt does not contain %q:\n%s", want, prompt)
		}
	}
}

func TestCoachQueryPromptTellsAgentToGenerateObjectivesWhenDocumentsExist(t *testing.T) {
	prompt := CoachQueryPrompt(CoachQueryInput{
		Message: "继续学习",
		Documents: []CoachDocument{
			{ID: "doc-hello", FolderID: "folder-go", Filename: "hello.md"},
		},
	})

	for _, want := range []string{
		"create_learning_objectives",
		"doc-hello",
		"生成学习小节",
		"first_objective_id",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt does not contain %q:\n%s", want, prompt)
		}
	}
}
