# 04 · Agent 系统设计

> 编排框架使用 **eino**，运行在 Go 进程内；不额外启动 Python / JS agent 服务。

本系统不是让 AI 一路主动上课，而是围绕费曼学习法建立一套“用户输出 -> 系统验证 -> 必要时补讲 -> 再输出 -> 写入学习状态”的闭环。

费曼学习法在 Verve 里的产品定义是：

- 用户必须尝试用自己的话讲明白一个小目标；
- 系统负责听、判断、指出漏洞，而不是一上来替用户讲；
- 只有当用户讲错、讲漏、讲不出来时，教学 agent 才介入；
- 学习路线允许在学习过程中被质疑和修正。

## 角色总览

| Agent | 一句话职责 |
|---|---|
| Learning Planner | 根据用户目标和资料生成初始学习路线 |
| Feynman Evaluator | 听用户解释，判断是否真正理解 |
| Teaching Agent | 在用户讲错、讲漏、讲不出来时做针对性补讲 |
| Learning Examiner | 写入学习状态、学习日志、薄弱点和复习信息 |
| Curriculum Reviewer | 处理用户对路线、内容、顺序、难度、资料映射的质疑，并修改路线 |
| Learning Orchestrator | 决定当前该调用谁，控制 session 流程，避免多个 agent 同时对用户说话 |

## 4.1 Learning Planner

Planner 只负责学习路线的初次生成，不负责后续陪学。

输入：

- 用户的一句话学习目标；
- 用户上传或选择的文档资料；
- 可选：用户背景、期望难度、学习目的。

输出：

- 学习路线；
- 阶段列表；
- 每个阶段下的小目标；
- 每个小目标关联的资料来源；
- 第一个建议开始的小目标。

职责：

- 判断用户目标对应的学习范围；
- 把资料拆成可学习的小目标；
- 安排由浅入深的学习顺序；
- 避免一个小目标塞入多个认知点；
- 为每个小目标保留资料依据，方便后续查看和质疑。

Planner 不做这些事：

- 不负责判断用户是否掌握；
- 不负责补讲；
- 不负责写学习日志；
- 不负责在用户质疑后强行维护原路线。

## 4.2 Feynman Evaluator

Feynman Evaluator 是费曼学习闭环的核心 agent。

它不是教学 agent。它的职责不是先讲课，而是听用户讲。

触发时机：

- 用户点击一个学习小目标；
- 用户查看资料后点击“开始费曼讲解”；
- 用户提交自己的解释、代码说明、运行结果或类比说明。

输入：

- 当前学习目标；
- 当前小目标；
- 小目标关联的资料片段；
- 用户提交的解释；
- 历史掌握状态。

输出：

```json
{
  "verdict": "pass | partial | fail",
  "mastery_after": "none | seen | heard | explained | written | verified",
  "correct_points": ["用户讲对的点"],
  "missing_points": ["用户漏掉的点"],
  "wrong_points": ["用户讲错的点"],
  "next_action": "pass_to_examiner | ask_user_retry | call_teaching_agent"
}
```

判断原则：

- 用户能用简单语言讲清楚，才算 explained；
- 用户能写出可工作的代码或步骤，才算 written；
- 用户能结合运行结果、例子或反例验证，才算 verified；
- 只会复述术语，不算真正掌握；
- 解释中缺关键因果关系，只能算 partial；
- 解释方向错误或无法说明，则 fail。

## 4.3 Teaching Agent

Teaching Agent 只在需要时出现。

它不是默认入口，也不替代用户输出。它的存在是为了解决用户暴露出来的具体漏洞。

触发时机：

- Feynman Evaluator 判定为 fail；
- 用户说不出来；
- 用户的解释缺少关键概念；
- 用户混淆了概念边界；
- 用户要求系统解释某个卡点。

输入：

- 当前小目标；
- 用户刚才的错误解释或空白回答；
- Evaluator 给出的 missing_points / wrong_points；
- 相关资料片段。

输出：

- 为什么用户刚才的说法不对；
- 正确理解应该是什么；
- 一个尽量生活化或代码化的类比；
- 一个很小的重讲任务。

要求：

- 不重新上一整节课；
- 只补当前漏洞；
- 不连续追问；
- 补讲后必须回到用户重新解释；
- 不直接替用户完成费曼输出。

## 4.4 Learning Examiner

Learning Examiner 是学习状态维护者。

它不负责聊天，也不负责总结气氛。它维护的是平台里的学习状态机。

职责：

- 根据 Evaluator 的判定更新小目标 mastery_level；
- 更新 learning objective 的状态；
- 写入 LearningJournal；
- 更新 LearningProfile；
- 记录薄弱点；
- 标记需要复习的概念；
- 生成下一步建议。

它写入的内容应该是学习事实，而不是聊天流水。

示例：

```json
{
  "objective_id": "xxx",
  "mastery_after": "explained",
  "evidence": "用户能解释 panic/recover 的调用边界，但还不能结合代码验证",
  "weak_points": ["defer 执行顺序", "recover 生效位置"],
  "next_recommendation": "用一个小代码例子验证 recover 只能在 defer 中生效"
}
```

