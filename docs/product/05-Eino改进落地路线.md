# 05 · Eino 改进落地路线

> 配套文档：[04-学习闭环与当前 Agent 边界](./04-学习闭环与当前Agent边界.md)。
> 本篇是 Eino 改进的唯一落地路线：先把 Eino 用法推进到可维护、可配置、可编排，再逐步释放 workflow / preset / callbacks 这些能力。

## 1. 总体判断

Verve 当前已经开始使用 Eino ADK，但还停留在“多个独立 `ChatModelAgent` + 手写 prompt + 手写 SSE + 前端 phase 驱动”的阶段。继续把 Eino 当成普通 LLM 调用封装，会让 agent 越加越多、prompt 越写越散，最后很难判断一次学习会话到底由谁负责。

这条路线要避免把三类问题混成一团：

| 类型 | 内容 | 是否应该进入第一阶段 |
|---|---|---|
| 架构卫生 | prompt 拆分、结构化上下文、callbacks、workflow | 应该 |
| 产品功能 | 学习风格 preset、onboarding、设置页 | 保留 `preset_key`，暂缓完整 UI |
| 开发治理 | `.agents/skills/`、skill registry、多工具分发 | 只做最小沉淀 |

第一阶段不要追求“Eino 能力全用上”，而是把现有学习闭环推进到三个状态：prompt 可 review、上下文可结构化、后续 workflow 有明确入口。

## 2. 第一原则

### 2.1 把 preset 当成扩展点，而不是完整产品功能

学习风格 preset 是 prompt 结构化的合理动机。它说明为什么 prompt 不应该继续散落在 `const string` 里：后续同一个 Tutor / Guide / Examiner 可能要根据学习风格生成略有差异的指令。

但这里要区分两件事：

- `preset_key` 扩展点：第一阶段就可以保留，成本很低；
- 完整 preset 产品：onboarding、设置页、5 种风格卡片、切换记录、效果分析，第一阶段暂缓。

如果过早做完整 preset 产品，问题会变成：

- 需要新增数据模型和 UI；
- 需要解释每种风格的真实差异；
- 需要证明用户真的会主动选择；
- 需要处理切换风格后的历史会话语义；
- 需要评估“风格”到底影响 Tutor、Guide、Examiner 还是整个 Orchestrator。

这些都是产品问题，不应该堵住 prompt 结构化。但 prompt 结构化应该为它们留门。

### 2.2 先让现有 agent 交接变清楚，再上 workflow

当前学习闭环里还没有稳定的 server-side workflow state。这里提到的 `RecoveryWorkflow` 不应该直接理解成“马上把 Tutor + Examiner 串起来”。

更准确的第一步是写清楚 workflow contract：

- 输入是什么；
- 输出是什么；
- 哪些状态落库；
- 哪些事件通过 SSE 发给前端；
- 失败后如何恢复；
- 页面刷新后如何继续；
- 哪些阶段仍然只是 UI 展示。

没有 contract 就直接上 Eino workflow，容易把现在的前端 state 问题搬到后端。更好的路线是：先把 prompt 和上下文边界理顺，再写 workflow contract，最后实现 workflow。

### 2.3 保留用户输出优先的学习逻辑

Verve 的费曼学习定义不是“AI 主动上课”，而是：

1. 用户先读资料；
2. 用户尝试复述；
3. 系统判断；
4. 失败或不完整时才补讲；
5. 补讲后回到用户再次输出；
6. 通过后写入学习状态和学习日志。

任何 Eino 改造都不能把这个闭环改成“Coach 直接决定并讲课”。Eino workflow 的价值是让这个闭环更可靠：状态可恢复、agent 交接可追踪、学习事实可沉淀。

## 3. 推荐落地路线

### P0：Prompt 拆分 + 简单参数化

目标：把现在集中在 `server/infrastructure/llm/agent.go` 的 instruction 拆出来，让 agent 创建代码变薄，同时建立最小参数化能力。P0 不改变产品行为，但要为 preset 和上下文结构化开好口子。

建议新增：

```text
server/infrastructure/llm/prompts/
  guide.go
  objective_generator.go
  coach.go
  tutor.go
  examiner.go
  preset.go
```

第一阶段做四件事：

- 每个 agent 的 instruction 从 `agent.go` 移到独立文件；
- 给 prompt 增加 `Version` / `Name` / `PresetKey` 这类元信息；
- 每个 prompt 暴露简单函数，例如 `TutorInstruction(input TutorPromptInput) string`；
- agent 创建处仍然拿到最终 string，不引入重型模板系统。

