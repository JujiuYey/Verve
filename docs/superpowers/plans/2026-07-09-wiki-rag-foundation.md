# Wiki RAG Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first Verve RAG foundation so learning agents can retrieve markdown-aware, wiki-scoped knowledge before `wiki_agent_instances` are introduced.

**Architecture:** Add a backend `rag` domain that turns wiki markdown into structured chunks, embeds those chunks, stores chunk metadata in PostgreSQL, stores vectors in Qdrant, and exposes scoped retrieval as both an HTTP API and a Coach agent tool. The first version is structured RAG, not full Agentic RAG: retrieval is deterministic and scoped by wiki root, while the agent-facing tool creates the later upgrade path for multi-step Agentic RAG.

**Tech Stack:** Go 1.25, Fiber, Bun/PostgreSQL, MinIO-backed wiki documents, Qdrant, OpenAI-compatible embedding models from existing system model configuration, Eino ADK tools.

---

## Scope

This plan covers the RAG foundation only:

- Markdown-aware chunking.
- Chunk metadata persistence.
- Embedding model adapter.
- Qdrant vector storage.
- Wiki-root scoped retrieval.
- Coach tool integration.

This plan deliberately excludes `wiki_agent_instances`, multi-agent orchestration, Graph RAG, and server-side Feynman workflow changes. Those should become separate plans after retrieval is reliable.

## File Structure

- Create `server/migrations/rag/001_schema.sql`: PostgreSQL schema for chunk metadata and indexing jobs.
- Create `server/app/rag/models/db/chunk.go`: Bun model for `rag_wiki_chunks`.
- Create `server/app/rag/models/db/index_job.go`: Bun model for `rag_index_jobs`.
- Create `server/app/rag/models/payload/retrieval.go`: request/response payloads for retrieval and indexing endpoints.
- Create `server/app/rag/repository/chunk.go`: chunk metadata repository.
- Create `server/app/rag/repository/index_job.go`: indexing job repository.
- Create `server/app/rag/service/markdown_chunker.go`: markdown-aware chunker.
- Create `server/app/rag/service/root_resolver.go`: resolves any folder/document to a root folder and folder path.
- Create `server/app/rag/service/embedding.go`: embedding adapter using active default embedding model.
- Create `server/infrastructure/vector/qdrant.go`: Qdrant collection and point operations.
- Create `server/app/rag/service/indexer.go`: document indexing orchestration.
- Create `server/app/rag/service/retriever.go`: scoped retrieval orchestration.
- Create `server/app/rag/handlers/rag.go`: admin/dev indexing and retrieval endpoints.
- Create `server/app/rag/router/rag.go`: route registration.
- Modify `server/infrastructure/database/database.go`: register RAG models and repositories.
- Modify `server/router/router.go`: register RAG routes.
- Modify `server/app/wiki/handlers/document.go`: trigger indexing after upload/update/delete.
- Modify `server/app/learning/tools/coach_tools.go`: add `search_wiki_knowledge`.
- Modify `server/app/learning/handlers/coach.go`: pass RAG service/vector client into Coach tools after service wiring is available.

## Data Model

### `rag_wiki_chunks`

The PostgreSQL table owns chunk metadata and source traceability. Qdrant owns vector similarity search.

```sql
CREATE TABLE rag_wiki_chunks (
  id VARCHAR(32) PRIMARY KEY,
  root_folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
  folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
  document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
  document_title VARCHAR(255) NOT NULL,
  folder_path TEXT NOT NULL,
  heading_path TEXT NOT NULL,
  chunk_index INTEGER NOT NULL,
  block_type VARCHAR(32) NOT NULL,
  content TEXT NOT NULL,
  content_hash VARCHAR(64) NOT NULL,
  token_count INTEGER NOT NULL DEFAULT 0,
  vector_point_id VARCHAR(64) NOT NULL,
  embedding_model VARCHAR(128) NOT NULL,
  indexed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uk_rag_chunks_doc_hash UNIQUE (document_id, content_hash),
  CONSTRAINT uk_rag_chunks_doc_index UNIQUE (document_id, chunk_index)
);
```

### `rag_index_jobs`

The job table gives indexing visibility without introducing a queue worker in the first version.

