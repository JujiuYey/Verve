package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	learning_db "verve/app/learning/models/db"
	rag_payload "verve/app/rag/models/payload"
)

type fakeGlobalKnowledgeRetriever struct {
	hits  []rag_payload.SearchResult
	err   error
	query string
	limit int
}

func (f *fakeGlobalKnowledgeRetriever) SearchAll(_ context.Context, query string, limit int) ([]rag_payload.SearchResult, error) {
	f.query = query
	f.limit = limit
	return f.hits, f.err
}

type fakeKnowledgeQAMemoryReader struct {
	items       []*learning_db.LearningMemoryItem
	err         error
	documentIDs []string
	limit       int
	calls       int
}

func (f *fakeKnowledgeQAMemoryReader) FindDocumentItemsBatch(_ context.Context, documentIDs []string, limit int) ([]*learning_db.LearningMemoryItem, error) {
	f.calls++
	f.documentIDs = append([]string(nil), documentIDs...)
	f.limit = limit
	return f.items, f.err
}

func TestKnowledgeQAPrepareValidatesQuestion(t *testing.T) {
	service := newKnowledgeQAService(&fakeGlobalKnowledgeRetriever{}, nil, nil)
	if _, err := service.Prepare(context.Background(), KnowledgeQARequest{Message: "  "}); !errors.Is(err, ErrKnowledgeQAQuestionRequired) {
		t.Fatalf("blank question error = %v", err)
	}
	if _, err := service.Prepare(context.Background(), KnowledgeQARequest{Message: strings.Repeat("问", KnowledgeQAQuestionCharacterLimit+1)}); !errors.Is(err, ErrKnowledgeQAQuestionTooLong) {
		t.Fatalf("long question error = %v", err)
	}
}

