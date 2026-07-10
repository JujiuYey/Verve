# 07 · Eino 学习 Workflow 契约

> 配套文档：
> [04-学习闭环与当前 Agent 边界](./04-学习闭环与当前Agent边界.md)，
> [05-Eino 改进落地路线](./05-Eino改进落地路线.md)。
>
> 本篇只定义未来 server-side workflow 的契约和进入条件，不实现 workflow。

## 1. 结论

当前不要马上实现 Eino workflow。

现在的 Verve 已经有一个可跑的费曼闭环，但它仍然由前端 phase、HTTP 接口、SSE 流和多个 service/handler 拼起来：

- `/learn/feynman-practice/$objectiveId` 的页面态是 `reading / answering / teaching`；
- `SessionHandler.Create` 创建练习会话；
- `SessionHandler.Exercise` 调用 Examiner 判定并写学习状态；
- `SessionHandler.Chat` 调用 Tutor 做补讲并通过 SSE 输出；
- 前端自己决定何时进入 teaching、何时把 Tutor 旁注追加到 Markdown。

如果现在直接实现 workflow，很容易只是把前端 phase 搬到后端，收益不够。workflow 只有在能解决下面这些问题时才值得做：

1. 页面刷新后能恢复到真实学习进度；
2. Examiner / Tutor 失败后能安全重试；
3. Tutor 与 Examiner 的交接有明确状态；
4. pass 后学习状态、学习画像、学习日志写回保持一致；
5. Markdown 旁注写回有用户确认边界，不由 agent 静默改教材。

## 2. 推荐名称

推荐未来 workflow 命名为 `FeynmanSessionWorkflow`。

不推荐 `RecoveryWorkflow`，因为“补救”只是 partial / fail 后的一段分支，不是完整会话边界。

不推荐 `LearningSessionWorkflow`，因为这个名字太泛，会把 Coach 调度、目标生成、阅读、复述、判定、补讲、复习都卷进来。

`FeynmanSessionWorkflow` 的边界更清楚：它只负责一个 objective 的一次费曼练习会话。

## 3. 入口

第一版 workflow 只应该从练习工作台进入：

```text
/learn/feynman-practice/$objectiveId
    ↓
POST /api/learning/session
    ↓
FeynmanSessionWorkflow(objective_id, user_id)
```

Coach 不应该直接拥有这个 workflow。Coach 的职责仍然是外侧调度：

```text
LearningCoach
    ↓ navigate_to_practice action
/learn/feynman-practice/$objectiveId
    ↓
FeynmanSessionWorkflow
```

如果未来 Coach tool 需要创建练习会话，它也只能创建 session 或返回 `objective_id`，不能替用户跳过阅读和首次复述。

## 4. 输入

workflow 启动输入：

| 字段 | 来源 | 说明 |
|---|---|---|
| `user_id` | auth context | 权限和状态写回边界 |
| `objective_id` | route / Coach action | 本次只练一个学习小节 |
| `session_id` | `learning_sessions` | 已存在则恢复；不存在则创建 |
| `source_document_id` | `learning_objectives` | 可选，用于读取源 Markdown |
| `source_folder_id` | `learning_objectives` | 可选，用于学习画像和日志 |

用户动作输入：

| 动作 | 输入 | 说明 |
|---|---|---|
| `read_started` | `session_id` | 用户进入阅读，不代表已掌握 |
| `submit_answer` | `answer`, `prompt_type` | 触发 Examiner |
| `request_teaching` | `examiner_result_id` 或最近判定 | 触发 Tutor |
| `retry_answer` | `answer` | 补讲后再次复述 |
| `confirm_note_writeback` | `note`, `document_id` | 用户确认后才写 Markdown |

## 5. 输出

workflow 输出分两类：前端事件和持久化事实。

前端事件通过 SSE 传给页面：

| 输出 | 用途 |
|---|---|
| `phase_changed` | 告诉前端当前 server-owned phase |
| `reasoning` | 展示 agent 思考摘要或模型 reasoning |
| `stream_chunk` | Tutor 可见正文 |
| `tool_call` / `tool_result` | 工具调用透明度 |
| `exercise` | Examiner 判定结果 |
| `note_proposed` | Tutor 生成的 Markdown 旁注草稿 |
| `error` | 可恢复或不可恢复错误 |

持久化事实写入数据库或文档：

| 输出 | 当前对应 |
|---|---|
| `learning_sessions.status` | active / completed / abandoned |
| `learning_messages` | Tutor 对话和 assistant 输出 |
| `learning_exercises` | 用户作答与 Examiner verdict |
| `learning_objectives.mastery_level` | Examiner 判定后的掌握层级 |
| `learning_profiles` | 已掌握内容、薄弱点、下一目标 |
| `learning_journals` | 当日学习事实和本次改进建议 |
| Wiki Markdown | 用户确认后的学习旁注 |

## 6. 状态机

未来 server-owned 状态不应该直接等同于前端三个 phase。建议状态如下：

| 状态 | 含义 | 前端展示 |
|---|---|---|
| `reading` | 用户阅读源资料和 Guide | 阅读页 |
| `answering` | 等待用户首次复述 | 复述页 |
| `evaluating` | Examiner 正在判定 | 复述页 loading |
| `passed` | 本轮通过，学习事实已写回 | 结果面板 |
| `teaching` | Tutor 正在补讲 | 教学页 streaming |
| `retrying` | 补讲后等待用户再次复述 | 复述页 |
| `note_reviewing` | 等待用户确认旁注写回 | 教学页 |
| `completed` | 本次 session 完整结束 | 可返回 Coach 或继续下一节 |
| `failed` | workflow 出现不可自动恢复错误 | 错误态 |

允许的主路径：

