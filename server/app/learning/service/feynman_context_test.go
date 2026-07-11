package service

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	rag_db "verve/app/rag/models/db"
	rag_payload "verve/app/rag/models/payload"
	wiki_db "verve/app/wiki/models/db"
)

type fakeFeynmanDocumentSource struct {
	document        *wiki_db.Document
	markdown        string
	searchResults   []rag_payload.SearchResult
	neighbors       []*rag_db.WikiChunk
	searchCalls     int
	neighborCalls   int
	searchDocument  string
	searchQuery     string
	neighborDoc     string
	neighborIndexes []int
}

func (f *fakeFeynmanDocumentSource) LoadDocument(context.Context, string, string) (*wiki_db.Document, string, error) {
	return f.document, f.markdown, nil
}

func (f *fakeFeynmanDocumentSource) SearchDocument(_ context.Context, documentID, query string, _ int) ([]rag_payload.SearchResult, error) {
	f.searchCalls++
	f.searchDocument = documentID
	f.searchQuery = query
	return f.searchResults, nil
}

func (f *fakeFeynmanDocumentSource) FindNeighbors(_ context.Context, documentID string, indexes []int, _ int) ([]*rag_db.WikiChunk, error) {
	f.neighborCalls++
	f.neighborDoc = documentID
	f.neighborIndexes = append([]int(nil), indexes...)
	return f.neighbors, nil
}

func TestFeynmanContextUsesFullMarkdownAtCharacterLimit(t *testing.T) {
	prefix := "# 总览\n\n## 第一章\n\n### 细节\n\n```go\n# 不是标题\n```\n\n附录\n------\n\n"
	markdown := prefix + strings.Repeat("界", FullDocumentCharacterLimit-utf8.RuneCountInString(prefix))
	source := &fakeFeynmanDocumentSource{
		document: &wiki_db.Document{ID: "doc-1", Filename: "类型系统.md"},
		markdown: markdown,
	}

	got, err := NewFeynmanContextBuilder(source).Build(context.Background(), "user-1", "doc-1", "我的解释")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if got.Mode != FeynmanContextModeFull {
		t.Fatalf("mode = %q", got.Mode)
	}
	if got.FullText != markdown {
		t.Fatalf("full text was changed or truncated")
	}
	if !reflect.DeepEqual(got.Outline, []string{"总览", "总览 > 第一章", "总览 > 第一章 > 细节", "总览 > 附录"}) {
		t.Fatalf("outline = %#v", got.Outline)
	}
	if !got.ContextSufficient {
		t.Fatalf("full document context should be sufficient")
	}
	if source.searchCalls != 0 || source.neighborCalls != 0 {
		t.Fatalf("retrieval calls = search:%d neighbors:%d", source.searchCalls, source.neighborCalls)
	}
}

func TestFeynmanContextUsesDocumentScopedRAGWithoutTruncatingFullText(t *testing.T) {
	markdown := "# 总览\n\n## 基础\n\n## 进阶\n\n### 陷阱\n\n" + strings.Repeat("a", FullDocumentCharacterLimit)
	source := &fakeFeynmanDocumentSource{
		document: &wiki_db.Document{ID: "doc-2", Filename: "并发.md"},
		markdown: markdown,
		searchResults: []rag_payload.SearchResult{
			{ChunkID: "hit-3", DocumentID: "doc-2", ChunkIndex: 3, Score: 0.91, HeadingPath: "总览 > 进阶", Content: "命中三"},
			{ChunkID: "hit-1", DocumentID: "doc-2", ChunkIndex: 1, Score: 0.83, HeadingPath: "总览 > 基础", Content: "命中一"},
		},
		neighbors: []*rag_db.WikiChunk{
			{ID: "neighbor-4", DocumentID: "doc-2", ChunkIndex: 4, HeadingPath: "总览 > 进阶 > 陷阱", Content: "相邻四"},
			{ID: "stored-hit-3", DocumentID: "doc-2", ChunkIndex: 3, HeadingPath: "总览 > 进阶", Content: "命中三"},
			{ID: "neighbor-2", DocumentID: "doc-2", ChunkIndex: 2, HeadingPath: "总览 > 基础", Content: "相邻二"},
		},
	}

	got, err := NewFeynmanContextBuilder(source).Build(context.Background(), "user-1", "doc-2", "  channel 会阻塞发送者  ")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if got.Mode != FeynmanContextModeRAG || got.FullText != "" {
		t.Fatalf("mode/full text = %q/%q", got.Mode, got.FullText)
	}
	if source.searchDocument != "doc-2" || source.searchQuery != "channel 会阻塞发送者" {
		t.Fatalf("search scope/query = %q/%q", source.searchDocument, source.searchQuery)
	}
	if source.neighborDoc != "doc-2" || !reflect.DeepEqual(source.neighborIndexes, []int{3, 1}) {
		t.Fatalf("neighbor request = %q/%v", source.neighborDoc, source.neighborIndexes)
	}
	if !reflect.DeepEqual(got.Outline, []string{"总览", "总览 > 基础", "总览 > 进阶", "总览 > 进阶 > 陷阱"}) {
		t.Fatalf("outline = %#v", got.Outline)
	}
	wantIndexes := []int{1, 2, 3, 4}
	gotIndexes := make([]int, 0, len(got.Evidence))
	for _, item := range got.Evidence {
		gotIndexes = append(gotIndexes, item.ChunkIndex)
	}
	if !reflect.DeepEqual(gotIndexes, wantIndexes) {
		t.Fatalf("evidence indexes = %v", gotIndexes)
	}
	if !got.ContextSufficient {
		t.Fatalf("expanded RAG context should be sufficient")
	}
}

