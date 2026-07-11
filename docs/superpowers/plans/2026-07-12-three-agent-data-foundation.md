# Three-Agent Data Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the Learning-side persistence foundation for explicit Listener, Teacher, and Curator turns while keeping the existing Feynman review API working.

**Architecture:** A `learning_turns` row is the idempotent owner of one user request and one selected agent. The existing review endpoint becomes the first Listener producer: it persists the processing turn and user message before model execution, then atomically persists the assistant message, structured review, and completed status; Teacher receives its structured intervention model now, while Wiki versioning and Curator execution remain a later phase.

**Tech Stack:** Go 1.24, Bun, PostgreSQL, Fiber, React/TypeScript, TanStack Query

---

### Task 1: Add the Learning turn schema and models

**Files:**
- Create: `server/migrations/learning/004_three_agent_data_foundation.sql`
- Create: `server/app/learning/models/db/turn.go`
- Create: `server/app/learning/models/db/intervention.go`
- Modify: `server/app/learning/models/db/session.go`
- Modify: `server/app/learning/models/db/message.go`
- Modify: `server/app/learning/models/db/review.go`
- Modify: `server/infrastructure/database/database.go`

- [ ] **Step 1: Write model metadata tests that require the new tables and fields**

Add `server/app/learning/models/db/turn_test.go` assertions for Bun table names and the `turn_id`, `request_id`, `agent_type`, `status`, and teaching-intervention fields.

- [ ] **Step 2: Run the model test and verify RED**

Run: `cd server && go test ./app/learning/models/db`

Expected: FAIL because `LearningTurn` and `LearningTeachingIntervention` do not exist.

- [ ] **Step 3: Add models and migration**

Create `LearningTurn` with `listener`, `teacher`, and `curator` agent constants plus `processing`, `completed`, and `failed` status constants. Create `LearningTeachingIntervention` with JSONB slices/maps. Give every business field a concise Chinese trailing comment.

The migration must:

```sql
CREATE TABLE learning_turns (...);
ALTER TABLE learning_messages ADD COLUMN turn_id VARCHAR(32);
ALTER TABLE learning_explanation_reviews ADD COLUMN turn_id VARCHAR(32);
CREATE TABLE learning_teaching_interventions (...);
CREATE UNIQUE INDEX uq_learning_memory_events_source
  ON learning_memory_events(source_type, source_id, event_type)
  WHERE source_id IS NOT NULL;
```

Backfill legacy messages and reviews into deterministic legacy turns before making `turn_id` non-null. Then drop `learning_sessions.message_count`, `learning_messages.agent_type`, and the duplicated persisted review columns. Keep derived compatibility fields on `LearningExplanationReview` as `scanonly` fields.

- [ ] **Step 4: Register the new Bun models**

Register `LearningTurn` and `LearningTeachingIntervention` in `DatabaseService`.

- [ ] **Step 5: Run model tests and verify GREEN**

Run: `cd server && go test ./app/learning/models/db`

Expected: PASS.

### Task 2: Persist idempotent Listener turn lifecycle

**Files:**
- Create: `server/app/learning/repository/turn.go`
- Create: `server/app/learning/repository/turn_test.go`
- Modify: `server/app/learning/repository/review.go`
- Modify: `server/app/learning/repository/review_test.go`
- Modify: `server/app/learning/repository/memory.go`
- Modify: `server/infrastructure/database/database.go`

- [ ] **Step 1: Write failing repository tests**

Cover these behaviors:

```text
BeginListenerTurn inserts one processing listener turn and one user message in a transaction.
The same (session_id, request_id) returns the existing turn instead of inserting duplicates.
CompleteListenerTurn inserts the assistant message and review, then marks the turn completed.
FailTurn records failed status and error details.
FindBySession hydrates review compatibility fields through turn/session/user-message joins.
CreateEvent is idempotent for (source_type, source_id, event_type).
```

- [ ] **Step 2: Run repository tests and verify RED**

