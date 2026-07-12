# Three-Agent Learning Experience Design

## Status

Approved product decisions:

- The learner explicitly selects `FeynmanListener`, `LearningTeacher`, or `WikiCurator` for every input.
- The practice flow does not use a Supervisor or model-based router.
- The backend keeps one chronological timeline, while the practice page filters history into three Agent-specific top-level tabs.
- The page has four top-level tabs: Reading article, Start explaining, Teaching supplement, and Revise document.
- Each Agent tab fixes the target Agent and keeps its own local draft; there is no nested Agent selector.
- WikiCurator proposals, diffs, confirmation, cancellation, conflicts, and failures render inline in the originating turn.
- Wiki writes use proposal, user confirmation, and deterministic backend execution.
- Wiki document versions and RAG index versions are bound.

This specification extends the accepted data model in `docs/temp/three-agent-data-model.md` and the shipped Learning foundation in commit `878cd6f`.

## Goal

Deliver a complete three-Agent practice experience in which a learner can:

1. Ask the Listener to review an explanation.
2. Ask the Teacher for a grounded teaching intervention when stuck.
3. Ask the Curator to propose a Wiki edit, inspect the diff, and explicitly confirm or cancel it.
4. Reload the page and recover the same ordered turn history and artifacts.
5. Trust that teaching and review use the current Wiki version and never silently use stale RAG evidence.

The feature is complete only when all three Agent paths have real backend behavior and are available through their top-level practice tabs. A schema-only, API-only, or UI-only phase is not the completed feature.

## Non-Goals

- No Supervisor, automatic Agent routing, or model-selected tools for choosing an Agent.
- No session-level `active_agent` or Agent state machine.
- No multi-document practice session.
- No automatic Wiki write without confirmation.
- No general Wiki revision browser, rollback UI, or background orphan-object cleaner in this release.
- No removal of the existing `/review` compatibility endpoint.
- No redesign of the Wiki management page or learning entry page.
- No `learning_journals` cleanup.

## Existing Baseline

The repository already contains:

- `learning_sessions`, `learning_turns`, `learning_messages`, `learning_explanation_reviews`, and `learning_teaching_interventions`.
- Idempotent Listener turn creation and completion through the current review handler.
- Current document loading and document-scoped RAG context for Listener review.
- Wiki documents stored in MinIO with metadata in PostgreSQL.
- Markdown-aware chunks in PostgreSQL and vectors in Qdrant.
- A per-document RAG indexing flow.
- shadcn `Tabs`, `Badge`, `Alert`, and `Button` components.
- `ai-elements` message, artifact, and confirmation components.

Missing pieces are Wiki revisions, version-aware RAG, Teacher and Curator services, a unified Turn API, unified timeline responses, and the four-tab practice UI.

## Architecture Decision

Use one explicit Turn API rather than three unrelated submit APIs or a routing Agent:

```text
Agent-specific top-level tab and shared composer component
  -> POST /api/learning/session/:id/turns
      -> listener service
      -> teacher service
      -> curator service
```

The request names the selected Agent. The handler performs authentication, session ownership validation, idempotent turn creation, and deterministic dispatch. Each Agent service owns its prompt, context mapping, structured output parsing, and artifact persistence.

The existing `/api/learning/session/:id/review` endpoint remains as a compatibility adapter to the Listener path. It must not maintain a separate implementation.

## Agent Responsibilities

### FeynmanListener

The Listener:

- Receives the learner's explanation.
- Uses the latest document text or current-version RAG evidence.
- Returns the existing structured explanation review.
- Writes one assistant message and one `learning_explanation_reviews` artifact.
- May create explanation evidence and misconception memory.
- Cannot write teaching interventions or Wiki changes.

### LearningTeacher

The Teacher:

- Receives the learner's question or description of where they are stuck.
- Uses the latest document text or current-version RAG evidence.
- May use prior Listener reviews and document-scoped learning memory to understand the learner's context.
- Returns a direct teaching response plus structured knowledge gaps, key points, examples, evidence, and a concise explanation summary.
- Writes one assistant message and one `learning_teaching_interventions` artifact.
- May create a memory event meaning that help was requested or received.
- Cannot create mastery memory, Listener reviews, or Wiki changes.

The Teacher structured model output is:

