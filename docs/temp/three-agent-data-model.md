# 三 Agent 练习数据模型设计

## 状态

- 三个 Agent 由用户在 UI 中明确选择：已确认。
- 费曼练习内部不使用 Supervisor 自动路由：已确认。
- `LearningTeacher` 使用独立的教学干预表：已确认。
- `WikiCurator` 使用“生成方案、用户确认、后端执行”的两阶段流程：已确认。
- Wiki 文档版本与 RAG 索引版本绑定：已确认。
- 本文只设计数据模型，不处理 Agent、页面或 API 的重命名。

## 目标

费曼练习页面提供三个由用户主动选择的 Agent：

```text
FeynmanListener
LearningTeacher
WikiCurator
```

每次用户输入只交给当前选中的一个 Agent。后端不再调用另一个模型判断应该选择哪个 Agent。

数据模型不遵循“一个 Agent 对应一张表”的机械规则。只有当 Agent 产生了需要结构化保存、查询和约束的业务事实时，才建立专用表。

## 数据关系总览

```text
learning_sessions
└── learning_turns
    ├── learning_messages
    ├── learning_explanation_reviews       FeynmanListener
    ├── learning_teaching_interventions    LearningTeacher
    └── Wiki 修改来源关联                  WikiCurator

wiki_documents
├── wiki_document_change_requests
├── wiki_document_revisions
└── rag_index_jobs
    └── rag_wiki_chunks
```

## learning_sessions

`learning_sessions` 继续表示用户围绕一篇 Wiki 文档进行的一次完整练习。

当前字段中的以下内容继续保留：

- `id`
- `user_id`
- `document_id`
- `status`
- `summary`
- `started_at`
- `ended_at`
- `created_at`
- `updated_at`

不增加以下字段：

- `phase`
- `active_agent`
- Supervisor 运行状态

Agent 由用户在每次输入时明确选择，不属于会话的自动状态机。

当前 `message_count` 没有实际维护逻辑，也不是三个 Agent 运行所需的事实。建议删除；如果后续确实需要统计消息数量，应通过消息查询聚合，或者在同一事务中原子维护，不能保留一个可能失真的计数。

## learning_turns

新增 `learning_turns`，表示“一次用户输入交给一个 Agent 处理”的业务轮次。

建议字段：

```text
id              VARCHAR(32)  主键
session_id      VARCHAR(32)  学习会话ID
request_id      VARCHAR(64)  客户端请求幂等标识
agent_type      VARCHAR(32)  listener / teacher / curator
status          VARCHAR(20)  processing / completed / failed
error_code      VARCHAR(64)  可为空
error_message   TEXT         可为空
started_at      TIMESTAMP
completed_at    TIMESTAMP    可为空
created_at      TIMESTAMP
updated_at      TIMESTAMP
```

必要约束：

```text
FOREIGN KEY (session_id) REFERENCES learning_sessions(id) ON DELETE CASCADE
UNIQUE (session_id, request_id)
CHECK (agent_type IN ('listener', 'teacher', 'curator'))
CHECK (status IN ('processing', 'completed', 'failed'))
```

这张表负责：

- 保证一次输入只对应一个 Agent。
- 在调用 Agent 前保存用户输入和处理状态。
- 记录失败轮次并支持重试。
- 使用同一个 `request_id` 防止重复消息和重复 Wiki 修改。
- 为练习日志提供稳定的 Agent 协作时间线。

## learning_messages

`learning_messages` 保存用户和 Agent 的完整对话内容。

需要增加：

```text
turn_id VARCHAR(32) NOT NULL
```

约束：

```text
FOREIGN KEY (turn_id) REFERENCES learning_turns(id) ON DELETE CASCADE
```

当前 `agent_type` 的数据库约束只允许 `tutor`、`examiner`、`guide`，不支持新的三个 Agent。建议由 `learning_turns.agent_type` 作为该轮处理 Agent 的唯一事实来源，避免在 turn 和 message 中重复保存同一含义。

消息中的 `role` 继续区分：

- `user`
- `assistant`
- `system`

同一 turn 至少包含一条用户消息；成功后包含一条 Agent 回复。工具执行过程是否保存为独立消息，等实际 Agent 工具协议确定后再决定，不提前扩展 `role`。

## FeynmanListener 与 learning_explanation_reviews

`learning_explanation_reviews` 只保存 `FeynmanListener` 对用户解释产生的结构化审阅。

建议增加：

```text
turn_id VARCHAR(32) NOT NULL UNIQUE
```

