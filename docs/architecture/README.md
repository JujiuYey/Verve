# 技术设计文档(Architecture)

> 多 Agent 学习平台 · 技术方案

本目录讲"**功能怎么实现、接口怎么设计、agent 怎么编排**"。需求见 [`../prd/`](../prd/),产品形态与表结构见 [`../product/`](../product/)。

## 技术栈速查

| 层 | 技术 |
|---|---|
| 前端 | React 19 + Vite + TanStack Router/Query + shadcn/ui + AI SDK + streamdown |
| 后端 | Go + Fiber v2 |
| Agent | **eino `adk`**(多 agent + transfer + 流式) |
| ORM / DB | bun + Postgres |
| 队列 | asynq(Redis) |
| 其它 | JWT 认证、SSE 流式、MinIO |

## 目录

| 文档 | 负责 | 状态 |
|---|---|---|
| [00-架构总览](00-架构总览.md) | 分层、请求流、跨切面约定 | 初稿 |
| [01-后端模块与功能实现](01-后端模块与功能实现.md) | learning 模块结构 + 每个功能的实现路径 | 初稿 |
| [02-API接口设计](02-API接口设计.md) | REST 接口清单 + 详情 | 初稿 |
| [03-Agent编排设计](03-Agent编排设计.md) | eino adk 多 agent、transfer、工具、prompt | 初稿 |
| [04-核心流程时序](04-核心流程时序.md) | 建目标→路线、上课→验证→写状态 | 初稿 |

## 贴合现有代码的约定(沿用,不另起一套)

- **API 前缀** `/api`,模块路由 `/<module>/<entity>`(单数),如 `/api/learning/goal`。
- **统一响应** `Response[T]{code,data,message}`(成功 `code=0`);用 `response.XxxCtx` 快捷方法。
- **认证**:受保护路由走 `AuthMiddleware`,`user_id` 从 `c.Locals("user_id")` 取。
- **SSE**:`c.Context().SetBodyStreamWriter` + `data: {...}\n\n`,以 `data: [DONE]\n\n` 收尾。
- **分页**:`GET /page` + `response.PaginateCtx`。
- **ID**:`varchar(32)`,沿用现有应用层生成方式。
- **eino**:`adk.NewChatModelAgent` + `adk.NewRunner(EnableStreaming:true)`;工具用 `utils.InferTool`。