```json
{
  "response": "direct teaching response",
  "question_summary": "learner's sticking point",
  "knowledge_gaps": ["missing prerequisite"],
  "explanation_summary": "what was taught",
  "key_points": ["key point"],
  "examples": ["example"],
  "evidence": [{"document_version": 2, "chunk_id": "...", "heading_path": "..."}]
}
```

### WikiCurator

The Curator:

- Receives a concrete edit instruction.
- Loads the complete current Markdown and its `current_version`.
- Returns only a structured proposal: proposed complete Markdown and a change summary.
- Does not receive MinIO, database, or indexing write tools.
- Writes one assistant message and one `wiki_document_change_requests` artifact.
- Cannot apply its own proposal.

The backend validates non-empty Markdown and size limits, computes a unified diff using a maintained diff library, and persists both proposed content and diff. Curator input is limited to documents of at most 60,000 Unicode code points. Larger documents return a clear unsupported-size error instead of truncating content or generating a partial full-document proposal.

## Unified Turn API

### Submit Turn

```http
POST /api/learning/session/:id/turns
```

```json
{
  "request_id": "client-generated UUID",
  "agent_type": "listener | teacher | curator",
  "content": "learner input",
  "replaces_change_request_id": "optional conflicted request ID"
}
```

Validation rules:

- The session must belong to the current user and be `active`.
- `request_id`, `agent_type`, and trimmed `content` are required.
- The route session document is the only document an Agent may access.
- The same `(session_id, request_id)` with different content or Agent type returns HTTP 409.

Idempotency behavior:

- New request: create a processing turn and user message before invoking an Agent.
- Existing processing request: return HTTP 409 with `turn_status=processing`.
- Existing completed request: return the persisted timeline item without invoking the model.
- Existing failed request with identical input and Agent: reset the same turn to processing and retry.

The success response is one timeline item:

```json
{
  "turn": {
    "id": "...",
    "agent_type": "teacher",
    "status": "completed",
    "created_at": "...",
    "completed_at": "..."
  },
  "user_message": {"id": "...", "content": "..."},
  "assistant_message": {"id": "...", "content": "..."},
  "artifact": {
    "type": "teaching_intervention",
    "data": {}
  }
}
```

`artifact.type` is exactly one of:

- `explanation_review`
- `teaching_intervention`
- `wiki_change_request`

### Read Session Timeline

`GET /api/learning/session/:id` keeps existing `session`, `messages`, and `reviews` fields for compatibility and adds:

```json
{
  "timeline": [
    {
      "turn": {},
      "user_message": {},
      "assistant_message": {},
      "artifact": {"type": "...", "data": {}}
    }
  ]
}
```

Timeline rows are ordered by `turn.created_at ASC, turn.id ASC`. A failed turn has a user message, no assistant message, and no artifact. The response exposes the persisted failure code but not internal model or storage error details.

## Wiki Version Model

### wiki_documents

Add:

- `current_version BIGINT NOT NULL DEFAULT 1`
- `content_hash VARCHAR(64) NULL`

`file_path` always points to the immutable object for `current_version`. `content_hash` is temporarily nullable only for legacy documents that have not yet been snapshotted.

Every content write uses the same document-version service. This includes Curator apply and the existing `PUT /api/wiki/documents/:id/content` editor save path; the existing editor may not overwrite the current MinIO object in place or bypass revision and RAG job creation.

### wiki_document_change_requests

Store:

- document and requesting user
- source type and source turn ID
- idempotent request ID
- base version
- original instruction
- optional replaced change-request ID for conflict regeneration
- model change summary
- complete proposed Markdown
- backend-generated unified diff
- status: `proposed`, `applied`, `failed`, `cancelled`, or `conflict`
- applied version and failure metadata

There is one change request per Curator turn. Learning depends on Wiki, so `source_id` is validated in the service layer rather than through a reverse database foreign key to `learning_turns`.

### wiki_document_revisions

Store immutable revision metadata:

- document and version
- immutable MinIO object path
- SHA-256 content hash
- byte size
- optional change request
- changing user and summary
- creation time

`(document_id, version)` is unique.

New uploads create revision 1 immediately. Legacy documents are snapshotted lazily before their first Curator apply: the service reads the current object, writes it to an immutable revision-1 path, calculates its hash, inserts revision 1, and fills `wiki_documents.content_hash`.

Deleting a Wiki document removes Qdrant vectors and every known revision object path before deleting PostgreSQL metadata. The background cleanup non-goal applies only to unreferenced objects left by failed cross-system writes.

## Curator Confirmation Flow

