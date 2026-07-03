---
name: eino-agents
description: Add or modify Verve backend Eino learning agents. Use when changing server/infrastructure/llm/agent.go, server/infrastructure/llm/prompts, learning agent tools, Coach/Tutor/Guide/Examiner/ObjectiveGenerator behavior, SSE agent streaming, or tests around Verve learning agents.
---

# Eino Agents

Use this skill for backend agent work in Verve. Keep changes grounded in the live repository; product direction docs are context, but current Go files are the source of truth.

## Start Here

1. Read `server/infrastructure/llm/agent.go` to identify the current agent constructors.
2. Read the relevant files under `server/infrastructure/llm/prompts/` before editing prompt text or prompt input shape.
3. Read the caller before changing agent behavior:
   - Coach entry: `server/app/learning/handlers/coach.go`
   - Tutor and Examiner flow: `server/app/learning/handlers/session.go`
   - Guide service: `server/app/learning/service/guide.go`
   - Objective generation service: `server/app/learning/service/objective_generation.go`
   - Examiner service: `server/app/learning/service/examiner.go`
4. Read tool factories before adding tool-dependent behavior:
   - Coach tools: `server/app/learning/tools/coach_tools.go`
   - Tutor tools: `server/app/learning/tools/learning_tools.go`
5. Read SSE vocabulary before changing streaming output: `server/app/learning/handlers/sse_events.go`.

## Current Agent Boundaries

- `Guide`: reads Markdown plus an objective and returns structured JSON for guide content.
- `ObjectiveGenerator`: reads Wiki Markdown and returns structured JSON learning objectives.
- `LearningCoach`: chooses the next learning step from runtime context and tools; may emit `<ACTION>{"type":"navigate_to_practice",...}</ACTION>`.
- `Tutor`: runs Feynman-style coaching inside a practice session.
- `Examiner`: judges an exercise answer and returns structured JSON state facts.

Do not create a new agent when one of these boundaries can be extended cleanly. Prefer extending prompt input, tools, or service mapping first.

## Prompt Rules

- Keep static system instructions in `server/infrastructure/llm/prompts/*.go`, not in `agent.go`.
- Use exported prompt render functions from agent constructors, for example `prompts.CoachPrompt(prompts.Input{})`.
- Put runtime-context rendering behind typed prompt input in the prompt package. Service code may map database models into prompt structs, but should not own long Markdown rendering.
- Preserve JSON-only constraints for structured agents and the Coach `<ACTION>` contract unless the caller and tests are updated together.
- Keep `preset_key` lightweight. It is an extension point for future prompt variants, not a user-facing preset product by itself.

## Constructor Rules

- Keep constructors in `server/infrastructure/llm/agent.go`.
- Keep existing constructor names and signatures stable unless all callers are updated:
  - `NewGuideAgent(ctx)`
  - `NewObjectiveGeneratorAgent(ctx)`
  - `NewCoachAgent(ctx, tools)`
  - `NewTutorAgent(ctx, tools)`
  - `NewExaminerAgent(ctx, tools)`
- Use `NewStructuredChatModel` for JSON-only agents (`Guide`, `ObjectiveGenerator`).
- Use `NewChatModel` for conversational/tool agents (`LearningCoach`, `Tutor`, `Examiner`).
- Do not change model selection, tool wiring, or streaming behavior as part of a prompt-only refactor.

## Tool Rules

- Add Coach discovery/navigation tools in `server/app/learning/tools/coach_tools.go`.
- Add Tutor practice-session tools in `server/app/learning/tools/learning_tools.go`.
- Keep tool input/output structs near the tool factory that owns them.
- If a prompt tells an agent to call a tool, verify that the exact tool name and output fields exist.
- For Coach navigation, keep `create_learning_objectives` and `first_objective_id` aligned with `BuildCoachQuery`, `ParseCoachAction`, and the frontend action consumer.

## SSE Rules

- Treat `server/app/learning/handlers/sse_events.go` as the current event-name source of truth.
- Do not introduce new SSE event types without checking the frontend `LearningStreamEvent` consumer.
- If Coach still emits `<ACTION>`, keep final action parsing in `server/app/learning/service/coach.go` and final event emission in `server/app/learning/handlers/coach.go`.

## Test Checklist

Run focused tests for the surface you touch:

```bash
cd server
go test ./infrastructure/llm/... ./app/learning/service
```

Also run handler tests when changing SSE, stream parsing, or handler flow:

```bash
cd server
go test ./app/learning/handlers
```

Add or update tests in the nearest package:

- Prompt rendering: `server/infrastructure/llm/prompts/prompts_test.go`
- Coach query and action parsing: `server/app/learning/service/coach_test.go`
- Guide parsing/query behavior: `server/app/learning/service/guide_test.go`
- Objective generation parsing: `server/app/learning/service/objective_generation_test.go`
- Examiner parsing/state behavior: `server/app/learning/service/examiner_test.go`
- SSE stream behavior: `server/app/learning/handlers/sse_events_test.go`

## Stop Conditions

Pause and clarify before:

- adding preset UI, onboarding cards, settings pages, database preset fields, or migrations;
- implementing Eino workflow, callbacks, interrupt/resume, or workflow-as-tool;
- changing frontend routes or learning page layout;
- creating another project skill beyond `eino-agents`.
