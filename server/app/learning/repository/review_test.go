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
	c.recorder.selectQuery = query
	return c.recorder.rows, nil
}

type reviewRows struct {
	values [][]driver.Value
	index  int
}

func (r *reviewRows) Columns() []string {
	return []string{
		"id", "turn_id", "heard_summary",
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

func TestReviewRepositoryFindBySessionOrdersOldestFirst(t *testing.T) {
	createdAt := time.Date(2026, 7, 11, 9, 30, 0, 0, time.UTC)
	recorder := &reviewQueryRecorder{rows: &reviewRows{values: [][]driver.Value{{
		"review-1", "turn-1", "heard", `["clear"]`, `[]`, `[]`, "follow up", "summary", false, true, createdAt,
	}}}}
	repo, closeDB := newReviewRepositoryForTest(recorder)
	defer closeDB()

	reviews, err := repo.FindBySession(context.Background(), "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(recorder.selectQuery, `JOIN learning_turns AS lt`) {
		t.Fatalf("query does not join turns: %s", recorder.selectQuery)
	}
	if !strings.Contains(recorder.selectQuery, `lt.session_id = 'session-1'`) {
		t.Fatalf("query does not filter session: %s", recorder.selectQuery)
	}
	if !strings.Contains(recorder.selectQuery, `ORDER BY ler.created_at ASC, ler.id ASC`) {
		t.Fatalf("query does not order deterministically oldest first: %s", recorder.selectQuery)
	}
	if len(reviews) != 1 || reviews[0].ID != "review-1" {
		t.Fatalf("reviews = %#v", reviews)
	}
	if reviews[0].TurnID != "turn-1" {
		t.Fatalf("derived review = %#v", reviews[0])
	}
	if len(reviews[0].ClearPoints) != 1 || reviews[0].ClearPoints[0] != "clear" {
		t.Fatalf("clear points = %#v", reviews[0].ClearPoints)
	}
}

func TestReviewRepositoryFindByTurnUsesTurnIdentity(t *testing.T) {
	createdAt := time.Date(2026, 7, 11, 9, 30, 0, 0, time.UTC)
	recorder := &reviewQueryRecorder{rows: &reviewRows{values: [][]driver.Value{{
		"review-1", "turn-1", "heard", `[]`, `[]`, `[]`, "follow up", "summary", false, true, createdAt,
	}}}}
	repo, closeDB := newReviewRepositoryForTest(recorder)
	defer closeDB()

	review, err := repo.FindByTurn(context.Background(), "turn-1")
	if err != nil {
		t.Fatal(err)
	}
	if review.ID != "review-1" || !strings.Contains(recorder.selectQuery, `WHERE (ler.turn_id = 'turn-1')`) {
		t.Fatalf("review=%#v query=%s", review, recorder.selectQuery)
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