```text
Curator turn
  -> proposed change request
  -> inline diff
  -> user applies or cancels
```

Endpoints:

```http
POST /api/wiki/change-requests/:id/apply
POST /api/wiki/change-requests/:id/cancel
```

Apply sequence:

1. Validate current user, document access, request ownership, and `proposed` status.
2. Ensure legacy revision 1 exists when required.
3. Read and validate proposed content.
4. Write proposed content to a new immutable MinIO object path.
5. Start a PostgreSQL transaction.
6. Lock the document and change request.
7. Recheck `current_version == base_version`.
8. On mismatch, mark the request `conflict` and commit without updating the document.
9. Insert the next revision.
10. Update document version, path, hash, and size.
11. Mark the request `applied`.
12. Create a RAG job fixed to the new document version and object path.
13. Commit.
14. Trigger asynchronous indexing for that exact job.

If MinIO succeeds but the transaction fails, the database never points to the new object. The unreferenced object is acceptable for this release and can be removed by a later cleanup job.

Repeated apply for an already applied request returns the persisted revision result. Cancel is idempotent for an already cancelled request. Applied, cancelled, conflict, and failed requests cannot transition back to proposed.

A conflict is never overwritten or mutated into a new proposal. The UI creates a new Curator turn with a new `request_id`, the same instruction, and `replaces_change_request_id` referencing the conflicted request.

## Version-Aware RAG

Add `document_version` to:

- `rag_index_jobs`
- `rag_wiki_chunks`
- Qdrant point payloads
- retrieval result payloads

Vector point IDs include document ID, version, and chunk identity. Existing version-1 Qdrant payloads do not require an eager backfill because PostgreSQL performs the authoritative version check.

Index jobs add `superseded` to their status set and allow only one non-superseded row per `(document_id, document_version)`. A job records the immutable revision object path when created and never reloads `wiki_documents.file_path` while running. During migration, existing jobs become version 1 and all but the newest job per document become superseded.

Publish sequence:

1. Read the job's immutable revision object.
2. Chunk and embed that exact version.
3. Before publishing, compare job version with the locked document current version.
4. If different, mark the job `superseded` and do not publish.
5. Read and retain the previous PostgreSQL chunk point IDs for cleanup.
6. Upsert new Qdrant points with `document_version`.
7. In PostgreSQL, replace the document's chunk rows with the new version and complete the job.
8. Delete the retained old Qdrant point IDs best-effort after the current version is available.

Retrieval uses two guards:

- Qdrant uses document or root-folder scope and returns more candidates than the requested result count.
- PostgreSQL joins every candidate chunk against `wiki_documents.current_version` before returning it.

The PostgreSQL check is authoritative. A failed old-vector cleanup can reduce recall until a later cleanup succeeds, but it cannot make stale content valid.

When the current version is not indexed:

- Never return chunks from an older version.
- Documents within the existing full-document character limit use current Markdown directly.
- Larger documents return a typed `index_not_ready` context result.
- The practice page shows that knowledge retrieval is still syncing and offers retry after the job fails.

The existing document indexing endpoint retries the current version rather than creating a duplicate job, and a focused status endpoint supports the practice page:

```text
GET  /api/rag/wiki/documents/:id/index-status
POST /api/rag/wiki/documents/:id/index
```

Legacy documents, chunks, and jobs are assigned version 1 during migration. Existing Qdrant points remain usable because their PostgreSQL chunk rows supply version 1; newly indexed points include `document_version` in both stores.

## Practice Page Experience

The practice page exposes four top-level tabs:

```text
[ Reading article ] [ Start explaining ] [ Teaching supplement ] [ Revise document ]
```

Rules:

- Reading article renders the current source document without an Agent composer.
- Start explaining fixes `agent_type=listener` and shows only Listener turns.
- Teaching supplement fixes `agent_type=teacher` and shows only Teacher turns.
- Revise document fixes `agent_type=curator` and shows only Curator turns and proposal actions.
- The backend session timeline remains unified and chronological; filtering is presentation-only and does not create separate sessions.
- Each Agent tab keeps an independent local draft so switching tasks does not carry input into another Agent.
- While a request is processing, disable duplicate submission. The learner may still inspect another top-level tab.
- Completing the session disables all new Agent turns.
- The completion action is available only in Start explaining after at least one Listener turn.
- Current-version index synchronization and retry messaging appear in Teaching supplement, where grounded retrieval is required.
- Each Agent tab count reflects only that Agent's turns.

