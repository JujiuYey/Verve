# 04 · 学习闭环与当前 Agent 边界

> 本篇定义 Verve 学习功能的产品闭环和当前真实 Agent 边界。
> 后续 Eino 改进路线见：[05-Eino 改进落地路线](./05-Eino改进落地路线.md)。

## 1. 核心原则

Verve 的学习功能不是“AI 一路主动上课”，而是围绕费曼学习法建立一个可验证的闭环：

1. 用户先读资料；
2. 用户尝试用自己的话讲清楚一个小目标；
3. 系统判断用户是否真正理解；
4. 讲错、讲漏、讲不出来时，系统才补讲；
5. 补讲后回到用户再次输出；
6. 通过后写入学习状态、学习记忆和学习日志。

这个闭环有两个底线：

- 系统不能一上来替用户讲完整课程；
- agent 的存在是为了推进学习事实，不是为了制造多角色聊天感。

## 2. 当前真实 Agent

当前 Verve 后端的 Eino ADK agent 集中在 `server/infrastructure/llm/agent.go`，真实边界如下：

| Agent | 当前职责 | 当前状态 |
|---|---|---|
| `Guide` | 阅读 Markdown 资料，生成导学摘要、掌握目标、复述小点和易错点 | 独立 `ChatModelAgent`，结构化输出 |
| `ObjectiveGenerator` | 把 Wiki Markdown 拆成学习小节 | 独立 `ChatModelAgent`，结构化输出 |
| `LearningCoach` | 根据 Wiki 资料、学习小节、学习记忆、legacy 画像和学习记录决定下一步 | 独立 `ChatModelAgent`，带 Coach tools，输出可选 `<ACTION>` |
| `Tutor` | 用户没讲清楚时做费曼式补讲和小练习 | 独立 `ChatModelAgent`，带 learning tools |
| `Examiner` | 根据用户作答判断 `pass / partial / fail`，产出学习事实 | 独立 `ChatModelAgent`，当前调用处可能不传 tools |

这五个 agent 是当前 source of truth。后续规划不能默认存在 `Learning Planner`、`Feynman Evaluator`、`Teaching Agent`、`Curriculum Reviewer`、`Learning Orchestrator` 这些独立角色，除非新的 workflow contract 明确把它们重新定义并落到代码边界。

## 3. 当前学习入口

### 3.1 `/learn/feynman`

`/learn/feynman` 是学习调度入口。

用户可以只说“继续学习”。后端 `LearningCoach` 应该根据真实上下文决定下一步：

- 查询 Wiki 文件夹和文档；
- 查询已有学习小节；
- 查询学习记忆，必要时参考 legacy 学习画像；
- 查询最近学习记录；
- 必要时从文档生成学习小节；
- 能确定小节时输出 `navigate_to_practice` action。

这里的重点是“让 agent 查上下文再决策”，不是前端硬编码下一步。

### 3.2 `/learn/feynman-practice/$objectiveId`

`/learn/feynman-practice/$objectiveId` 是具体小节的学习工作台。

当前工作台仍然以页面 phase 承载流程：

- `reading`：阅读源 Markdown；
- `answering`：用户复述；
- `teaching`：失败后请求 Tutor 补讲；
- 通过后写入学习记录和学习记忆；
- Tutor 生成的学习旁注可以追加回 Markdown。

这里的 phase 是当前实现，不等于最终 server-side workflow。未来是否把这些状态迁移到后端，要由 workflow contract 判断。

## 4. 当前闭环

当前最小闭环可以理解成：

```text
Wiki Markdown
    ↓
ObjectiveGenerator 生成学习小节
    ↓
Guide 生成导学内容
    ↓
用户阅读资料
    ↓
用户复述
    ↓
Examiner 判定 pass / partial / fail
    ↓
pass: 写入学习状态、学习记忆和学习日志
partial / fail: Tutor 针对漏洞补讲
    ↓
用户重新复述
```

`LearningCoach` 位于这个闭环外侧，负责决定用户下一步应该进入哪个小节或是否需要先生成学习小节。它应该优先参考 `learning_memory_items` 这类证据型学习记忆；`learning_profiles` 只作为兼容旧流程的投影，不再是用户画像的产品 source of truth。

## 5. 当前不成立的角色假设

以下概念可以作为未来讨论素材，但不是当前架构事实：

| 概念 | 当前处理方式 |
|---|---|
| `Learning Planner` | 当前由 `ObjectiveGenerator` 承担“从资料拆学习小节”的一部分能力，不是完整规划 agent |
| `Feynman Evaluator` | 当前由 `Examiner` 承担判定职责 |
| `Teaching Agent` | 当前由 `Tutor` 承担补讲职责 |
| `Curriculum Reviewer` | 当前没有独立实现 |
| `Learning Orchestrator` | 当前没有独立 server-side workflow；一部分流程仍在前端 phase 和 service/handler 中 |

不要为了匹配这些旧角色名而新增 agent。新增 agent 的前提应该是：现有 `Guide` / `ObjectiveGenerator` / `LearningCoach` / `Tutor` / `Examiner` 无法自然扩展，并且新的边界能降低复杂度。

## 6. 下一步架构方向

当前不是直接进入大一统 orchestrator，而是按下面顺序收敛：

1. 拆分 prompt，让 agent 文案可 review、可参数化；
2. 把 Coach 上下文改成 typed input，减少 service 层字符串拼接；
3. 沉淀 `.agents/skills/eino-agents`，让后续 agent 改动有项目本地规范；
4. 写 workflow contract，判断哪些 phase 应该后端化；
5. contract 证明有价值后，再实现 server-side workflow。

这个方向的目标不是增加 agent 数量，而是让现有学习闭环更稳定：上下文更清楚、状态更可恢复、agent 交接更可追踪。

## 7. 非目标

- 不默认新增一组理想化 agent 角色；
- 不做多个 agent 同时面向用户聊天；
- 不把费曼学习做成普通问答；
- 不把 Tutor 变成默认讲课入口；
- 不在没有 workflow contract 的情况下直接实现 Orchestrator；
- 不把用户侧 preset、interrupt/resume、workflow-as-tool 提前塞进第一阶段。
