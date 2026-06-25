
# RAG Query Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现用户提问功能：用户提问 → 根据文档权限检索 Qdrant chunks → 结合上下文 LLM 生成答案

**Architecture:**
```
用户提问
    ↓
Query Embedding（复用现有 EmbeddingService）
    ↓
Folder Tree Expansion（获取所有子文件夹）
    ↓
获取所有相关文档的 chunks
    ↓
按用户权限过滤（FolderPermission 继承模型）
    ↓
Qdrant 向量检索（ANN search）
    ↓
结合检索结果 + 用户问题 → LLM 生成答案
```

**Tech Stack:** Go, Qdrant, EmbeddingService, Fiber, existing RagSession/RagMessage models

---

## File Structure

**New Files to Create:**
- `server/app/ai/service/retrieval.go` - RAG 检索服务核心逻辑
- `server/app/ai/repository/folder_expander.go` - 文件夹树展开和权限过滤
- `server/infrastructure/qdrant/chunk_search.go` - Qdrant 向量检索方法

**Existing Files to Modify:**
- `server/infrastructure/qdrant/chunk_dao.go` - 添加 SearchChunksByVector 方法
- `server/app/ai/handlers/chat.go` - 扩展 ChatHandler 支持 RAG 检索
- `server/app/wiki/repository/folder.go` - 添加递归获取子文件夹方法
- `server/app/wiki/repository/folder_permission.go` - 添加批量权限检查方法

---

## Task 1: Add Chunk Search by Vector in Qdrant

**Files:**
- Modify: `server/infrastructure/qdrant/chunk_dao.go`

- [ ] **Step 1: Add SearchChunksByVector method**

在 `ChunkDAO` 中添加向量检索方法：

```go
// SearchChunksByVector 根据向量搜索最相似的 chunks
// filter: 可选的 Qdrant filter 条件
// limit: 返回结果数量限制
func (d *ChunkDAO) SearchChunksByVector(ctx context.Context, queryVector []float32, filter *qdrantpb.Filter, limit uint64) ([]*ChunkInfo, error)
```

实现逻辑：
1. 调用 Qdrant 的 `SearchPoints` API
2. 使用余弦相似度（collection 已配置）
3. 返回匹配度最高的 chunks

- [ ] **Step 2: Run verification**

确保代码编译通过：
```bash
cd server && go build ./...
```

---

## Task 2: Add Folder Tree Expansion and Permission Filter

**Files:**
- Create: `server/app/ai/repository/folder_expander.go`
- Modify: `server/app/wiki/repository/folder.go`

- [ ] **Step 1: Add GetAllSubFolderIDs to folder repository**

在 `FolderRepository` 接口添加：
```go
GetAllSubFolderIDs(ctx context.Context, parentID string) ([]string, error)
```

实现：递归获取所有子文件夹 ID（包括自身）

- [ ] **Step 2: Add BatchCheckPermissions to folder permission repository**

在 `FolderPermissionRepository` 接口添加：
```go
BatchCheckPermissions(ctx context.Context, folderIDs []string, userID string) (map[string]bool, error)
```

返回每个文件夹ID对应的权限状态（用户有权限返回 true）

- [ ] **Step 3: Create folder_expander.go**

```go
type FolderExpander struct {
    folderRepo     repository.FolderRepository
    permissionRepo repository.FolderPermissionRepository
}

// ExpandAndFilter 获取用户有权限访问的文件夹列表（递归展开子文件夹）
func (e *FolderExpander) ExpandAndFilter(ctx context.Context, folderID string, userID string) ([]string, error)
```

逻辑：
1. 调用 `GetAllSubFolderIDs` 展开文件夹树
2. 调用 `BatchCheckPermissions` 批量检查权限
3. 返回用户有权限的所有文件夹 ID

- [ ] **Step 4: Run verification**

```bash
cd server && go build ./...
```

---

## Task 3: Create RAG Retrieval Service

**Files:**
- Create: `server/app/ai/service/retrieval.go`

