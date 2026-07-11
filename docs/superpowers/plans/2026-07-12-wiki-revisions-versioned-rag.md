# Wiki Revisions and Version-Aware RAG Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every Wiki content write create an immutable revision and make RAG indexing and retrieval honor the document's current version.

**Architecture:** A Wiki `DocumentVersionService` writes immutable MinIO objects, then commits document metadata, revision metadata, and a version-pinned RAG job in PostgreSQL. The RAG indexer reads only the job's object path and publishes chunks only when the job version still equals the document current version.

**Tech Stack:** Go 1.25, Fiber, Bun, PostgreSQL, MinIO, Qdrant

---

### Task 1: Add Wiki and RAG version schemas

**Files:**
- Create: `server/migrations/wiki/003_document_versions.sql`
- Create: `server/migrations/rag/002_document_versions.sql`
- Create: `server/app/wiki/models/db/revision.go`
- Create: `server/app/wiki/models/db/change_request.go`
- Modify: `server/app/wiki/models/db/document.go`
- Modify: `server/app/rag/models/db/index_job.go`
- Modify: `server/app/rag/models/db/chunk.go`
- Modify: `server/app/rag/models/payload/retrieval.go`
- Modify: `server/infrastructure/database/database.go`
- Test: `server/app/wiki/models/db/wiki_model_test.go`
- Create: `server/app/rag/models/db/version_test.go`

- [ ] **Step 1: Write failing model tests**

Assert the new models and fields compile and expose `current_version`, `content_hash`, `document_version`, `object_path`, and the approved change-request statuses.

```go
func TestDocumentVersionModels(t *testing.T) {
    if _, ok := reflect.TypeOf(Document{}).FieldByName("CurrentVersion"); !ok {
        t.Fatal("Document.CurrentVersion is missing")
    }
    if _, ok := reflect.TypeOf(DocumentRevision{}).FieldByName("ObjectPath"); !ok {
        t.Fatal("DocumentRevision.ObjectPath is missing")
    }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/wiki/models/db ./app/rag/models/db`

Expected: FAIL because revision models and version fields do not exist.

- [ ] **Step 3: Add migrations and Bun models**

Use these stable model shapes:

```go
type DocumentRevision struct {
    ID string
    DocumentID string
    Version int64
    ObjectPath string
    ContentHash string
    FileSize int64
    ChangeRequestID *string
    ChangedBy string
    ChangeSummary string
    CreatedAt time.Time
}

type DocumentChangeRequest struct {
    ID string
    DocumentID string
    RequestedBy string
    SourceType string
    SourceID string
    RequestID string
    ReplacesChangeRequestID *string
    BaseVersion int64
    Instruction string
    ChangeSummary string
    ProposedContent string
    ProposedDiff string
    Status string
    ErrorMessage *string
    AppliedVersion *int64
    CreatedAt time.Time
    UpdatedAt time.Time
    AppliedAt *time.Time
}
```

Give every business field a concise Chinese trailing comment. The Wiki migration assigns existing documents version 1 without inventing hashes or revisions. The RAG migration assigns existing chunks and jobs version 1, adds job `object_path`, adds `superseded`, changes chunk uniqueness to include version, marks all but the newest historical job per document superseded with `row_number()`, and adds a partial unique index on `(document_id, document_version) WHERE status <> 'superseded'`.

- [ ] **Step 4: Register the new models and verify GREEN**

