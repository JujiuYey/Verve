package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	learning_db "verve/app/learning/models/db"
)

type reviewQueryRecorder struct {
	execQuery   string
	selectQuery string
	rows        driver.Rows
}

func (r *reviewQueryRecorder) Connect(context.Context) (driver.Conn, error) {
	return reviewQueryConn{recorder: r}, nil
}

func (r *reviewQueryRecorder) Driver() driver.Driver { return reviewQueryDriver{recorder: r} }

type reviewQueryDriver struct{ recorder *reviewQueryRecorder }

func (d reviewQueryDriver) Open(string) (driver.Conn, error) {
	return reviewQueryConn{recorder: d.recorder}, nil
}

type reviewQueryConn struct{ recorder *reviewQueryRecorder }

func (c reviewQueryConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c reviewQueryConn) Close() error                        { return nil }
func (c reviewQueryConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c reviewQueryConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.recorder.execQuery = query
	return driver.RowsAffected(1), nil
}
func (c reviewQueryConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.HasPrefix(query, "INSERT ") {
		c.recorder.execQuery = query
		return &reviewInsertRows{createdAt: time.Now()}, nil
	}
	c.recorder.selectQuery = query
	return c.recorder.rows, nil
}

type reviewInsertRows struct {
	createdAt time.Time
	done      bool
}

func (r *reviewInsertRows) Columns() []string { return []string{"created_at"} }
func (r *reviewInsertRows) Close() error      { return nil }
func (r *reviewInsertRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	dest[0] = r.createdAt
	r.done = true
	return nil
}

type reviewRows struct {
	values [][]driver.Value
	index  int
}

func (r *reviewRows) Columns() []string {
	return []string{
		"id", "session_id", "document_id", "user_id", "explanation", "heard_summary",
		"clear_points", "confusing_points", "misconceptions", "follow_up_question",
		"explanation_summary", "ready_to_wrap_up", "context_sufficient", "created_at",
	}
}
func (r *reviewRows) Close() error { return nil }
func (r *reviewRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func newReviewRepositoryForTest(recorder *reviewQueryRecorder) (*ReviewRepository, func()) {
	sqldb := sql.OpenDB(recorder)
	db := bun.NewDB(sqldb, pgdialect.New())
	return NewReviewRepository(db), func() { _ = db.Close() }
}

func TestReviewRepositoryCreateAssignsIDAndPersistsReview(t *testing.T) {
	recorder := &reviewQueryRecorder{}
	repo, closeDB := newReviewRepositoryForTest(recorder)
	defer closeDB()

	review := &learning_db.LearningExplanationReview{
		SessionID:          "session-1",
		DocumentID:         "document-1",
		UserID:             "user-1",
		Explanation:        "A channel coordinates communication.",
		HeardSummary:       "Channels coordinate senders and receivers.",
		ClearPoints:        []string{"coordination", "blocking"},
		ConfusingPoints:    []string{"buffer capacity"},
		Misconceptions:     []string{"all sends are asynchronous"},
		FollowUpQuestion:   "What happens when the buffer is full?",
		ExplanationSummary: "The learner understands coordination but not full buffers.",
		ReadyToWrapUp:      false,
		ContextSufficient:  true,
	}

	if err := repo.Create(context.Background(), review); err != nil {
		t.Fatal(err)
	}
	if len(review.ID) != 32 || strings.Contains(review.ID, "-") {
		t.Fatalf("review ID = %q, want 32-character UUID without hyphens", review.ID)
	}

	for _, want := range []string{
		`"session_id"`, `"document_id"`, `"user_id"`, `"explanation"`, `"heard_summary"`,
		`"clear_points"`, `"confusing_points"`, `"misconceptions"`, `"follow_up_question"`,
		`"explanation_summary"`, `"ready_to_wrap_up"`, `"context_sufficient"`,
		`'session-1'`, `'document-1'`, `'user-1'`,
		`'A channel coordinates communication.'`,
		`'Channels coordinate senders and receivers.'`,
		`'["coordination","blocking"]'`, `'["buffer capacity"]'`,
		`'["all sends are asynchronous"]'`,
		`'What happens when the buffer is full?'`,
		`'The learner understands coordination but not full buffers.'`,
		`FALSE`, `TRUE`,
	} {
		if !strings.Contains(recorder.execQuery, want) {
			t.Fatalf("insert query does not contain %q: %s", want, recorder.execQuery)
		}
	}
}

func TestReviewRepositoryFindBySessionOrdersOldestFirst(t *testing.T) {
	createdAt := time.Date(2026, 7, 11, 9, 30, 0, 0, time.UTC)
	recorder := &reviewQueryRecorder{rows: &reviewRows{values: [][]driver.Value{{
		"review-1", "session-1", "document-1", "user-1", "explanation", "heard",
		`["clear"]`, `[]`, `[]`, "follow up", "summary", false, true, createdAt,
	}}}}
	repo, closeDB := newReviewRepositoryForTest(recorder)
	defer closeDB()

	reviews, err := repo.FindBySession(context.Background(), "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(recorder.selectQuery, `WHERE (session_id = 'session-1')`) {
		t.Fatalf("query does not filter session: %s", recorder.selectQuery)
	}
	if !strings.Contains(recorder.selectQuery, `ORDER BY "created_at" ASC`) {
		t.Fatalf("query does not order oldest first: %s", recorder.selectQuery)
	}
	if len(reviews) != 1 || reviews[0].ID != "review-1" {
		t.Fatalf("reviews = %#v", reviews)
	}
	if len(reviews[0].ClearPoints) != 1 || reviews[0].ClearPoints[0] != "clear" {
		t.Fatalf("clear points = %#v", reviews[0].ClearPoints)
	}
}

func TestReviewRepositoryFindBySessionReturnsNonNilEmptySlice(t *testing.T) {
	recorder := &reviewQueryRecorder{rows: &reviewRows{}}
	repo, closeDB := newReviewRepositoryForTest(recorder)
	defer closeDB()

	reviews, err := repo.FindBySession(context.Background(), "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if reviews == nil || len(reviews) != 0 {
		t.Fatalf("reviews = %#v, want non-nil empty slice", reviews)
	}
}
