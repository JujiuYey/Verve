# Project Rules for AI Agents

## Frontend Test Policy

For frontend work under `web/`, do not add tests by default.

Only write frontend tests when the change touches one of these areas:

- Reusable global UI components shared across the app.
- Logic-heavy hooks, utilities, state stores, or data transformation code.
- Behavior that is difficult to verify manually and has meaningful regression risk.

Do not write tests for ordinary page layout, feature-specific sidebar/menu composition, visual arrangement, or route-level UI wiring unless the user explicitly asks for them. These tests usually add maintenance cost without proving useful product behavior.

## Frontend Visual Verification

Do not use screenshots, browser automation, or visual self-check loops for frontend UI work unless the user explicitly asks for them.

Agents should focus on implementing the UI changes cleanly, then report what changed and ask the user to inspect the interface directly. The user is responsible for visual acceptance and design direction.

When finishing frontend UI work, include the local URL or relevant route if available, and use concise language such as: "请你在页面里验收一下效果，我根据你的反馈继续改。"

## Frontend Component Sources

For frontend UI work under `web/`, use the project's established component sources instead of hand-rolling custom markup.

Prefer components in this order:

- shadcn/ui components installed in `web/src/components/ui`.
- shadcn-related project skill, CLI, or MCP documentation when adding, composing, or checking component APIs.
- Project-specific reusable components in `web/src/components/sag-ui`.

Before building a custom component, check whether shadcn/ui or `sag-ui` already provides the needed behavior. If a required component is missing, do not silently invent a one-off replacement. Tell the user what is missing, where you checked, and whether you recommend adding a shadcn component, reusing/extending `sag-ui`, or implementing a small local component.

## Go Model Field Documentation

For Go structs under any `models/db` or `models/payload` directory, every business field must have a clear field-level comment.

- Write concise Chinese trailing comments in the same style as the existing model files.
- Describe the field's business meaning; include enum values or units only when necessary.
- Keep comments synchronized whenever fields are added, renamed, repurposed, or removed.
- Prefer short comments such as “用户ID”“会话状态”“输入 token 数”; do not write paragraph-style field documentation.
