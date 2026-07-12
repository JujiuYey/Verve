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

type turnQueryRecorder struct {
	execQueries        []string
	selectQuery        string
	turnInsertAffected int64
	statusAffected     int64
	rows               driver.Rows
	rowQueue           []driver.Rows
	commits            int
	rollbacks          int
}

func (r *turnQueryRecorder) Connect(context.Context) (driver.Conn, error) {
	return turnQueryConn{recorder: r}, nil
}

func (r *turnQueryRecorder) Driver() driver.Driver { return turnQueryDriver{recorder: r} }

type turnQueryDriver struct{ recorder *turnQueryRecorder }

func (d turnQueryDriver) Open(string) (driver.Conn, error) {
	return turnQueryConn{recorder: d.recorder}, nil
}

type turnQueryConn struct{ recorder *turnQueryRecorder }

func (c turnQueryConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c turnQueryConn) Close() error                        { return nil }
func (c turnQueryConn) Begin() (driver.Tx, error)           { return turnQueryTx{recorder: c.recorder}, nil }
func (c turnQueryConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.recorder.execQueries = append(c.recorder.execQueries, query)
	switch {
	case strings.HasPrefix(query, "INSERT INTO \"learning_turns\""):
		return driver.RowsAffected(c.recorder.turnInsertAffected), nil
	case strings.HasPrefix(query, "UPDATE \"learning_turns\""):
		return driver.RowsAffected(c.recorder.statusAffected), nil
	default:
		return driver.RowsAffected(1), nil
	}
}
func (c turnQueryConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.recorder.selectQuery = query
	if len(c.recorder.rowQueue) > 0 {
		rows := c.recorder.rowQueue[0]
		c.recorder.rowQueue = c.recorder.rowQueue[1:]
		return rows, nil
	}
	return c.recorder.rows, nil
}

type turnQueryTx struct{ recorder *turnQueryRecorder }

func (t turnQueryTx) Commit() error   { t.recorder.commits++; return nil }
func (t turnQueryTx) Rollback() error { t.recorder.rollbacks++; return nil }

type learningTurnRows struct {
	values [][]driver.Value
	index  int
}

type learningMessageContentRows struct {
	value string
	done  bool
}