这里的“简单参数化”不是完整 Eino `ChatTemplate` 改造，也不是产品 preset 系统。它只是把未来会变化的东西显式变成输入：

```go
type TutorPromptInput struct {
    PresetKey string
}

type GuidePromptInput struct {
    PresetKey string
}
```

如果某个 agent 暂时没有变量，也可以先只接收 `PresetKey` 或使用默认 input。重点是从第一天起避免回到裸 `const string`。

不要在 P0 做：

- 不做 onboarding；
- 不做设置页；
- 不做 5 种学习风格；
- 不改数据库；
- 不重写 SSE；
- 不引入完整 workflow；
- 不引入复杂模板 DSL。

P0 验收标准：

- `agent.go` 只负责创建 agent，不再承载大段 prompt 文案；
- 现有 Guide / ObjectiveGenerator / Coach / Tutor / Examiner 行为保持一致；
- prompt 文件能被单独 review；
- prompt 函数已经有 `preset_key` 扩展点；
- 后续要加真实 preset 时，不需要再移动 prompt 文件。

### P1：把 Coach 的上下文拼接改成结构化输入层

当前 `BuildCoachQuery` 把 Wiki 文件夹、文档、学习小节、画像、日志拼成一段 Markdown。这能跑，但后续越加越难控制 token、字段和行为边界。

P1 不要求马上改成 Eino `ChatTemplate`。更稳的做法是先把 Coach 的业务上下文变成 prompt input：

```go
type CoachPromptInput struct {
    UserMessage string
    Folders     []FolderPromptItem
    Documents   []DocumentPromptItem
    Objectives  []ObjectivePromptItem
    Profiles    []ProfilePromptItem
    Journals    []JournalPromptItem
}
```

然后由 prompt 包负责把 `CoachPromptInput` 渲染成最终文本。这样 P0 的“简单参数化”就自然扩展到了 Coach：不是否定结构化，而是先用 Go 类型把结构立起来。

这样做的好处：

- service 层不再关心 prompt 文案；
- prompt 层能集中控制字段裁剪；
- 未来迁移到 Eino `ChatTemplate` 时，不需要再整理业务数据结构；
- tests 可以断言输入结构，而不是只断言一大段字符串里包含某些片段。

P1 验收标准：

- `BuildCoachQuery` 不再直接写大段业务 prompt；
- prompt 渲染逻辑有单元测试；
- 对空 folders、空 documents、已有 objectives、recent journal 这几类关键分支都有覆盖；
- Coach 输出 `<ACTION>` 的约束仍然保留。

### P2：只写一个开发侧 skill：`eino-agents`

05 里列了多个 skill：`eino-agents`、`learning-domain`、`sse-events`、`sag-ui`、`verify-frontend`。第一阶段不要拆这么多。

先写一个 `.agents/skills/eino-agents/SKILL.md`，覆盖最常见的开发动作：

- 新增 agent 要改哪些文件；
- prompt 放在哪里；
- tool 放在哪里；
- handler / service 如何调用 agent；
- SSE event 如何映射；
- 日志字段怎么写；
- 哪些情况下不能新增 agent，而应该扩展现有 agent。

暂缓：

- `learning-domain`：等领域模型稳定后再拆；
- `sse-events`：等 SSE contract 改动变多后再拆；
- `sag-ui`：属于前端组件治理，不应该混入 Eino 改造；
- `verify-frontend`：AGENTS.md 已经有约定，暂时不需要独立 skill。

P2 验收标准：

- 后续 agent 相关改动可以先读 `eino-agents`；
- skill 不描述理想架构，只描述当前 Verve 真实文件位置和改动顺序；
- skill 中明确“不要为了显得 agent-native 而新增 agent”。

### P3：设计 workflow contract，再实现 workflow

只有当前三步完成后，才进入 workflow。

不要一上来实现 `RecoveryWorkflow`。先新增一份 contract 文档，回答下面这些问题：

| 问题 | 必须明确 |
|---|---|
| workflow 名称 | 是 `RecoveryWorkflow`、`FeynmanSessionWorkflow`，还是 `LearningSessionWorkflow` |
| 入口 | 从 `/learn/feynman-practice/$objectiveId` 进入，还是 Coach tool 触发 |
| 输入 | objective、source markdown、user answer、session id、profile、journal |
| 输出 | verdict、teaching advice、next action、journal update、markdown note |
| 状态 | reading / answering / evaluating / teaching / retrying / passed |
| SSE | 每个状态变化对应什么 event |
| 恢复 | 刷新页面后根据什么恢复 |
| 写回 | Markdown 旁注由谁决定，是否需要用户确认 |

