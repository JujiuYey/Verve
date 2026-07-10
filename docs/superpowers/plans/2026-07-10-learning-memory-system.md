# Learning Memory System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current "profile as a mutable state table" model with a learning memory system that records learning facts, extracts stable memory items, and serves a derived memory view to Coach and the learner.

**Architecture:** Keep existing learning flows working while introducing a new memory layer beside `learning_profiles`. Exercises, journals, messages, and Wiki content remain source facts; new memory tables store extracted facts and summary projections. `learning_profiles` becomes legacy/cache-like context and stops being the primary product concept.

**Tech Stack:** Go, Fiber, Bun ORM, PostgreSQL JSONB, React, TanStack Query, existing Verve learning repositories and prompt packages.

---

## Product Decision

The product concept is not "user profile is a row." The product concept is:

> Learning memory is an evidence-backed view of what Verve remembers from the user's learning history.

This means:

- Source of truth: learning messages, exercises, journals, objectives, Wiki documents, user notes, and later code/output artifacts.
- Memory event: a normalized fact that says something happened or was observed.
- Memory item: a stable, deduplicated learning fact extracted from events.
- Memory summary: a compact projection used by Coach and UI.
- Profile page: a reader-facing "learning memory" page, not an editable profile table.

## File Structure

- Create `server/migrations/learning/003_memory.sql`
  - Adds `learning_memory_events`, `learning_memory_items`, and `learning_memory_summaries`.
- Create `server/app/learning/models/db/memory.go`
  - Bun models for memory events, items, and summaries.
- Create `server/app/learning/repository/memory.go`
  - Repository methods for appending events, listing items, and upserting summaries.
- Modify `server/infrastructure/database/database.go`
  - Register memory models and expose `Memories`.
- Create `server/app/learning/service/memory.go`
  - Service that records exercise events and derives memory items from examiner output.
- Modify `server/app/learning/handlers/session.go`
  - On exercise submit, write memory events/items instead of treating profile as the primary write target.
- Modify `server/app/learning/service/coach.go`
  - Add memory context mapping for Coach.
- Modify `server/infrastructure/llm/prompts/coach.go`
  - Render `## 学习记忆` as the main learner context. Keep legacy profile context behind it temporarily.
- Modify `server/app/learning/tools/coach_tools.go`
  - Replace or supplement `get_learning_profile` with `search_learning_memory`.
- Modify `web/src/pages/learning/profile/index.tsx`
  - Rename page copy from "我的画像" to "学习记忆".
- Create `web/src/api/learning/memory.ts`
  - API hook for memory summary/items once the backend endpoint is added.

## Phase 1: Add Memory Storage Without Changing User Flow

### Task 1: Add Memory Schema

**Files:**
- Create: `server/migrations/learning/003_memory.sql`

- [ ] Add tables:

```sql
CREATE TABLE learning_memory_events (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    document_id VARCHAR(32) REFERENCES wiki_documents(id) ON DELETE SET NULL,
    objective_id VARCHAR(32) REFERENCES learning_objectives(id) ON DELETE SET NULL,
    session_id VARCHAR(32) REFERENCES learning_sessions(id) ON DELETE SET NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(32),
    event_type VARCHAR(40) NOT NULL,
    content TEXT NOT NULL,
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_learning_memory_events_user_time ON learning_memory_events(user_id, occurred_at DESC);
CREATE INDEX idx_learning_memory_events_folder_time ON learning_memory_events(folder_id, occurred_at DESC);
CREATE INDEX idx_learning_memory_events_objective ON learning_memory_events(objective_id);

CREATE TABLE learning_memory_items (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    document_id VARCHAR(32) REFERENCES wiki_documents(id) ON DELETE SET NULL,
    objective_id VARCHAR(32) REFERENCES learning_objectives(id) ON DELETE SET NULL,
    kind VARCHAR(40) NOT NULL,
    statement TEXT NOT NULL,
    evidence_event_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    confidence VARCHAR(20) NOT NULL DEFAULT 'observed',
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_learning_memory_items_user_kind ON learning_memory_items(user_id, kind);
CREATE INDEX idx_learning_memory_items_folder ON learning_memory_items(folder_id);
CREATE INDEX idx_learning_memory_items_objective ON learning_memory_items(objective_id);

CREATE TABLE learning_memory_summaries (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    generated_from_event_id VARCHAR(32),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_learning_memory_summary_folder UNIQUE (user_id, folder_id)
);
```