```sql
CREATE TABLE rag_index_jobs (
  id VARCHAR(32) PRIMARY KEY,
  document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
  root_folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE SET NULL,
  status VARCHAR(32) NOT NULL,
  error_message TEXT,
  chunk_count INTEGER NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_rag_index_jobs_status CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);
```

## Task 1: Add RAG Schema and Repository Models

**Files:**
- Create: `server/migrations/rag/001_schema.sql`
- Create: `server/app/rag/models/db/chunk.go`
- Create: `server/app/rag/models/db/index_job.go`
- Modify: `server/infrastructure/database/database.go`

- [ ] **Step 1: Write schema migration**

Create `server/migrations/rag/001_schema.sql` with the two tables from the Data Model section plus these indexes:

```sql
CREATE INDEX idx_rag_chunks_root_folder ON rag_wiki_chunks(root_folder_id);
CREATE INDEX idx_rag_chunks_document ON rag_wiki_chunks(document_id);
CREATE INDEX idx_rag_chunks_folder ON rag_wiki_chunks(folder_id);
CREATE INDEX idx_rag_chunks_heading ON rag_wiki_chunks(heading_path);
CREATE INDEX idx_rag_jobs_document_status ON rag_index_jobs(document_id, status);
CREATE INDEX idx_rag_jobs_status_created ON rag_index_jobs(status, created_at DESC);
```

- [ ] **Step 2: Create Bun models**

Create `server/app/rag/models/db/chunk.go`:

```go
package db

import (
	"time"

	"github.com/uptrace/bun"
)

type WikiChunk struct {
	bun.BaseModel `bun:"table:rag_wiki_chunks,alias:rwc"`

	ID             string    `bun:"id,pk,type:varchar(32)" json:"id"`
	RootFolderID   string    `bun:"root_folder_id,type:varchar(32),notnull" json:"root_folder_id"`
	FolderID       string    `bun:"folder_id,type:varchar(32),notnull" json:"folder_id"`
	DocumentID     string    `bun:"document_id,type:varchar(32),notnull" json:"document_id"`
	DocumentTitle  string    `bun:"document_title,notnull" json:"document_title"`
	FolderPath     string    `bun:"folder_path,notnull" json:"folder_path"`
	HeadingPath    string    `bun:"heading_path,notnull" json:"heading_path"`
	ChunkIndex     int       `bun:"chunk_index,notnull" json:"chunk_index"`
	BlockType       string    `bun:"block_type,notnull" json:"block_type"`
	Content         string    `bun:"content,notnull" json:"content"`
	ContentHash     string    `bun:"content_hash,notnull" json:"content_hash"`
	TokenCount      int       `bun:"token_count,notnull" json:"token_count"`
	VectorPointID   string    `bun:"vector_point_id,notnull" json:"vector_point_id"`
	EmbeddingModel  string    `bun:"embedding_model,notnull" json:"embedding_model"`
	IndexedAt       time.Time `bun:"indexed_at,nullzero,notnull,default:current_timestamp" json:"indexed_at"`
	CreatedAt       time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
```

Create `server/app/rag/models/db/index_job.go`:

```go
package db

import (
	"time"

	"github.com/uptrace/bun"
)

type IndexJob struct {
	bun.BaseModel `bun:"table:rag_index_jobs,alias:rij"`

	ID           string     `bun:"id,pk,type:varchar(32)" json:"id"`
	DocumentID   string     `bun:"document_id,type:varchar(32),notnull" json:"document_id"`
	RootFolderID *string    `bun:"root_folder_id,type:varchar(32)" json:"root_folder_id"`
	Status       string     `bun:"status,notnull" json:"status"`
	ErrorMessage *string    `bun:"error_message" json:"error_message"`
	ChunkCount   int        `bun:"chunk_count,notnull" json:"chunk_count"`
	StartedAt    *time.Time `bun:"started_at" json:"started_at"`
	FinishedAt   *time.Time `bun:"finished_at" json:"finished_at"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
