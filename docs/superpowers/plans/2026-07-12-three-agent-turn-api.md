# Teacher, Curator, and Unified Turn API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Add real Teacher and Curator Agents and expose all three explicit Agent paths through one idempotent Turn API and recoverable timeline.

**Architecture:** A Learning `TurnService` validates a selected Agent, owns the shared turn lifecycle, and dispatches to three focused processors. Each processor persists exactly one typed artifact; TimelineRepository assembles turns, messages, and artifacts for API replay.

**Tech Stack:** Go 1.25, Fiber, Bun, Eino ADK, PostgreSQL

---

### Task 1: Generalize turn identity and persistence

**Files:**
- Modify: `server/app/learning/repository/turn.go`
- Modify: `server/app/learning/repository/turn_test.go`
- Create: `server/app/learning/repository/intervention.go`
- Create: `server/app/learning/repository/intervention_test.go`
- Modify: `server/app/learning/models/db/intervention.go`

- [x] **Step 1: Write failing generic-turn tests**

Cover `BeginTurn(sessionID, requestID, agentType, content)`, payload mismatch conflict, failed retry, assistant-message persistence, and one intervention per Teacher turn.

```go
type BeginTurnInput struct {
    SessionID string
    RequestID string
    AgentType string
    Content string
}
```

- [x] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/learning/repository -run 'Turn|Intervention'`

Expected: FAIL because only Listener-specific persistence exists.

- [x] **Step 3: Implement generic turn methods**

Replace `BeginListenerTurn` internals with `BeginTurn`; keep the old method as a small adapter. Add transactional completion methods for review, intervention, and change-request references. Store Teacher evidence as a JSON array of typed evidence records.

- [x] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/learning/repository`

- [x] **Step 5: Commit**

```bash
git add server/app/learning/models/db/intervention.go server/app/learning/repository
git commit -m "refactor(learning): generalize agent turn persistence"
```

### Task 2: Add LearningTeacher prompt, Agent, and service

**Files:**
- Create: `server/infrastructure/llm/prompts/learning_teacher.go`
- Modify: `server/infrastructure/llm/prompts/prompts_test.go`
- Modify: `server/infrastructure/llm/agent.go`
- Create: `server/app/learning/service/teacher.go`
- Create: `server/app/learning/service/teacher_test.go`

- [x] **Step 1: Write failing prompt and service tests**

Verify untrusted boundaries, JSON-only output, no mastery claim, current-version evidence mapping, output validation, and `index_not_ready` behavior.

```go
type TeachingResult struct {
    Response string `json:"response"`
    QuestionSummary string `json:"question_summary"`
    KnowledgeGaps []string `json:"knowledge_gaps"`
    ExplanationSummary string `json:"explanation_summary"`
    KeyPoints []string `json:"key_points"`
    Examples []string `json:"examples"`
    Evidence []TeachingEvidence `json:"evidence"`
}
```

- [x] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./infrastructure/llm/prompts ./app/learning/service -run 'Teacher|Teaching'`

- [x] **Step 3: Implement `NewLearningTeacherAgent` and `TeacherService`**

Use `NewStructuredChatModel`, the existing Feynman document source, prior reviews, and memory. The service returns structured data only; the turn processor stores `result.Response` as the assistant message.

- [x] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./infrastructure/llm/... ./app/learning/service`

- [x] **Step 5: Commit**

```bash
git add server/infrastructure/llm server/app/learning/service/teacher.go server/app/learning/service/teacher_test.go
git commit -m "feat(learning): add grounded teaching agent"
```

### Task 3: Add WikiCurator proposal generation

**Files:**
- Modify: `server/go.mod`
- Modify: `server/go.sum`
- Create: `server/infrastructure/llm/prompts/wiki_curator.go`
- Modify: `server/infrastructure/llm/prompts/prompts_test.go`
- Modify: `server/infrastructure/llm/agent.go`
- Create: `server/app/learning/service/curator.go`
- Create: `server/app/learning/service/curator_test.go`

- [x] **Step 1: Write failing Curator tests**

Verify full-document input, the 60,000-code-point limit, untrusted boundaries, JSON-only proposed Markdown, non-empty validation, deterministic unified diff, and absence of write tools.

```go
type CuratorResult struct {
    ChangeSummary string `json:"change_summary"`
    ProposedContent string `json:"proposed_content"`
}
```

- [x] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./infrastructure/llm/prompts ./app/learning/service -run Curator`

- [x] **Step 3: Implement Curator with `github.com/sergi/go-diff/diffmatchpatch`**

Generate unified patch text on the backend. Persist a `DocumentChangeRequest` with `source_type=learning_turn`, the turn ID, base version, instruction, optional `replaces_change_request_id`, proposed content, and diff. Do not expose a write tool to Eino.

- [x] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./infrastructure/llm/... ./app/learning/service`