保留的结构化结果：

- `heard_summary`
- `clear_points`
- `confusing_points`
- `misconceptions`
- `follow_up_question`
- `explanation_summary`
- `ready_to_wrap_up`
- `context_sufficient`
- `created_at`

建议删除以下重复字段：

- `user_id`
- `document_id`
- `session_id`
- `explanation`

这些信息可以通过 `turn -> session -> user message` 获得。当前重复保存但没有一致性约束，可能出现 review 的用户、文档或解释与实际会话不一致。

Listener 审阅成功后，可以继续生成：

- `learning_memory_events`
- `learning_memory_items`

只有用户自己的解释和 Listener 的观察可以作为掌握、含混和误解的学习证据。

## LearningTeacher 与 learning_teaching_interventions

新增 `learning_teaching_interventions`，表示用户卡住后发生的一次结构化教学干预。

这张表存在的原因不是“Teacher 是一个 Agent”，而是教学干预本身是需要查询和复用的业务事实。

建议字段：

```text
id                   VARCHAR(32)  主键
turn_id              VARCHAR(32)  对应教学轮次
question_summary     TEXT         用户卡住的问题摘要
knowledge_gaps       JSONB        缺少的前置知识
explanation_summary  TEXT         本次教学内容摘要
key_points           JSONB        讲解关键点
examples             JSONB        使用的示例
evidence             JSONB        文档版本或 RAG chunk 依据
created_at           TIMESTAMP
```

必要约束：

```text
FOREIGN KEY (turn_id) REFERENCES learning_turns(id) ON DELETE CASCADE
UNIQUE (turn_id)
```

不重复保存以下内容：

- 完整用户问题：保存在 user message。
- 完整 Teacher 回复：保存在 assistant message。
- `user_id`、`document_id`、`session_id`：通过 turn 和 session 获得。

不增加 `understood`、`mastered` 等字段。Teacher 完成讲解不等于用户已经掌握。只有用户之后重新进行费曼解释，并由 Listener 观察到，才能形成掌握证据。

Teacher 可以形成“用户曾在这里请求帮助”的学习事件，但不能直接创建“已经掌握”的 `learning_memory_items`。

## learning_memory_events 与 learning_memory_items

现有学习记忆结构可以继续使用，但需要保证重试不会重复创建同一份证据。

建议为事件增加幂等约束：

```text
UNIQUE (source_type, source_id, event_type)
```

来源建议：

```text
FeynmanListener 审阅
  source_type = 'explanation_review'
  source_id   = learning_explanation_reviews.id

LearningTeacher 教学干预
  source_type = 'teaching_intervention'
  source_id   = learning_teaching_interventions.id
```

Teacher 对应的事件只能表示“请求过帮助”“接受过某项讲解”等历史事实，不能自动转化为掌握记忆。

WikiCurator 的文档修改不写入学习记忆表。它属于 Wiki 变更和审计事实。

## WikiCurator 的边界

`WikiCurator` 不能因为用户切换到该 Agent 就直接覆盖文档。

用户选择 Agent 只表示进入修改场景，不表示已经确认某一份具体修改内容。确认流程为：

```text
用户选择 WikiCurator
→ 用户说明希望如何修改
→ Agent 读取当前文档并生成修改方案
→ 保存 change request
→ UI 展示 diff
→ 用户确认或取消
→ 确定性后端服务执行修改
→ 创建文档 revision
→ 创建对应文档版本的 RAG 索引任务
```

真正写入 MinIO、更新数据库和创建索引任务的是普通后端服务，不是模型继续自由调用写入工具。

## wiki_document_change_requests

新增 Wiki 修改申请表，保存 Agent 提出的具体修改和用户确认状态。

建议字段：

```text
id                  VARCHAR(32)  主键
document_id         VARCHAR(32)  Wiki文档ID
requested_by        VARCHAR(32)  用户ID
source_type         VARCHAR(32)  来源类型，当前为 learning_turn
source_id           VARCHAR(32)  learning_turn ID
request_id          VARCHAR(64)  幂等标识
base_version        BIGINT       生成方案时的文档版本
instruction         TEXT         用户修改要求
change_summary      TEXT         Agent生成的修改摘要
proposed_content    TEXT         修改后的完整Markdown
proposed_diff       TEXT         展示给用户的差异
status              VARCHAR(20)  proposed / approved / applying / applied / failed / cancelled / conflict
error_message       TEXT         可为空
applied_version     BIGINT       可为空
created_at          TIMESTAMP
updated_at          TIMESTAMP
applied_at          TIMESTAMP    可为空
```

