# 删除 learning_journals 旧业务

## 状态

- 业务是否废弃：已确认。
- 删除范围：已确认。
- 数据库迁移处理方式：实施前确认。

## 结论

`learning_journals` 代表的是“按用户、Wiki 文件夹和日期汇总一条学习日志”的旧业务。这个模型不再符合当前以练习会话为事实来源的设计，应整体删除。

不再补充 `JournalRepository.Create` 的调用，也不再引入 `LearningJournalAgent` 为这张表生成内容。

## 当前代码事实

当前存在以下代码：

- `learning_journals` 数据库表。
- `LearningJournal` 数据模型。
- `JournalRepository` 查询和创建方法。
- `GET /api/learning/journal/page` 查询接口。
- 前端 `journalApi`。
- `LearningCoach` 运行上下文中的最近日志。
- `list_learning_journals` Agent 工具。
- Coach Prompt 中的日志渲染逻辑。
- 相关服务和 SSE 测试。

仓库中没有任何地方调用 `JournalRepository.Create`。费曼练习不会写入这张表。

## 为什么不是补写入逻辑

当前系统已经保存了更接近事实的数据：

- `learning_sessions`：一次练习会话。
- `learning_messages`：真实对话消息、Agent 类型和工具调用。
- `learning_explanation_reviews`：每轮解释的清晰点、含混点、误解和追问。
- `learning_memory_events`：练习产生的学习事件和证据。
- `learning_memory_items`：从事件沉淀出的长期理解、薄弱点和误区。

如果再按文件夹和日期生成 `learned`、`evidence`、`weak_points`、`next_step`，会形成第二套汇总事实，并产生以下问题：

- 与学习记忆系统职责重复。
- 多次练习被合并后无法还原真实会话。
- 汇总内容与底层证据可能不同步。
- 页面字段被旧表结构限制。

因此缺少写入不是待修复 Bug，而是这条旧业务从未形成闭环的表现。

## 练习日志页面如何保留

“练习日志”页面继续存在，但它是练习事实数据的查询和展示视图，不对应一张 `learning_journals` 汇总表。

当前页面继续使用 mock 数据，用于验证以下内容是否真正有价值：

- 有效练习贡献图。
- 最近练习会话。
- 已掌握、待解决和需要纠正的内容。
- Agent 协作时间线。
- Wiki 修改记录。
- 下一步行动。

在 Multi-Agent 实际运行并产生数据前，不设计真实查询接口，也不为了适配 mock 页面提前增加数据库字段。

## 已确认删除范围

删除时必须完整清理整条依赖链：

### 数据库

- `learning_journals` 表、索引和表注释。

### 后端

- `server/app/learning/models/db/journal.go`
- `server/app/learning/repository/journal.go`
- `server/app/learning/handlers/journal.go`
- Learning Router 中的 `/journal/page` 路由。
- `DatabaseService.Journals` 注册和初始化。
- Coach Runtime Context 中的 Journals 字段及映射。
- Coach Prompt 中的 `CoachJournal` 和日志渲染逻辑。
- `list_learning_journals` 工具及输入类型。
- 只服务于旧日志业务的相关测试。

### 前端

- `web/src/api/learning/journal.ts`
- `web/src/api/learning/index.ts` 中对应导出。

练习日志 mock 页面和 `/learn/journal` 路由保留。

## 实施约束

- 删除数据库表之前必须先删除所有运行时读取逻辑，否则 Coach 请求会访问不存在的表。
- 不要误删学习会话、解释审阅或学习记忆数据。
- 不要因为删除旧 API 而删除练习日志页面。
- 不在本任务中设计新的练习日志接口。

## 尚未确认

当前仍需在实施前确认数据库迁移方式：

- 如果开发数据库允许重建，推荐直接从基础迁移中删除 `learning_journals`，再重建开发数据库。
- 如果需要保留现有数据库迁移历史，则新增迁移执行 `DROP TABLE learning_journals`。

该选择只影响数据库迁移操作，不影响业务删除结论。

## 不在本文范围

- Multi-Agent 实现。
- LearningSupervisor 和“开始”页面重命名。
- 练习日志最终查询字段。
- Agent 模型配置。