```

- [ ] **Step 3: Register models in database service**

Modify `server/infrastructure/database/database.go` imports and `RegisterModel` calls:

```go
rag_db "verve/app/rag/models/db"
rag_repo "verve/app/rag/repository"
```

Add fields:

```go
RAGChunks *rag_repo.ChunkRepository
RAGJobs   *rag_repo.IndexJobRepository
```

Register models:

```go
db.RegisterModel((*rag_db.WikiChunk)(nil))
db.RegisterModel((*rag_db.IndexJob)(nil))
```

Initialize repositories after they are created in Task 2:

```go
RAGChunks: rag_repo.NewChunkRepository(db),
RAGJobs:   rag_repo.NewIndexJobRepository(db),
```

- [ ] **Step 4: Verify compile fails until repositories exist**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./infrastructure/database
```

Expected: compile failure mentioning missing `verve/app/rag/repository` until Task 2 creates it.

## Task 2: Add Chunk and Job Repositories

**Files:**
- Create: `server/app/rag/repository/chunk.go`
- Create: `server/app/rag/repository/index_job.go`
- Test: `server/app/rag/repository/chunk_test.go`

- [ ] **Step 1: Create chunk repository**

Create `server/app/rag/repository/chunk.go`:

```go
package repository

import (
	"context"

	rag_db "verve/app/rag/models/db"

	"github.com/uptrace/bun"
)

type ChunkRepository struct {
	db *bun.DB
}

func NewChunkRepository(db *bun.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) ReplaceDocumentChunks(ctx context.Context, documentID string, chunks []*rag_db.WikiChunk) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().Model((*rag_db.WikiChunk)(nil)).Where("document_id = ?", documentID).Exec(ctx); err != nil {
			return err
		}
		if len(chunks) == 0 {
			return nil
		}
		_, err := tx.NewInsert().Model(&chunks).Exec(ctx)
		return err
	})
}

func (r *ChunkRepository) FindByPointIDs(ctx context.Context, pointIDs []string) ([]*rag_db.WikiChunk, error) {
	if len(pointIDs) == 0 {
		return []*rag_db.WikiChunk{}, nil
	}
	var chunks []*rag_db.WikiChunk
	err := r.db.NewSelect().
		Model(&chunks).
		Where("vector_point_id IN (?)", bun.In(pointIDs)).
		Scan(ctx)
	return chunks, err
}

func (r *ChunkRepository) DeleteByDocument(ctx context.Context, documentID string) error {
	_, err := r.db.NewDelete().Model((*rag_db.WikiChunk)(nil)).Where("document_id = ?", documentID).Exec(ctx)
	return err
}
```

- [ ] **Step 2: Create index job repository**

Create `server/app/rag/repository/index_job.go`:

```go
package repository

import (
	"context"
	"time"

	rag_db "verve/app/rag/models/db"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type IndexJobRepository struct {
	db *bun.DB
}

func NewIndexJobRepository(db *bun.DB) *IndexJobRepository {
	return &IndexJobRepository{db: db}
}

func (r *IndexJobRepository) CreatePending(ctx context.Context, documentID string) (*rag_db.IndexJob, error) {
	job := &rag_db.IndexJob{
		ID:         compactUUID(),
		DocumentID: documentID,
		Status:     "pending",
	}
	_, err := r.db.NewInsert().Model(job).Exec(ctx)
	return job, err
}

func (r *IndexJobRepository) MarkRunning(ctx context.Context, jobID string, rootFolderID string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "running").
		Set("root_folder_id = ?", rootFolderID).
		Set("started_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func (r *IndexJobRepository) MarkCompleted(ctx context.Context, jobID string, chunkCount int) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "completed").
		Set("chunk_count = ?", chunkCount).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func (r *IndexJobRepository) MarkFailed(ctx context.Context, jobID string, message string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*rag_db.IndexJob)(nil)).
		Set("status = ?", "failed").
		Set("error_message = ?", message).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", jobID).
		Exec(ctx)
	return err
}

func compactUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:32]
}
```

Add `strings` to the imports in this file.

- [ ] **Step 3: Run repository compile**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/repository ./infrastructure/database
```

Expected: pass after imports are corrected.

## Task 3: Implement Markdown-Aware Chunking

**Files:**
- Create: `server/app/rag/service/markdown_chunker.go`
- Test: `server/app/rag/service/markdown_chunker_test.go`

- [ ] **Step 1: Write chunker tests**

Create tests that verify headings are preserved, code fences are not split, and empty markdown yields no chunks.

```go
func TestMarkdownChunkerPreservesHeadingPath(t *testing.T) {
	chunker := NewMarkdownChunker(1200)
	chunks := chunker.Chunk("# Go\n\n## Channel\n\nChannel is a typed pipe.\n\n### Close\n\nClosing broadcasts completion.")
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].HeadingPath != "Go > Channel" {
		t.Fatalf("heading path = %q", chunks[0].HeadingPath)
	}
	if chunks[1].HeadingPath != "Go > Channel > Close" {
		t.Fatalf("heading path = %q", chunks[1].HeadingPath)
	}
}