- [ ] **Step 1: Create RetrievalService**

```go
type RetrievalService struct {
    embeddingService *EmbeddingService
    chunkDAO        *qdrant.ChunkDAO
    folderExpander  *FolderExpander
    documentRepo    repository.DocumentRepository
}

// SearchResult 检索结果
type SearchResult struct {
    ChunkInfo   *qdrant.ChunkInfo
    Score       float32
}

// Search 执行 RAG 检索
func (s *RetrievalService) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error)
```

**SearchRequest 结构：**
```go
type SearchRequest struct {
    Query      string   // 用户问题
    FolderID   string   // 可选：限定文件夹
    DocumentID string   // 可选：限定文档
    UserID     string   // 用户ID（用于权限过滤）
    Limit      int      // 返回 chunks 数量
}
```

逻辑流程：
1. 调用 `EmbeddingService.GetEmbedding` 将 query 转为向量
2. 如果指定了 `FolderID`，调用 `FolderExpander.ExpandAndFilter` 获取有权限的文件夹列表
3. 构建 Qdrant filter：只检索指定文件夹/文档的 chunks
4. 调用 `ChunkDAO.SearchChunksByVector` 执行向量检索
5. 返回 top-k 结果

- [ ] **Step 2: Run verification**

```bash
cd server && go build ./...
```

---

## Task 4: Integrate RAG into Chat Handler

**Files:**
- Modify: `server/app/ai/handlers/chat.go`

- [ ] **Step 1: Add RAG retrieval to Chat flow**

在 `Chat` handler 中集成检索：
1. 接收 `folder_id` 和 `document_id` 参数（可选）
2. 从 `c.Locals("user_id")` 获取当前用户 ID
3. 调用 `RetrievalService.Search` 获取相关 chunks
4. 将检索到的 `chunk_text` 组装成上下文
5. 将上下文 + 用户问题一起发给 Agent

**修改请求结构：**
```go
var req struct {
    Query      string `json:"query"`
    FolderID   string `json:"folder_id"`   // 可选
    DocumentID string `json:"document_id"` // 可选
}
```

- [ ] **Step 2: Run verification**

```bash
cd server && go build ./...
```

---

## Task 5: Create Document Repository Method

**Files:**
- Modify: `server/app/wiki/repository/document_repo.go`

- [ ] **Step 1: Add GetDocumentsByFolderIDs method**

```go
// GetDocumentsByFolderIDs 根据文件夹ID列表获取所有文档
func (r *documentRepository) GetDocumentsByFolderIDs(ctx context.Context, folderIDs []string) ([]*wiki_db.Document, error)
```

用于后续过滤（如果需要）

- [ ] **Step 2: Run verification**

```bash
cd server && go build ./...
```

---

## Task 6: End-to-End Verification

**Files:**
- Integration test (manual or test file)

- [ ] **Step 1: Manual API test**

```bash
# 1. 创建测试文件夹和上传文档
# 2. 为当前用户添加文件夹权限
# 3. 调用 embedding 接口处理文档
# 4. 调用 chat 接口提问
# 5. 验证返回结果包含正确的上下文
```

**验证点：**
- [ ] query embedding 正常生成
- [ ] 文件夹树正确展开（包含子文件夹）
- [ ] 权限过滤正确（无权限文件夹的 chunks 不被检索）
- [ ] 向量检索返回相关 chunks
- [ ] LLM 生成的回答基于检索到的上下文

---

## Dependencies

1. **前提条件**（已完成）：
   - ✅ 文档 embedding 存储到 Qdrant
   - ✅ EmbeddingService 已实现
   - ✅ RagSession/RagMessage 模型已存在

2. **本计划新增**：
   - Qdrant 向量检索方法
   - 文件夹树展开逻辑
   - 权限批量检查
   - RAG 检索服务

---

## Verification Commands

```bash
# 编译验证
cd server && go build ./...

# 单测（如果有）
cd server && go test ./app/ai/... -v
```
