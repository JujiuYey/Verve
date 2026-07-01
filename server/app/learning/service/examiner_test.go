package service

import (
	"reflect"
	"strings"
	"testing"

	learning_db "verve/app/learning/models/db"
)

func TestParseLearningExaminerOutputIncludesStateFacts(t *testing.T) {
	raw := `一些前置说明
{"verdict":"pass","mastery_after":"explained","feedback":"讲清楚了值必须有类型","evidence":"能说明类型由编译期检查","weak_points":["缺少反例","编译期检查解释偏泛"],"next_recommendation":"补一个 int x = \"hello\" 的反例","review_required":false}`

	result, err := parseLearningExaminerOutput(raw)
	if err != nil {
		t.Fatalf("parseLearningExaminerOutput returned error: %v", err)
	}

	if result.Verdict != "pass" {
		t.Fatalf("verdict = %q", result.Verdict)
	}
	if result.Evidence != "能说明类型由编译期检查" {
		t.Fatalf("evidence = %q", result.Evidence)
	}
	if !reflect.DeepEqual(result.WeakPoints, []string{"缺少反例", "编译期检查解释偏泛"}) {
		t.Fatalf("weak points = %#v", result.WeakPoints)
	}
	if result.NextRecommendation != "补一个 int x = \"hello\" 的反例" {
		t.Fatalf("next recommendation = %q", result.NextRecommendation)
	}
	if result.ReviewRequired {
		t.Fatalf("review required = true")
	}
}

func TestParseLearningExaminerOutputBackfillsMissingStateFacts(t *testing.T) {
	result, err := parseLearningExaminerOutput(`{"verdict":"partial","mastery_after":"heard","feedback":"还缺一个例子"}`)
	if err != nil {
		t.Fatalf("parseLearningExaminerOutput returned error: %v", err)
	}

	if result.Evidence == "" {
		t.Fatalf("expected evidence to fall back to feedback")
	}
	if result.NextRecommendation == "" {
		t.Fatalf("expected next recommendation to be generated")
	}
	if result.WeakPoints == nil {
		t.Fatalf("expected weak points to be normalized to an empty slice")
	}
}

func TestMergeLearningProfileStateDeduplicatesWeakPointsAndTopics(t *testing.T) {
	currentLevel := "heard"
	profile := &learning_db.LearningProfile{
		CurrentLevel:    &currentLevel,
		CompletedTopics: []string{"值类型与字面量"},
		WeakPoints:      []string{"缺少反例"},
	}
	result := &JudgeResult{
		Verdict:            "pass",
		MasteryAfter:       "explained",
		Evidence:           "能说明值必须有类型",
		WeakPoints:         []string{"缺少反例", "编译期检查解释偏泛"},
		NextRecommendation: "补一个反例",
	}
	obj := &learning_db.LearningObjective{Title: "值类型与字面量"}

	MergeLearningProfileState(profile, obj, result)

	if profile.CurrentLevel == nil || *profile.CurrentLevel != "explained" {
		t.Fatalf("current level = %v", profile.CurrentLevel)
	}
	if !reflect.DeepEqual(profile.CompletedTopics, []string{"值类型与字面量"}) {
		t.Fatalf("completed topics = %#v", profile.CompletedTopics)
	}
	if !reflect.DeepEqual(profile.WeakPoints, []string{"缺少反例", "编译期检查解释偏泛"}) {
		t.Fatalf("weak points = %#v", profile.WeakPoints)
	}
	if profile.NextGoal == nil || *profile.NextGoal != "补一个反例" {
		t.Fatalf("next goal = %v", profile.NextGoal)
	}
	if profile.VerificationHabits == nil || !strings.Contains(*profile.VerificationHabits, "能说明值必须有类型") {
		t.Fatalf("verification habits = %v", profile.VerificationHabits)
	}
}