- [ ] Run migration through the existing project migration path.
- [ ] Verify tables exist in the local database.

### Task 2: Add Go Models And Repository

**Files:**
- Create: `server/app/learning/models/db/memory.go`
- Create: `server/app/learning/repository/memory.go`
- Modify: `server/infrastructure/database/database.go`

- [ ] Add model structs:

```go
type LearningMemoryEvent struct {
    bun.BaseModel `bun:"table:learning_memory_events,alias:lme"`
    ID string `bun:"id,pk,type:varchar(32)" json:"id"`
    UserID string `bun:"user_id,notnull" json:"user_id"`
    FolderID *string `bun:"folder_id" json:"folder_id,omitempty"`
    DocumentID *string `bun:"document_id" json:"document_id,omitempty"`
    ObjectiveID *string `bun:"objective_id" json:"objective_id,omitempty"`
    SessionID *string `bun:"session_id" json:"session_id,omitempty"`
    SourceType string `bun:"source_type,notnull" json:"source_type"`
    SourceID *string `bun:"source_id" json:"source_id,omitempty"`
    EventType string `bun:"event_type,notnull" json:"event_type"`
    Content string `bun:"content,notnull" json:"content"`
    Evidence map[string]interface{} `bun:"evidence,type:jsonb" json:"evidence"`
    OccurredAt time.Time `bun:"occurred_at,nullzero,notnull,default:current_timestamp" json:"occurred_at"`
    CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}
```

- [ ] Add repository methods:
  - `CreateEvent(ctx, event)`
  - `CreateItem(ctx, item)`
  - `FindItemsByUser(ctx, userID, folderID string, limit int)`
  - `FindSummaryByFolder(ctx, userID, folderID string)`
  - `UpsertSummary(ctx, summary)`

- [ ] Register models in `database.go`.
- [ ] Add `Memories *learning_repo.MemoryRepository` to `DatabaseService`.
- [ ] Test with repository unit tests using existing Bun test style if present; otherwise cover through service tests in Task 3.

## Phase 2: Record Memory From Exercises

### Task 3: Build Memory Service

**Files:**
- Create: `server/app/learning/service/memory.go`
- Test: `server/app/learning/service/memory_test.go`

- [ ] Define a service API:

```go
type MemoryService struct {
    db *database.DatabaseService
}

func NewMemoryService(db *database.DatabaseService) *MemoryService {
    return &MemoryService{db: db}
}

func (s *MemoryService) RecordExerciseJudgement(ctx context.Context, userID string, obj *learning_db.LearningObjective, sessionID string, result *JudgeResult) error
```

- [ ] On every exercise judgement, append a `learning_memory_events` row with:
  - `source_type = "exercise"`
  - `event_type = "examiner_judgement"`
  - `content = result.Feedback`
  - evidence containing verdict, mastery, weak points, evidence text, and improvement suggestion.

- [ ] Derive memory items conservatively:
  - If `verdict == "pass"`, create `kind = "mastered_concept"` with statement `用户已经能解释：<objective title>`.
  - If `result.Evidence` is non-empty, create `kind = "verification_evidence"` with the evidence text.
  - Do not create long-term "weakness" items in Phase 2; weak points remain event evidence until repeated patterns are implemented.

- [ ] Test that `RecordExerciseJudgement` writes one event and the expected stable items.
- [ ] Test that weak points are stored as evidence but not promoted to permanent memory items.

### Task 4: Stop Treating Profile As The Primary Write Target

**Files:**
- Modify: `server/app/learning/handlers/session.go`
- Modify: `server/app/learning/service/examiner.go`
- Test: `server/app/learning/handlers/session_test.go` or existing handler tests.

- [ ] In `Exercise`, call `MemoryService.RecordExerciseJudgement` after the exercise row is created.
- [ ] Keep `learning_profiles` update temporarily for compatibility, but mark it as legacy in code comments.
- [ ] Remove any new writes that put improvement suggestions into `profile.NextGoal`.
- [ ] Update UI helper text from "写入学习画像" to "写入学习记忆".

## Phase 3: Feed Coach From Memory

### Task 5: Add Memory Context To Coach

**Files:**
- Modify: `server/app/learning/service/coach.go`
- Modify: `server/app/learning/handlers/coach.go`
- Modify: `server/infrastructure/llm/prompts/coach.go`
- Test: `server/infrastructure/llm/prompts/prompts_test.go`

