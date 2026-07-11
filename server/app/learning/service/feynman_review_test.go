package service

import (
	"reflect"
	"testing"
)

func TestParseFeynmanReviewOutputAcceptsFencedJSON(t *testing.T) {
	raw := "```json\n" + `{
  "heard_summary": "我听到你说 channel 会让发送和接收同步",
  "clear_points": ["说清了发送与接收的配对关系"],
  "confusing_points": ["缓冲区满时的情况还不明确"],
  "misconceptions": [],
  "follow_up_question": "有缓冲 channel 已满时，发送会发生什么？",
  "explanation_summary": "学习者能说明无缓冲 channel 的同步关系，但尚未覆盖有缓冲情况。",
  "ready_to_wrap_up": false,
  "context_sufficient": true
}` + "\n```"

	got, err := parseFeynmanReviewOutput(raw)
	if err != nil {
		t.Fatalf("parseFeynmanReviewOutput returned error: %v", err)
	}
	if got.HeardSummary != "我听到你说 channel 会让发送和接收同步" {
		t.Fatalf("heard summary = %q", got.HeardSummary)
	}
	if !reflect.DeepEqual(got.ClearPoints, []string{"说清了发送与接收的配对关系"}) {
		t.Fatalf("clear points = %#v", got.ClearPoints)
	}
	if got.ReadyToWrapUp {
		t.Fatalf("ready to wrap up = true")
	}
}

func TestParseFeynmanReviewOutputNormalizesNilArrays(t *testing.T) {
	raw := `{"heard_summary":"我听到你解释了接口的隐式实现","explanation_summary":"学习者提到了方法集匹配。","follow_up_question":"","ready_to_wrap_up":true,"context_sufficient":true}`

	got, err := parseFeynmanReviewOutput(raw)
	if err != nil {
		t.Fatalf("parseFeynmanReviewOutput returned error: %v", err)
	}
	if got.ClearPoints == nil || got.ConfusingPoints == nil || got.Misconceptions == nil {
		t.Fatalf("arrays were not normalized: %#v", got)
	}
}

func TestParseFeynmanReviewOutputRejectsEmptyAndInvalidOutputs(t *testing.T) {
	for _, raw := range []string{"", "not json", `{}`, `{"heard_summary":""}`} {
		if got, err := parseFeynmanReviewOutput(raw); err == nil {
			t.Fatalf("parseFeynmanReviewOutput(%q) = %#v, nil error", raw, got)
		}
	}
}

func TestParseFeynmanReviewOutputRetainsFalseContextSufficient(t *testing.T) {
	raw := `{"heard_summary":"我听到你说调度器会公平执行所有 goroutine","clear_points":[],"confusing_points":["公平的含义没有说明"],"misconceptions":[],"follow_up_question":"你说的公平具体指什么？","explanation_summary":"当前解释包含无法由检索片段核对的全局断言。","ready_to_wrap_up":false,"context_sufficient":false}`

	got, err := parseFeynmanReviewOutput(raw)
	if err != nil {
		t.Fatalf("parseFeynmanReviewOutput returned error: %v", err)
	}
	if got.ContextSufficient {
		t.Fatalf("context sufficient was changed to true")
	}
}