- [x] **Step 5: Commit**

```bash
git add server/go.mod server/go.sum server/infrastructure/llm server/app/learning/service/curator.go server/app/learning/service/curator_test.go
git commit -m "feat(learning): add safe wiki curator proposals"
```

### Task 4: Add timeline assembly

**Files:**
- Create: `server/app/learning/models/payload/timeline.go`
- Create: `server/app/learning/repository/timeline.go`
- Create: `server/app/learning/repository/timeline_test.go`
- Modify: `server/infrastructure/database/database.go`

- [x] **Step 1: Write failing timeline tests**

Test deterministic ordering, three artifact variants, failed turns without assistant messages, and omission of internal error messages.

```go
type TimelineItem struct {
    Turn *LearningTurn `json:"turn"`
    UserMessage *LearningMessage `json:"user_message"`
    AssistantMessage *LearningMessage `json:"assistant_message,omitempty"`
    Artifact *TurnArtifact `json:"artifact,omitempty"`
}
```

- [x] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/learning/repository -run Timeline`

- [x] **Step 3: Implement batched timeline queries**

Load turns once, messages once, and each artifact table once for the session; assemble in Go by turn ID. Do not issue per-turn queries.

- [x] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/learning/repository`

- [x] **Step 5: Commit**

```bash
git add server/app/learning/models/payload/timeline.go server/app/learning/repository/timeline.go server/app/learning/repository/timeline_test.go server/infrastructure/database/database.go
git commit -m "feat(learning): assemble unified agent timeline"
```

### Task 5: Implement TurnService orchestration

**Files:**
- Create: `server/app/learning/service/turn.go`
- Create: `server/app/learning/service/turn_test.go`
- Modify: `server/app/learning/service/memory.go`
- Modify: `server/app/learning/service/memory_test.go`

- [x] **Step 1: Write failing orchestration tests**

Cover explicit dispatch, no Supervisor, completed replay, processing conflict, failed retry, artifact-specific completion, failure marking, Teacher help memory events, and Curator no-memory behavior.

- [x] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/learning/service -run Turn`

- [x] **Step 3: Implement processors behind one service**

```go
type TurnRequest struct {
    RequestID string
    AgentType string
    Content string
    ReplacesChangeRequestID *string
}
type AgentProcessor interface {
    Process(ctx context.Context, input AgentInput) (*AgentOutput, error)
}
func (s *TurnService) Submit(ctx context.Context, userID, sessionID string, req TurnRequest) (*TimelineItem, error)
```

Use a fixed map keyed by `listener`, `teacher`, and `curator`; never ask a model to select the processor.

- [x] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/learning/service`

- [x] **Step 5: Commit**

```bash
git add server/app/learning/service
git commit -m "feat(learning): orchestrate explicit agent turns"
```

### Task 6: Expose the unified Turn API and compatibility adapter

**Files:**
- Modify: `server/app/learning/models/payload/session.go`
- Modify: `server/app/learning/handlers/session.go`
- Modify: `server/app/learning/handlers/session_test.go`
- Modify: `server/app/learning/router/learning.go`
- Modify: `server/router/router.go`

- [x] **Step 1: Write failing handler tests**

Cover request validation, ownership, all Agent types, 409 mismatch/processing/conflict, timeline in session detail, stable error codes, and `/review` delegation to Listener.

- [x] **Step 2: Run tests and verify RED**

Run: `cd server && go test ./app/learning/handlers`

- [x] **Step 3: Add endpoint and dependency wiring**

```text
POST /api/learning/session/:id/turns
```

Construct the three processors and TurnService in router wiring with existing DB, MinIO, Retriever, and the Wiki change-request repository. Keep handlers thin.

- [x] **Step 4: Run tests and verify GREEN**

Run: `cd server && go test ./app/learning/handlers ./app/learning/router`

- [x] **Step 5: Commit**

```bash
git add server/app/learning server/router/router.go
git commit -m "feat(api): expose explicit learning agent turns"
```

### Task 7: Verify Plan 2

- [x] **Step 1: Format and run all backend tests**

```bash
cd server
gofmt -w app/learning infrastructure/llm infrastructure/database router
go test -count=1 ./...
```

- [x] **Step 2: Verify boundaries**

Run: `git diff --check && rg -n 'Supervisor|active_agent' server/app/learning server/infrastructure/llm`

Expected: no implementation introduces Supervisor routing or session active-Agent state.
