package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	learning_service "verve/app/learning/service"
)

type fakeKnowledgeQAProcessor struct {
	prepared   *learning_service.PreparedKnowledgeQA
	prepareErr error
	answer     *learning_service.KnowledgeQAAnswer
	answerErr  error
}

func (f *fakeKnowledgeQAProcessor) Prepare(context.Context, learning_service.KnowledgeQARequest) (*learning_service.PreparedKnowledgeQA, error) {
	return f.prepared, f.prepareErr
}

func (f *fakeKnowledgeQAProcessor) Generate(context.Context, *learning_service.PreparedKnowledgeQA) (*learning_service.KnowledgeQAAnswer, error) {
	return f.answer, f.answerErr
}

func TestKnowledgeQAHandlerRejectsInvalidQuestion(t *testing.T) {
	handler := newKnowledgeQAHandler(&fakeKnowledgeQAProcessor{prepareErr: learning_service.ErrKnowledgeQAQuestionRequired})
	response := performKnowledgeQARequest(t, handler, `{"message":""}`)
	defer response.Body.Close()
	if response.StatusCode != 400 {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestKnowledgeQAHandlerReturnsRetrievalFailureBeforeSSE(t *testing.T) {
	handler := newKnowledgeQAHandler(&fakeKnowledgeQAProcessor{prepareErr: errors.New("qdrant secret detail")})
	response := performKnowledgeQARequest(t, handler, `{"message":"事务是什么"}`)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != 500 || !strings.Contains(string(body), "检索 Wiki 失败") {
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	if strings.Contains(string(body), "secret detail") {
		t.Fatalf("internal error leaked: %s", body)
	}
}

func TestKnowledgeQAHandlerStreamsSourcesStatusAnswerAndDone(t *testing.T) {
	handler := newKnowledgeQAHandler(&fakeKnowledgeQAProcessor{
		prepared: &learning_service.PreparedKnowledgeQA{Sources: []learning_service.KnowledgeQASource{{
			DocumentID: "doc-1", DocumentTitle: "database.md", Score: 0.91,
		}}},
		answer: &learning_service.KnowledgeQAAnswer{KnowledgeAnswer: "知识回答", LearningAdvice: "学习建议"},
	})
	response := performKnowledgeQARequest(t, handler, `{"message":"事务是什么"}`)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	text := string(body)
	if response.StatusCode != 200 || response.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("status=%d content-type=%q", response.StatusCode, response.Header.Get("Content-Type"))
	}
	positions := []int{
		strings.Index(text, `"type":"sources"`),
		strings.Index(text, `"type":"status"`),
		strings.Index(text, `"type":"answer"`),
		strings.Index(text, "data: [DONE]"),
	}
	for index, position := range positions {
		if position < 0 || (index > 0 && position <= positions[index-1]) {
			t.Fatalf("unexpected SSE order %v:\n%s", positions, text)
		}
	}
	if strings.Contains(text, "reasoning") || strings.Contains(text, "tool_call") || strings.Contains(text, "action") {
		t.Fatalf("legacy event leaked:\n%s", text)
	}
}

func TestKnowledgeQAHandlerStreamsSanitizedGenerationFailure(t *testing.T) {
	handler := newKnowledgeQAHandler(&fakeKnowledgeQAProcessor{
		prepared:  &learning_service.PreparedKnowledgeQA{},
		answerErr: errors.New("provider https://secret.example failed"),
	})
	response := performKnowledgeQARequest(t, handler, `{"message":"事务是什么"}`)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	text := string(body)
	if !strings.Contains(text, `"type":"error"`) || !strings.Contains(text, "生成回答失败，请重试") || !strings.Contains(text, "data: [DONE]") {
		t.Fatalf("body = %s", text)
	}
	if strings.Contains(text, "secret.example") {
		t.Fatalf("internal error leaked: %s", text)
	}
}

func performKnowledgeQARequest(t *testing.T, handler *KnowledgeQAHandler, body string) *http.Response {
	t.Helper()
	app := fiber.New()
	app.Post("/api/learning/ask", handler.Ask)
	request := httptest.NewRequest("POST", "/api/learning/ask", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
