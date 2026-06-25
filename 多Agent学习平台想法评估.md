# 多 Agent 学习平台想法评估

日期：2026-06-22

## 1. 本轮讨论的核心判断

这个项目有搞头。

它不应该只是把现有 Go 教程做成一个网页，也不应该只是一个普通的 AI 聊天学习工具。更准确的方向是：

> 一个多 agent 驱动的个人学习平台。

它的核心价值不是“AI 能回答问题”，而是平台持续知道用户：

- 正在学什么；
- 已经学到哪一步；
- 哪些概念还不稳；
- 今天应该练什么；
- 学习结果是否通过代码、测试或解释验证；
- 下一步该进入哪个小目标。

这比单次问答更有产品价值。

## 2. 为什么 Next.js 合适

Next.js 适合作为这个产品的全栈应用层。

它可以负责：

- 页面和交互界面；
- 课程阅读页面；
- 学习路线展示；
- 聊天和陪练界面；
- 用户登录；
- 学习进度；
- 学习日志；
- API Route / Server Action；
- agent streaming 输出展示。

但 Next.js 本身不应该承担全部长期 agent runtime。

如果只是 MVP，可以先把 agent 调用写在 Next.js 的 API Route 里。但随着功能变复杂，尤其出现多 agent、长任务、状态恢复、代码运行、学习记录检查时，应该把 agent 执行层单独抽象出来。

推荐边界：

```text
Next.js App
- 前端页面
- 用户交互
- 课程阅读
- 学习进度
- 聊天流式展示
- 发起 agent run

Agent Runtime
- LangChain / LangGraph / Deep Agents
- 多 agent 编排
- 状态机
- checkpoint
- 工具调用
- human-in-the-loop

Database / Storage
- 用户画像
- 学习路线
- 课程内容
- 学习状态
- 每日学习日志
- agent run logs
- 练习和验证结果
```

## 3. 后端是否用 LangChain

可以用 LangChain，但要注意：

LangChain 不是完整后端，它是 agent 引擎或 agent 编排层的一部分。

真正的后端仍然需要：

- Next.js API / Server Actions；
- 数据库；
- 用户系统；
- 任务状态；
- agent run 记录；
- 文件或对象存储；
- 后续可能还需要队列 / worker。

更合适的表达是：

```text
后端应用层：Next.js
agent 编排层：LangChain / LangGraph / Deep Agents
数据层：Postgres + ORM
执行层：worker / sandbox / queue
```

对于这个项目，LangGraph 或 Deep Agents 可能比裸 LangChain 更适合，因为项目天然需要：

- 多 agent 分工；
- 长任务状态；
- checkpoint；
- 学习过程记录；
- agent 之间交接；
- 文件 / memory / human approval。

## 4. 多 agent 的初步分工

目前设想的三类 agent 是合理的。

### 4.1 学科专家 agent

例如：

- Go 专家；
- Rust 专家；
- Python 专家；
- Web 后端专家；
- 数据库专家。

职责：

- 制定学习路线；
- 判断学习顺序；
- 拆分阶段目标；
- 识别该学科的常见卡点；
- 为用户生成当前阶段最合适的小节目标。

示例：

```text
用户：我要学 Rust。
Rust 专家 agent：
- 判断用户背景；
- 生成 Rust 入门路线；
- 安排 ownership、borrowing、lifetime 的学习顺序；
- 给出第一阶段目标。
```

### 4.2 费曼教学 agent

这是当前 Go 仓库里最有价值的教学方式，可以迁移到平台里。

职责：

- 每次只推进一个认知点；
- 先诊断，再讲解；
- 不连续追问；
- 讲完给一个小练习；
- 要求用户用命令或输出验证；
- 帮用户把概念讲出来。

稳定教学循环：

```text
读取学习状态
→ 选择一个小目标
→ 一个热身诊断问题
→ 短讲解
→ 小练习
→ 一个验证命令
→ 反思
→ 生成学习日志草稿
```

### 4.3 学习结果检查 / 写入 agent

这个 agent 不应该只是“总结对话”。

它应该成为平台的学习状态维护者。

职责：

- 判断用户今天是否真的掌握；
- 区分“看过”“听过”“能讲出来”“能写出来”“能验证”；
- 把学习事实写入 profile；
- 把每日学习记录写入 journal；
- 更新薄弱点；
- 推荐下一步；
- 标记需要复习的概念。