func TestFeynmanContextRejectsEmptyMarkdown(t *testing.T) {
	source := &fakeFeynmanDocumentSource{
		document: &wiki_db.Document{ID: "doc-empty", Filename: "empty.md"},
		markdown: " \n\t ",
	}

	_, err := NewFeynmanContextBuilder(source).Build(context.Background(), "user-1", "doc-empty", "解释")
	if err == nil || !strings.Contains(err.Error(), "markdown") || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("error = %v", err)
	}
}

func TestFeynmanContextMarksInsufficientRAGEvidence(t *testing.T) {
	source := &fakeFeynmanDocumentSource{
		document: &wiki_db.Document{ID: "doc-large", Filename: "large.md"},
		markdown: strings.Repeat("x", FullDocumentCharacterLimit+1),
		searchResults: []rag_payload.SearchResult{
			{ChunkID: "only", DocumentID: "doc-large", ChunkIndex: 7, Score: 0.9, Content: "唯一证据"},
		},
		neighbors: []*rag_db.WikiChunk{
			{ID: "before", DocumentID: "doc-large", ChunkIndex: 6, Content: "前一个邻块"},
			{ID: "after", DocumentID: "doc-large", ChunkIndex: 8, Content: "后一个邻块"},
		},
	}

	got, err := NewFeynmanContextBuilder(source).Build(context.Background(), "user-1", "doc-large", "解释")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if got.ContextSufficient {
		t.Fatalf("one primary hit must remain insufficient despite neighbor expansion")
	}
	if strings.TrimSpace(got.ContextInsufficiencyReason) == "" {
		t.Fatalf("missing insufficiency reason")
	}
}

func TestFeynmanContextRequiresTwoRelevantPrimaryHits(t *testing.T) {
	tests := []struct {
		name       string
		scores     []float64
		sufficient bool
	}{
		{name: "two low score hits", scores: []float64{MinimumFeynmanEvidenceScore - 0.01, 0.2}, sufficient: false},
		{name: "one qualifying hit", scores: []float64{MinimumFeynmanEvidenceScore, MinimumFeynmanEvidenceScore - 0.01}, sufficient: false},
		{name: "two qualifying hits", scores: []float64{MinimumFeynmanEvidenceScore, 0.95}, sufficient: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &fakeFeynmanDocumentSource{
				document: &wiki_db.Document{ID: "doc-scores", Filename: "scores.md"},
				markdown: strings.Repeat("x", FullDocumentCharacterLimit+1),
				searchResults: []rag_payload.SearchResult{
					{ChunkID: "first", DocumentID: "doc-scores", ChunkIndex: 1, Score: tt.scores[0], Content: "第一条"},
					{ChunkID: "second", DocumentID: "doc-scores", ChunkIndex: 2, Score: tt.scores[1], Content: "第二条"},
				},
			}

			got, err := NewFeynmanContextBuilder(source).Build(context.Background(), "user-1", "doc-scores", "解释")
			if err != nil {
				t.Fatalf("Build returned error: %v", err)
			}
			if got.ContextSufficient != tt.sufficient {
				t.Fatalf("context sufficient = %t, want %t for scores %v", got.ContextSufficient, tt.sufficient, tt.scores)
			}
		})
	}
}

func TestMarkdownHeadingPathsUsesCommonMarkBoundaries(t *testing.T) {
	markdown := "# Root\n\n    # indented code\n\n```go\n# fenced\n``` trailing text\n# still fenced\n```\n\n# C#\n"

	got := markdownHeadingPaths(markdown)
	want := []string{"Root", "C#"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("outline = %#v, want %#v", got, want)
	}
}
