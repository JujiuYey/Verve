# Whole-Document Feynman Practice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace objective-based pass/fail exercises with document-scoped, multi-turn Feynman explanation practice backed by adaptive full-document/RAG context.

**Architecture:** A learning session belongs directly to one Wiki document. Short documents are supplied in full; long documents use the full heading outline plus document-filtered RAG hits and adjacent chunks. A structured Feynman reviewer returns listener-oriented feedback without pass/fail or mastery levels, and learning memory records explanation evidence rather than objective completion.

**Tech Stack:** Go 1.x, Fiber, Bun/PostgreSQL, Eino ADK, MinIO, Qdrant, React 19, TanStack Query/Router, shadcn/ui.

---

## File Structure

### Backend additions

- `server/app/rag/service/retriever.go`: add document-scoped vector retrieval.
- `server/app/rag/repository/chunk.go`: load neighboring chunks in document order.
- `server/app/learning/service/feynman_context.go`: choose full-text or document-RAG context.
- `server/app/learning/service/feynman_review.go`: render and parse structured listener feedback.
- `server/infrastructure/llm/prompts/feynman_reviewer.go`: own reviewer instructions and typed runtime prompt input.
- `server/app/learning/models/db/review.go`: persist one explanation/review turn.
- `server/app/learning/repository/review.go`: store and list review turns by session.

### Backend removals

- Objective model/repository/handler/service/prompt/tests.
- Examiner service/prompt/tests and Tutor-only learning tools.
- Objective-specific exercise model/repository/handler/API behavior.

### Frontend changes

- Keep route identity `/_layout/learn/feynman-practice/$documentId` and remove `objectiveId` search state.
- Replace objective outline, mastery badges, verdict panels, and teaching fallback with document reading plus explanation/review turns.
- Navigate directly from Wiki and Coach actions using `document_id`.

## Task 1: Add Document-Scoped RAG With Neighbor Expansion

**Files:**
- Modify: `server/app/rag/models/payload/retrieval.go`
- Modify: `server/app/rag/repository/chunk.go`
- Modify: `server/app/rag/service/retriever.go`
- Test: `server/app/rag/service/retriever_test.go`
- Test: `server/app/rag/repository/chunk_test.go`

- [ ] **Step 1: Write failing retriever tests**

Add tests proving that `SearchDocument` rejects a blank document ID and sends this Qdrant filter:

```go
map[string]any{
    "must": []map[string]any{
        {"key": "document_id", "match": map[string]any{"value": "doc-1"}},
    },
}
```

Also assert that returned `SearchResult` values include `ChunkIndex`.

- [ ] **Step 2: Verify the tests fail**

Run: `cd server && go test ./app/rag/service -run 'TestRetrieverSearchDocument' -count=1`

Expected: FAIL because `SearchDocument` and `ChunkIndex` do not exist.

- [ ] **Step 3: Implement document-scoped retrieval**

Add this public API while reusing the existing embedding and vector-search path:

```go
func (r *Retriever) SearchDocument(
    ctx context.Context,
    documentID string,
    query string,
    limit int,
) ([]rag_payload.SearchResult, error)
```

Refactor the existing root search into a private method accepting a Qdrant filter. Map `WikiChunk.ChunkIndex` into `SearchResult.ChunkIndex`.

- [ ] **Step 4: Add adjacent-chunk repository behavior test-first**

Add and test:

```go
func (r *ChunkRepository) FindNeighbors(
    ctx context.Context,
    documentID string,
    indexes []int,
    radius int,
) ([]*rag_db.WikiChunk, error)
```

It must return unique chunks ordered by `chunk_index ASC`, restricted to the same document, including each hit index plus `radius=1` neighbors.

- [ ] **Step 5: Run focused tests**

Run: `cd server && go test ./app/rag/... -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add server/app/rag server/infrastructure/vector
git commit -m "feat(rag): add document-scoped context retrieval"
```

## Task 2: Build Adaptive Feynman Context and Reviewer