如果 contract 写完发现 workflow 只是把现有前端 phase 挪到后端，那就先停住。只有当它能解决“刷新恢复、失败重试、agent 交接、状态一致性”这些问题时，才值得上。

## 4. 对 05 中几个具体点的修正

### 4.1 “Tools 分布不均”要写得更准确

当前问题不是简单的“某些 agent 没有 tools”。更准确地说：

- Coach / Tutor 已经通过 tools 读取或操作部分学习上下文；
- Examiner 的构造函数支持 tools，但当前调用时传的是 `nil`；
- Guide / ObjectiveGenerator 当前没有 tools 参数；
- 大量上下文仍然靠 service 层拼进 query，而不是由 agent 按需查询。

所以第一阶段应该先统一“上下文进入 agent 的方式”，而不是盲目给每个 agent 都加 tool。

### 4.2 “SSE 是手搓的”不是马上要替换

当前 SSE 虽然是自定义事件，但已经承载了前端需要的 `reasoning`、`stream_chunk`、`tool_call`、`tool_result`、`action`、`error`。短期不要为了“更 Eino”而重写。

更现实的目标是：

- 固化事件类型；
- 给 workflow 新增状态事件；
- 避免前端解析隐藏约定；
- 保持 `ai-elements` 能稳定消费流式内容。

### 4.3 `compose.WithCallbacks` 可以先做轻量版

callbacks 的第一阶段目标不是完整 observability 平台，而是替代分散的 `log.Printf`。

先记录这些字段就够：

- agent name；
- session id；
- objective id；
- user id；
- run id；
- started / finished / failed；
- token 或 output chars；
- tool call name；
- error。

如果一开始就接 trace、metrics、dashboard，范围会变大。

## 5. 一个更小的第一张任务卡

如果要马上开始动手，第一张任务卡建议这样写：

### 标题

拆分 Eino agent prompts，并建立最小参数化入口

### 范围

- 从 `server/infrastructure/llm/agent.go` 移出 5 段 instruction；
- 新增 `server/infrastructure/llm/prompts/` 包；
- 每个 prompt 暴露一个函数和 input 结构；
- input 中至少保留 `preset_key` 或对应默认值；
- `agent.go` 继续创建原来的 5 个 `ChatModelAgent`；
- 不改变 handler、service、SSE、数据库、前端。

### 不做

- 不新增学习风格 UI；
- 不新增数据库字段；
- 不实现 workflow；
- 不接 interrupt/resume；
- 不改 Coach 行为；
- 不改 Guide / Examiner 的 tool 能力。

### 验收

- 后端测试通过；
- `agent.go` 不再包含长 prompt；
- prompt 文案迁移后内容等价；
- 新增 prompt 包的命名和 input 能支撑后续 `preset_key`；
- diff 能被单独 review，不和产品功能混在一起。

## 6. 后续能力的进入条件

### 6.1 用户侧 preset 什么时候进入

当 P0 的 prompt input 已经稳定、至少一个 agent 能根据 `preset_key` 改变指令时，再考虑用户侧 preset 产品。

进入条件：

- 已经能证明两个 preset 在 Tutor 或 Guide 上有清晰行为差异；
- 差异不是换文案，而是影响提问方式、例子密度、讲解顺序或验证标准；
- 默认 preset 可以覆盖大多数用户，不选择也能正常学习；
- 切换 preset 只影响新会话，不回写历史判断。

### 6.2 Interrupt/Resume 将来怎么进入

Eino 支持 interrupt/resume，但 Verve 不应该把它当成一个独立按钮功能。它适合在 workflow state 稳定后进入，用来表达“agent 等待用户输入，并能从 checkpoint 恢复”。

将来做它时，至少要先有：

- checkpoint id 存储；
- session 状态落库；
- 前端知道当前等待谁输入；
- SSE 能表达 interrupted；
- resume API 能携带用户输入；
- 失败后能重新进入同一个 checkpoint。

所以 interrupt/resume 不进入第一张任务卡，但保留为 workflow 成熟后的自然下一步。

## 7. 结论

05 可以保留为方向文档，但实际执行时应按下面顺序收敛：

1. 先整理 prompt，并建立简单参数化入口；
2. 再整理上下文输入，让 Coach 从字符串拼接转向 typed input；
3. 再沉淀一个 `eino-agents` skill；
4. 再写 workflow contract；
5. 最后把用户侧 preset、interrupt/resume、workflow as tool 作为自然增强接进来。

这条路线不是为了少做，而是为了让每一步都把 Verve 往更 agent-native 的方向推一点：prompt 可配置、上下文有结构、会话可编排、用户学习闭环不变形。