Run: `cd server && go test ./app/wiki/models/db ./app/rag/models/db`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/migrations/wiki server/migrations/rag server/app/wiki/models server/app/rag/models server/infrastructure/database/database.go
git commit -m "feat(wiki): add immutable document version schema"
```

### Task 2: Add revision, change-request, and version transaction repositories

**Files:**
- Create: `server/app/wiki/repository/revision.go`
- Create: `server/app/wiki/repository/change_request.go`
- Create: `server/app/wiki/repository/version.go`
- Create: `server/app/wiki/repository/version_test.go`
- Modify: `server/app/wiki/repository/document_repo.go`
- Modify: `server/infrastructure/database/database.go`

- [ ] **Step 1: Write failing transaction tests**

Cover initial revision creation, lazy revision-1 snapshot, direct edit, change-request apply, repeated apply, cancel, and version conflict. Verify the repository locks both request and document with `FOR UPDATE`.

```go
type ApplyVersionInput struct {
    ChangeRequestID *string
    DocumentID string
    ExpectedVersion int64
    ObjectPath string
    ContentHash string
    FileSize int64
    ChangedBy string
    ChangeSummary string
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/wiki/repository -run 'Version|ChangeRequest'`

Expected: FAIL because the repositories do not exist.

- [ ] **Step 3: Implement repository contracts**

Implement:

```go
CreateInitial(ctx, document, revision, job) error
EnsureLegacyRevision(ctx, snapshot) error
CreateProposal(ctx, request) error
FindChangeRequest(ctx, id string) (*DocumentChangeRequest, error)
ListRevisionObjectPaths(ctx, documentID string) ([]string, error)
ApplyChangeRequest(ctx, input ApplyVersionInput) (*DocumentRevision, *IndexJob, error)
ApplyDirectEdit(ctx, input ApplyVersionInput) (*DocumentRevision, *IndexJob, error)
CancelChangeRequest(ctx, id, userID string) error
```

Return sentinel errors `ErrVersionConflict`, `ErrChangeRequestNotProposed`, and `ErrChangeRequestForbidden`. Repeated apply returns the existing revision rather than inserting another row.

- [ ] **Step 4: Run repository tests and verify GREEN**

Run: `cd server && go test ./app/wiki/repository`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/app/wiki/repository server/infrastructure/database/database.go
git commit -m "feat(wiki): persist document revisions atomically"
```

### Task 3: Implement the document-version service and immutable storage paths

**Files:**
- Create: `server/app/wiki/service/document_version.go`
- Create: `server/app/wiki/service/document_version_test.go`
- Modify: `server/infrastructure/storage/minio.go`

- [ ] **Step 1: Write failing service tests**

Verify SHA-256 hashing, immutable path generation, lazy legacy snapshot, MinIO-before-database ordering, orphan-safe database failure, direct edits, apply replay, and conflicts.

```go
func revisionObjectPath(documentID string, version int64, filename string) string {
    return fmt.Sprintf("documents/%s/revisions/%d/%s", documentID, version, filename)
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/wiki/service`

Expected: FAIL because `DocumentVersionService` does not exist.

- [ ] **Step 3: Implement service methods**

```go
CreateInitial(ctx context.Context, input InitialDocumentInput) (*Document, *IndexJob, error)
SaveDirectEdit(ctx context.Context, userID, documentID, content string) (*DocumentRevision, *IndexJob, error)
ApplyChangeRequest(ctx context.Context, userID, requestID string) (*DocumentRevision, *IndexJob, error)
CancelChangeRequest(ctx context.Context, userID, requestID string) error
```

Add a MinIO copy helper only if the SDK can perform a server-side copy; otherwise read and write the legacy object through existing content APIs. Never overwrite `Document.FilePath` in place.

- [ ] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/wiki/service`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/app/wiki/service server/infrastructure/storage/minio.go
git commit -m "feat(wiki): version every document content write"
```

### Task 4: Route uploads, manual saves, apply, and cancel through versioning

**Files:**
- Modify: `server/app/wiki/handlers/document.go`
- Modify: `server/app/wiki/handlers/document_delete_test.go`
- Create: `server/app/wiki/handlers/document_version_test.go`
- Create: `server/app/wiki/handlers/change_request.go`
- Create: `server/app/wiki/handlers/change_request_test.go`
- Modify: `server/app/wiki/router/document.go`
- Modify: `server/router/router.go`

- [ ] **Step 1: Write failing handler tests**

Require upload and `PUT /documents/:id/content` to call the version service and start `ProcessJob(job.ID)`, not call `PutFileContent` or `IndexDocument(documentID)`. Cover apply success/replay/conflict and cancel. Extend delete tests to require Qdrant deletion, every revision object deletion, and then PostgreSQL deletion in that order.

- [ ] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/wiki/handlers`

Expected: FAIL because handlers still overwrite the current object and change-request routes do not exist.

- [ ] **Step 3: Wire versioned handlers**

Expose:

```text
POST /api/wiki/change-requests/:id/apply
POST /api/wiki/change-requests/:id/cancel
```

Map `ErrVersionConflict` to HTTP 409 and ownership failures to HTTP 403. Keep existing Wiki routes and response envelopes stable.

Return apply results as `{change_request, revision, index_job}` so the practice timeline can update immediately without guessing the applied version.

- [ ] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/wiki/handlers ./app/wiki/router`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/app/wiki/handlers server/app/wiki/router server/router/router.go
git commit -m "refactor(wiki): route saves through document versions"
```

### Task 5: Pin RAG jobs and publish only the current version

**Files:**
- Modify: `server/app/rag/repository/index_job.go`
- Modify: `server/app/rag/repository/chunk.go`
- Create: `server/app/rag/repository/index_job_test.go`
- Modify: `server/app/rag/repository/chunk_test.go`
- Modify: `server/app/rag/service/indexer.go`
- Modify: `server/app/rag/service/indexer_test.go`
- Modify: `server/app/rag/handlers/rag.go`
- Create: `server/app/rag/handlers/rag_test.go`
- Modify: `server/app/rag/router/rag.go`
- Modify: `server/infrastructure/vector/qdrant.go`
- Modify: `server/infrastructure/vector/qdrant_test.go`

- [ ] **Step 1: Write failing version-pinning tests**

Verify jobs retain `document_version` and `object_path`, the indexer reads that path, point payloads include version, point IDs are version-derived, stale jobs become superseded, current jobs publish chunks before old point cleanup, and a failed current-version job is reset rather than duplicated.

- [ ] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/rag/repository ./app/rag/service ./infrastructure/vector`

Expected: FAIL on missing version-aware contracts.

- [ ] **Step 3: Implement version-aware indexing**

Use these repository methods:

```go
CreatePending(ctx, documentID string, version int64, objectPath string) (*IndexJob, error)
FindCurrentVersion(ctx, documentID string) (*IndexJob, error)
RetryCurrentVersion(ctx, documentID string) (*IndexJob, error)
MarkSuperseded(ctx, jobID string) error
PublishVersion(ctx, job *IndexJob, chunks []*WikiChunk) (oldPointIDs []string, superseded bool, err error)
DeletePoints(ctx, collection string, pointIDs []string) error
```

`PublishVersion` locks `wiki_documents`, compares the job version, replaces PostgreSQL chunks and completes the job in one transaction. Delete only the returned old point IDs after commit.

Expose `GET /api/rag/wiki/documents/:id/index-status`. Keep the existing POST index route, but make it process or reset the current-version job and reject completed/superseded versions instead of inserting a duplicate.

- [ ] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/rag/repository ./app/rag/service ./infrastructure/vector`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/app/rag server/infrastructure/vector
git commit -m "feat(rag): pin indexes to wiki document versions"
```

### Task 6: Reject stale retrieval results

**Files:**
- Modify: `server/app/rag/repository/chunk.go`
- Modify: `server/app/rag/repository/chunk_test.go`
- Modify: `server/app/rag/service/retriever.go`
- Modify: `server/app/rag/service/retriever_test.go`
- Modify: `server/app/learning/service/feynman_context.go`
- Modify: `server/app/learning/service/feynman_context_test.go`

- [ ] **Step 1: Write failing stale-result tests**

Make `FindByPointIDs` join `wiki_documents` and require `rwc.document_version = d.current_version`. Verify Retriever over-fetches candidates, preserves score order after filtering, and returns typed `ErrIndexNotReady` for a long current document with no current chunks.

- [ ] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/rag/repository ./app/rag/service ./app/learning/service`

Expected: FAIL because stale chunks are currently accepted.

- [ ] **Step 3: Implement authoritative current-version filtering**

Request `min(limit*4, 48)` Qdrant candidates, filter through PostgreSQL, and truncate valid results to the requested limit. Add `DocumentVersion` to `SearchResult` and Feynman evidence.

- [ ] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/rag/repository ./app/rag/service ./app/learning/service`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/app/rag server/app/learning/service
git commit -m "fix(rag): exclude stale wiki evidence"
```

### Task 7: Verify Plan 1

**Files:** all Plan 1 files

- [ ] **Step 1: Format and run complete backend verification**

```bash
cd server
gofmt -w app/wiki app/rag app/learning/service infrastructure/database infrastructure/storage infrastructure/vector router
go test -count=1 ./...
```

- [ ] **Step 2: Check migration and diff quality**

Run: `git diff --check && rg -n 'COMMENT ON' server/migrations/wiki/003_document_versions.sql server/migrations/rag/002_document_versions.sql`

Expected: no diff errors and concise table/field metadata comments are present.

- [ ] **Step 3: Confirm scope**

Confirm no Teacher/Curator prompt, unified Turn endpoint, or frontend selector was added in Plan 1.
