package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	learning_db "verve/app/learning/models/db"
	learning_service "verve/app/learning/service"
	rag_db "verve/app/rag/models/db"
	rag_payload "verve/app/rag/models/payload"
	wiki_db "verve/app/wiki/models/db"
)

type fakeSessionStore struct {
	sessions map[string]*learning_db.LearningSession
	created  *learning_db.LearningSession
	updated  *learning_db.LearningSession
	findErr  error
}

func (f *fakeSessionStore) FindOne(_ context.Context, id string) (*learning_db.LearningSession, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	session, ok := f.sessions[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copy := *session
	return &copy, nil
}

func (f *fakeSessionStore) Create(_ context.Context, session *learning_db.LearningSession) error {
	session.ID = "session-new"
	f.created = session
	return nil
}

func (f *fakeSessionStore) Update(_ context.Context, session *learning_db.LearningSession) error {
	f.updated = session
	return nil
}

type fakeDocumentStore struct {
	doc *wiki_db.Document
	err error
}

func (f fakeDocumentStore) FindOne(context.Context, string) (*wiki_db.Document, error) {
	return f.doc, f.err
}

func (f fakeDocumentStore) FindAccessible(context.Context, string, string) (*wiki_db.Document, error) {
	return f.doc, f.err
}

type fakeMessageStore struct {
	messages []*learning_db.LearningMessage
}

func (f fakeMessageStore) FindBySession(context.Context, string) ([]*learning_db.LearningMessage, error) {
	return f.messages, nil
}

type fakeReviewStore struct {
	reviews []*learning_db.LearningExplanationReview
	created *learning_db.LearningExplanationReview
}

func (f *fakeReviewStore) Create(_ context.Context, review *learning_db.LearningExplanationReview) error {
	review.ID = "review-new"
	f.created = review
	f.reviews = append(f.reviews, review)
	return nil
}

func (f *fakeReviewStore) FindBySession(context.Context, string) ([]*learning_db.LearningExplanationReview, error) {
	return f.reviews, nil
}

type fakeFeynmanReviewer struct {
	request learning_service.FeynmanReviewRequest
}

func (f *fakeFeynmanReviewer) Review(_ context.Context, request learning_service.FeynmanReviewRequest) (*learning_service.FeynmanReview, error) {
	f.request = request
	return &learning_service.FeynmanReview{
		HeardSummary:       "听到你解释了类型会约束操作",
		ClearPoints:        []string{"值有具体类型"},
		ConfusingPoints:    []string{},
		Misconceptions:     []string{},
		FollowUpQuestion:   "这种约束在什么时候检查？",
		ExplanationSummary: "值的类型决定可用操作",
		ReadyToWrapUp:      false,
		ContextSufficient:  true,
	}, nil
}

type fakeMemoryRecorder struct {
	called         bool
	err            error
	readErr        error
	items          []*learning_db.LearningMemoryItem
	readUserID     string
	readDocumentID string
}

func (f *fakeMemoryRecorder) FindDocumentItems(_ context.Context, userID, documentID string, _ int) ([]*learning_db.LearningMemoryItem, error) {
	f.readUserID = userID
	f.readDocumentID = documentID
	return f.items, f.readErr
}

func (f *fakeMemoryRecorder) RecordExplanationReview(_ context.Context, _ string, _ *learning_db.LearningSession, _ *learning_db.LearningExplanationReview) error {
	f.called = true
	return f.err
}

func TestSessionCreateUsesDocument(t *testing.T) {
	store := &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{}}
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  store,
		Documents: fakeDocumentStore{doc: &wiki_db.Document{ID: "doc-1"}},
		Messages:  fakeMessageStore{},
		Reviews:   &fakeReviewStore{},
		Reviewer:  &fakeFeynmanReviewer{},
		Memory:    &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)

	req := httptest.NewRequest("POST", "/session", strings.NewReader(`{"document_id":"doc-1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if store.created == nil || store.created.DocumentID != "doc-1" || store.created.UserID != "user-1" {
		t.Fatalf("created session = %#v", store.created)
	}
	var body struct {
		Data struct {
			SessionID string `json:"session_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Data.SessionID != "session-new" {
		t.Fatalf("session id = %q", body.Data.SessionID)
	}
}