**Files:**
- Create: `server/app/learning/service/feynman_context.go`
- Create: `server/app/learning/service/feynman_context_test.go`
- Create: `server/app/learning/service/feynman_review.go`
- Create: `server/app/learning/service/feynman_review_test.go`
- Create: `server/infrastructure/llm/prompts/feynman_reviewer.go`
- Modify: `server/infrastructure/llm/prompts/prompts_test.go`
- Modify: `server/infrastructure/llm/agent.go`

- [ ] **Step 1: Write failing context-policy tests**

Define the desired API through tests:

```go
type FeynmanDocumentContext struct {
    DocumentID string
    Title      string
    Outline    []string
    Evidence   []FeynmanEvidenceChunk
    FullText   string
    Mode       string // full | rag
}

func BuildFeynmanDocumentContext(
    ctx context.Context,
    source FeynmanDocumentSource,
    documentID string,
    explanation string,
) (FeynmanDocumentContext, error)
```

Tests must prove:

- a document at or below `FullDocumentCharacterLimit` uses `Mode="full"` and does not call retrieval;
- a larger document uses `Mode="rag"`, retains every Markdown heading in `Outline`, and includes document-scoped hits plus neighbors;
- empty Markdown returns a descriptive error;
- insufficient RAG evidence is represented explicitly rather than silently truncating.

- [ ] **Step 2: Verify context tests fail**

Run: `cd server && go test ./app/learning/service -run 'TestBuildFeynmanDocumentContext' -count=1`

Expected: FAIL because the context builder does not exist.

- [ ] **Step 3: Implement the minimal context builder**

Use a small dependency interface so tests use real context-selection logic without MinIO or Qdrant:

```go
type FeynmanDocumentSource interface {
    LoadDocument(ctx context.Context, documentID string) (*wiki_db.Document, string, error)
    SearchDocument(ctx context.Context, documentID, query string, limit int) ([]rag_payload.SearchResult, error)
    FindNeighbors(ctx context.Context, documentID string, indexes []int, radius int) ([]*rag_db.WikiChunk, error)
}
```

Keep the threshold as one named constant. Build the outline from Markdown headings and preserve heading paths in evidence.

- [ ] **Step 4: Write failing reviewer-output tests**

Define the structured response:

```go
type FeynmanReview struct {
    HeardSummary       string   `json:"heard_summary"`
    ClearPoints        []string `json:"clear_points"`
    ConfusingPoints    []string `json:"confusing_points"`
    Misconceptions     []string `json:"misconceptions"`
    FollowUpQuestion   string   `json:"follow_up_question"`
    ExplanationSummary string  `json:"explanation_summary"`
    ReadyToWrapUp      bool     `json:"ready_to_wrap_up"`
    ContextSufficient  bool     `json:"context_sufficient"`
}
```

Tests must parse fenced JSON, normalize nil arrays, reject an empty review, and retain `context_sufficient=false`.

- [ ] **Step 5: Implement prompt and reviewer service**

Add `NewFeynmanReviewerAgent(ctx)` using `NewStructuredChatModel`. The system prompt must require listener-oriented feedback, exactly one natural follow-up question when clarification is needed, no pass/fail language, and no mandatory code unless a runtime claim cannot be checked conceptually.

The runtime query must include document title, outline, full text or RAG evidence, prior review turns, and the new explanation.

- [ ] **Step 6: Run focused tests**

