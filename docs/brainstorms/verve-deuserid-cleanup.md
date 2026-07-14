---
date: 2026-07-14
topic: verve-deuserid-cleanup
---

# Verve 代码清理：去用户管理 / 去 user_id

## 目标

让 Verve 变成单用户自部署项目。从 Go + SQL + Vue 三层去掉：
- `sys_users` / JWT / 登录态
- 业务表里所有 `user_id` / `created_by` / `updated_by` / `requested_by` / `changed_by` 列
- 前端登录页、账号菜单、`auth` store、`Authorization: Bearer` 拦截器

只动代码逻辑与数据迁移。不动 README、不加 LICENSE、不改品牌文案、不加 .github 模板、不重写 docker-compose。

## 数据库迁移

新追加两个迁移文件（不改历史 SQL，保持社区可复现）：

### `server/migrations/learning/006_drop_user_id.sql`（先跑）

幂等清理：
- `ALTER TABLE ... DROP CONSTRAINT IF EXISTS fk_*_user`（去掉外键）
- `DROP COLUMN IF EXISTS user_id / created_by / updated_by / requested_by / changed_by`
- `DROP INDEX IF EXISTS idx_*_user_id`
- 重建 `learning_memory_summaries` 上的两条部分唯一索引：
  - DROP `uq_learning_memory_summaries_user_folder`
  - DROP `uq_learning_memory_summaries_user_global`
  - 新建 `uq_learning_memory_summaries_folder ON (folder_id) WHERE folder_id IS NOT NULL`

### `server/migrations/system/999_drop_user_auth.sql`（后跑）

- `TRUNCATE TABLE sys_users;`
- `DROP TABLE IF EXISTS sys_users CASCADE;`
- `DROP TABLE IF EXISTS wiki_folder_permissions CASCADE;`（幻表，全仓库无 CREATE TABLE，仅 COMMENT 引用，顺手清理）

顺序硬约束：必须 006 在前 999 在后，否则外键依赖会拒绝 `DROP TABLE sys_users`。

## 后端 Go

### 直接删除（user/auth 模块）

```
server/app/system/handlers/{user,auth}.go
server/app/system/router/{user,auth}.go
server/app/system/repository/user.go
server/app/system/models/db/user.go
server/app/system/models/payload/{user,auth}.go
server/middleware/auth.go
server/utils/jwt.go
```

注意 `wiki/handlers/folder.go` 引用了 `system_repo.NewUserRepository(db)` 作为 `userRepo` 字段，R9 后会破坏编译 — 同时清理。

### 保留

`server/app/system/handlers/{models,platforms,agent_model_configs}.go` 仍是 Verve 实例级配置入口，不删；只需去掉内部依赖 user 上下文的代码（实际上 model-config 与 user 解耦，留空即可）。

`server/app/system/handlers/models.go` 是 `ModelHandler`（AI 模型配置），**不要按文件名误删**。

### Router 改动

`server/router/router.go`：
- 移除 `/api/system/auth`、`/api/system/users` 路由挂载
- 移除所有 `middleware.AuthMiddleware()` 与 `middleware.RoleMiddleware(...)` 包裹
- 其他模块（wiki/rag/learning/system 的 model-config 类）的路由保留

### Handler / Service / Model 清理

权威查询命令（grep 而非硬编码行号）：

```bash
grep -rn 'c\.Locals("user_id")\|\bUserID\b\|user_id' \
  server/app/wiki server/app/learning
```

这个 grep 的命中列表是 P1 的实施清单。所有命中点删除：
- handler 里 `c.Locals("user_id")` 读取
- service 入参的 `UserID` 字段、session ownership 校验（`session.UserID != userID`）
- model 字段 + `gorm`/`bun` tag
- JSON 序列化里的 `user_id` field（前端 FolderTreeNode 等）

特别注意：`server/app/wiki/handlers/folder.go` 里有 `userRepo *system_repo.UserRepository` 字段 + `FolderTreeNode.UserID *string` JSON 字段；`server/app/learning/models/db/review.go` 仍残留 `bun:",scanonly"` 的 `SessionID/DocumentID/UserID/Explanation` 字段，这两处是 grep 不会高亮但必须手改的隐藏耦合。

### 测试 fixture 改写

`server/app` 下实测 `user_id` / `UserID` 命中 137 次 / 35 文件。改写原则：
- 把 `UserID: "user-1"` 全部删掉
- 原"跨用户不能跨 session 访问"的鉴权用例（如果有）也一并删除——单用户场景下无意义
- "session-by-id 查询"的正常用例保留，只是不再需要 user ID 入参

`server/app/learning/models/db/review_test.go` 等如果引用已 scanonly 字段，同步清理。

## 前端 Vue/TS

### 删除

```
web/src/routes/login.tsx
web/src/routes/_layout/system/user.tsx
web/src/routes/_layout/common/account.tsx
web/src/layout/sidebar/user.tsx
web/src/pages/system/user/                 # 整个目录
web/src/pages/login/                       # 整个目录，含 login-form.tsx、login-img/
web/src/api/system/user.ts
web/src/api/auth/                          # 整个目录
web/src/stores/auth.ts
```

删除后必须让 TanStack Router CLI 重新生成 `web/src/routeTree.gen.ts`，否则 build 失败。

### 修改

| 文件 | 改动 |
|---|---|
| `web/src/routes/_layout.tsx` | 移除路由级 auth 守卫 |
| `web/src/layout/sidebar/{index.tsx,menu.ts,nav-system.tsx}` | 移用户菜单项 |
| `web/src/utils/request/index.ts` | 去掉 `Authorization: Bearer` 头、`useAuthStore.getState().token` 调用、401 token-refresh 队列（80+ 行） |
| `web/src/utils/request/sse.ts` | 同上 |
| `web/src/api/learning/session.ts` | 同上 |
| `web/src/api/wiki/folder.ts` | 删 `import type { User } from "../system/user"`；删 `Folder`/`FolderTreeNode` 的 `user_id`/`created_by`/`updated_by`/`created_by_user`/`updated_by_user` |
| `web/src/layout/index.tsx` | 不再调 `useAuthStore` |
| `web/src/api/learning/session.ts` 等 8+ 文件 | 改写 `useAuthStore` / `authApi` 消费方 |

`useAuthStore` / `authApi` 的所有消费方用 grep 找出后批量改写。

## 执行阶段（按依赖顺序）

| 阶段 | 内容 | 验收 |
|---|---|---|
| **P0 数据层** | 写 `006_drop_user_id.sql` 与 `999_drop_user_auth.sql`，本地 psql 跑通 | `\d` 无 user_id 列与 sys_users 表 |
| **P1 后端 Go** | 删除 user/auth 文件、改 router.go、按 grep 清单清理 handler/service/model/测试 | `go build ./...` + `go test ./...` 通过 |
| **P2 前端** | 删除登录/账号/用户管理页、改 layout 与 api 拦截器、重生成 `routeTree.gen.ts` | `pnpm build` 通过 |

每阶段独立 commit，便于 review 与回滚。

## 验收总表

- 干净 PostgreSQL 跑 `psql ... < all migrations` 全部成功
- `grep -rn 'user_id\|UserID\|JWT_SECRET\|sys_users' server/ web/` 无业务命中
- `grep -rn 'c\.Locals("user_id")' server/` 无命中
- `go build ./...` 通过
- `go test ./...` 通过
- `pnpm build` 通过
- 端到端冒烟通过：上传 wiki 文档 → RAG 问答 → 学习会话 + Feynman