func TestSessionCreateRejectsMissingDocument(t *testing.T) {
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{}},
		Documents: fakeDocumentStore{err: fmt.Errorf("lookup document: %w", sql.ErrNoRows)},
		Messages:  fakeMessageStore{}, Reviews: &fakeReviewStore{}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)
	req := httptest.NewRequest("POST", "/session", strings.NewReader(`{"document_id":"missing"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestSessionCreateReturnsInternalForDocumentStoreFailure(t *testing.T) {
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{}},
		Documents: fakeDocumentStore{err: errors.New("database unavailable")},
		Messages:  fakeMessageStore{}, Reviews: &fakeReviewStore{}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)
	req := httptest.NewRequest("POST", "/session", strings.NewReader(`{"document_id":"doc-1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestSessionCreateRejectsAnotherUsersPrivateDocument(t *testing.T) {
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{}},
		Documents: fakeDocumentStore{err: errDocumentForbidden},
		Messages:  fakeMessageStore{}, Reviews: &fakeReviewStore{}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)
	req := httptest.NewRequest("POST", "/session", strings.NewReader(`{"document_id":"doc-private"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestSessionReviewUsesPriorTurnsAndRecordsMemoryBestEffort(t *testing.T) {
	prior := &learning_db.LearningExplanationReview{ID: "review-1", Explanation: "值有类型", CreatedAt: time.Now()}
	reviews := &fakeReviewStore{reviews: []*learning_db.LearningExplanationReview{prior}}
	reviewer := &fakeFeynmanReviewer{}
	memory := &fakeMemoryRecorder{
		err:   errors.New("memory unavailable"),
		items: []*learning_db.LearningMemoryItem{{ID: "memory-1", Kind: "misconception", Statement: "曾把值和变量混为一谈"}},
	}
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{"session-1": {ID: "session-1", UserID: "user-1", DocumentID: "doc-1", Status: "active"}}},
		Documents: fakeDocumentStore{doc: &wiki_db.Document{ID: "doc-1"}}, Messages: fakeMessageStore{}, Reviews: reviews, Reviewer: reviewer, Memory: memory,
	})
	app := sessionTestApp(handler)
	req := httptest.NewRequest("POST", "/session/session-1/review", strings.NewReader(`{"explanation":"A value has a concrete type."}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(reviewer.request.PriorTurns) != 1 || reviewer.request.UserID != "user-1" || reviewer.request.DocumentID != "doc-1" {
		t.Fatalf("reviewer request = %#v", reviewer.request)
	}
	if memory.readUserID != "user-1" || memory.readDocumentID != "doc-1" {
		t.Fatalf("memory scope = user:%q document:%q", memory.readUserID, memory.readDocumentID)
	}
	if len(reviewer.request.MemoryItems) != 1 || reviewer.request.MemoryItems[0].ID != "memory-1" {
		t.Fatalf("reviewer memory = %#v", reviewer.request.MemoryItems)
	}
	if reviews.created == nil || reviews.created.Explanation != "A value has a concrete type." || reviews.created.SessionID != "session-1" {
		t.Fatalf("stored review = %#v", reviews.created)
	}
	if !memory.called {
		t.Fatal("memory recorder was not called")
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	data := body["data"].(map[string]any)
	if _, exists := data["verdict"]; exists {
		t.Fatal("review response contains verdict")
	}
	if _, exists := data["mastery_after"]; exists {
		t.Fatal("review response contains mastery_after")
	}
}

func TestSessionReviewContinuesWithEmptyMemoryWhenReadFails(t *testing.T) {
	reviewer := &fakeFeynmanReviewer{}
	memory := &fakeMemoryRecorder{readErr: errors.New("memory unavailable")}
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{"session-1": {ID: "session-1", UserID: "user-1", DocumentID: "doc-1", Status: "active"}}},
		Documents: fakeDocumentStore{}, Messages: fakeMessageStore{}, Reviews: &fakeReviewStore{}, Reviewer: reviewer, Memory: memory,
	})
	app := sessionTestApp(handler)
	req := httptest.NewRequest("POST", "/session/session-1/review", strings.NewReader(`{"explanation":"值有类型"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if reviewer.request.MemoryItems == nil || len(reviewer.request.MemoryItems) != 0 {
		t.Fatalf("reviewer memory = %#v, want non-nil empty", reviewer.request.MemoryItems)
	}
}

func TestSessionReviewRejectsAnotherUser(t *testing.T) {
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{"session-1": {ID: "session-1", UserID: "other", DocumentID: "doc-1"}}},
		Documents: fakeDocumentStore{}, Messages: fakeMessageStore{}, Reviews: &fakeReviewStore{}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)
	req := httptest.NewRequest("POST", "/session/session-1/review", strings.NewReader(`{"explanation":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestSessionReviewRejectsInactiveSessions(t *testing.T) {
	for _, status := range []string{"completed", "abandoned"} {
		t.Run(status, func(t *testing.T) {
			reviewer := &fakeFeynmanReviewer{}
			handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
				Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{"session-1": {ID: "session-1", UserID: "user-1", DocumentID: "doc-1", Status: status}}},
				Documents: fakeDocumentStore{}, Messages: fakeMessageStore{}, Reviews: &fakeReviewStore{}, Reviewer: reviewer, Memory: &fakeMemoryRecorder{},
			})
			app := sessionTestApp(handler)
			req := httptest.NewRequest("POST", "/session/session-1/review", strings.NewReader(`{"explanation":"值有类型"}`))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != fiber.StatusConflict {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			if reviewer.request.Explanation != "" {
				t.Fatal("inactive session reached reviewer")
			}
		})
	}
}

func TestSessionLookupDistinguishesMissingRowsFromDatabaseFailure(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want int
	}{
		{name: "missing", err: sql.ErrNoRows, want: fiber.StatusNotFound},
		{name: "database failure", err: errors.New("database unavailable"), want: fiber.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
				Sessions: &fakeSessionStore{findErr: tt.err}, Documents: fakeDocumentStore{}, Messages: fakeMessageStore{},
				Reviews: &fakeReviewStore{}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
			})
			app := sessionTestApp(handler)
			resp, err := app.Test(httptest.NewRequest("GET", "/session/session-1", nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestSessionFindOneReturnsOrderedReviewHistory(t *testing.T) {
	reviews := []*learning_db.LearningExplanationReview{{ID: "review-1"}, {ID: "review-2"}}
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions:  &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{"session-1": {ID: "session-1", UserID: "user-1", DocumentID: "doc-1"}}},
		Documents: fakeDocumentStore{}, Messages: fakeMessageStore{messages: []*learning_db.LearningMessage{}},
		Reviews: &fakeReviewStore{reviews: reviews}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)
	resp, err := app.Test(httptest.NewRequest("GET", "/session/session-1", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Data struct {
			Reviews []learning_db.LearningExplanationReview `json:"reviews"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.Reviews) != 2 || body.Data.Reviews[0].ID != "review-1" || body.Data.Reviews[1].ID != "review-2" {
		t.Fatalf("reviews = %#v", body.Data.Reviews)
	}
}

func TestSessionCompleteSummarizesAllReviewTurnsWithoutNextObjective(t *testing.T) {
	store := &fakeSessionStore{sessions: map[string]*learning_db.LearningSession{"session-1": {ID: "session-1", UserID: "user-1", DocumentID: "doc-1"}}}
	handler := NewSessionHandlerWithDependencies(SessionHandlerDependencies{
		Sessions: store, Documents: fakeDocumentStore{}, Messages: fakeMessageStore{},
		Reviews: &fakeReviewStore{reviews: []*learning_db.LearningExplanationReview{
			{ID: "review-1", ExplanationSummary: "值有类型"},
			{ID: "review-2", ExplanationSummary: "类型决定可用操作"},
		}}, Reviewer: &fakeFeynmanReviewer{}, Memory: &fakeMemoryRecorder{},
	})
	app := sessionTestApp(handler)
	resp, err := app.Test(httptest.NewRequest("POST", "/session/session-1/complete", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if store.updated == nil || store.updated.Status != "completed" || store.updated.Summary == nil || *store.updated.Summary != "值有类型\n类型决定可用操作" {
		t.Fatalf("updated session = %#v", store.updated)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	data := body["data"].(map[string]any)
	if _, exists := data["next_objective"]; exists {
		t.Fatal("complete response contains next_objective")
	}
}

type fakeDocumentContent struct {
	path  string
	calls int
}

func (f *fakeDocumentContent) GetFileContent(_ context.Context, path string) (string, error) {
	f.path = path
	f.calls++
	return "# Go\n", nil
}

type fakeFolderStore struct {
	folder *wiki_db.Folder
	err    error
}

func (f fakeFolderStore) FindOne(context.Context, string) (*wiki_db.Folder, error) {
	return f.folder, f.err
}

type fakeDocumentRetriever struct{ documentID string }

func (f *fakeDocumentRetriever) SearchDocument(_ context.Context, documentID, _ string, _ int) ([]rag_payload.SearchResult, error) {
	f.documentID = documentID
	return []rag_payload.SearchResult{}, nil
}

type fakeDocumentChunks struct{ documentID string }

func (f *fakeDocumentChunks) FindNeighbors(_ context.Context, documentID string, _ []int, _ int) ([]*rag_db.WikiChunk, error) {
	f.documentID = documentID
	return []*rag_db.WikiChunk{}, nil
}

func TestFeynmanDocumentSourceUsesDocumentFileAndScopedRAG(t *testing.T) {
	files := &fakeDocumentContent{}
	retriever := &fakeDocumentRetriever{}
	chunks := &fakeDocumentChunks{}
	source := &feynmanDocumentSource{
		documents: fakeDocumentStore{doc: &wiki_db.Document{ID: "doc-1", FolderID: "folder-public", FilePath: "wiki/doc-1.md"}},
		folders:   fakeFolderStore{folder: &wiki_db.Folder{ID: "folder-public"}},
		files:     files, retriever: retriever, chunks: chunks,
	}
	doc, markdown, err := source.LoadDocument(context.Background(), "user-1", "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID != "doc-1" || markdown != "# Go\n" || files.path != "wiki/doc-1.md" {
		t.Fatalf("loaded = doc:%#v markdown:%q path:%q", doc, markdown, files.path)
	}
	if _, err := source.SearchDocument(context.Background(), "doc-1", "value", 6); err != nil {
		t.Fatal(err)
	}
	if _, err := source.FindNeighbors(context.Background(), "doc-1", []int{2}, 1); err != nil {
		t.Fatal(err)
	}
	if retriever.documentID != "doc-1" || chunks.documentID != "doc-1" {
		t.Fatalf("scope = retriever:%q chunks:%q", retriever.documentID, chunks.documentID)
	}
}

func TestFeynmanDocumentSourceRejectsPrivateFolderBeforeReadingContent(t *testing.T) {
	owner := "other-user"
	files := &fakeDocumentContent{}
	source := &feynmanDocumentSource{
		documents: fakeDocumentStore{doc: &wiki_db.Document{ID: "doc-private", FolderID: "folder-private", FilePath: "secret.md"}},
		folders:   fakeFolderStore{folder: &wiki_db.Folder{ID: "folder-private", UserID: &owner}},
		files:     files,
	}
	if _, _, err := source.LoadDocument(context.Background(), "user-1", "doc-private"); !errors.Is(err, errDocumentForbidden) {
		t.Fatalf("error = %v", err)
	}
	if files.calls != 0 {
		t.Fatalf("content reads = %d", files.calls)
	}
}

func TestFeynmanDocumentSourceAllowsFolderOwner(t *testing.T) {
	owner := "user-1"
	files := &fakeDocumentContent{}
	source := &feynmanDocumentSource{
		documents: fakeDocumentStore{doc: &wiki_db.Document{ID: "doc-owned", FolderID: "folder-owned", FilePath: "owned.md"}},
		folders:   fakeFolderStore{folder: &wiki_db.Folder{ID: "folder-owned", UserID: &owner}},
		files:     files,
	}
	if _, _, err := source.LoadDocument(context.Background(), "user-1", "doc-owned"); err != nil {
		t.Fatal(err)
	}
	if files.calls != 1 {
		t.Fatalf("content reads = %d", files.calls)
	}
}

func sessionTestApp(handler *SessionHandler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", "user-1"); return c.Next() })
	app.Post("/session", handler.Create)
	app.Get("/session/:id", handler.FindOne)
	app.Post("/session/:id/review", handler.Review)
	app.Post("/session/:id/complete", handler.Complete)
	return app
}