Run: `cd server && go test ./app/learning/service ./infrastructure/llm/prompts -count=1`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add server/app/learning/service/feynman_* server/infrastructure/llm
git commit -m "feat(learning): add Feynman listener review engine"
```

## Task 3: Rebuild the Learning Schema Around Documents

**Files:**
- Rewrite: `server/migrations/learning/001_schema.sql`
- Rewrite: `server/migrations/learning/003_memory.sql`
- Delete: `server/migrations/learning/002_guides.sql`
- Delete: `server/migrations/learning/004_remove_guides.sql`
- Modify: `server/app/learning/models/db/session.go`
- Create: `server/app/learning/models/db/review.go`
- Modify: `server/app/learning/models/db/memory.go`
- Create: `server/app/learning/repository/review.go`
- Create: `server/app/learning/repository/review_test.go`
- Modify: `server/infrastructure/database/database.go`

- [ ] **Step 1: Write failing repository tests**

Using Bun's PostgreSQL query formatter or the repository's established test pattern, verify that review creation persists `session_id`, `document_id`, `user_id`, explanation text, JSON arrays, summaries, and booleans, and that listing is ordered by `created_at ASC`.

- [ ] **Step 2: Verify repository tests fail**

Run: `cd server && go test ./app/learning/repository -run 'TestReviewRepository' -count=1`

Expected: FAIL because the model and repository do not exist.

- [ ] **Step 3: Define the clean development schema**

The base learning schema must contain:

```sql
CREATE TABLE learning_sessions (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    summary TEXT,
    message_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE learning_explanation_reviews (
    id VARCHAR(32) PRIMARY KEY,
    session_id VARCHAR(32) NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,
    document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    explanation TEXT NOT NULL,
    heard_summary TEXT NOT NULL,
    clear_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    confusing_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    misconceptions JSONB NOT NULL DEFAULT '[]'::jsonb,
    follow_up_question TEXT NOT NULL DEFAULT '',
    explanation_summary TEXT NOT NULL,
    ready_to_wrap_up BOOLEAN NOT NULL DEFAULT FALSE,
    context_sufficient BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

The clean SQL schema omits objective, exercise, profile, mastery, and verdict structures. Keep messages, journals, and learning memory. Remove `objective_id` from memory events/items and retain `document_id`, `folder_id`, and `session_id` evidence links. Keep the old Go models and repositories temporarily so the repository remains compilable until Task 5 removes all callers.

- [ ] **Step 4: Implement models and repositories**

Add `DocumentID` to `LearningSession` for the new flow while temporarily retaining `ObjectiveID` in Go until Task 5 removes old callers. Add `LearningExplanationReview` matching the schema and register `ReviewRepository` on `DatabaseService`. Do not remove old Go repositories in this task.

- [ ] **Step 5: Run model and repository tests**

Run: `cd server && go test ./app/learning/models/db ./app/learning/repository ./infrastructure/database -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add server/migrations/learning server/app/learning/models server/app/learning/repository server/infrastructure/database
git commit -m "refactor(learning): rebuild sessions around documents"
```

## Task 4: Cut Session APIs Over to Multi-Turn Reviews

**Files:**
- Modify: `server/app/learning/models/payload/session.go`
- Modify: `server/app/learning/handlers/session.go`
- Create: `server/app/learning/handlers/session_test.go`
- Modify: `server/app/learning/router/learning.go`
- Rewrite: `server/app/learning/service/memory.go`
- Rewrite: `server/app/learning/service/memory_test.go`
- Delete: `server/app/learning/service/examiner.go`
- Delete: `server/app/learning/service/examiner_test.go`
- Delete: `server/infrastructure/llm/prompts/examiner.go`
- Delete: `server/infrastructure/llm/prompts/tutor.go`
- Delete: `server/app/learning/tools/learning_tools.go`
- Modify: `server/infrastructure/llm/agent.go`
- Modify: `server/infrastructure/llm/prompts/prompts_test.go`

- [ ] **Step 1: Write failing handler tests**

Tests must establish these contracts:

```json
POST /api/learning/session
{"document_id":"doc-1"}

POST /api/learning/session/:id/review
{"explanation":"A value has a concrete type, and the type determines valid operations."}
```

The create endpoint rejects a missing document, creates a document-bound session, and returns `session_id`. The review endpoint rejects another user's session, builds context from the current document and prior turns, stores the new review, records learning memory best-effort, and returns `FeynmanReview` without verdict/mastery fields.

- [ ] **Step 2: Verify handler tests fail**

Run: `cd server && go test ./app/learning/handlers -run 'TestSession(Create|Review)' -count=1`

Expected: FAIL because the document payload and review endpoint do not exist.

- [ ] **Step 3: Implement dependency-injected session handling**

Construct `SessionHandler` with database, MinIO content loading, document-scoped retriever, and chunk repository dependencies. Add a private source adapter implementing `FeynmanDocumentSource` through `Documents.FindOne`, `MinIOService.GetFileContent`, `Retriever.SearchDocument`, and `RAGChunks.FindNeighbors`.

Persist every user explanation and structured review. `FindOne` returns session, messages, and ordered reviews. `Complete` summarizes the accumulated explanation and marks the session completed; it must not select a next objective.

- [ ] **Step 4: Rewrite learning-memory recording test-first**

Replace `RecordExerciseJudgement` with:

```go
func (s *MemoryService) RecordExplanationReview(
    ctx context.Context,
    userID string,
    session *learning_db.LearningSession,
    review *learning_db.LearningExplanationReview,
) error
```

Record clear relationships as `explanation_evidence`, explicit misconceptions as `misconception`, and clarified follow-ups as evidence events. Never create `mastered_concept` from a single response.

- [ ] **Step 5: Remove old Tutor/Examiner runtime paths**

Delete the exercise endpoint, tutor chat endpoint, examiner service/prompt, tutor prompt, and objective-based learning tools. Keep generic SSE utilities only where Coach still uses them.

- [ ] **Step 6: Run backend learning tests**

Run: `cd server && go test ./app/learning/... ./infrastructure/llm/... -count=1`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add server/app/learning server/infrastructure/llm
git commit -m "feat(learning): make Feynman review document-native"
```

## Task 5: Simplify Coach and Remove the Objective Subsystem

**Files:**
- Modify: `server/app/learning/handlers/coach.go`
- Modify: `server/app/learning/service/coach.go`
- Modify: `server/app/learning/service/coach_test.go`
- Modify: `server/app/learning/tools/coach_tools.go`
- Modify: `server/infrastructure/llm/prompts/coach.go`
- Modify: `server/infrastructure/llm/prompts/prompts_test.go`
- Delete: `server/app/learning/handlers/objective.go`
- Delete: `server/app/learning/models/db/objective.go`
- Delete: `server/app/learning/repository/objective.go`
- Delete: `server/app/learning/service/objective_generation.go`
- Delete: `server/app/learning/service/objective_generation_test.go`
- Delete: `server/infrastructure/llm/prompts/objective_generator.go`
- Delete: `server/app/learning/handlers/exercise.go`
- Delete: `server/app/learning/models/db/exercise.go`
- Delete: `server/app/learning/repository/exercise.go`
- Delete: `server/app/learning/models/db/profile.go`
- Delete: `server/app/learning/repository/profile.go`
- Modify: `server/infrastructure/database/database.go`
- Modify: `web/src/pages/system/agent-config/agent-definitions.tsx`

- [ ] **Step 1: Write failing Coach contract tests**

Update expected actions to:

```go
CoachAction{
    Type:       "navigate_to_practice",
    DocumentID: "doc-1",
    Label:      "开始费曼练习",
}
```

Tests must prove the prompt lists real documents and memory but never mentions objectives, mastery, `create_learning_objectives`, or `first_objective_id`.

- [ ] **Step 2: Verify Coach tests fail**

Run: `cd server && go test ./app/learning/service ./infrastructure/llm/prompts -run 'Coach' -count=1`

Expected: FAIL because actions and prompts still use `objective_id`.

- [ ] **Step 3: Implement document navigation**

Remove objectives from `CoachRuntimeContext`, prompt inputs, and tool lists. Remove `list_objectives`, `create_learning_objectives`, and objective session tools. Coach may inspect folders, documents, RAG, memory, and journals, then emit:

```text
<ACTION>{"type":"navigate_to_practice","document_id":"doc-1","label":"开始费曼练习"}</ACTION>
```

- [ ] **Step 4: Delete the objective/exercise subsystem**

Remove all remaining ObjectiveGenerator, objective API/router, exercise API/router, legacy profile repositories/models, tests, and database registrations. Remove ObjectiveGenerator, Tutor, and Examiner cards from agent configuration; retain Coach and RAG-related agents.

- [ ] **Step 5: Run a repository-wide residual scan**

Run:

```bash
rg -n "LearningObjective|objective_id|objectiveId|ObjectiveGenerator|mastery_level|pass\|partial\|fail" server web/src
```

Expected: no active learning-flow matches. Historical design documents may still describe the removed system.

- [ ] **Step 6: Run backend tests and frontend type-check**

Run: `cd server && go test ./... -count=1`

Run: `cd web && pnpm lint:type`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add server web/src/pages/system/agent-config/agent-definitions.tsx
git commit -m "refactor(learning): remove objective-based practice"
```

## Task 6: Rebuild the Feynman Workbench Around Explanation Turns

**Files:**
- Modify: `web/src/api/learning/session.ts`
- Delete: `web/src/api/learning/objective.ts`
- Delete: `web/src/api/learning/exercise.ts`
- Delete: `web/src/api/learning/profile.ts`
- Modify: `web/src/api/learning/index.ts`
- Modify: `web/src/routes/_layout/learn/feynman-practice/$documentId.tsx`
- Modify: `web/src/pages/wiki/index/index.tsx`
- Modify: `web/src/pages/learning/feynman/index.tsx`
- Modify: `web/src/pages/learning/feynman/_components/coach-workspace.tsx`
- Rewrite: `web/src/pages/learning/feynman-practice/index.tsx`
- Rewrite: `web/src/pages/learning/feynman-practice/_components/practice-panel.tsx`
- Modify: `web/src/pages/learning/feynman-practice/_components/source-panel.tsx`
- Delete: `web/src/pages/learning/feynman-practice/_components/objective-outline.tsx`
- Delete: `web/src/pages/learning/feynman-practice/_components/study-info-panel.tsx`
- Delete: `web/src/pages/learning/feynman-practice/_components/teaching-panel.tsx`
- Modify: `web/src/pages/learning/feynman-practice/_shared.ts`

- [ ] **Step 1: Update TypeScript API contracts**

Define:

```ts
export interface CreateSessionRequest {
  document_id: string;
}

export interface FeynmanReview {
  heard_summary: string;
  clear_points: string[];
  confusing_points: string[];
  misconceptions: string[];
  follow_up_question: string;
  explanation_summary: string;
  ready_to_wrap_up: boolean;
  context_sufficient: boolean;
}
```

Add `useReviewExplanation(sessionId)` and remove exercise/objective contracts.

- [ ] **Step 2: Make every entry document-native**

Wiki document clicks navigate directly to `$documentId` without generating objectives. Coach actions carry `document_id` and navigate directly. Remove `objectiveId` route search validation.

- [ ] **Step 3: Rebuild the workbench**

Keep the existing Markdown reader and catalog. The explanation view must:

- invite the learner to explain the article to a complete beginner;
- allow one long explanation or multiple follow-up turns;
- show ordered prior explanation/review turns;
- render heard summary, clear points, confusion, misconceptions, and one follow-up question;
- offer `继续补充` and `结束本次练习` commands;
- avoid pass/fail badges, mastery labels, required code, or objective navigation.

Use installed shadcn/ui and existing `ai-elements` primitives. Do not add frontend tests for this route-level UI change under the project test policy.

- [ ] **Step 4: Remove obsolete components and state**

Delete objective outline, study-info/mastery panel, teaching fallback panel, `buildPrompt`, objective switching, verdict labels, and phase logic that represents passing/failing.

- [ ] **Step 5: Run frontend checks**

Run:

```bash
cd web
pnpm lint:type
pnpm exec oxfmt --check \
  src/api/learning \
  src/pages/learning/feynman \
  src/pages/learning/feynman-practice \
  src/pages/wiki/index/index.tsx \
  'src/routes/_layout/learn/feynman-practice/$documentId.tsx'
```

Expected: PASS.

- [ ] **Step 6: Run final repository verification**

Run: `cd server && go test ./... -count=1`

Run: `git diff --check`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add web server docs/superpowers
git commit -m "feat(web): deliver whole-document Feynman practice"
```
