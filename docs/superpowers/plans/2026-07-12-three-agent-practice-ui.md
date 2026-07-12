# Unified Timeline and Agent Selector UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Listener-only practice panel with one recoverable timeline, shared composer, explicit three-Agent selector, and inline Curator confirmation.

**Architecture:** The page consumes server `TimelineItem` values and keeps only the selected Agent, draft text, and in-flight request identity locally. Pure timeline reconciliation and artifact-state functions are tested separately from layout components.

**Tech Stack:** React 19, TypeScript, TanStack Query, shadcn/ui, ai-elements, Tailwind CSS

---

### Task 1: Add frontend contracts and API hooks

**Files:**
- Modify: `web/src/api/learning/session.ts`
- Modify: `web/src/api/wiki/document.ts`
- Modify: `web/src/api/rag/wiki.ts`

- [ ] **Step 1: Add exact discriminated-union types**

```ts
export type LearningAgentType = "listener" | "teacher" | "curator";
export type TurnArtifact =
  | { type: "explanation_review"; data: LearningExplanationReview }
  | { type: "teaching_intervention"; data: LearningTeachingIntervention }
  | { type: "wiki_change_request"; data: WikiDocumentChangeRequest };
export interface TimelineItem {
  turn: LearningTurn;
  user_message: LearningMessage;
  assistant_message?: LearningMessage;
  artifact?: TurnArtifact;
}
```

- [ ] **Step 2: Add hooks**

Add `useSubmitTurn`, `useApplyWikiChangeRequest`, `useCancelWikiChangeRequest`, `useDocumentIndexStatus`, and `useRetryDocumentIndex`. Extend index status with `superseded` and `document_version`. `SessionDetail` includes `timeline` while keeping existing compatibility fields.

- [ ] **Step 3: Run type checking**

Run: `pnpm --dir web lint:type`

Expected: PASS before page migration because the new fields are additive.

- [ ] **Step 4: Commit**

```bash
git add web/src/api/learning/session.ts web/src/api/wiki/document.ts web/src/api/rag/wiki.ts
git commit -m "feat(web): add three-agent timeline contracts"
```

### Task 2: Build and test timeline reconciliation

**Files:**
- Delete: `web/src/pages/learning/feynman-practice/_lib/review-turns.ts`
- Delete: `web/src/pages/learning/feynman-practice/_lib/review-turns.test.ts`
- Create: `web/src/pages/learning/feynman-practice/_lib/timeline.ts`
- Create: `web/src/pages/learning/feynman-practice/_lib/timeline.test.ts`

- [ ] **Step 1: Write failing reconciliation tests**

Cover server ordering, one local placeholder consumed per persisted request ID, stale server snapshots, artifact status replacement, and conflict regeneration input.

```ts
export function mergeTimeline(current: TimelineItem[], server: TimelineItem[]): TimelineItem[];
export function curatorRegenerationInput(item: TimelineItem): {
  content: string;
  replaces_change_request_id: string;
};
```

- [ ] **Step 2: Run tests and verify RED**

Run: `pnpm --dir web exec vitest run src/pages/learning/feynman-practice/_lib/timeline.test.ts`

- [ ] **Step 3: Implement pure functions and verify GREEN**

Run the same command and expect PASS.

- [ ] **Step 4: Commit**

```bash
git add web/src/pages/learning/feynman-practice/_lib
git commit -m "test(web): define agent timeline reconciliation"
```

### Task 3: Add focused Agent UI components

**Files:**
- Create: `web/src/pages/learning/feynman-practice/_components/agent-selector.tsx`
- Create: `web/src/pages/learning/feynman-practice/_components/agent-composer.tsx`
- Create: `web/src/pages/learning/feynman-practice/_components/agent-timeline.tsx`
- Create: `web/src/pages/learning/feynman-practice/_components/listener-artifact.tsx`
- Create: `web/src/pages/learning/feynman-practice/_components/teacher-artifact.tsx`
- Create: `web/src/pages/learning/feynman-practice/_components/curator-artifact.tsx`
- Delete: `web/src/pages/learning/feynman-practice/_components/practice-panel.tsx`