Run: `cd server && go test ./app/learning/repository`

Expected: FAIL because the turn repository and revised review queries do not exist.

- [ ] **Step 3: Implement transactional persistence**

Use `bun.DB.RunInTx`. `BeginListenerTurn` assigns 32-character UUIDs, inserts the turn with `ON CONFLICT DO NOTHING`, and only inserts the user message when the turn was newly created. `CompleteListenerTurn` rejects non-processing turns, stores the JSON-encoded structured reply as the assistant message, inserts the one-to-one review, and sets `completed_at`. `FailTurn` stores a stable error code and bounded error message.

- [ ] **Step 4: Implement derived review reads and memory idempotency**

Join `learning_explanation_reviews -> learning_turns -> learning_sessions` and the turn's user message so API callers still receive `session_id`, `document_id`, `user_id`, and `explanation`. Make memory event creation return the existing event ID on uniqueness conflict so memory items never reference a discarded UUID.

- [ ] **Step 5: Run repository tests and verify GREEN**

Run: `cd server && go test ./app/learning/repository`

Expected: PASS.

### Task 3: Route the existing Feynman Listener endpoint through turns

**Files:**
- Modify: `server/app/learning/models/payload/session.go`
- Modify: `server/app/learning/handlers/session.go`
- Modify: `server/app/learning/handlers/session_test.go`
- Modify: `server/app/learning/service/memory.go`
- Modify: `server/app/learning/service/memory_test.go`
- Modify: `web/src/api/learning/session.ts`
- Modify: `web/src/pages/learning/feynman-practice/index.tsx`

- [ ] **Step 1: Write failing handler tests**

Require `request_id`, verify a turn is begun before reviewer invocation, verify success completes the same turn, verify reviewer failure marks it failed, verify completed duplicates return the existing review without invoking the model, and verify processing duplicates return HTTP 409.

- [ ] **Step 2: Run handler tests and verify RED**

Run: `cd server && go test ./app/learning/handlers`

Expected: FAIL because `SessionHandler` has no turn lifecycle dependency.

- [ ] **Step 3: Implement the handler lifecycle**

Add a focused `listenerTurnStore` dependency. The handler validates `request_id`, starts or resumes the idempotent turn, calls the reviewer only for a new/failed turn, completes persistence before recording best-effort memory, and marks the turn failed when model execution or completion persistence fails.

- [ ] **Step 4: Align memory provenance**

Use `source_type = explanation_review`, `source_id = review.id`, and preserve `event_type = explanation_review`. Keep mastery evidence restricted to Listener reviews.

- [ ] **Step 5: Add a stable client request ID**

Extend `ReviewExplanationRequest` with `request_id`. Create the UUID once per user submission in `submitExplanation` and pass it through `mutateAsync`, so network retries reuse the same identity for that submission.

- [ ] **Step 6: Run focused backend and frontend checks**

Run:

```bash
cd server && go test ./app/learning/handlers ./app/learning/service ./app/learning/repository ./app/learning/models/db
pnpm --dir web lint:type
```

Expected: both commands PASS.

### Task 4: Verify the first phase as a whole

**Files:**
- Verify all modified files

- [ ] **Step 1: Format changed Go and frontend files**

Run:

```bash
cd server && gofmt -w app/learning infrastructure/database
pnpm --dir web exec oxfmt --write src/api/learning/session.ts src/pages/learning/feynman-practice/index.tsx
```

- [ ] **Step 2: Run the complete backend suite**

Run: `cd server && go test ./...`

Expected: PASS.

- [ ] **Step 3: Run frontend type checking and diff checks**

Run:

```bash
pnpm --dir web lint:type
git diff --check
git status --short
```

Expected: type checking and diff check PASS; status contains only this phase's plan and implementation files.

- [ ] **Step 4: Review phase boundary**

Confirm this phase does not add Wiki change requests, document revisions, RAG version columns, Curator write execution, agent routing, page layout changes, or agent renames. Those remain in the next implementation plan.
