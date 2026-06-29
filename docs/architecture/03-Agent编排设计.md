# 03 · Agent 编排设计

> 基于 **eino `adk`**(现有 `app/ai/model/agent.go`、`chat.go` 已在用)。adk 提供 `NewChatModelAgent`、`NewRunner(EnableStreaming:true)`,以及多 agent `transfer` 机制。

## 四个 agent

| Agent | 角色 | 主要工具 | 输出 |
|---|---|---|---|
| **Planner** | 把目标拆成学习路线 | 写 path + objectives | 结构化路线 |
| **Tutor** | 费曼式陪练讲解 | 读 objective / profile | 对话 + 练习 |
| **Examiner** | 判定掌握、维护状态 | 写 exercise / profile / journal、更新 mastery | verdict + 反馈 |
| **Orchestrator** | 调度何时调谁 | — | 流程控制 |

## 关键取舍:MVP 用 service 显式编排,不靠 LLM supervisor

adk 支持两种编排:

- **A. supervisor + transfer**:Orchestrator 当根 agent,由 LLM 决策 `transfer` 到子 agent。
- **B. service 显式编排**:后端代码按确定流程依次调用各 agent 的独立 Runner。

**MVP 选 B**,理由:

- MVP 流程是**确定的**(生成路线 → 诊断 → 讲解 → 练习 → 验证 → 小结),用代码编排比让 LLM 当 supervisor **更可控、更省 token、更易调试**;
- Orchestrator 退化为 **service 层状态机**(`orchestrator.go`),按 session 的 `phase` 决定调 Tutor 还是 Examiner;
- adk 的 `transfer` / supervisor 留作**后续**(开放主题、自由对话时再启用)。

```text
orchestrator.go(service 状态机)
  phase = diagnose / explain / practice  → 调 Tutor Runner(流式)
  phase = verify(收到 exercise 提交)     → 调 Examiner(判定)
  本节结束                                → Examiner 写 profile / journal,推进 objective
```

## agent 定义(对齐现有 `model/agent.go` 写法)

```go
// 每个 agent 一个构造器,复用现有 NewChatModel(modelRepo) 拿 chatModel
adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "Tutor",
    Description: "费曼式陪练",
    Instruction: tutorInstruction,   // system prompt,见下
    Model:       chatModel,
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{ Tools: learningTutorTools },
    },
})
// Runner:adk.NewRunner(ctx, adk.RunnerConfig{EnableStreaming:true, Agent:a})
```

## 各 agent 的 Instruction 要点(system prompt)

- **Planner**:输入"学习目标 + 用户背景";输出阶段化路线(JSON);每个小目标可在**一节课内**完成;标注常见卡点;简洁不啰嗦。
- **Tutor**(教学铁律):每次只推进**一个**认知点;先诊断后讲解;不连续追问;讲完给一个小练习并要求验证;用提问帮用户把概念讲出来;**不空泛鼓励**。
- **Examiner**:对照掌握分级(`none→…→verified`)判定;区分"能讲 / 能写 / 能验证";给具体不敷衍的反馈;输出 `verdict + mastery_after`;经工具把事实写入 profile / journal。

## 工具(eino `utils.InferTool`,对齐现有 `tools/system.go`)

learning 工具(注入 repository,内部调 `FindOne/Create/Update`):

| 工具 | 作用 |
|---|---|
| `get_objective` / `get_profile` | 读当前上下文 |
| `record_exercise` | 写 `learning_exercises` |
| `update_mastery` | 更新 `learning_objectives.mastery_level` |
| `upsert_profile` | 写 `learning_profiles` |
| `write_journal` | 写 `learning_journals` |

```go
utils.InferTool("update_mastery", "更新小目标掌握层级",
  func(ctx context.Context, in *UpdateMasteryInput) (*UpdateMasteryOutput, error) {
     // repo.Update(...)
  })
```

## 上下文注入与流式

- 每次构造 Runner 前,service 拼接:当前 objective、最近 profile(掌握 / 薄弱点)、本 session 历史 messages(**裁剪长度**控 token)。
- 流式:`Runner.Query → AsyncIterator`,handler 转 SSE(同 `chat.go`),扩展 `phase` / `exercise` 事件。
- 落库:每条 user/assistant 消息写 `learning_messages`(`agent_type`、token,token 取自 eino usage)。

## 成本控制(对齐 D2)

MVP 系统预置 key;裁剪历史上下文;Examiner 判定尽量一次完成;记录 token 备后续计量 / 限额。