func TestMarkdownChunkerKeepsCodeFenceTogether(t *testing.T) {
	chunker := NewMarkdownChunker(80)
	chunks := chunker.Chunk("## Example\n\n```go\nfunc main() {\n println(\"hi\")\n}\n```\n\nAfter text.")
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	if !strings.Contains(chunks[0].Content, "func main") {
		t.Fatalf("code fence was split away: %q", chunks[0].Content)
	}
}
```

- [ ] **Step 2: Implement chunker**

Create `server/app/rag/service/markdown_chunker.go`:

```go
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

type MarkdownChunk struct {
	HeadingPath string
	BlockType   string
	Content     string
	ContentHash string
	TokenCount  int
}

type MarkdownChunker struct {
	maxChars int
}

func NewMarkdownChunker(maxChars int) *MarkdownChunker {
	if maxChars <= 0 {
		maxChars = 1800
	}
	return &MarkdownChunker{maxChars: maxChars}
}

func (c *MarkdownChunker) Chunk(markdown string) []MarkdownChunk {
	lines := strings.Split(markdown, "\n")
	headings := make([]string, 0, 4)
	chunks := make([]MarkdownChunk, 0)
	var buf strings.Builder
	blockType := "paragraph"
	inFence := false

	flush := func() {
		content := strings.TrimSpace(buf.String())
		if content == "" {
			buf.Reset()
			return
		}
		path := strings.Join(headings, " > ")
		if path == "" {
			path = "Document"
		}
		chunks = append(chunks, MarkdownChunk{
			HeadingPath: path,
			BlockType:   blockType,
			Content:     content,
			ContentHash: hashText(path + "\n" + content),
			TokenCount:  estimateTokens(content),
		})
		buf.Reset()
		blockType = "paragraph"
	}

	headingRE := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			blockType = "code"
			buf.WriteString(line + "\n")
			continue
		}
		if !inFence {
			if match := headingRE.FindStringSubmatch(trimmed); match != nil {
				flush()
				level := len(match[1])
				title := strings.TrimSpace(match[2])
				if level <= len(headings) {
					headings = headings[:level-1]
				}
				headings = append(headings, title)
				continue
			}
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || regexp.MustCompile(`^\d+\.\s+`).MatchString(trimmed) {
				blockType = "list"
			}
			if buf.Len() > c.maxChars && trimmed == "" {
				flush()
				continue
			}
		}
		buf.WriteString(line + "\n")
	}
	flush()
	return chunks
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func estimateTokens(text string) int {
	runes := []rune(text)
	return (len(runes) + 3) / 4
}
```

- [ ] **Step 3: Run chunker tests**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/service -run TestMarkdownChunker -v
```

Expected: pass.

## Task 4: Add Root Folder Resolution

**Files:**
- Create: `server/app/rag/service/root_resolver.go`
- Test: `server/app/rag/service/root_resolver_test.go`

- [ ] **Step 1: Implement root resolver contract**

Create a resolver that walks `wiki_folders.parent_id` until the root. It returns:

```go
type FolderScope struct {
	RootFolderID string
	FolderPath   string
}
```

The implementation should load parent folders through `db.Folders.FindOne(ctx, id)` and build paths such as `Go Tutorial/Concurrency/Channel`.

- [ ] **Step 2: Add cycle and missing-parent protection**

Use a `seen := map[string]bool{}` guard. If a folder points to itself or a cycle is found, return an error:

```go
return FolderScope{}, fmt.Errorf("folder cycle detected at %s", folder.ID)
```

If a parent id exists but cannot be loaded, return the repository error.