Timeline rendering:

- Each Agent tab filters the unified timeline by `turn.agent_type` and preserves chronological order within that filtered history.
- Every item displays Agent identity and turn status.
- Listener artifact uses the current review presentation.
- Teacher artifact displays the response, knowledge gaps, key points, examples, and evidence.
- Curator artifact uses `ai-elements` artifact and confirmation primitives to show summary, unified diff, apply, cancel, conflict, failure, and applied revision.
- Curator confirmation remains inside the originating timeline item.
- A conflict item offers `Regenerate from latest version`, which submits a new Curator turn rather than mutating the old request.
- Refresh reconstructs the timeline entirely from `GET /session/:id`.

Use existing shadcn components first, then existing `ai-elements`. No new general-purpose chat framework or one-off replacement for an available component is allowed.

## Security Boundaries

- Session ownership is checked before every turn operation.
- Document access is checked through the existing Wiki folder ownership rules.
- Change request apply and cancel require both document access and request ownership.
- Model-generated Markdown is data, never executable instructions.
- Document text, RAG evidence, prior messages, memory, and user input are rendered as untrusted prompt sections.
- Curator cannot call write tools.
- Internal error messages are logged server-side and reduced to stable error codes in API responses.

## Failure Semantics

- Agent initialization, context, model, parsing, or persistence failure marks the turn failed and retains the user message.
- Curator generation failure creates no change request.
- Apply version mismatch marks the request conflict and returns HTTP 409.
- Apply persistence failure marks the request failed when safe to do so; it never advances the document.
- RAG failure does not roll back an applied Wiki revision. The job is retryable and the UI reports the unsynchronized index.
- A late old-version index job becomes superseded.
- Best-effort cleanup failures never make stale evidence eligible for retrieval.

## Testing Strategy

Backend work follows TDD.

Required coverage:

- Turn request validation, ownership, input identity, completed replay, processing conflict, and failed retry.
- Teacher prompt boundaries, output parsing, intervention persistence, and memory restrictions.
- Curator prompt boundaries, output validation, deterministic diff generation, and no-write behavior.
- Legacy revision snapshot, apply success, repeated apply, cancel, access denial, version conflict, MinIO failure, and transaction failure.
- RAG job version pinning, superseded jobs, current-version chunk replacement, Qdrant payloads, and stale-result rejection.
- Timeline assembly and artifact mapping.
- Frontend timeline merge, Agent filtering, artifact mapping, independent drafts, and Curator state transitions.

Per project policy, do not add tests for page layout, tab composition, or ordinary route wiring. Do not use screenshots, browser automation, or visual self-check loops. The user performs visual acceptance in the running page.

## Delivery Decomposition

The feature uses one product specification and three sequential implementation plans.

### Plan 1: Wiki Revisions and Version-Aware RAG

Delivers Wiki revision persistence, legacy snapshot behavior, versioned indexing, current-version retrieval, and the apply/cancel backend service. It is independently testable without Agent or UI changes.

### Plan 2: Teacher, Curator, and Unified Turn API

Delivers Agent constructors, prompts, services, artifact persistence, unified submission, completed replay, and timeline assembly. It depends on Plan 1 for Curator proposals and revision apply behavior.

### Plan 3: Unified Timeline and Agent Practice UI

Delivers the shared composer component, four top-level tabs, Agent-filtered histories, three artifact renderers, inline Curator confirmation, conflict regeneration, and index synchronization states. It depends on Plans 1 and 2.

No plan may be described as the completed three-Agent feature by itself.

## Acceptance Criteria

The feature is complete when:

- The practice page visibly offers Reading article, Start explaining, Teaching supplement, and Revise document as four top-level tabs.
- Each Agent tab sends input to exactly its fixed Agent without a nested selector.
- Listener, Teacher, and Curator each persist the correct artifact in one backend timeline and appear only in their corresponding filtered history.
- Switching Agent tabs preserves independent drafts, and refresh restores the persisted timeline and artifact states.
- Curator never writes before explicit confirmation.
- Apply creates an immutable revision and a version-pinned RAG job.
- Version conflicts cannot overwrite newer Markdown.
- Retrieval never returns an older document version as current evidence.
- Failed and superseded states are visible and retry behavior is deterministic.
- Backend tests, frontend logic tests, frontend type checks, formatting checks, and full Go tests pass.
- The user visually accepts the page in the local application.