func (r *learningMessageContentRows) Columns() []string { return []string{"content"} }
func (r *learningMessageContentRows) Close() error      { return nil }
func (r *learningMessageContentRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

func (r *learningTurnRows) Columns() []string {
	return []string{"id", "session_id", "request_id", "agent_type", "status", "error_code", "error_message", "started_at", "completed_at", "created_at", "updated_at"}
}
func (r *learningTurnRows) Close() error { return nil }
func (r *learningTurnRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func newTurnRepositoryForTest(recorder *turnQueryRecorder) (*TurnRepository, func()) {
	sqldb := sql.OpenDB(recorder)
	db := bun.NewDB(sqldb, pgdialect.New())
	return NewTurnRepository(db), func() { _ = db.Close() }
}

func TestTurnRepositoryBeginListenerTurnCreatesTurnAndUserMessageAtomically(t *testing.T) {
	recorder := &turnQueryRecorder{turnInsertAffected: 1}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()

	result, err := repo.BeginListenerTurn(context.Background(), "session-1", "request-1", "值有具体类型")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Created || result.Turn.AgentType != learning_db.LearningAgentListener || result.Turn.Status != learning_db.LearningTurnProcessing {
		t.Fatalf("result = %#v", result)
	}
	if len(recorder.execQueries) != 2 {
		t.Fatalf("exec queries = %#v", recorder.execQueries)
	}
	for _, want := range []string{`'session-1'`, `'request-1'`, `'listener'`, `'processing'`} {
		if !strings.Contains(recorder.execQueries[0], want) {
			t.Fatalf("turn insert missing %q: %s", want, recorder.execQueries[0])
		}
	}
	for _, want := range []string{`"turn_id"`, `'user'`, `'值有具体类型'`} {
		if !strings.Contains(recorder.execQueries[1], want) {
			t.Fatalf("message insert missing %q: %s", want, recorder.execQueries[1])
		}
	}
	if recorder.commits != 1 || recorder.rollbacks != 0 {
		t.Fatalf("commits=%d rollbacks=%d", recorder.commits, recorder.rollbacks)
	}
}

func TestTurnRepositoryBeginTurnPersistsSelectedAgent(t *testing.T) {
	recorder := &turnQueryRecorder{turnInsertAffected: 1}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()

	result, err := repo.BeginTurn(context.Background(), BeginTurnInput{
		SessionID: "session-1", RequestID: "request-1", AgentType: learning_db.LearningAgentTeacher, Content: "请解释 channel",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Created || result.Turn.AgentType != learning_db.LearningAgentTeacher {
		t.Fatalf("result = %#v", result)
	}
	if !strings.Contains(recorder.execQueries[0], `'teacher'`) || !strings.Contains(recorder.execQueries[1], `'请解释 channel'`) {
		t.Fatalf("queries = %#v", recorder.execQueries)
	}
}

func TestTurnRepositoryBeginTurnRejectsIdempotencyPayloadMismatch(t *testing.T) {
	now := time.Now()
	recorder := &turnQueryRecorder{
		turnInsertAffected: 0,
		rowQueue: []driver.Rows{
			&learningTurnRows{values: [][]driver.Value{{
				"turn-existing", "session-1", "request-1", "listener", "completed", nil, nil,
				now, now, now, now,
			}}},
			&learningMessageContentRows{value: "原始解释"},
		},
	}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()

	_, err := repo.BeginTurn(context.Background(), BeginTurnInput{
		SessionID: "session-1", RequestID: "request-1", AgentType: learning_db.LearningAgentTeacher, Content: "不同问题",
	})
	if err != ErrTurnRequestConflict {
		t.Fatalf("error = %v", err)
	}
}

func TestTurnRepositoryBeginListenerTurnReturnsExistingWithoutDuplicateMessage(t *testing.T) {
	now := time.Now()
	recorder := &turnQueryRecorder{
		turnInsertAffected: 0,
		rowQueue: []driver.Rows{
			&learningTurnRows{values: [][]driver.Value{{
				"turn-existing", "session-1", "request-1", "listener", "completed", nil, nil,
				now, now, now, now,
			}}},
			&learningMessageContentRows{value: "值有具体类型"},
		},
	}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()

	result, err := repo.BeginListenerTurn(context.Background(), "session-1", "request-1", "值有具体类型")
	if err != nil {
		t.Fatal(err)
	}
	if result.Created || result.Turn.ID != "turn-existing" || result.Turn.Status != learning_db.LearningTurnCompleted {
		t.Fatalf("result = %#v", result)
	}
	if len(recorder.execQueries) != 1 {
		t.Fatalf("duplicate request executed extra writes: %#v", recorder.execQueries)
	}
}

func TestTurnRepositoryCompleteListenerTurnWritesResultAndCompletesProcessingTurn(t *testing.T) {
	recorder := &turnQueryRecorder{statusAffected: 1}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()
	review := &learning_db.LearningExplanationReview{
		HeardSummary: "听到了类型约束", ClearPoints: []string{"类型"}, ConfusingPoints: []string{},
		Misconceptions: []string{}, FollowUpQuestion: "何时检查？", ExplanationSummary: "理解类型约束", ContextSufficient: true,
	}

	if err := repo.CompleteListenerTurn(context.Background(), "session-1", "turn-1", `{"heard_summary":"听到了类型约束"}`, review); err != nil {
		t.Fatal(err)
	}
	if review.TurnID != "turn-1" || len(review.ID) != 32 {
		t.Fatalf("review = %#v", review)
	}
	if len(recorder.execQueries) != 3 {
		t.Fatalf("exec queries = %#v", recorder.execQueries)
	}
	if !strings.HasPrefix(recorder.execQueries[0], `UPDATE "learning_turns"`) || !strings.Contains(recorder.execQueries[0], `status = 'processing'`) {
		t.Fatalf("status update = %s", recorder.execQueries[0])
	}
	if !strings.Contains(recorder.execQueries[1], `'assistant'`) || !strings.Contains(recorder.execQueries[2], `"turn_id"`) {
		t.Fatalf("result writes = %#v", recorder.execQueries[1:])
	}
}

func TestTurnRepositoryCompleteListenerTurnRejectsNonProcessingTurn(t *testing.T) {
	recorder := &turnQueryRecorder{statusAffected: 0}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()

	err := repo.CompleteListenerTurn(context.Background(), "session-1", "turn-1", `{}`, &learning_db.LearningExplanationReview{})
	if err != ErrTurnNotProcessing {
		t.Fatalf("error = %v", err)
	}
	if recorder.commits != 0 || recorder.rollbacks != 1 || len(recorder.execQueries) != 1 {
		t.Fatalf("commits=%d rollbacks=%d queries=%#v", recorder.commits, recorder.rollbacks, recorder.execQueries)
	}
}

func TestTurnRepositoryRetryFailedTurnReturnsItToProcessing(t *testing.T) {
	recorder := &turnQueryRecorder{statusAffected: 1}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()

	if err := repo.RetryFailedTurn(context.Background(), "turn-1"); err != nil {
		t.Fatal(err)
	}
	if len(recorder.execQueries) != 1 {
		t.Fatalf("exec queries = %#v", recorder.execQueries)
	}
	query := recorder.execQueries[0]
	for _, want := range []string{`status = 'processing'`, `error_code = NULL`, `error_message = NULL`, `completed_at = NULL`, `status = 'failed'`} {
		if !strings.Contains(query, want) {
			t.Fatalf("retry update missing %q: %s", want, query)
		}
	}
}

func TestTurnRepositoryCompleteTeacherTurnWritesMessageAndIntervention(t *testing.T) {
	recorder := &turnQueryRecorder{statusAffected: 1}
	repo, closeDB := newTurnRepositoryForTest(recorder)
	defer closeDB()
	intervention := &learning_db.LearningTeachingIntervention{
		QuestionSummary: "channel 阻塞", ExplanationSummary: "讲解同步点",
		KnowledgeGaps: []string{}, KeyPoints: []string{"发送接收配对"}, Examples: []string{},
		Evidence: []learning_db.LearningTeachingEvidence{{ChunkID: "chunk-1", DocumentVersion: 2}},
	}

	if err := repo.CompleteTeacherTurn(context.Background(), "session-1", "turn-1", "先看同步点。", intervention); err != nil {
		t.Fatal(err)
	}
	if intervention.TurnID != "turn-1" || len(intervention.ID) != 32 {
		t.Fatalf("intervention = %#v", intervention)
	}
	if len(recorder.execQueries) != 3 || !strings.Contains(recorder.execQueries[2], `learning_teaching_interventions`) {
		t.Fatalf("queries = %#v", recorder.execQueries)
	}
}