- [ ] **Step 3: Run service tests**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/service -run TestRootResolver -v
```

Expected: pass with stubbed folder repository tests.

## Task 5: Add Embedding Adapter

**Files:**
- Create: `server/app/rag/service/embedding.go`
- Test: `server/app/rag/service/embedding_test.go`
- Modify: `server/app/system/repository/model.go` if the repository lacks a helper for default embedding model lookup.

- [ ] **Step 1: Define embedding interface**

Create:

```go
type Embedder interface {
	EmbedTexts(ctx context.Context, texts []string) (EmbeddingResult, error)
}

type EmbeddingResult struct {
	Model      string
	Dimension  int
	Embeddings [][]float32
}
```

- [ ] **Step 2: Implement OpenAI-compatible embedding request**

Use the active default `sys_models.model_type = 'embedding'` and its enabled platform. The request path is:

```text
POST {platform.base_url}/embeddings
Authorization: Bearer {api_key}
Content-Type: application/json
```

Request body:

```json
{
  "model": "embedding-model-name",
  "input": ["chunk text 1", "chunk text 2"]
}
```

Response parsing should support:

```json
{
  "data": [
    { "embedding": [0.1, 0.2, 0.3] }
  ],
  "model": "embedding-model-name"
}
```

- [ ] **Step 3: Keep key handling compatible with current system config**

If the existing platform repository does not expose decrypted API keys, the first implementation may read `api_key_ciphertext` directly only if current runtime already stores plaintext there. If it is encrypted, return a clear error:

```go
return EmbeddingResult{}, errors.New("embedding model platform key decryption is not wired")
```

Do not hard-code API keys in this adapter.

- [ ] **Step 4: Unit test request shape**

Use `httptest.Server` to assert method, path, authorization header, model, and input array. Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/service -run TestOpenAICompatibleEmbedder -v
```

Expected: pass.

## Task 6: Add Qdrant Vector Store

**Files:**
- Create: `server/infrastructure/vector/qdrant.go`
- Test: `server/infrastructure/vector/qdrant_test.go`
- Modify: `server/main.go`
- Modify: `server/router/router.go`

- [ ] **Step 1: Define vector store interface**

Create:

```go
type ScoredPoint struct {
	PointID string
	Score   float64
}

type Store interface {
	EnsureCollection(ctx context.Context, collection string, dimension int) error
	Upsert(ctx context.Context, collection string, points []Point) error
	Search(ctx context.Context, collection string, vector []float32, filter map[string]any, limit int) ([]ScoredPoint, error)
	DeleteByDocument(ctx context.Context, collection string, documentID string) error
}

type Point struct {
	ID      string
	Vector  []float32
	Payload map[string]any
}
```

- [ ] **Step 2: Implement Qdrant HTTP calls**

Use `config.GetQdrantConfig().URL`. Use one collection name:

```go
const WikiChunkCollection = "verve_wiki_chunks"
```

Required Qdrant endpoints:

```text
PUT /collections/{collection}
PUT /collections/{collection}/points?wait=true
POST /collections/{collection}/points/search
POST /collections/{collection}/points/delete?wait=true
```

Search filter must include:

```json
{
  "must": [
    { "key": "root_folder_id", "match": { "value": "..." } }
  ]
}
```

- [ ] **Step 3: Wire vector store into app startup**

In `server/main.go`, create the Qdrant client from config and pass it into route setup. Keep the app startable even if Qdrant is down; the retrieval/indexing endpoint should return a clear runtime error when vector calls fail.

- [ ] **Step 4: Unit test Qdrant request shape**

Use `httptest.Server` and run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./infrastructure/vector -v
```

Expected: pass.

## Task 7: Implement Document Indexer

**Files:**
- Create: `server/app/rag/service/indexer.go`
- Modify: `server/app/wiki/handlers/document.go`
- Test: `server/app/rag/service/indexer_test.go`

- [ ] **Step 1: Define indexer API**

Create:

```go
type Indexer struct {
	chunks   *repository.ChunkRepository
	jobs     *repository.IndexJobRepository
	folders  wiki_repo.FolderRepository
	docs     *wiki_repo.DocumentRepository
	minio    *storage.MinIOService
	chunker  *MarkdownChunker
	resolver *RootResolver
	embedder Embedder
	vectors  vector.Store
}

