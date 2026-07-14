# Verve

面向个人使用的开源 AI 知识库与学习助理。基于 RAG 检索增强生成,内建三
类学习陪练 Agent(Coach 调度、Tutor 教学、Curator 文档修订),本地直接
运行,无需用户体系。

## ✨ 功能特性

- **📚 Wiki 知识库**: Markdown 文档按文件夹组织,每次编辑生成不可变版本,
  历史可回溯
- **🧠 RAG 智能问答**: 文档向量化入库,问答命中并附带引用片段
- **🎓 Coach 陪练调度**: 基于知识库上下文,通过 SSE 流式推送学习建议
- **📖 Feynman 练习**: 学习者解释 → 倾听 Agent 复述评估 → 追问闭环
- **✍️ Curator 文档修订**: 学习过程中对文档提出精确修改建议
- **📓 学习记忆**: 按 Wiki 文件夹维度聚类,持续累积学习反馈

## 🏗️ 项目架构

```
verve/
├── server/                      # Go 后端服务 (Fiber + Bun ORM)
│   ├── main.go                  # 应用入口
│   ├── config/                  # 配置模块 (database / minio / qdrant / redis)
│   ├── app/
│   │   ├── file/                # 文件下载 (公开)
│   │   ├── learning/            # 学习会话 / Coach / Feynman / 记忆
│   │   ├── rag/                 # RAG 索引与检索
│   │   ├── system/              # 模型 / 平台 / Agent 配置
│   │   └── wiki/                # 文件夹 / 文档 / 修订 / 变更申请
│   ├── infrastructure/
│   │   ├── agent/               # Eino Agent 实现
│   │   ├── llm/                 # LLM 客户端与 Prompt 模板
│   │   ├── storage/             # MinIO 对象存储
│   │   ├── vector/              # Qdrant
│   │   └── database/            # Bun DB 客户端
│   ├── migrations/              # SQL 迁移 (append-only)
│   ├── models/db/               # 数据库模型 (Bun)
│   ├── models/payload/          # API 请求/响应模型
│   ├── repository/              # 数据访问层
│   ├── router/                  # 路由定义 + 统一中间件
│   └── utils/                   # 通用工具
├── web/                         # React 19 前端 (Vite + TanStack Router + shadcn/ui)
│   └── src/
│       ├── api/                 # 接口封装 (按模块)
│       ├── app/                 # 应用入口 (dashboard 等)
│       ├── components/          # 通用组件 (含 sag-ui / ui / ai-elements)
│       ├── constants/           # 常量
│       ├── hooks/               # 自定义 Hooks
│       ├── layout/              # 布局
│       ├── lib/                 # 通用工具库
│       ├── pages/               # 页面 (按模块)
│       ├── routes/              # 路由 (TanStack file-based)
│       ├── stores/              # 应用状态
│       └── utils/               # 工具函数
└── docker-compose.yml           # 一键部署 Postgres / Redis / Qdrant / MinIO
```

## 🛠️ 技术栈

### 后端

| 技术 | 用途 |
|------|------|
| **Go 1.21+** | 开发语言 |
| **Fiber v2** | Web 框架 |
| **Bun ORM** | PostgreSQL ORM |
| **Qdrant** | 向量数据库 (RAG 检索) |
| **Redis** | 缓存 & 任务队列 |
| **MinIO** | 对象存储 (文档) |
| **Eino** | AI Agent 框架 (Cloudwego) |
| **OpenAI-compatible LLM** | Chat / Embedding |

### 前端

| 技术 | 用途 |
|------|------|
| **React 19** | 前端框架 |
| **TypeScript** | 类型安全 |
| **Vite** | 构建工具 |
| **TanStack Router** | 文件路由 |
| **TanStack Query** | 服务端状态 |
| **shadcn/ui** (new-york) | UI 组件 |
| **TailwindCSS v4** | 样式 |

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ / pnpm 10+
- PostgreSQL 14+
- Qdrant 1.x
- Redis 6+
- MinIO (或任意 S3 兼容存储)

### Docker 快速启动依赖

```bash
docker-compose up -d
```

### 本地开发

```bash
git clone https://github.com/your-org/verve.git
cd verve

# 后端
cd server
go mod download
cp .env.example .env
# 编辑 .env 配置 LLM_API_KEY / LLM_BASE_URL 等
go run main.go
# → http://localhost:8080

# 前端 (另一个终端)
cd ../web
pnpm install
cp .env.example .env
pnpm dev
# → http://localhost:5173
```

## ⚙️ 环境变量

### 后端 (`server/.env`)

| 变量 | 说明 |
|------|------|
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | PostgreSQL 连接 |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | Redis 连接 |
| `QDRANT_URL` | Qdrant HTTP 地址 |
| `MINIO_ENDPOINT` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` / `MINIO_BUCKET_NAME` | 对象存储 |
| `BUNDEBUG` | 开发环境设为 `1` 打印 SQL |

### LLM (`server/.env`)

| 变量 | 说明 |
|------|------|
| `LLM_API_KEY` | **必填**,启动时若为空则拒绝启动 |
| `LLM_BASE_URL` | OpenAI-compatible endpoint,留空则走系统默认 |

> 旧的 `JWT_SECRET` / `DEFAULT_PASSWORD` 等鉴权相关变量已废弃,
> 启动时不再读取,可以从 `.env` 中删除。

### 前端 (`web/.env`)

```
VITE_API_BASE_URL=http://localhost:8080/api
```

## 📡 API 速查

后端路由全部开放(本地直接使用,默认只监听本机)。

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/api/wiki/folders` | 文件夹列表/树 |
| POST/PUT/DELETE | `/api/wiki/folders` | 文件夹 CRUD |
| POST | `/api/wiki/documents` | 文档上传 (multipart) |
| GET | `/api/wiki/documents/:id` | 文档详情 / 内容 |
| POST | `/api/rag/chat` | RAG 问答 |
| POST | `/api/learning/sessions` | 启动 Feynman 学习会话 |
| POST | `/api/learning/coach/chat` | 陪练对话 (SSE) |
| GET/POST | `/api/system/platforms` | 模型平台配置 |
| GET/POST | `/api/system/models` | 可用模型列表 |

## 🗄 数据库

迁移文件位于 `server/migrations/{system,wiki,rag,learning}/`,按编号追加执行:

```bash
psql "$DATABASE_URL" -f server/migrations/system/001_schema.sql
psql "$DATABASE_URL" -f server/migrations/wiki/001_schema.sql
# ...
```

Verve 以追加模式管理 schema(历史迁移视为不可变快照)。

## 🤝 贡献

1. Fork 仓库
2. 新建特性分支 (`git checkout -b feature/your-feature`)
3. 提交改动 (`git commit -m 'feat(scope): description'`)
4. 推送到远端 (`git push origin feature/your-feature`)
5. 发起 Pull Request

## 📄 License

基于 [MIT License](LICENSE) 开源。
