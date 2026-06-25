# SAG-WIKI

一个基于 RAG（检索增强生成）的全栈智能应用平台，支持知识库管理、AI 问答、Agent 任务执行等功能。

## ✨ 功能特性

- **🧠 RAG 智能问答**: 基于向量检索的检索增强生成，提供精准的知识问答服务
- **📚 知识库管理**: 支持文档上传、管理和权限控制
- **🤖 AI Agent**: 智能代理系统，可执行复杂的多步骤任务
- **🔐 用户认证**: 完整的用户、角色、权限管理体系
- **🏢 组织架构**: 支持部门管理和层级结构
- **💬 会话管理**: 保留历史对话上下文
- **📊 任务队列**: 异步任务处理机制

## 🏗️ 项目架构

```
sag-wiki/
├── server/                      # Go 后端服务
│   ├── main.go                  # 应用入口
│   ├── config/                  # 配置模块
│   │   ├── database.go         # 数据库配置
│   │   ├── minio.go            # MinIO 对象存储配置
│   │   ├── qdrant.go           # Qdrant 向量数据库配置
│   │   └── redis.go            # Redis 配置
│   ├── handlers/               # HTTP 处理器
│   │   ├── ai/                 # AI 相关
│   │   │   ├── agent.go        # Agent 处理器
│   │   │   └── rag.go          # RAG 处理器
│   │   ├── common/             # 公共功能
│   │   │   └── file.go         # 文件处理
│   │   ├── system/             # 系统管理
│   │   │   ├── auth.go         # 认证
│   │   │   ├── department.go   # 部门管理
│   │   │   ├── queue.go        # 队列管理
│   │   │   ├── role.go         # 角色管理
│   │   │   └── user.go         # 用户管理
│   │   └── wiki/               # 知识库
│   │       ├── document.go     # 文档管理
│   │       └── knowledge_base.go # 知识库管理
│   ├── infrastructure/          # 基础设施层
│   │   ├── agent/              # Agent 实现
│   │   │   ├── agent.go
│   │   │   ├── eino_agent.go
│   │   │   ├── setup.go
│   │   │   ├── tool_registry.go
│   │   │   └── tools/          # Agent 工具
│   │   ├── llm/                # LLM 客户端
│   │   │   ├── eino_client.go
│   │   │   └── llm_client.go
│   │   ├── qdrant/             # Qdrant 客户端
│   │   ├── queue/              # 任务队列
│   │   └── storage/            # 存储服务
│   │       └── minio.go
│   ├── middleware/             # 中间件
│   │   └── auth.go             # 认证中间件
│   ├── migrations/             # 数据库迁移
│   ├── models/                 # 数据模型
│   │   ├── api/                # API 响应模型
│   │   └── db/                 # 数据库模型
│   ├── repository/             # 数据访问层
│   ├── router/                 # 路由定义
│   ├── services/               # 业务逻辑层
│   │   ├── document_processor.go
│   │   └── rag_engine.go
│   └── utils/                  # 工具函数
│       ├── env.go
│       └── jwt.go
├── web/                        # Vue 3 前端应用
│   ├── src/
│   │   ├── api/               # API 调用
│   │   ├── assets/            # 静态资源
│   │   ├── components/        # 公共组件
│   │   │   ├── ai-elements/  # AI 相关组件
│   │   │   └── sag-ui/       # UI 组件库
│   │   ├── composables/       # 组合式函数
│   │   ├── layout/            # 布局组件
│   │   ├── pages/             # 页面组件
│   │   ├── router/            # 路由配置
│   │   ├── stores/            # Pinia 状态管理
│   │   ├── types/             # TypeScript 类型
│   │   ├── utils/             # 工具函数
│   │   └── views/             # 视图组件
│   │       ├── ai/           # AI 相关页面
│   │       ├── common/       # 公共页面
│   │       ├── tasks/        # 任务管理
│   │       └── wiki/         # 知识库页面
│   └── package.json
└── docker-compose.yml          # Docker 编排配置
```

## 🛠️ 技术栈

### 后端