func (s *Indexer) IndexDocument(ctx context.Context, documentID string) error
func (s *Indexer) DeleteDocumentIndex(ctx context.Context, documentID string) error
```

- [ ] **Step 2: Implement index flow**

`IndexDocument` should:

```text
1. create pending job
2. load wiki document
3. load markdown from MinIO
4. resolve root_folder_id and folder_path
5. chunk markdown
6. embed chunk contents
7. ensure Qdrant collection using embedding dimension
8. upsert Qdrant points with payload root_folder_id, folder_id, document_id, heading_path
9. replace PostgreSQL chunk metadata for the document
10. mark job completed
```

On error, mark the job failed and include the error message.

- [ ] **Step 3: Trigger indexing from wiki document mutations**

Modify `server/app/wiki/handlers/document.go`:

- After successful upload, call `go indexer.IndexDocument(context.Background(), doc.ID)` if an indexer is configured.
- After successful content update, call the same indexing method.
- After delete, call `indexer.DeleteDocumentIndex(c.Context(), docID)`.

Keep upload/update successful even if async indexing fails. The job table records failures.

- [ ] **Step 4: Run service tests**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/service -run TestIndexer -v
```

Expected: pass with fake embedder and fake vector store.

## Task 8: Implement Scoped Retriever

**Files:**
- Create: `server/app/rag/service/retriever.go`
- Create: `server/app/rag/models/payload/retrieval.go`
- Test: `server/app/rag/service/retriever_test.go`

- [ ] **Step 1: Define retrieval payloads**

Create:

```go
type SearchRequest struct {
	RootFolderID string `json:"root_folder_id"`
	Query        string `json:"query"`
	Limit        int    `json:"limit"`
}

type SearchResult struct {
	ChunkID       string  `json:"chunk_id"`
	Score         float64 `json:"score"`
	RootFolderID  string  `json:"root_folder_id"`
	FolderID      string  `json:"folder_id"`
	DocumentID    string  `json:"document_id"`
	DocumentTitle string  `json:"document_title"`
	FolderPath    string  `json:"folder_path"`
	HeadingPath   string  `json:"heading_path"`
	Content        string  `json:"content"`
}
```

- [ ] **Step 2: Implement retriever**

`Search(ctx, rootFolderID, query string, limit int)` should:

```text
1. reject empty root_folder_id
2. reject empty query
3. embed the query
4. search Qdrant with root_folder_id filter
5. load chunk metadata by vector point ids
6. preserve Qdrant score order
7. return at most limit results
```

Default limit is 6. Maximum limit is 12.

- [ ] **Step 3: Run retriever tests**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/service -run TestRetriever -v
```

Expected: pass.

## Task 9: Add RAG HTTP Routes

**Files:**
- Create: `server/app/rag/handlers/rag.go`
- Create: `server/app/rag/router/rag.go`
- Modify: `server/router/router.go`

- [ ] **Step 1: Add handlers**

Expose:

```text
POST /api/rag/wiki/documents/:id/index
DELETE /api/rag/wiki/documents/:id/index
POST /api/rag/wiki/search
```

The search body is:

```json
{
  "root_folder_id": "folder id",
  "query": "channel close 怎么理解",
  "limit": 6
}
```

- [ ] **Step 2: Register routes**

Add `rag_router.SetupRAGRoutes(protected.Group("/"), dbService, minioService, vectorStore)` in `server/router/router.go`.

- [ ] **Step 3: Verify route compile**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/rag/... ./router
```

Expected: pass.

## Task 10: Add Coach Retrieval Tool

**Files:**
- Modify: `server/app/learning/tools/coach_tools.go`
- Modify: `server/app/learning/handlers/coach.go`
- Modify: `server/infrastructure/llm/prompts/coach.go`
- Test: `server/app/learning/tools/coach_tools_test.go`

- [ ] **Step 1: Add tool input and output**

Add:

```go
type SearchWikiKnowledgeInput struct {
	RootFolderID string `json:"root_folder_id" jsonschema_description:"限定检索的 Wiki 根目录 ID"`
	Query        string `json:"query" jsonschema_description:"要检索的学习问题或概念"`
	Limit        int    `json:"limit" jsonschema_description:"最多返回多少个片段,默认 6"`
}
```

- [ ] **Step 2: Add `search_wiki_knowledge` tool**

