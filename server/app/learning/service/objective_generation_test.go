package service

import "testing"

func TestParseObjectiveGenerationOutputNormalizesObjectives(t *testing.T) {
	raw := "```json\n{\"objectives\":[{\"title\":\" 值类型 \",\"detail\":\"讲清值一定有类型\"},{\"title\":\"\",\"detail\":\"跳过\"}]}\n```"

	result, err := parseObjectiveGenerationOutput(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(result.Objectives) != 1 {
		t.Fatalf("objectives len = %d", len(result.Objectives))
	}
	if result.Objectives[0].Title != "值类型" {
		t.Fatalf("title = %q", result.Objectives[0].Title)
	}
	if result.Objectives[0].Detail != "讲清值一定有类型" {
		t.Fatalf("detail = %q", result.Objectives[0].Detail)
	}
}

func TestParseObjectiveGenerationOutputRepairsUnescapedQuotesInStringValues(t *testing.T) {
	raw := `{"objectives":[{"title":"字面量易错点辨析","detail":"能区分 "1"（字符串字面量）与 '1'（rune 字面量，值为字符 1 的 Unicode 码点 49），并能预判整数 1 与浮点 1.0 的类型差异。"}]}`

	result, err := parseObjectiveGenerationOutput(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(result.Objectives) != 1 {
		t.Fatalf("objectives len = %d", len(result.Objectives))
	}
	want := `能区分 "1"（字符串字面量）与 '1'（rune 字面量，值为字符 1 的 Unicode 码点 49），并能预判整数 1 与浮点 1.0 的类型差异。`
	if result.Objectives[0].Detail != want {
		t.Fatalf("detail = %q", result.Objectives[0].Detail)
	}
}