## 4.5 Curriculum Reviewer

Curriculum Reviewer 负责处理用户对学习路线和学习内容的质疑。

这是一个独立 agent，不属于 Planner，也不属于 Teaching Agent。

触发时机：

- 用户说“这个顺序不对”；
- 用户说“这个我已经会了”；
- 用户说“这个太难 / 太浅”；
- 用户说“这个和资料不一致”；
- 用户说“小目标拆得不合理”；
- 用户说“这节不应该先学这个”；
- 用户希望调整学习路线、阶段、小目标或资料映射。

输入：

- 当前学习目标；
- 当前学习路线；
- 当前小目标；
- 原始资料和资料片段；
- 用户学习状态；
- 用户提出的质疑或修改建议。

输出：

```json
{
  "decision": "accept | partially_accept | reject",
  "reason": "为什么这样判断",
  "changes": [
    {
      "type": "reorder | split | merge | skip | add | rewrite | remap_resource",
      "target_id": "被修改的小目标或阶段",
      "detail": "具体修改内容"
    }
  ],
  "current_objective_after_change": "调整后建议继续学习的小目标"
}
```

判断原则：

- 如果用户质疑成立，要修改路线，而不是解释原路线永远正确；
- 如果用户只掌握一部分，可以跳过已掌握部分，但保留未验证部分；
- 如果用户误解了路线安排，要解释为什么当前顺序仍然合理；
- 修改路线时必须保留资料依据；
- 调整当前小目标时，要避免用户 session 突然跳到无关内容。

## 4.6 Learning Orchestrator

Learning Orchestrator 是唯一直接控制 session 流程的调度层。

它决定当前该调用哪个 agent，并负责把 agent 输出转成产品动作。

职责：

- 判断当前是生成路线、开始学习、提交解释、补讲、重讲、通过、写日志，还是路线修订；
- 管理 agent 之间的输入输出；
- 避免多个 agent 同时对用户说话；
- 确保 Teaching Agent 只在需要时出现；
- 确保 Evaluator 的判断结果能进入 Examiner；
- 确保 Curriculum Reviewer 的修改能刷新路线和当前学习目标；
- 决定本节是否结束，以及下一步推荐什么。

Orchestrator 不是一个“聊天角色”。它更像学习 session 的状态机。

## 主流程：用户输出验证

```text
用户点击一个学习小目标
        ↓
查看对应资料 / 小目标内容
        ↓
用户点击“开始费曼讲解”
        ↓
用户用自己的话解释
        ↓
Feynman Evaluator 判断
        ↓
    pass
        ↓
Learning Examiner 写学习状态 + 学习日志
        ↓
Orchestrator 推荐下一步

    partial / fail / 用户说不出来
        ↓
Teaching Agent 针对漏洞补讲
        ↓
用户重新解释
        ↓
Feynman Evaluator 再判断
```

## 分支流程：用户质疑路线

```text
用户对路线 / 小目标 / 资料提出质疑
        ↓
Orchestrator 暂停当前学习推进
        ↓
Curriculum Reviewer 评估质疑
        ↓
    质疑成立
        ↓
修改阶段 / 小目标 / 顺序 / 难度 / 资料映射
        ↓
刷新当前路线和当前小目标

    部分成立
        ↓
局部修改，并说明保留哪些内容

    不成立
        ↓
解释当前安排为什么合理
        ↓
回到当前学习小目标
```

## 产品交互落点

当前学习目标详情页应该围绕“小目标验证”组织，而不是围绕“AI 讲课”组织。

建议页面动作：

- 查看配套资料；
- 开始费曼讲解；
- 提交我的解释；
- 查看系统评价；
- 需要补讲时进入针对性讲解；
- 重新讲一次；
- 通过后记录学习日志；
- 对路线提出质疑或申请调整。

右侧小目标面板里的“练习任务：用自己的话解释这个知识点”是正确方向。下一步应把它升级为可执行闭环：

```text
开始费曼讲解 -> 提交解释 -> AI 判断 -> 补讲或通过 -> 写入状态
```

## MVP 实现顺序

1. 在小目标详情中加入“开始费曼讲解 / 提交解释 / 查看判定”入口；
2. 将现有 Examiner 调整为 Feynman Evaluator 的语义，先完成 pass / partial / fail 判定；
3. partial / fail 时接入 Teaching Agent，生成针对性补讲；
4. pass 时写入 LearningJournal 和 LearningProfile；
5. 增加“质疑路线 / 调整小目标”入口；
6. 接入 Curriculum Reviewer，允许局部修改路线、小目标和资料映射；
7. 最后把这些分支收敛到 Learning Orchestrator 状态机。

## 非目标

- 不做多个 agent 同时聊天；
- 不做“AI 从头讲完整课程”的默认体验；
- 不把费曼学习法做成普通问答；
- 不把用户质疑路线视为异常；
- 不在 MVP 阶段追求复杂多轮 agent 协作，先跑通可验证的学习闭环。