它维护的是学习状态机，而不是普通聊天记录。

## 5. 建议增加一个 Orchestrator

除了上面三个 agent，建议增加一个编排器：

```text
Learning Orchestrator
```

它可以是一个 agent，也可以先是普通后端逻辑。

职责：

- 判断当前应该调用哪个 agent；
- 控制一个学习 session 的流程；
- 管理 agent 之间的输入输出；
- 避免多个 agent 同时对用户说话；
- 决定何时进入日志写入；
- 决定何时结束本节学习；
- 决定下一步推荐。

没有 Orchestrator，多 agent 容易变成一群角色同时聊天，用户体验会乱。

推荐流程：

```text
用户选择目标：我要学 Rust
        ↓
学科专家生成阶段路线
        ↓
Orchestrator 选择今天的一小节
        ↓
费曼教学 agent 陪学
        ↓
用户提交回答 / 运行代码
        ↓
检查 agent 判断学习结果
        ↓
写入学习状态和学习日志
        ↓
推荐下一步
```

## 6. 第一版 MVP 不要做太大

第一版不要一上来做完整平台、很多学科、很多 agent。

推荐只验证一个闭环：

```text
选择学科
→ 生成路线
→ 开始一节课
→ 做一个小练习
→ 验证结果
→ 写入学习状态
→ 推荐下一步
```

可以先只支持 Go 或 Rust。

如果这个闭环成立，再扩展到多学科。

## 7. 推荐的第一版技术栈

```text
Next.js App Router + TypeScript
Postgres
Drizzle 或 Prisma
LangChain JS / LangGraph / Deep Agents
SSE streaming
Markdown / MDX 课程内容
后续增加代码 sandbox
```

第一阶段可以先不做复杂代码执行环境。

学习验证可以先从这几类开始：

- 用户解释；
- 用户选择题；
- 用户填空；
- 用户粘贴本地运行结果；
- 用户提交最小代码片段；
- agent 判断输出是否合理。

后续再做：

- 浏览器内代码运行；
- Go / Rust sandbox；
- 自动测试；
- 远程容器执行；
- 项目级练习。

## 8. 数据模型草案

初步可以有这些核心对象：

```text
User
- id
- name
- preferences

Subject
- id
- name
- description

LearningPath
- id
- userId
- subjectId
- currentStage
- currentObjective

LessonSession
- id
- userId
- subjectId
- objective
- status
- startedAt
- endedAt

LearningProfile
- userId
- currentLevel
- completedTopics
- weakPoints
- verificationHabits
- nextGoal

LearningJournal
- id
- userId
- date
- learned
- evidence
- weakPoints
- nextStep

AgentRun
- id
- sessionId
- agentType
- input
- output
- toolCalls
- status
```

## 9. 这个项目真正难的地方

难点不在 Next.js，也不在 LangChain。

真正难的是：

- 如何判断用户是否真的学会；
- 如何让 agent 不要空泛鼓励；
- 如何避免一直问问题却不教学；
- 如何维护长期学习状态；
- 如何让多个 agent 协作但不打架；
- 如何让学习路线既个性化又不乱跳；
- 如何让用户感到每天都在推进。

如果这些做对了，技术栈可以逐步演进。

## 10. 后续评估问题

以后真正开始做之前，可以先回答这些问题：

1. 第一版是单用户自用，还是多用户 SaaS？
2. 第一版先支持 Go，还是 Rust？
3. 课程内容来自现有 Markdown，还是让学科专家 agent 动态生成？
4. 学习日志是用户确认后写入，还是 agent 自动写入？
5. 代码验证第一版要不要真的运行？
6. agent 执行是在 Next.js API Route 里，还是独立 worker？
7. 是否要保留“每次只推进一个认知点”的教学铁律？
8. 用户能否查看和修改自己的学习画像？
9. 学习路线是否允许 agent 根据表现自动调整？
10. 多 agent 之间的交接结果是否需要可视化？

## 11. 当前推荐结论

推荐继续评估。

不要先做大而全的平台。先做一个能证明价值的小闭环：

```text
一个学科
一个学习路线
一个教学 agent
一个检查写入 agent
一个学习 profile
一个每日 journal
一个完整 session 闭环
```

如果这个闭环能让用户明显感觉“它知道我学到哪了，也知道我下一步该怎么练”，这个项目就值得继续做。

