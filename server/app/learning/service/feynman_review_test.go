package service

import (
	"fmt"
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

	got, err := parseFeynmanReviewOutput(raw, true)
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
	raw := `{"heard_summary":"我听到你解释了接口的隐式实现","clear_points":null,"confusing_points":null,"misconceptions":null,"follow_up_question":"","explanation_summary":"学习者提到了方法集匹配。","ready_to_wrap_up":true,"context_sufficient":true}`

	got, err := parseFeynmanReviewOutput(raw, true)
	if err != nil {
		t.Fatalf("parseFeynmanReviewOutput returned error: %v", err)
	}
	if got.ClearPoints == nil || got.ConfusingPoints == nil || got.Misconceptions == nil {
		t.Fatalf("arrays were not normalized: %#v", got)
	}
}

func TestParseFeynmanReviewOutputRejectsEmptyAndInvalidOutputs(t *testing.T) {
	for _, raw := range []string{"", "not json", `{}`, `{"heard_summary":""}`} {
		if got, err := parseFeynmanReviewOutput(raw, false); err == nil {
			t.Fatalf("parseFeynmanReviewOutput(%q) = %#v, nil error", raw, got)
		}
	}
}

func TestParseFeynmanReviewOutputUsesAuthoritativeContextSufficiency(t *testing.T) {
	base := `{"heard_summary":"我听到你解释了调度公平性","clear_points":[],"confusing_points":[],"misconceptions":[],"follow_up_question":"","explanation_summary":"学习者对调度公平性作出断言。","ready_to_wrap_up":true,"context_sufficient":%t}`
	tests := []struct {
		name         string
		modelValue   bool
		runtimeValue bool
	}{
		{name: "runtime false overrides model true", modelValue: true, runtimeValue: false},
		{name: "runtime true overrides model false", modelValue: false, runtimeValue: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFeynmanReviewOutput(fmt.Sprintf(base, tt.modelValue), tt.runtimeValue)
			if err != nil {
				t.Fatalf("parseFeynmanReviewOutput returned error: %v", err)
			}
			if got.ContextSufficient != tt.runtimeValue {
				t.Fatalf("context sufficient = %t, want authoritative %t", got.ContextSufficient, tt.runtimeValue)
			}
		})
	}
}

func TestParseFeynmanReviewOutputRejectsMissingKeys(t *testing.T) {
	raw := `{"heard_summary":"我听到你的解释","clear_points":[],"confusing_points":[],"misconceptions":[],"follow_up_question":"","explanation_summary":"解释摘要","ready_to_wrap_up":true}`

	if got, err := parseFeynmanReviewOutput(raw, true); err == nil {
		t.Fatalf("missing context_sufficient accepted: %#v", got)
	}
}

func TestParseFeynmanReviewOutputRejectsNullScalar(t *testing.T) {
	raw := `{"heard_summary":"我听到你的解释","clear_points":[],"confusing_points":[],"misconceptions":[],"follow_up_question":"问题是什么？","explanation_summary":"解释摘要","ready_to_wrap_up":false,"context_sufficient":null}`

	if got, err := parseFeynmanReviewOutput(raw, true); err == nil {
		t.Fatalf("null scalar accepted: %#v", got)
	}
}

func TestParseFeynmanReviewOutputRequiresQuestionWhileNotReady(t *testing.T) {
	raw := `{"heard_summary":"我听到你的解释","clear_points":[],"confusing_points":["还有缺口"],"misconceptions":[],"follow_up_question":"  ","explanation_summary":"解释仍需澄清","ready_to_wrap_up":false,"context_sufficient":true}`

	if got, err := parseFeynmanReviewOutput(raw, true); err == nil {
		t.Fatalf("non-wrap review without a question accepted: %#v", got)
	}
}

func TestParseFeynmanReviewOutputRejectsQuestionWhenReady(t *testing.T) {
	raw := `{"heard_summary":"我听到你的解释","clear_points":["已经清楚"],"confusing_points":[],"misconceptions":[],"follow_up_question":"还要再举一个例子吗？","explanation_summary":"解释可以收束","ready_to_wrap_up":true,"context_sufficient":true}`

	if got, err := parseFeynmanReviewOutput(raw, true); err == nil {
		t.Fatalf("ready review with a follow-up question accepted: %#v", got)
	}
}
