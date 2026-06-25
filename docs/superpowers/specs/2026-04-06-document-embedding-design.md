# 文档 Embedding 向量化设计方案

**日期**: 2026-04-06
**状态**: 已批准

---

## 1. 目标

在 `web/src/pages/wiki/documents` 页面实现文档向量化功能：
- 点击按钮将文档文本 embedding 后存储到 Qdrant
- 在页面查看 chunk 列表和详情（展开查看完整文本和向量维度）

---

## 2. 整体流程

```
用户点击"处理文档"
    ↓
POST /api/wiki/documents/:id/process
    ↓
Reprocess handler → 任务入队 → 立即返回 202
    ↓
HandleDocumentProcess (Worker)
    ↓
1. 从 MinIO 读取文件内容
2. 文本分块 (先段落，超限按句子切)
3. 获取 embedding 模型配置 (ai_model_config 表)
4. 调用 embedding API 生成向量
5. 批量 Upsert 到 Qdrant
6. 更新文档状态为 completed (chunk_count)
```

---

## 3. 新增 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/wiki/documents/:id/process` | 处理文档（已有，入队） |
| `GET` | `/api/wiki/documents/:id/chunks` | 获取文档的 chunk 列表 |
| `DELETE` | `/api/wiki/documents/:id/chunks` | 删除文档的 chunks (重新处理时) |

---

## 4. 后端新增文件

### 4.1 Embedding Service
**文件**: `server/app/ai/service/embedding.go`

- 从 `ai_model_config` 表读取 `model_type='embedding'` 的默认配置
- 调用兼容 OpenAI embedding API 格式的接口
- `base_url` + `/embeddings`，`POST` 请求
- Request: `{"input": "文本", "model": "xxx"}`
- Response: `{"data": [{"embedding": [0.1, 0.2, ...]}]}`

### 4.2 Chunker Service
**文件**: `server/app/ai/service/chunker.go`

分块策略（两者结合）:
1. 按 `\n\n` 分段
2. 每段超 500 字符 → 按 `.。！？` 断句
3. 每 chunk 最小 50 字符（过滤碎屑）
4. 相邻短段落合并

### 4.3 Document Processor Service
**文件**: `server/app/ai/service/document_processor.go`

整合流程：
1. 从 MinIO 读取文件内容
2. 调用 Chunker 分块
3. 调用 EmbeddingService 生成向量
4. 批量 Upsert 到 Qdrant
5. 更新 PostgreSQL 文档状态

### 4.4 Qdrant Chunk DAO
**文件**: `server/infrastructure/qdrant/chunk_dao.go`

- `GetChunksByDocumentID(ctx, documentID)` - 查 chunk 列表
- `DeleteChunksByDocumentID(ctx, documentID)` - 删除文档所有 chunks
- `UpsertChunks(ctx, chunks)` - 批量写入

### 4.5 Handler 改动
**文件**: `server/app/wiki/handlers/document.go`

- 新增 `GetChunks` - 查 chunk 列表
- 新增 `DeleteChunks` - 删除 chunks
- `HandleDocumentProcess` 实现完整处理逻辑

---

## 5. Qdrant 数据结构

**Collection**: `documents` (已存在)
**VectorSize**: 2048
**Distance**: Cosine

### Point Payload 结构

```json
{
  "document_id": "uuid",
  "chunk_index": 0,
  "chunk_text": "完整的 chunk 文本内容...",
  "chunk_size": 1234,
  "filename": "文档名.pdf",
  "folder_id": "uuid"
}
```

### Chunk 查询 API 响应

```json
{
  "document_id": "uuid",
  "chunk_count": 5,
  "chunks": [
    {
      "chunk_id": "uuid",
      "chunk_index": 0,
      "chunk_text": "chunk 文本...",
      "chunk_size": 1234,
      "vector_dim": 2048
    }
  ]
}
```

---

## 6. 前端改动

### 6.1 DocumentsPage (`index.tsx`)
- 点击"处理"按钮调用 `documentApi.process()`
- 展示 processing 状态

### 6.2 DataTable (`_components/data-table.tsx`)
- 每行增加 chunk_count badge
- 每行可展开显示 chunk 列表
- chunk 详情弹窗：显示完整文本 + 向量维度

### 6.3 API 层 (`@/api/wiki/document.ts`)
- 新增 `getChunks(docId)` - 获取 chunk 列表
- 新增 `deleteChunks(docId)` - 删除 chunks

---

## 7. 错误处理

| 场景 | 处理方式 |
|------|----------|
| API 调用失败 | asynq 任务重试 (MaxRetry=3) |
| 分块过长 | 跳过该 chunk 并记录日志 |
| Qdrant 写入失败 | 任务重试 |
| 文档内容为空 | 更新状态为 failed，错误信息 |

---

## 8. 实现顺序

1. **Embedding Service** - 调用 embedding API
2. **Chunker Service** - 文本分块逻辑
3. **Qdrant Chunk DAO** - chunk 增删查
4. **Document Processor** - 整合流程
5. **Handler** - 实现 `HandleDocumentProcess` + 新增 API
6. **前端** - chunk 列表展示
