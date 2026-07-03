package prompts

import (
	"strings"
	"testing"
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
