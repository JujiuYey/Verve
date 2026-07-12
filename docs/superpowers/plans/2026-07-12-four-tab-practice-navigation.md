# Four-Tab Practice Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Promote Listener, Teacher, and Curator into three dedicated top-level practice tabs alongside Reading article, with filtered histories and independent drafts.

**Architecture:** Keep the server's unified chronological timeline and filter it by `turn.agent_type` in a tested frontend utility. Reuse the existing shadcn Tabs, AgentTimeline, and AgentComposer components; each Agent tab fixes its Agent type, while the page stores one draft per Agent.

**Tech Stack:** React 19, TypeScript, TanStack Query, shadcn/ui Tabs, Vitest, lucide-react

---

### Task 1: Add Agent Timeline Filtering

**Files:**
- Modify: `web/src/pages/learning/feynman-practice/_lib/timeline.ts`
- Modify: `web/src/pages/learning/feynman-practice/_lib/timeline.test.ts`

- [x] **Step 1: Write the failing filter test**

Update the test helper to accept an Agent and add a test that preserves chronological input order while returning only the requested Agent:

```ts
function item(
  id: string,
  requestId: string,
  status = "completed",
  agentType: LearningAgentType = "listener",
): TimelineItem {
  // existing item body
  result.turn.agent_type = agentType;
  return result;
}

it("filters timeline items by agent without changing order", () => {
  const timeline = [
    item("listener-1", "request-1", "completed", "listener"),
    item("teacher-1", "request-2", "completed", "teacher"),
    item("listener-2", "request-3", "completed", "listener"),
  ];
  expect(filterTimelineByAgent(timeline, "listener").map((entry) => entry.turn.id)).toEqual([
    "listener-1",
    "listener-2",
  ]);
});
```

- [x] **Step 2: Run the focused test and verify RED**

Run:

```bash
pnpm --dir web exec vitest run src/pages/learning/feynman-practice/_lib/timeline.test.ts
```

Expected: FAIL because `filterTimelineByAgent` is not exported.

- [x] **Step 3: Implement the minimal filter**

Add:

```ts
export function filterTimelineByAgent(
  timeline: TimelineItem[],
  agentType: LearningAgentType,
): TimelineItem[] {
  return timeline.filter((item) => item.turn.agent_type === agentType);
}
```

Import `LearningAgentType` with the existing timeline API types.

- [x] **Step 4: Run the focused test and verify GREEN**

Run the same Vitest command. Expected: all timeline tests pass.

### Task 2: Make the Composer Agent-Fixed

**Files:**
- Modify: `web/src/pages/learning/feynman-practice/_components/agent-composer.tsx`
- Delete: `web/src/pages/learning/feynman-practice/_components/agent-selector.tsx`

- [x] **Step 1: Remove nested Agent selection**

Rename the `selectedAgent` prop to `agentType`, remove `onAgentChange`, remove the `AgentSelector` import and render, and resolve labels from `copy[agentType]`.

Keep the completion button in the existing top action row when `canComplete` is true. The page passes `canComplete=false` for Teacher and Curator.

- [x] **Step 2: Delete the unused selector component**

Delete `agent-selector.tsx` and verify no `AgentSelector` imports remain:

```bash
rg -n "AgentSelector|agent-selector" web/src/pages/learning/feynman-practice
```

Expected: no matches.

### Task 3: Build Four Top-Level Tabs

**Files:**
- Modify: `web/src/pages/learning/feynman-practice/index.tsx`

- [x] **Step 1: Define tab-to-Agent mapping and independent drafts**

Add Teacher and Curator lucide icons. Replace `answer` and `selectedAgent` with:

```ts
type AgentDrafts = Record<LearningAgentType, string>;

const emptyDrafts = (): AgentDrafts => ({ listener: "", teacher: "", curator: "" });

const [drafts, setDrafts] = useState<AgentDrafts>(emptyDrafts);
const activeAgent: LearningAgentType | null =
  activeView === "listener" || activeView === "teacher" || activeView === "curator"
    ? activeView
    : null;
const answer = activeAgent ? drafts[activeAgent] : "";
```

Reset drafts with `setDrafts(emptyDrafts())` only when route identity changes. Keep `answerRef` synchronized with the active draft for request identity protection.

- [x] **Step 2: Submit with the fixed tab Agent**

Require an Agent from either `agentOverride` or `activeAgent`. On a normal successful submission, clear only that Agent's draft when it still equals the submitted value:

```ts
setDrafts((current) =>
  current[agentType] === submittedAnswer ? { ...current, [agentType]: "" } : current,
);
```

Keep Curator regeneration using its explicit `agentOverride` and content override.

- [x] **Step 3: Render the four triggers**

Use the existing `TabsList` and render:

```tsx
<TabsTrigger value="source"><BookOpenIcon />阅读文章</TabsTrigger>
<TabsTrigger value="listener"><MessageSquareTextIcon />开始讲解{listenerItems.length ? ` (${listenerItems.length})` : ""}</TabsTrigger>
<TabsTrigger value="teacher"><GraduationCapIcon />教学补充{teacherItems.length ? ` (${teacherItems.length})` : ""}</TabsTrigger>
<TabsTrigger value="curator"><FilePenIcon />修订文档{curatorItems.length ? ` (${curatorItems.length})` : ""}</TabsTrigger>
```

Compute each list with `filterTimelineByAgent`.

- [x] **Step 4: Render one fixed Agent workspace per tab**

Extract a page-local `AgentWorkspace` only if needed to avoid repeating the section, timeline, and composer markup. Each workspace receives its filtered items, fixed `agentType`, draft, and completion availability.

Show index pending/failed alerts only in the Teacher tab. Show the completed-session alert in all three Agent tabs. Keep Curator apply, cancel, and regenerate callbacks attached to its filtered items.

- [x] **Step 5: Format and type-check**

Run:

```bash
pnpm --dir web exec oxfmt --write \
  src/pages/learning/feynman-practice/index.tsx \
  src/pages/learning/feynman-practice/_components/agent-composer.tsx \
  src/pages/learning/feynman-practice/_lib/timeline.ts \
  src/pages/learning/feynman-practice/_lib/timeline.test.ts
pnpm --dir web lint:type
```

Expected: both commands exit successfully.

### Task 4: Final Verification

**Files:**
- Verify only; no new frontend layout tests.

- [x] **Step 1: Run focused logic tests**

```bash
pnpm --dir web exec vitest run src/pages/learning/feynman-practice/_lib/timeline.test.ts
```

Expected: all tests pass.

- [x] **Step 2: Run repository checks**

```bash
pnpm --dir web lint:type
git diff --check
```

Expected: both commands exit successfully.

- [x] **Step 3: Hand off visual acceptance**

Keep the existing frontend dev server running and report `http://127.0.0.1:5201/`. Per project policy, do not run screenshots or browser automation; ask the user to inspect the four tabs, filtered histories, independent drafts, and fixed Agent submissions directly.
