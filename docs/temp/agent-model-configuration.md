# 学习 Agent 模型配置接入

## 状态

- 当前问题：已确认。
- 本轮是否实施：明确不实施。
- 临时方案：继续全局固定使用 `MiniMax-M3`。
- 最终配置方式：后续讨论。
- 实施顺序：第四阶段。

## 当前代码事实

学习 Agent 当前通过以下方法创建模型：

- `NewChatModel`
- `NewStructuredChatModel`

这两个方法都在 `server/infrastructure/llm/chat.go` 中临时固定使用 `MiniMax-M3`。

仓库已经存在通用模型配置能力：

- `sys_agent_model_configs` 表按 `agent_key + scene_key` 绑定模型。
- 系统 Agent 配置页面已经定义 `coach` 等 Agent。
- Wiki RAG 的 embedding 场景已经通过 Agent 配置查找模型。

但学习 Agent 的聊天模型尚未读取这套配置。因此系统页面上选择的模型目前不会真正决定 `LearningCoach` 或 `FeynmanReviewer` 使用的模型。

## 本阶段决定

Multi-Agent 第一版继续统一使用 `MiniMax-M3`，不把模型配置接入混入 Multi-Agent 行为验证。

原因是当前首先需要验证：

- Supervisor 是否能正确选择子 Agent。
- Listener 和 Teacher 的边界是否有效。
- 用户显式触发 WikiCurator 的流程是否自然。
- 一轮只运行一个子 Agent 是否满足真实练习体验。

如果同时引入每个 Agent 的模型选择，行为差异将难以判断究竟来自编排、Prompt 还是模型本身。

## 未来目标

后续每个 Agent 应通过统一的模型配置解析器获取模型，不再在 `chat.go` 中写死模型名称。

候选配置标识如下，但尚未确认：

| Agent | 候选 agent_key | scene_key |
| --- | --- | --- |
| LearningSupervisor | `learning_supervisor` | `default` |
| FeynmanListener | `feynman_listener` | `default` |
| LearningTeacher | `learning_teacher` | `default` |
| WikiCurator | `wiki_curator` | `default` |

配置解析需要至少支持：

- 根据 Agent 和场景读取模型绑定。
- 读取供应商、模型名称、密钥配置和调用参数。
- 区分普通对话模型与要求结构化 JSON 的模型调用。
- 配置缺失时给出明确错误或受控的默认值。
- 配置更新后的生效策略。

## 需要避免的设计

- 不让每个 Agent 自己重复查询数据库和拼装模型客户端。
- 不在 Prompt 或 Agent 构造函数中硬编码供应商细节。
- 不继续扩散 `NewChatModel` 和 `NewStructuredChatModel` 这种全局固定模型入口。
- 不因为多个 Agent 暂时使用同一个模型，就把它们永久绑定为一个配置项。

## 尚未确认

- 每个子 Agent 是否必须独立配置模型。
- Supervisor 与子 Agent 是否允许共享配置。
- 结构化输出是否作为独立 scene，而不是独立 Agent key。
- 模型配置缺失时是启动失败、请求失败，还是回退默认模型。
- 是否需要为旧的 `coach` key 提供迁移或兼容读取。
- 配置是在每次请求读取，还是通过带失效机制的缓存读取。

## 不在本文范围

- 当前 Multi-Agent 的实现和行为测试。
- LearningSupervisor 和“开始”页面重命名。
- `learning_journals` 旧业务删除。
- Wiki RAG embedding 配置改造。