- [ ] Add `MemoryItems []*learning_db.LearningMemoryItem` and `MemorySummary *learning_db.LearningMemorySummary` to `CoachRuntimeContext`.
- [ ] In `buildRuntimeContext`, load memory items for the selected root folder or recent folders.
- [ ] Render memory before profile:

```text
## 学习记忆
- 已掌握: 用户已经能解释：值与类型的关系
- 验证证据: 能用 %T 和 %v 说明类型和值
```

- [ ] Keep `## 学习画像` temporarily as legacy context, but prompt should prefer learning memory when both exist.
- [ ] Update Coach instruction:
  - "优先参考学习记忆和最近练习记录"
  - "不要把一次练习缺口直接当成长期画像"

### Task 6: Add Coach Memory Tool

**Files:**
- Modify: `server/app/learning/tools/coach_tools.go`

- [ ] Add `search_learning_memory` tool:
  - input: `folder_id`, `query`, `limit`
  - output: matching memory items and summaries.
- [ ] Keep `get_learning_profile` one release cycle for compatibility.
- [ ] Update tool descriptions so Coach understands memory is the preferred learner context.

## Phase 4: Rename The Product Surface

### Task 7: Rename Profile Page To Learning Memory

**Files:**
- Modify: `web/src/pages/learning/profile/index.tsx`
- Modify: `web/src/layout/sidebar/menu.ts`
- Create: `web/src/api/learning/memory.ts`

- [ ] Sidebar title changes from `我的画像` to `学习记忆`.
- [ ] Page heading changes to `学习记忆`.
- [ ] Page copy explains: Verve remembers evidence-backed learning facts from answers, exercises, documents, and notes.
- [ ] Do not add frontend tests for this page copy/layout change, following project frontend test policy.
- [ ] Run `pnpm --dir /Users/jujiuyey/Projects/Verve/web lint:type`.

### Task 8: Add Memory Read API

**Files:**
- Create or modify: `server/app/learning/handlers/memory.go`
- Modify: learning routes file where handlers are registered.
- Create: `web/src/api/learning/memory.ts`

- [ ] Add endpoint:

```text
GET /api/learning/memory?folder_id=<id>
```

- [ ] Response shape:

```json
{
  "summary": "用户已经能解释 Go 值与类型的基础关系，但代码运行验证证据还少。",
  "items": [
    {
      "kind": "mastered_concept",
      "statement": "用户已经能解释：值与类型的关系",
      "confidence": "observed",
      "last_seen_at": "2026-07-10T00:00:00Z"
    }
  ]
}
```

- [ ] Frontend page reads this endpoint when folder context exists; if no folder is selected, show recent memory across folders.

## Phase 5: Degrade `learning_profiles`

### Task 9: Update Docs And Product Contracts

**Files:**
- Modify: `docs/product/04-学习闭环与当前Agent边界.md`
- Modify: `docs/product/07-Eino学习Workflow契约.md`

- [ ] Replace "学习画像" as primary state with "学习记忆".
- [ ] Document `learning_profiles` as legacy/cache/projection.
- [ ] State the rule: "Exercises write facts; memory service derives learner context; Coach queries memory."

### Task 10: Optional Later Migration

**Do not do this in the first implementation batch.**

Later, after UI and Coach read from memory:

- Stop creating new `learning_profiles` rows.
- Keep old rows for backwards compatibility.
- Add a data migration that converts existing `completed_topics`, `verification_habits`, and `weak_points` into memory events/items.
- Remove profile from Coach prompt only after memory coverage is verified.

## Verification Plan

- Backend:
  - `go test ./app/learning/service ./app/learning/handlers ./app/learning/repository ./infrastructure/llm/prompts`
  - `go test ./app/learning/...`
- Frontend:
  - `pnpm --dir /Users/jujiuyey/Projects/Verve/web lint:type`
- Safety:
  - `git diff --check`
  - Submit one exercise locally and confirm:
    - exercise row exists
    - memory event exists
    - memory item exists for pass verdict
    - Coach prompt renders `学习记忆`
    - UI says `学习记忆`, not `我的画像`

## Recommended Delivery Slices

1. Storage + service only: schema, models, repository, memory service tests.
2. Exercise write path: record memory events/items while keeping legacy profile.
3. Coach read path: prefer memory context, keep profile fallback.
4. UI rename/read API: make the product surface match the architecture.
5. Legacy cleanup: migrate or deprecate `learning_profiles` after memory is trusted.
