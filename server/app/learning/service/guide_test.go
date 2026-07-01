package service

import (
	"encoding/json"
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"
)

func TestParseGuideOutputStripsMarkdownFence(t *testing.T) {
	raw := "```json\n{\"summary\":\"先理解接口\",\"mastery_goals\":[\"能解释接口的作用\"],\"practice_points\":[{\"title\":\"接口是什么\",\"goal\":\"能说清接口是一组方法签名\",\"evidence\":[\"interface 定义了一组方法\"]}],\"reading_steps\":[\"先看定义\"],\"pitfalls\":[\"不要只背语法\"],\"self_check_questions\":[\"接口解决什么问题?\"],\"evidence\":[\"interface 定义了一组方法\"]}\n```"

	result, err := parseGuideOutput(raw)
	if err != nil {
		t.Fatalf("parseGuideOutput returned error: %v", err)
	}

	if result.Summary != "先理解接口" {
		t.Fatalf("summary = %q", result.Summary)
	}
	if len(result.MasteryGoals) != 1 || result.MasteryGoals[0] != "能解释接口的作用" {
		t.Fatalf("mastery goals = %#v", result.MasteryGoals)
	}
	if len(result.PracticePoints) != 1 || result.PracticePoints[0].Title != "接口是什么" {
		t.Fatalf("practice points = %#v", result.PracticePoints)
	}
	if len(result.Evidence) != 1 || result.Evidence[0] != "interface 定义了一组方法" {
		t.Fatalf("evidence = %#v", result.Evidence)
	}
}

func TestParseGuideOutputFindsJSONAfterThinkingBlock(t *testing.T) {
	raw := `<think>
我先分析资料:
1. Go 是静态类型
2. 字面量包括 '1' 和 "1"
{"draft":["这个对象在 thinking 里,不是最终答案"]}
</think>

{"summary":"先理解值类型与字面量","mastery_goals":["能说明值一定有类型"],"practice_points":[{"title":"字面量是什么","goal":"能说清字面量是直接写在代码里的值","evidence":["42、true、\"hello\" 都是字面量"]}],"reading_steps":["先看值和类型"],"pitfalls":["不要把 := 理解成动态类型"],"self_check_questions":["'1' 和 \"1\" 一样吗?"],"evidence":["Go 是静态类型语言"]}`

	result, err := parseGuideOutput(raw)
	if err != nil {
		t.Fatalf("parseGuideOutput returned error: %v", err)
	}

	if result.Summary != "先理解值类型与字面量" {
		t.Fatalf("summary = %q", result.Summary)
	}
	if len(result.PracticePoints) != 1 || result.PracticePoints[0].Title != "字面量是什么" {
		t.Fatalf("practice points = %#v", result.PracticePoints)
	}
}

func TestParseGuideOutputNormalizesMissingArrays(t *testing.T) {
	raw := `{"summary":"资料不足","mastery_goals":["先看定义"]}`

	result, err := parseGuideOutput(raw)
	if err != nil {
		t.Fatalf("parseGuideOutput returned error: %v", err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	body := string(payload)
	for _, want := range []string{
		`"practice_points":[]`,
		`"reading_steps":[]`,
		`"pitfalls":[]`,
		`"self_check_questions":[]`,
		`"evidence":[]`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("marshaled guide result does not contain %s: %s", want, body)
		}
	}
}

func TestBuildGuideQueryIncludesObjectiveAndMarkdown(t *testing.T) {
	detail := "理解接口定义和实现关系"
	stage := "Go 抽象能力"
	obj := &learning_db.LearningObjective{
		Title:      "接口基础",
		Detail:     &detail,
		StageTitle: &stage,
	}

	query := buildGuideQuery(obj, "# 接口基础\ninterface 是一组方法签名。")

	for _, want := range []string{
		"小目标:接口基础",
		"阶段:Go 抽象能力",
		"要点:理解接口定义和实现关系",
		"interface 是一组方法签名",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %q:\n%s", want, query)
		}
	}
}