```text
reading
  → answering
  → evaluating
  → passed
  → completed
```

允许的补讲路径：

```text
evaluating
  → teaching
  → retrying
  → evaluating
```

允许的旁注路径：

```text
teaching
  → note_reviewing
  → retrying
```

关键约束：

- Tutor 只能在 `partial / fail` 后进入，不能默认先讲课；
- `passed` 后才能写“已掌握”类状态；
- `note_reviewing` 不能自动写 Markdown，必须等用户确认；
- `reading` 只是展示源资料，不应该写 mastery。

## 7. SSE 契约

当前 SSE 事件类型定义在 `server/app/learning/handlers/sse_events.go`：

```text
reasoning
stream_chunk
message
tool_call
tool_result
exercise
action
error
```

第一版 workflow 可以复用这些事件，不急着新增事件类型。

建议做法：

- 状态变化先用 `message` 或 `exercise` payload 里的 `phase` 字段表达；
- Tutor 输出继续使用 `stream_chunk`；
- Examiner 判定继续使用 `exercise`；
- 工具调用继续使用 `tool_call` / `tool_result`；
- 不把 Coach 的 `action` 语义混进练习 workflow，`action` 仍然主要用于页面跳转。

只有当前端需要稳定订阅 server-owned phase 时，才新增：

```text
workflow_state
```

新增前必须同步修改：

- `server/app/learning/handlers/sse_events.go`;
- `web/src/api/learning/session.ts` 的 `LearningStreamEvent`;
- `web/src/pages/learning/feynman-practice/*` 的消费逻辑。

## 8. 恢复规则

workflow 必须能回答“刷新页面后恢复到哪里”。

恢复依据优先级：

1. `learning_sessions.status`;
2. 最近一条 `learning_exercises` 的 verdict 和 mastery_after；
3. 最近 Tutor assistant message；
4. 是否存在未确认的 note proposal；
5. objective 当前 mastery_level；
6. source document 是否仍存在。

恢复策略：

| 条件 | 恢复到 |
|---|---|
| session active，尚无 exercise | `reading` 或 `answering` |
| 最近 verdict 是 pass | `passed` 或 `completed` |
| 最近 verdict 是 partial / fail，已有 Tutor 输出 | `retrying` |
| 最近 verdict 是 partial / fail，无 Tutor 输出 | `teaching` |
| Tutor stream 中断 | `teaching`，允许重试 |
| Markdown 旁注写回失败 | `note_reviewing`，保留草稿 |
| objective / document 不存在 | `failed`，提示回到 Coach |

如果这些恢复规则没有对应持久化字段，先不要实现 workflow。

## 9. 写回边界

### 9.1 学习状态写回

Examiner 判定通过或部分通过后，可以写：

- `learning_exercises`;
- `learning_objectives.mastery_level`;
- `learning_profiles`;
- `learning_journals`。

写回应由后端完成，不能依赖前端自行拼状态。

### 9.2 Markdown 旁注写回

Markdown 是学习资料本体，不是 agent 草稿缓存。Tutor 可以生成旁注草稿，但不能静默写回。

正确边界：

1. Tutor 生成 `note_proposed`；
2. 前端展示给用户；
3. 用户点击确认；
4. 后端或现有 document API 追加到 Markdown；
5. 写回成功后刷新源文档缓存。

当前前端已经在 `appendTutorNoteToMarkdown` 中直接调用 document API。workflow 实现前可以保持现状。未来如果后端接管，必须保留用户确认动作。

## 10. 与当前前端 phase 的关系

当前前端类型是：

```ts
type WorkbenchPhase = "reading" | "answering" | "teaching";
```

这些 phase 目前是 UI 展示状态，不是权威学习状态。

未来迁移规则：

| 当前前端 phase | 是否应后端化 | 说明 |
|---|---|---|
| `reading` | 可以后端记录，但不急 | 阅读开始/完成本身不是 mastery |
| `answering` | 可以由 workflow 恢复 | 需要恢复用户是否正在复述 |
| `teaching` | 应后端化 | Tutor streaming、失败重试、旁注草稿都需要恢复 |

前端仍然可以保留 display phase，但权威状态应来自 workflow snapshot。

## 11. 实现进入条件

满足以下条件后，才开始写 workflow 实现计划：

1. P0 prompt 拆分已完成；
2. P1 Coach typed prompt input 已完成；
3. `.agents/skills/eino-agents` 已可用；
4. 本文 contract 被确认；
5. 已决定是否新增 `workflow_state` SSE 事件；
6. 已决定新增哪些持久化字段或表；
7. 已决定 Markdown 旁注由哪个 API 写回；
8. 已定义页面刷新后的恢复测试场景。

如果第 5-8 条没有答案，不要实现 Eino workflow。

## 12. 第一版实现建议

如果将来开始实现，第一版只做一个很小的切片：

- 不改 Coach；
- 不改 ObjectiveGenerator；
- 不做 interrupt/resume；
- 不做 workflow-as-tool；
- 不做多 agent 编排 UI；
- 只接管 Feynman practice session 的 `answer → evaluate → teach → retry`。

最小实现应该先证明：

- Tutor stream 中断后可以恢复；
- partial / fail 后可以重试；
- pass 后状态写回一致；
- Markdown 旁注不会未经确认写回。

做到这些，workflow 才比当前 handler + 前端 phase 有真实价值。

## 13. 当前非目标

- 不实现 workflow 代码；
- 不新增数据库迁移；
- 不新增 SSE event；
- 不改前端页面；
- 不改变 Coach `navigate_to_practice` action；
- 不把 `LearningCoach` 升级成总 orchestrator；
- 不新增 `Learning Planner`、`Teaching Agent`、`Curriculum Reviewer` 等旧概念角色。
