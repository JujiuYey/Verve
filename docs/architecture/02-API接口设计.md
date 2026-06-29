# 02 · API 接口设计

> 前缀 `/api`,受保护接口需 `Authorization: Bearer <token>`。响应统一 `Response[T]{code,data,message}`(成功 `code=0`)。分页用 `GET .../page`。

## 接口总表

### 账户(复用现有 auth / system)

| Method | Path | 说明 |
|---|---|---|
| POST | `/api/auth/login` | 登录,返回 token |
| POST | `/api/auth/register` | 注册(to C,按需补充) |
| GET | `/api/system/user/me` | 当前登录用户 |

### 学习目标

| Method | Path | 说明 |
|---|---|---|
| POST | `/api/learning/goal` | 创建目标 + 生成路线(SSE) |
| GET | `/api/learning/goal/page` | 目标分页列表 |
| GET | `/api/learning/goal/:id` | 目标详情(含路线 + 进度) |
| PUT | `/api/learning/goal` | 更新(标题 / 归档 / 完成,body 带 id) |
| DELETE | `/api/learning/goal/:id` | 删除目标 |

### 学习会话与陪练

| Method | Path | 说明 |
|---|---|---|
| POST | `/api/learning/session` | 开始一节(body: objective_id) |
| GET | `/api/learning/session/:id` | 会话详情(含历史消息) |
| POST | `/api/learning/session/:id/chat` | 陪练对话(**SSE**) |
| POST | `/api/learning/session/:id/exercise` | 提交练习,Examiner 判定 |
| POST | `/api/learning/session/:id/complete` | 结束本节(写画像 / 日志,推荐下一步) |

### 画像 / 日志 / 继续

| Method | Path | 说明 |
|---|---|---|
| GET | `/api/learning/goal/:id/profile` | 学习画像 |
| GET | `/api/learning/journal/page` | 学习日志(按天分页) |
| GET | `/api/learning/continue` | 继续上次 / 今日推荐 |

> 模型配置复用现有 `ai_model_config` 接口(MVP 系统预置,无需用户配置)。

## 关键接口详情

### POST `/api/learning/goal` — 创建目标 + 生成路线

请求:`{ "title": "我要学 Go 的并发" }`
处理:建 goal → Planner 生成路线 → 落 path + objectives(事务)。
响应(二选一,见 [03](03-Agent编排设计.md)):

- **SSE(推荐)**:流式推规划过程,末条 `data: {"type":"done","goal_id":"g_x","path":{...}}`。
- **同步 JSON**:
```json
{ "code":0, "message":"success", "data": {
  "goal_id":"g_x",
  "path": { "stages":[ { "title":"基础",
    "objectives":[ {"id":"o1","title":"goroutine 是什么","order":0} ] } ] } } }
```

### GET `/api/learning/goal/:id` — 目标详情(含路线 + 进度)
```json
{ "code":0, "data":{
  "goal": {"id":"g_x","title":"Go 并发","status":"active"},
  "progress": {"completed":3,"total":8},
  "current_objective_id":"o4",
  "stages":[ {"title":"基础","objectives":[
     {"id":"o1","title":"goroutine 是什么","status":"completed","mastery_level":"verified"} ]} ]
}}
```

### POST `/api/learning/session` — 开始一节
请求 `{ "objective_id":"o4" }` → 响应 `{ "code":0, "data":{"session_id":"s_x"} }`。

### POST `/api/learning/session/:id/chat` — 陪练对话(SSE)
请求 `{ "message":"我的回答……" }`(首次进入可为空,由 agent 先开场)。
SSE 事件(沿用现有 `chat.go` 结构 + learning 扩展 `phase` / `exercise`):
```text
data: {"type":"stream_chunk","content":"……","agent":"tutor"}
data: {"type":"message","content":"完整讲解","agent":"tutor","phase":"explain"}
data: {"type":"exercise","content":{"type":"code_snippet","prompt":"写一段……"},"agent":"tutor"}
data: {"type":"error","content":"……"}
data: [DONE]
```
- `type`:`stream_chunk` / `message` / `tool_result` / `exercise` / `action` / `error`。
- `phase`:`diagnose` / `explain` / `practice` / `verify`(驱动前端右栏四步进度)。

### POST `/api/learning/session/:id/exercise` — 提交练习验证
请求:
```json
{ "type":"code_snippet", "prompt":"写一段启动 3 个 goroutine 并等待",
  "user_answer":"package main……" }
```
响应:
```json
{ "code":0, "data":{
  "verdict":"pass", "mastery_after":"verified",
  "feedback":"正确,用了 WaitGroup……", "objective_id":"o4" }}
```

### POST `/api/learning/session/:id/complete` — 结束本节
处理:Examiner 汇总 → 更新 objective、profile、journal、推进 `path.current_objective_id`。
```json
{ "code":0, "data":{
  "summary":"本节掌握了 goroutine 调度……",
  "next_objective":{"id":"o5","title":"select 多路复用"} }}
```

### GET `/api/learning/continue` — 继续上次
```json
{ "code":0, "data":{
  "goal_id":"g_x","objective_id":"o4","title":"goroutine 调度","session_id":null }}
```

## 错误响应(统一)
```json
{ "code":400, "data":null, "message":"请求参数错误" }
```
- `400` 参数 / `401` 未登录 / `403` 越权(访问他人资源)/ `404` 不存在 / `500` 内部。

## 鉴权与数据隔离
- 除 `login` / `register` 外均需 JWT。
- 所有 `/learning/*` 资源校验 `user_id` 归属,防越权访问他人学习数据。
