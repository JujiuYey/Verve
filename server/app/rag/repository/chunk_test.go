package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type chunkQueryRecorder struct {
	query string
	args  []driver.NamedValue
	rows  driver.Rows
}

func (r *chunkQueryRecorder) Connect(context.Context) (driver.Conn, error) {
	return chunkQueryConn{recorder: r}, nil
}

func (r *chunkQueryRecorder) Driver() driver.Driver { return chunkQueryDriver{recorder: r} }

type chunkQueryDriver struct{ recorder *chunkQueryRecorder }

func (d chunkQueryDriver) Open(string) (driver.Conn, error) {
	return chunkQueryConn{recorder: d.recorder}, nil
}

type chunkQueryConn struct{ recorder *chunkQueryRecorder }

func (c chunkQueryConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c chunkQueryConn) Close() error                        { return nil }
func (c chunkQueryConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c chunkQueryConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.recorder.query = query
	c.recorder.args = append([]driver.NamedValue(nil), args...)
	return c.recorder.rows, nil
}

type chunkRows struct {
	values [][]driver.Value
	index  int
}

func (r *chunkRows) Columns() []string { return []string{"id", "document_id", "chunk_index"} }
func (r *chunkRows) Close() error      { return nil }
func (r *chunkRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func newChunkRepositoryForTest(recorder *chunkQueryRecorder) (*ChunkRepository, func()) {
	sqldb := sql.OpenDB(recorder)
	db := bun.NewDB(sqldb, pgdialect.New())
	return NewChunkRepository(db), func() { _ = db.Close() }
}

func TestChunkRepositoryFindNeighborsScopesExpandsDeduplicatesAndOrders(t *testing.T) {
	recorder := &chunkQueryRecorder{rows: &chunkRows{values: [][]driver.Value{
		{"c1", "doc-1", int64(1)},
		{"c2", "doc-1", int64(2)},
		{"c3", "doc-1", int64(3)},
		{"c4", "doc-1", int64(4)},
	}}}
	repo, closeDB := newChunkRepositoryForTest(recorder)
	defer closeDB()

	chunks, err := repo.FindNeighbors(context.Background(), "doc-1", []int{2, 3, 2}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(recorder.query, `rwc.document_id = 'doc-1'`) {
		t.Fatalf("query does not scope document: %s", recorder.query)
	}
	if !strings.Contains(recorder.query, `rwc.chunk_index IN (1, 2, 3, 4)`) {
		t.Fatalf("query does not use expanded deduplicated indexes: %s", recorder.query)
	}
	if !strings.Contains(recorder.query, `ORDER BY rwc.chunk_index ASC`) {
		t.Fatalf("query does not order by chunk index: %s", recorder.query)
	}
	gotIndexes := make([]int, 0, len(chunks))
	for _, chunk := range chunks {
		gotIndexes = append(gotIndexes, chunk.ChunkIndex)
	}
	if !reflect.DeepEqual(gotIndexes, []int{1, 2, 3, 4}) {
		t.Fatalf("chunk indexes = %#v", gotIndexes)
	}
}

func TestChunkRepositoryFindNeighborsReturnsEmptyWithoutIndexes(t *testing.T) {
	recorder := &chunkQueryRecorder{}
	repo, closeDB := newChunkRepositoryForTest(recorder)
	defer closeDB()

	chunks, err := repo.FindNeighbors(context.Background(), "doc-1", nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("chunks = %#v", chunks)
	}
	if recorder.query != "" {
		t.Fatalf("unexpected query: %s", recorder.query)
	}
}

func TestChunkRepositoryFindNeighborsNormalizesNegativeRadiusToZero(t *testing.T) {
	recorder := &chunkQueryRecorder{rows: &chunkRows{}}
	repo, closeDB := newChunkRepositoryForTest(recorder)
	defer closeDB()

	_, err := repo.FindNeighbors(context.Background(), "doc-1", []int{5}, -2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(recorder.query, `rwc.chunk_index IN (5)`) {
		t.Fatalf("query expanded a negative radius: %s", recorder.query)
	}
}