The tool must call `rag.Retriever.Search`. Tool output includes:

```json
{
  "results": [
    {
      "document_title": "channel.md",
      "folder_path": "Go Tutorial/Concurrency",
      "heading_path": "Channel > Close",
      "content": "..."
    }
  ]
}
```

- [ ] **Step 3: Update Coach prompt**

Update `server/infrastructure/llm/prompts/coach.go` so Coach behavior says:

```text
When the user asks to continue learning or asks a concept question, first identify the relevant wiki root/folder. If a root folder is known, call search_wiki_knowledge before planning the next step. Do not invent document content from filenames alone.
```

- [ ] **Step 4: Run learning tests**

Run:

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./app/learning/... ./infrastructure/llm/...
```

Expected: pass.

## Task 11: Manual Verification Flow

**Files:**
- No source files created in this task.

- [ ] **Step 1: Start dependencies**

Start PostgreSQL, MinIO, and Qdrant using the project’s existing local workflow. Confirm Qdrant is reachable:

```bash
curl http://localhost:6333/collections
```

Expected: JSON response with `result.collections`.

- [ ] **Step 2: Upload or update a markdown wiki document**

Use the existing `/wiki` UI or API. Confirm an index job exists:

```sql
SELECT document_id, status, chunk_count, error_message
FROM rag_index_jobs
ORDER BY created_at DESC
LIMIT 5;
```

Expected: latest job is `completed` and `chunk_count > 0`.

- [ ] **Step 3: Search within a wiki root**

Set the root folder id first, then run the request:

```bash
ROOT_FOLDER_ID="put-a-real-root-folder-id-here"
curl -X POST http://localhost:8000/api/rag/wiki/search \
  -H 'Content-Type: application/json' \
  -d "{\"root_folder_id\":\"${ROOT_FOLDER_ID}\",\"query\":\"这个文档的核心概念是什么\",\"limit\":6}"
```

Expected: response includes chunks with `document_title`, `folder_path`, `heading_path`, and `content`.

- [ ] **Step 4: Confirm scoped retrieval**

Repeat the search with a different `root_folder_id`.

Expected: results do not include chunks from the first wiki root.

## Task 12: Final Verification and Commit

**Files:**
- All files touched by Tasks 1-10.

- [ ] **Step 1: Run backend tests**

```bash
cd /Users/jujiuyey/Projects/Verve/server && go test ./...
```

Expected: pass.

- [ ] **Step 2: Run frontend type check only if API types are touched**

```bash
cd /Users/jujiuyey/Projects/Verve/web && pnpm lint:type
```

Expected: pass or only pre-existing unrelated failures.

- [ ] **Step 3: Check formatting**

```bash
cd /Users/jujiuyey/Projects/Verve && git diff --check
```

Expected: no whitespace errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/jujiuyey/Projects/Verve
git add server/migrations/rag server/app/rag server/infrastructure/vector server/infrastructure/database/database.go server/router/router.go server/main.go server/app/wiki/handlers/document.go server/app/learning/tools/coach_tools.go server/app/learning/handlers/coach.go server/infrastructure/llm/prompts/coach.go
git commit -m "feat: add wiki-scoped RAG foundation"
```

## Design Notes

- First version is structured RAG, not naive chunking. Chunks retain wiki root, folder path, document title, heading path, block type, and content hash.
- Retrieval must always be scoped by `root_folder_id`; global search is not part of this plan.
- Qdrant is used because the repo already has `server/config/qdrant.go`; PostgreSQL remains the source of truth for metadata.
- The RAG tool is added to Coach so the current learning flow can use retrieval before `wiki_agent_instances` exists.
- Agentic RAG begins after this plan: once `search_wiki_knowledge` is reliable, a future `WikiLearningAgent` can perform multi-step retrieval by calling the same tool repeatedly.

## Self-Review

- Spec coverage: the plan covers structured chunking, scoped retrieval, embedding, Qdrant storage, HTTP search, and Coach tool access.
- Scope check: `wiki_agent_instances`, Graph RAG, and workflow changes are explicitly outside this plan.
- Type consistency: chunk metadata fields match the schema, Bun model, Qdrant payload, and retrieval response names.
- Verification: backend tests, route compile, Qdrant manual smoke test, scoped retrieval check, and whitespace check are included.