| 技术 | 用途 |
|------|------|
| **Go 1.21+** | 开发语言 |
| **Gin** | Web 框架 |
| **GORM** | ORM 框架 |
| **Qdrant** | 向量数据库 (RAG 检索) |
| **Redis** | 缓存 & 任务队列 |
| **MinIO** | 对象存储 (文档存储) |
| **Eino** | AI/LLM 框架 |
| **JWT** | 身份认证 |

### 前端

| 技术 | 用途 |
|------|------|
| **Vue 3** | 前端框架 |
| **TypeScript** | 类型安全 |
| **Vite** | 构建工具 |
| **Pinia** | 状态管理 |
| **TailwindCSS** | CSS 框架 |
| **Vue Router** | 路由管理 |

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- pnpm 10+
- PostgreSQL 14+
- Qdrant (向量数据库)
- Redis 6+
- MinIO (对象存储)

### Docker 快速部署

```bash
# 启动所有服务
docker-compose up -d
```

### 本地开发

#### 1. 克隆项目

```bash
git clone https://github.com/your-repo/sag-wiki.git
cd sag-wiki
```

#### 2. 后端开发

```bash
cd server

# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置数据库、Redis 等

# 运行服务
go run main.go
```

后端服务运行在 `http://localhost:8080`

#### 3. 前端开发

```bash
cd web

# 安装依赖
pnpm install

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置 API 地址

# 启动开发服务器
pnpm dev
```

前端服务运行在 `http://localhost:5173`

## 📁 环境变量配置

### 后端 (.env)

```env
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=sag_wiki

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Qdrant
QDRANT_HOST=localhost
QDRANT_PORT=6333

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=your_access_key
MINIO_SECRET_KEY=your_secret_key
MINIO_BUCKET=sag-wiki

# JWT
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h

# LLM 配置
LLM_API_KEY=your_api_key
LLM_BASE_URL=https://api.openai.com/v1
```

### 前端 (.env)

```env
VITE_API_BASE_URL=http://localhost:8080/api
```

## 📡 API 文档

### 认证模块

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/auth/login | 用户登录 |
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/logout | 退出登录 |

### 系统管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET/POST | /api/users | 用户管理 |
| GET/POST | /api/roles | 角色管理 |
| GET/POST | /api/departments | 部门管理 |

### 知识库

| 方法 | 路径 | 描述 |
|------|------|------|
| GET/POST | /api/knowledge-bases | 知识库管理 |
| GET/POST | /api/documents | 文档管理 |

### AI 功能

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/rag/chat | RAG 问答 |
| POST | /api/rag/sessions | 会话管理 |
| POST | /api/agent/execute | Agent 执行 |

## 🔧 项目结构说明

### 后端架构 (分层架构)

```
server/
├── config/           # 配置层 - 负责所有配置加载
├── models/           # 模型层 - 数据模型定义
├── repository/       # 数据访问层 - 数据库操作
├── services/         # 业务逻辑层 - 核心业务逻辑
├── handlers/         # 处理器层 - HTTP 请求处理
├── middleware/       # 中间件层 - 认证、日志等
├── infrastructure/   # 基础设施层 - 外部服务集成
├── router/           # 路由层 - 路由定义
└── utils/            # 工具层 - 通用工具函数
```

### 前端架构

```
web/src/
├── api/              # API 接口封装
├── assets/           # 静态资源
├── components/       # 通用组件
├── composables/      # 组合式函数 (hooks)
├── layout/           # 布局组件
├── router/           # 路由配置
├── stores/           # Pinia 状态管理
├── types/            # TypeScript 类型定义
├── utils/            # 工具函数
└── views/            # 页面视图
```

## 📦 依赖服务

项目依赖以下外部服务，通过 Docker Compose 可一键部署：

| 服务 | 端口 | 用途 |
|------|------|------|
| PostgreSQL | 5432 | 主数据库 |
| Qdrant | 6333 | 向量数据库 |
| Redis | 6379 | 缓存 & 队列 |
| MinIO | 9000 | 对象存储 |

## 🤝 贡献指南

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 📄 License

本项目基于 MIT License 开协议开源。

## 🙏 致谢

- [Gin](https://gin-gonic.com/) - 高性能 Go Web 框架
- [Vue 3](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Qdrant](https://qdrant.tech/) - 高性能向量数据库
- [Eino](https://github.com/cloudwego/eino) - Go AI 框架