必要约束：

```text
FOREIGN KEY (document_id) REFERENCES wiki_documents(id) ON DELETE CASCADE
FOREIGN KEY (requested_by) REFERENCES sys_users(id) ON DELETE CASCADE
UNIQUE (document_id, request_id)
```

Wiki 模块已经被 Learning 模块依赖。为了避免 Wiki migration 反向依赖 `learning_turns` 形成循环依赖，`source_id` 不建立数据库外键，由服务层校验来源。

用户确认时必须再次检查：

```text
wiki_documents.current_version == change_request.base_version
```

如果不相等，说明生成方案后文档已经发生变化。申请标记为 `conflict`，不得覆盖新内容，必须基于最新版本重新生成方案。

## wiki_document_revisions

新增不可变的 Wiki 文档版本表。

建议字段：

```text
id                 VARCHAR(32)  主键
document_id        VARCHAR(32)  Wiki文档ID
version            BIGINT       文档版本号
object_path        TEXT         该版本在MinIO中的不可变路径
content_hash       VARCHAR(64)  内容哈希
file_size          BIGINT       文件大小
change_request_id  VARCHAR(32)  修改申请ID，可为空
changed_by         VARCHAR(32)  用户ID
change_summary     TEXT         修改摘要
created_at         TIMESTAMP
```

必要约束：

```text
UNIQUE (document_id, version)
```

`wiki_documents` 增加：

```text
current_version BIGINT NOT NULL DEFAULT 1
content_hash    VARCHAR(64)
```

`wiki_documents.file_path` 指向当前版本的不可变对象路径。旧版本不覆盖，保证可以审计和恢复。

## Wiki 修改事务

MinIO 和 PostgreSQL 无法共享数据库事务，因此写入顺序必须避免数据库指向不存在的对象：

```text
1. 校验 change request 状态和 base_version
2. 将新内容写入新的不可变 MinIO object_path
3. 开启数据库事务
4. 再次锁定并校验 wiki_documents.current_version
5. 插入 wiki_document_revisions
6. 更新 wiki_documents 当前版本、路径、哈希和大小
7. 标记 change request 为 applied
8. 创建指定 document_version 的 rag_index_jobs
9. 提交数据库事务
```

数据库事务失败时，MinIO 中可能留下未被引用的对象，但不会出现数据库指向缺失文件的情况。未引用对象可以由后续清理任务删除。

## RAG 文档版本

`rag_index_jobs` 增加：

```text
document_version BIGINT NOT NULL
```

`rag_wiki_chunks` 增加：

```text
document_version BIGINT NOT NULL
```

索引任务创建时固定文档版本和对应的 revision object path，不能在执行过程中重新读取 `wiki_documents.file_path`，否则并发更新时可能索引到错误版本。

索引完成、准备发布 chunk 和向量之前，再次检查：

```text
job.document_version == wiki_documents.current_version
```

如果不相等，说明该任务已经过期，应标记为 `superseded`，不能替换当前文档的索引。

检索时只允许返回当前文档版本的 chunk。当前版本尚未完成索引时：

- 不使用旧版本 chunk 作为最新依据。
- 费曼练习可以直接读取最新文档全文。
- 页面显示“文档已更新，知识检索尚未同步”。
- 允许重新执行当前版本的索引任务。

## 三个 Agent 的写入规则

| Agent | 必须写入 | 可以写入 | 禁止写入 |
| --- | --- | --- | --- |
| FeynmanListener | turn、消息、解释审阅 | 学习记忆事件和条目 | Wiki 文档 |
| LearningTeacher | turn、消息、教学干预 | 请求帮助类记忆事件 | 掌握记忆、解释审阅、Wiki 文档 |
| WikiCurator | turn、消息、Wiki change request | 用户确认后的 revision 和索引任务 | 解释审阅、掌握记忆、未经确认的文档写入 |

## 实施顺序

建议按依赖顺序实施：

1. 增加 `learning_turns` 并调整 `learning_messages`。
2. 让 `learning_explanation_reviews` 关联 turn。
3. 增加 `learning_teaching_interventions`。
4. 增加 Wiki change request 和 revision 版本模型。
5. 给 RAG job、chunk 和向量 payload 增加文档版本。
6. 实现 Wiki 修改确认和版本冲突处理。
7. 最后接入三个 Agent 的具体读写行为。

`learning_journals` 旧业务删除和练习日志真实查询不属于本文实施范围。