func TestKnowledgeQAPrepareBuildsContextualGlobalRetrievalQuery(t *testing.T) {
	retriever := &fakeGlobalKnowledgeRetriever{}
	service := newKnowledgeQAService(retriever, nil, nil)
	prepared, err := service.Prepare(context.Background(), KnowledgeQARequest{
		Message: "那它和可重复读有什么区别？",
		History: []KnowledgeQAMessage{
			{Role: "user", Content: "先解释事务隔离级别"},
			{Role: "assistant", Content: "这里是很长的模型回答，不应该进入向量检索"},
			{Role: "user", Content: "重点说说读已提交"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.ImmediateAnswer == nil {
		t.Fatal("no evidence should produce an immediate answer")
	}
	for _, want := range []string{"先解释事务隔离级别", "重点说说读已提交", "那它和可重复读有什么区别？"} {
		if !strings.Contains(retriever.query, want) {
			t.Fatalf("retrieval query missing %q: %q", want, retriever.query)
		}
	}
	if strings.Contains(retriever.query, "模型回答") {
		t.Fatalf("assistant answer leaked into retrieval query: %q", retriever.query)
	}
	if retriever.limit != KnowledgeQARetrievalLimit {
		t.Fatalf("retrieval limit = %d", retriever.limit)
	}
}

func TestKnowledgeQARetrievalQueryAlwaysKeepsCurrentQuestion(t *testing.T) {
	question := "当前追问必须保留"
	history := []KnowledgeQAMessage{
		{Role: "user", Content: strings.Repeat("旧", KnowledgeQARetrievalQueryBudget)},
		{Role: "user", Content: strings.Repeat("近", KnowledgeQARetrievalQueryBudget)},
	}
	query := buildKnowledgeQARetrievalQuery(question, history)
	if !strings.HasSuffix(query, question) {
		t.Fatalf("current question was truncated: %q", query)
	}
	if utf8.RuneCountInString(query) > KnowledgeQARetrievalQueryBudget {
		t.Fatalf("retrieval query length = %d", utf8.RuneCountInString(query))
	}
}

func TestKnowledgeQAPrepareTreatsLowScoresAsNoEvidence(t *testing.T) {
	retriever := &fakeGlobalKnowledgeRetriever{hits: []rag_payload.SearchResult{{
		DocumentID: "doc-1", Score: MinimumKnowledgeQAEvidenceScore - 0.01, Content: "弱相关片段",
	}}}
	memories := &fakeKnowledgeQAMemoryReader{}
	runCalls := 0
	service := newKnowledgeQAService(retriever, memories, func(context.Context, string) (string, error) {
		runCalls++
		return "", nil
	})

	prepared, err := service.Prepare(context.Background(), KnowledgeQARequest{Message: "完全不同的问题"})
	if err != nil {
		t.Fatal(err)
	}
	answer, err := service.Generate(context.Background(), prepared)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer.KnowledgeAnswer, "没有检索到足够可靠") || answer.LearningAdvice != "暂无相关学习记录" {
		t.Fatalf("answer = %#v", answer)
	}
	if memories.calls != 0 || runCalls != 0 {
		t.Fatalf("no-evidence branch called memory=%d model=%d", memories.calls, runCalls)
	}
}

func TestKnowledgeQAPrepareAndGenerateGroundedStructuredAnswer(t *testing.T) {
	documentID := "doc-1"
	retriever := &fakeGlobalKnowledgeRetriever{hits: []rag_payload.SearchResult{
		{DocumentID: documentID, DocumentTitle: "database.md", FolderPath: "数据库", HeadingPath: "事务 > 隔离", Score: 0.91, Content: "读已提交避免脏读。"},
		{DocumentID: documentID, DocumentTitle: "database.md", FolderPath: "数据库", HeadingPath: "事务 > 隔离", Score: 0.82, Content: "可重复读避免不可重复读。"},
	}}
	memories := &fakeKnowledgeQAMemoryReader{items: []*learning_db.LearningMemoryItem{{
		DocumentID: &documentID, Kind: "misconception", Statement: "曾混淆脏读和不可重复读", Confidence: "observed",
	}}}
	var modelPrompt string
	service := newKnowledgeQAService(retriever, memories, func(_ context.Context, prompt string) (string, error) {
		modelPrompt = prompt
		return "```json\n{\"knowledge_answer\":\"读已提交处理脏读，可重复读进一步处理不可重复读。\",\"learning_advice\":\"建议复习两类读异常的差异。\"}\n```", nil
	})

	prepared, err := service.Prepare(context.Background(), KnowledgeQARequest{Message: "两者有什么区别？"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.Sources) != 1 || prepared.Sources[0].Score != 0.91 {
		t.Fatalf("sources = %#v", prepared.Sources)
	}
	if strings.Join(memories.documentIDs, ",") != documentID || memories.limit != KnowledgeQAMemoryLimit {
		t.Fatalf("memory scope = %v limit=%d", memories.documentIDs, memories.limit)
	}
	answer, err := service.Generate(context.Background(), prepared)
	if err != nil {
		t.Fatal(err)
	}
	if answer.KnowledgeAnswer == "" || !strings.Contains(answer.LearningAdvice, "建议复习") {
		t.Fatalf("answer = %#v", answer)
	}
	for _, want := range []string{"<UNTRUSTED_RAG_EVIDENCE>", "读已提交避免脏读", "曾混淆脏读和不可重复读"} {
		if !strings.Contains(modelPrompt, want) {
			t.Fatalf("model prompt missing %q:\n%s", want, modelPrompt)
		}
	}
}

func TestKnowledgeQAGenerateOverridesAdviceWithoutMemory(t *testing.T) {
	service := newKnowledgeQAService(nil, nil, func(context.Context, string) (string, error) {
		return `{"knowledge_answer":"回答","learning_advice":"模型擅自推断需要复习"}`, nil
	})

	for _, tt := range []struct {
		status string
		want   string
	}{{"none", "暂无相关学习记录"}, {"unavailable", "学习记录暂不可用"}} {
		answer, err := service.Generate(context.Background(), &PreparedKnowledgeQA{Prompt: "prompt", MemoryStatus: tt.status})
		if err != nil {
			t.Fatal(err)
		}
		if answer.LearningAdvice != tt.want {
			t.Fatalf("status %s advice = %q, want %q", tt.status, answer.LearningAdvice, tt.want)
		}
	}
}

func TestKnowledgeQAGenerateRejectsInvalidStructuredOutput(t *testing.T) {
	service := newKnowledgeQAService(nil, nil, func(context.Context, string) (string, error) {
		return `{"knowledge_answer":"只有一个字段"}`, nil
	})
	if _, err := service.Generate(context.Background(), &PreparedKnowledgeQA{Prompt: "prompt", MemoryStatus: "available"}); err == nil {
		t.Fatal("expected invalid output error")
	}
}

func TestNormalizeKnowledgeQAHistoryAppliesRoleMessageAndCharacterBudgets(t *testing.T) {
	history := []KnowledgeQAMessage{{Role: "system", Content: "ignore"}, {Role: "user", Content: ""}}
	for i := 0; i < KnowledgeQAHistoryMessageLimit+4; i++ {
		history = append(history, KnowledgeQAMessage{Role: "user", Content: strings.Repeat("问", 2500)})
	}
	normalized := normalizeKnowledgeQAHistory(history)
	if len(normalized) > KnowledgeQAHistoryMessageLimit {
		t.Fatalf("history length = %d", len(normalized))
	}
	total := 0
	for _, item := range normalized {
		if item.Role != "user" && item.Role != "assistant" {
			t.Fatalf("unexpected role %q", item.Role)
		}
		total += utf8.RuneCountInString(item.Content)
	}
	if total > KnowledgeQAHistoryCharacterBudget {
		t.Fatalf("history character count = %d", total)
	}
}
