package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	wiki_db "verve/app/wiki/models/db"
)

func TestTeacherRejectsRAGContextThatIsNotReady(t *testing.T) {
	source := &fakeFeynmanDocumentSource{
		document: &wiki_db.Document{ID: "doc-1", Filename: "并发.md", CurrentVersion: 3},
		markdown: "# 并发\n" + strings.Repeat("内容", FullDocumentCharacterLimit),
	}
	service := newTeacherService(source, func(context.Context, string) (string, error) {
		t.Fatal("agent must not run without current-version evidence")
		return "", nil
	})

	_, err := service.Teach(context.Background(), TeachingRequest{UserID: "user-1", DocumentID: "doc-1", Question: "channel 为什么阻塞？"})
	if !errors.Is(err, ErrIndexNotReady) {
		t.Fatalf("error = %v", err)
	}
}

func TestParseTeachingResultValidatesResponseAndEvidence(t *testing.T) {
	result, err := parseTeachingResult(`{"response":"先看无缓冲 channel。","question_summary":"阻塞原因","knowledge_gaps":[],"explanation_summary":"解释同步点","key_points":["发送接收配对"],"examples":[],"evidence":[{"chunk_id":"chunk-1","document_version":3,"chunk_index":2,"heading_path":"并发 > channel","content":"无缓冲 channel 同步发送和接收。"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if result.Response == "" || len(result.Evidence) != 1 || result.Evidence[0].DocumentVersion != 3 {
		t.Fatalf("result = %#v", result)
	}
	if _, err := parseTeachingResult(`{"response":"","question_summary":"x","knowledge_gaps":[],"explanation_summary":"x","key_points":[],"examples":[],"evidence":[]}`); err == nil {
		t.Fatal("empty response must be rejected")
	}
}