- [ ] **Step 1: Build `AgentSelector` from shadcn ToggleGroup**

Use `type="single"`, never allow an empty selection, include Ear, GraduationCap, and FilePen icons, and disable it during a request or completed session.

- [ ] **Step 2: Build the shared composer**

Reuse `FeynmanAnswerEditor`. Map selected Agent to explicit placeholder, submit label, and progress text. Preserve the existing end-session controls and Listener wrap-up guidance.

- [ ] **Step 3: Build the timeline shell**

Use `ai-elements/Message` for user and assistant content, a Badge for Agent identity, and stable dimensions so status changes do not shift the composer.

- [ ] **Step 4: Build artifact renderers**

Move current review presentation into `ListenerArtifact`. Render Teacher structured lists and evidence without nested cards. Render Curator diff inside `Artifact` and actions inside `Confirmation`; map proposed to approval-requested, applied to accepted, cancelled to rejected, and conflict/failed to Alert states.

- [ ] **Step 5: Format and type-check components**

```bash
pnpm --dir web exec oxfmt --write src/pages/learning/feynman-practice/_components
pnpm --dir web lint:type
```

- [ ] **Step 6: Commit**

```bash
git add web/src/pages/learning/feynman-practice/_components
git commit -m "feat(web): add three-agent practice components"
```

### Task 4: Wire the page to unified turns

**Files:**
- Modify: `web/src/pages/learning/feynman-practice/index.tsx`

- [ ] **Step 1: Replace review-only local state**

Use:

```ts
const [selectedAgent, setSelectedAgent] = useState<LearningAgentType>("listener");
const [timeline, setTimeline] = useState<TimelineItem[]>([]);
```

Reset selected Agent to Listener only when route identity changes or the page reloads. Merge `sessionDetail.timeline` after each refetch.

- [ ] **Step 2: Submit selected Agent turns**

Generate one request UUID per submit, add a local processing item, call `useSubmitTurn`, and replace it from the response. Disable selector and composer while the request identity is active.

- [ ] **Step 3: Wire Curator apply, cancel, and regenerate**

Apply and cancel update the originating artifact from the API response and invalidate both session detail and Wiki document content. Conflict regeneration submits a new Curator turn with a new request UUID and the original instruction.

- [ ] **Step 4: Preserve session completion semantics**

Count Listener artifacts for wrap-up messaging, but render all Agent turns. Do not let Teacher or Curator artifacts alone mark Listener understanding or mastery.

- [ ] **Step 5: Show current-version index status**

Display a compact Alert when the current version is pending/running or failed. A failed state calls `useRetryDocumentIndex`; pending/running states do not expose stale evidence or a redundant retry button.

- [ ] **Step 6: Run type and logic checks**

```bash
pnpm --dir web exec vitest run src/pages/learning/feynman-practice/_lib/timeline.test.ts
pnpm --dir web lint:type
```

- [ ] **Step 7: Commit**

```bash
git add web/src/pages/learning/feynman-practice/index.tsx
git commit -m "feat(web): wire explicit agent selection into practice"
```

### Task 5: Verify the complete three-Agent feature

**Files:** all files from Plans 1-3

- [ ] **Step 1: Run frontend verification**

```bash
pnpm --dir web exec vitest run src/pages/learning/feynman-practice/_lib/timeline.test.ts
pnpm --dir web lint:type
pnpm --dir web exec oxfmt --check src/api/learning/session.ts src/api/wiki/document.ts src/pages/learning/feynman-practice
```

- [ ] **Step 2: Run backend verification**

Run: `cd server && go test -count=1 ./...`

- [ ] **Step 3: Check the full diff**

Run: `git diff --check && git status --short`

Expected: no formatting errors and only planned implementation files remain.

- [ ] **Step 4: Start the application for user acceptance**

Start the existing backend and frontend development commands on available ports and report the Feynman practice route. Do not take screenshots or run browser automation; ask the user to inspect the selector, timeline, Teacher response, Curator diff, apply/cancel, and refresh behavior directly.
