# 文档 Embedding 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现文档向量化功能 - 点击按钮将文档文本 embedding 后存储到 Qdrant，并在页面查看 chunk 详情

**Architecture:** Worker 异步处理：读取 MinIO 文件 → 分块 → 调用 embedding API → 批量写入 Qdrant → 更新状态。前端通过 Qdrant 直接查询 chunk 列表展示。

**Tech Stack:** Go (asynq, bun, qdrant-go-client), React, shadcn/ui

---

## File Structure

```
server/
├── app/
│   ├── ai/
│   │   └── service/
│   │       ├── embedding.go        # [CREATE] Embedding API 调用
│   │       ├── chunker.go           # [CREATE] 文本分块逻辑
│   │       └── document_processor.go # [CREATE] 文档处理主流程
│   └── wiki/
│       ├── handlers/
│       │   └── document.go          # [MODIFY] 新增 GetChunks, DeleteChunks
│       └── router/
│           └── document.go          # [MODIFY] 注册新路由
├── infrastructure/
│   ├── qdrant/
│   │   ├── qdrant_client.go         # [MODIFY] 新增 DeleteByDocumentID
│   │   └── chunk_dao.go             # [CREATE] Qdrant chunk 增删查
│   └── queue/
│       └── task_queue.go            # [MODIFY] 实现 HandleDocumentProcess
web/src/
├── api/wiki/
│   └── document.ts                  # [MODIFY] 新增 getChunks, deleteChunks API
└── pages/wiki/documents/
    ├── _components/
    │   └── data-table.tsx           # [MODIFY] chunk 列表展示
    └── index.tsx                    # [MODIFY] 处理后刷新列表
```

---

## Task 1: Embedding Service

**Files:**
- Create: `server/app/ai/service/embedding.go`
- Modify: `server/app/ai/repository/model_config.go` (新增 FindDefaultByType)

- [ ] **Step 1: 在 model_config repository 新增 FindDefaultByType 方法**

```go
// repository/model_config.go

// FindDefaultByType 获取指定类型的默认模型配置
func (r *modelConfigRepository) FindDefaultByType(ctx context.Context, modelType string) (*ai_db.ModelConfig, error) {
    var config ai_db.ModelConfig

    err := r.db.NewSelect().
        Model(&config).
        Where("is_default = ?", true).
        Where("model_type = ?", modelType).
        Scan(ctx)

    if err != nil {
        return nil, fmt.Errorf("获取默认模型配置失败: %w", err)
    }

    return &config, nil
}
```

- [ ] **Step 2: 创建 Embedding Service**

```go
// service/embedding.go
package service

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    ai_db "sag-wiki/app/ai/models/db"
    "sag-wiki/app/ai/repository"
)

type EmbeddingService struct {
    repo repository.ModelConfigRepository
    httpClient *http.Client
}

func NewEmbeddingService(repo repository.ModelConfigRepository) *EmbeddingService {
    return &EmbeddingService{
        repo: repo,
        httpClient: &http.Client{Timeout: 60 * time.Second},
    }
}

// GetEmbedding 调用 embedding API 获取向量
func (s *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
    config, err := s.repo.FindDefaultByType(ctx, ai_db.ModelTypeEmbedding)
    if err != nil {
        return nil, fmt.Errorf("获取 embedding 配置失败: %w", err)
    }

    return s.CallEmbeddingAPI(ctx, config, text)
}

// CallEmbeddingAPI 调用 embedding API
func (s *EmbeddingService) CallEmbeddingAPI(ctx context.Context, config *ai_db.ModelConfig, text string) ([]float32, error) {
    url := fmt.Sprintf("%s/embeddings", config.BaseURL)

    reqBody := map[string]interface{}{
        "input": text,
        "model": config.Model,
    }

    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("序列化请求失败: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))

    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("调用 embedding API 失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("embedding API 返回错误状态码: %d", resp.StatusCode)
    }

    var result struct {
        Data []struct {
            Embedding []float32 `json:"embedding"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    if len(result.Data) == 0 {
        return nil, fmt.Errorf("embedding 返回为空")
    }

    return result.Data[0].Embedding, nil
}

// BatchEmbeddings 批量获取 embeddings
func (s *EmbeddingService) BatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
    results := make([][]float32, 0, len(texts))
    for _, text := range texts {
        emb, err := s.GetEmbedding(ctx, text)
        if err != nil {
            return nil, err
        }
        results = append(results, emb)
    }
    return results, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add server/app/ai/service/embedding.go server/app/ai/repository/model_config.go
git commit -m "feat: add embedding service for API calls"
```

---

## Task 2: Chunker Service

**Files:**
- Create: `server/app/ai/service/chunker.go`

- [ ] **Step 1: 创建 Chunker Service**

```go
// service/chunker.go
package service

import (
    "regexp"
    "strings"
)

type Chunker struct {
    maxChunkSize int
    minChunkSize int
}

func NewChunker(maxChunkSize, minChunkSize int) *Chunker {
    return &Chunker{
        maxChunkSize: maxChunkSize,
        minChunkSize: minChunkSize,
    }
}

// Chunk 代表一个文本块
type Chunk struct {
    Index     int
    Text      string
    Size      int // 字符数
}

// ChunkText 对文本进行分块
func (c *Chunker) ChunkText(text string) []*Chunk {
    if text == "" {
        return nil
    }

    // 1. 先按双换行分段
    paragraphs := strings.Split(text, "\n\n")

    chunks := make([]*Chunk, 0)
    currentText := ""
    currentSize := 0

    for _, para := range paragraphs {
        para = strings.TrimSpace(para)
        if para == "" {
            continue
        }

        paraSize := len(para)

        // 如果单个段落就超过 maxChunkSize，需要进一步切分
        if paraSize > c.maxChunkSize {
            // 先保存当前累积的
            if currentText != "" {
                chunks = append(chunks, &Chunk{
                    Index: len(chunks),
                    Text:  currentText,
                    Size:  currentSize,
                })
                currentText = ""
                currentSize = 0
            }

            // 按句子切分大段落
            subChunks := c.splitBySentences(para)
            chunks = append(chunks, subChunks...)
            continue
        }

        // 尝试添加到当前 chunk
        if currentSize+paraSize <= c.maxChunkSize {
            if currentText != "" {
                currentText += "\n\n"
                currentSize += 2
            }
            currentText += para
            currentSize += paraSize
        } else {
            // 保存当前，开始新的
            if currentText != "" {
                chunks = append(chunks, &Chunk{
                    Index: len(chunks),
                    Text:  currentText,
                    Size:  currentSize,
                })
            }
            currentText = para
            currentSize = paraSize
        }
    }

    // 处理最后一个 chunk
    if currentText != "" {
        chunks = append(chunks, &Chunk{
            Index: len(chunks),
            Text:  currentText,
            Size:  currentSize,
        })
    }

    // 合并过短的 chunks
    chunks = c.mergeSmallChunks(chunks)

    // 重新编号
    for i, chunk := range chunks {
        chunk.Index = i
    }

    return chunks
}

// splitBySentences 将长段落按句子切分
func (c *Chunker) splitBySentences(text string) []*Chunk {
    // 按中英文句号、问号、感叹号切分
    re := regexp.MustCompile(`[.。！？!?]`)
    sentences := re.Split(text, -1)

    chunks := make([]*Chunk, 0)
    currentText := ""
    currentSize := 0

    for _, sentence := range sentences {
        sentence = strings.TrimSpace(sentence)
        if sentence == "" {
            continue
        }

        sentenceSize := len(sentence)

        if currentSize+sentenceSize <= c.maxChunkSize {
            if currentText != "" {
                currentText += "."
                currentSize += 1
            }
            currentText += sentence
            currentSize += sentenceSize
        } else {
            // 保存当前
            if currentText != "" {
                chunks = append(chunks, &Chunk{
                    Index: len(chunks),
                    Text:  currentText,
                    Size:  currentSize,
                })
            }

            // 如果单个句子就超过限制，直接按长度截断
            if sentenceSize > c.maxChunkSize {
                subChunks := c.splitByLength(sentence)
                chunks = append(chunks, subChunks...)
            } else {
                currentText = sentence
                currentSize = sentenceSize
            }
        }
    }

    // 处理最后一个
    if currentText != "" {
        chunks = append(chunks, &Chunk{
            Index: len(chunks),
            Text:  currentText,
            Size:  currentSize,
        })
    }

    return chunks
}

// splitByLength 按长度强制切分
func (c *Chunker) splitByLength(text string) []*Chunk {
    chunks := make([]*Chunk, 0)
    for i := 0; i < len(text); i += c.maxChunkSize {
        end := i + c.maxChunkSize
        if end > len(text) {
            end = len(text)
        }
        chunkText := text[i:end]
        if len(chunkText) >= c.minChunkSize {
            chunks = append(chunks, &Chunk{
                Index: len(chunks),
                Text:  chunkText,
                Size:  len(chunkText),
            })
        }
    }
    return chunks
}

// mergeSmallChunks 合并过短的相邻 chunks
func (c *Chunker) mergeSmallChunks(chunks []*Chunk) []*Chunk {
    if len(chunks) == 0 {
        return chunks
    }

    result := make([]*Chunk, 0, len(chunks))
    current := chunks[0]

    for i := 1; i < len(chunks); i++ {
        chunk := chunks[i]
        // 如果当前 chunk 太小，和下一个合并
        if current.Size < c.minChunkSize {
            current.Text += "\n" + chunk.Text
            current.Size += 1 + chunk.Size
        } else {
            result = append(result, current)
            current = chunk
        }
    }
    result = append(result, current)

    return result
}

// ValidateChunkSize 验证向量维度是否符合预期
func (c *Chunker) ValidateChunkSize(vectorDim int) error {
    // 目前主要支持 2048 维
    if vectorDim != 2048 {
        return fmt.Errorf("不支持的向量维度: %d", vectorDim)
    }
    return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add server/app/ai/service/chunker.go
git commit -m "feat: add text chunker service with paragraph and sentence splitting"
```

---

## Task 3: Qdrant Chunk DAO

**Files:**
- Create: `server/infrastructure/qdrant/chunk_dao.go`
- Modify: `server/infrastructure/qdrant/qdrant_client.go` (确认现有方法)

- [ ] **Step 1: 阅读 qdrant_client.go 确认现有方法**

Read `server/infrastructure/qdrant/qdrant_client.go` - 确认 `SearchByVector`, `UpsertPoints`, `CreatePoint` 方法存在。

- [ ] **Step 2: 创建 Chunk DAO**

```go
// chunk_dao.go
package qdrant

import (
    "context"
    "fmt"

    qdrantpb "github.com/qdrant/go-client/qdrant"
)

const (
    ChunkCollectionName = "documents"
)

// ChunkInfo chunk 信息
type ChunkInfo struct {
    ChunkID     string
    ChunkIndex  int
    ChunkText   string
    ChunkSize   int
    DocumentID  string
    Filename    string
    FolderID    string
    VectorDim   int
}

// ChunkDAO chunk 数据访问对象
type ChunkDAO struct {
    client     *QdrantClient
    collection string
}

func NewChunkDAO() (*ChunkDAO, error) {
    client, err := NewQdrantClient()
    if err != nil {
        return nil, err
    }
    return &ChunkDAO{
        client:     client,
        collection: ChunkCollectionName,
    }, nil
}

// GetChunksByDocumentID 根据文档 ID 获取所有 chunks
func (d *ChunkDAO) GetChunksByDocumentID(ctx context.Context, documentID string) ([]*ChunkInfo, error) {
    // 使用 filter 查询该文档的所有 chunks
    filter := &qdrantpb.Filter{
        Must: []*qdrantpb.Condition{
            {
                ConditionOneOf: &qdrantpb.Condition_Field{
                    Field: &qdrantpb.FieldCondition{
                        Key: "document_id",
                        Match: &qdrantpb.Match{
                            MatchValue: &qdrantpb.Match_Keyword{Keyword: documentID},
                        },
                    },
                },
            },
        },
    }

    // 先搜索获取 ID 列表（limit 设为大值）
    searchReq := &qdrantpb.SearchPoints{
        CollectionName: d.collection,
        Vector:         make([]float32, VectorSize), // 零向量
        Limit:          1000,
        WithPayload:    &qdrantpb.WithPayloadSelector{SelectorOptions: &qdrantpb.WithPayloadSelector_Enable{Enable: true}},
        Filter:         filter,
    }

    pointsClient := d.client.client.GetPointsClient()
    searchResp, err := pointsClient.Search(ctx, searchReq)
    if err != nil {
        return nil, fmt.Errorf("搜索 chunks 失败: %w", err)
    }

    chunks := make([]*ChunkInfo, 0, len(searchResp.Result))
    for _, point := range searchResp.Result {
        chunkInfo := &ChunkInfo{
            VectorDim: VectorSize,
        }

        // 提取 ID
        if point.Id != nil {
            if uuid, ok := point.Id.PointIdOptions.(*qdrantpb.PointId_Uuid); ok {
                chunkInfo.ChunkID = uuid.Uuid
            }
        }

        // 提取 payload
        if point.Payload != nil {
            for k, v := range point.Payload {
                val := d.client.extractValue(v)
                switch k {
                case "document_id":
                    chunkInfo.DocumentID, _ = val.(string)
                case "chunk_index":
                    if f, ok := val.(float64); ok {
                        chunkInfo.ChunkIndex = int(f)
                    }
                case "chunk_text":
                    chunkInfo.ChunkText, _ = val.(string)
                case "chunk_size":
                    if f, ok := val.(float64); ok {
                        chunkInfo.ChunkSize = int(f)
                    }
                case "filename":
                    chunkInfo.Filename, _ = val.(string)
                case "folder_id":
                    chunkInfo.FolderID, _ = val.(string)
                }
            }
        }

        chunks = append(chunks, chunkInfo)
    }

    return chunks, nil
}

// DeleteChunksByDocumentID 根据文档 ID 删除所有 chunks
func (d *ChunkDAO) DeleteChunksByDocumentID(ctx context.Context, documentID string) error {
    // 先获取所有 chunks 的 ID
    chunks, err := d.GetChunksByDocumentID(ctx, documentID)
    if err != nil {
        return err
    }

    if len(chunks) == 0 {
        return nil
    }

    // 构建 ID 列表
    pointIDs := make([]*qdrantpb.PointId, 0, len(chunks))
    for _, chunk := range chunks {
        pointIDs = append(pointIDs, &qdrantpb.PointId{
            PointIdOptions: &qdrantpb.PointId_Uuid{Uuid: chunk.ChunkID},
        })
    }

    // 删除
    pointsClient := d.client.client.GetPointsClient()
    _, err = pointsClient.Delete(ctx, &qdrantpb.DeletePoints{
        CollectionName: d.collection,
        Points:        pointIDs,
    })

    if err != nil {
        return fmt.Errorf("删除 chunks 失败: %w", err)
    }

    return nil
}

// UpsertChunks 批量写入 chunks
func (d *ChunkDAO) UpsertChunks(ctx context.Context, chunks []*ChunkInfo, vectors [][]float32) error {
    if len(chunks) != len(vectors) {
        return fmt.Errorf("chunks 数量 (%d) 与 vectors 数量 (%d) 不匹配", len(chunks), len(vectors))
    }

    points := make([]*PointStruct, 0, len(chunks))
    for i, chunk := range chunks {
        payload := map[string]interface{}{
            "document_id": chunk.DocumentID,
            "chunk_index": chunk.ChunkIndex,
            "chunk_text":  chunk.ChunkText,
            "chunk_size":  chunk.ChunkSize,
            "filename":     chunk.Filename,
            "folder_id":    chunk.FolderID,
        }

        point := CreatePoint(chunk.ChunkID, vectors[i], payload)
        points = append(points, point)
    }

    return d.client.UpsertPoints(ctx, points)
}
```

- [ ] **Step 3: Commit**

```bash
git add server/infrastructure/qdrant/chunk_dao.go
git commit -m "feat: add Qdrant chunk DAO for CRUD operations"
```

---

## Task 4: Document Processor Service

**Files:**
- Create: `server/app/ai/service/document_processor.go`
- Modify: `server/infrastructure/queue/task_queue.go` (TaskWorker 需要注入依赖)

- [ ] **Step 1: 创建 Document Processor Service**

```go
// service/document_processor.go
package service

import (
    "context"
    "fmt"
    "log"

    "github.com/google/uuid"

    qdrantdao "sag-wiki/infrastructure/qdrant"
    "sag-wiki/infrastructure/storage"
    "sag-wiki/app/ai/repository"
    wiki_db "sag-wiki/app/wiki/models/db"
    wiki_repo "sag-wiki/app/wiki/repository"
)

type DocumentProcessor struct {
    embeddingService *EmbeddingService
    chunker          *Chunker
    chunkDAO         *qdrantdao.ChunkDAO
    minioService     *storage.MinIOService
    documentRepo     wiki_repo.DocumentRepository
}

func NewDocumentProcessor(
    embeddingService *EmbeddingService,
    chunkDAO *qdrantdao.ChunkDAO,
    minioService *storage.MinIOService,
    documentRepo wiki_repo.DocumentRepository,
) *DocumentProcessor {
    return &DocumentProcessor{
        embeddingService: embeddingService,
        chunker:          NewChunker(500, 50),
        chunkDAO:         chunkDAO,
        minioService:     minioService,
        documentRepo:     documentRepo,
    }
}

// ProcessResult 处理结果
type ProcessResult struct {
    ChunkCount int
    Error      error
}

// ProcessDocument 处理单个文档
func (p *DocumentProcessor) ProcessDocument(ctx context.Context, docID, filePath, filename, folderID string) *ProcessResult {
    log.Printf("📄 开始处理文档: %s", docID)

    // 1. 从 MinIO 读取文件内容
    content, err := p.minioService.GetFileContent(ctx, filePath)
    if err != nil {
        log.Printf("❌ 读取文件内容失败: %v", err)
        return &ProcessResult{Error: fmt.Errorf("读取文件内容失败: %w", err)}
    }

    // 2. 文本分块
    chunks := p.chunker.ChunkText(content)
    if len(chunks) == 0 {
        log.Printf("⚠️  文档内容为空: %s", docID)
        return &ProcessResult{Error: fmt.Errorf("文档内容为空")}
    }
    log.Printf("📝 分块完成，共 %d 个 chunks", len(chunks))

    // 3. 生成 embeddings
    chunkTexts := make([]string, len(chunks))
    for i, chunk := range chunks {
        chunkTexts[i] = chunk.Text
    }

    vectors := make([][]float32, 0, len(chunkTexts))
    for i, text := range chunkTexts {
        emb, err := p.embeddingService.GetEmbedding(ctx, text)
        if err != nil {
            log.Printf("❌ 生成 embedding 失败 (chunk %d): %v", i, err)
            return &ProcessResult{Error: fmt.Errorf("生成 embedding 失败: %w", err)}
        }
        vectors = append(vectors, emb)
    }
    log.Printf("✅ 生成 %d 个 embeddings", len(vectors))

    // 4. 构建 chunk infos
    chunkInfos := make([]*qdrantdao.ChunkInfo, len(chunks))
    for i, chunk := range chunks {
        chunkInfos[i] = &qdrantdao.ChunkInfo{
            ChunkID:    uuid.New().String(),
            ChunkIndex: i,
            ChunkText:  chunk.Text,
            ChunkSize:  chunk.Size,
            DocumentID: docID,
            Filename:   filename,
            FolderID:   folderID,
        }
    }

    // 5. 写入 Qdrant
    if err := p.chunkDAO.UpsertChunks(ctx, chunkInfos, vectors); err != nil {
        log.Printf("❌ 写入 Qdrant 失败: %v", err)
        return &ProcessResult{Error: fmt.Errorf("写入 Qdrant 失败: %w", err)}
    }
    log.Printf("✅ 写入 Qdrant 成功")

    // 6. 更新文档状态
    if err := p.documentRepo.UpdateStatus(ctx, docID, wiki_db.DocumentStatusCompleted, len(chunks), nil); err != nil {
        log.Printf("⚠️  更新文档状态失败: %v", err)
    }

    log.Printf("✅ 文档处理完成: %s, chunks: %d", docID, len(chunks))
    return &ProcessResult{ChunkCount: len(chunks)}
}

// ReprocessDocument 重新处理文档（删除旧 chunks 后重新处理）
func (p *DocumentProcessor) ReprocessDocument(ctx context.Context, docID, filePath, filename, folderID string) *ProcessResult {
    // 1. 删除旧的 chunks
    if err := p.chunkDAO.DeleteChunksByDocumentID(ctx, docID); err != nil {
        log.Printf("⚠️  删除旧 chunks 失败: %v", err)
    }

    // 2. 重新处理
    return p.ProcessDocument(ctx, docID, filePath, filename, folderID)
}
```

- [ ] **Step 2: Commit**

```bash
git add server/app/ai/service/document_processor.go
git commit -m "feat: add document processor service integrating embedding and chunking"
```

---

## Task 5: Task Queue Worker 实现

**Files:**
- Modify: `server/infrastructure/queue/task_queue.go`

- [ ] **Step 1: 修改 TaskWorker 结构和构造函数**

```go
// task_queue.go

// TaskWorker 新增字段
type TaskWorker struct {
    server           *asynq.Server
    mux              *asynq.ServeMux
    dbService        *database.DatabaseService
    // 新增
    documentProcessor *ai_service.DocumentProcessor
}
```

- [ ] **Step 2: 修改 NewTaskWorker**

```go
// NewTaskWorker 新增 embeddingService 参数
func NewTaskWorker(
    redisAddr, redisPassword string,
    redisDB int,
    dbService *database.DatabaseService,
    embeddingService *ai_service.EmbeddingService,
    chunkDAO *qdrantdao.ChunkDAO,
    minioService *storage.MinIOService,
    documentRepo wiki_repo.DocumentRepository,
) *TaskWorker {
    // ... existing server setup ...

    worker := &TaskWorker{
        server:           server,
        mux:              mux,
        dbService:        dbService,
        documentProcessor: ai_service.NewDocumentProcessor(
            embeddingService,
            chunkDAO,
            minioService,
            documentRepo,
        ),
    }

    mux.HandleFunc(TypeDocumentProcess, worker.HandleDocumentProcess)
    return worker
}
```

- [ ] **Step 3: 实现 HandleDocumentProcess**

```go
// HandleDocumentProcess 实现文档处理
func (tw *TaskWorker) HandleDocumentProcess(ctx context.Context, task *asynq.Task) error {
    var payload DocumentProcessPayload
    if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return fmt.Errorf("解析任务载荷失败: %w", err)
    }

    log.Printf("🔄 开始处理文档任务: %s", payload.DocumentID)

    // 调用 document processor
    result := tw.documentProcessor.ReprocessDocument(
        ctx,
        payload.DocumentID,
        payload.FilePath,
        payload.Filename,
        payload.FolderID,
    )

    if result.Error != nil {
        // 更新状态为失败
        errMsg := result.Error.Error()
        tw.documentProcessor.DocumentRepo().UpdateStatus(
            ctx,
            payload.DocumentID,
            wiki_db.DocumentStatusFailed,
            0,
            &errMsg,
        )
        return result.Error
    }

    log.Printf("✅ 文档处理成功: %s, chunks: %d", payload.DocumentID, result.ChunkCount)
    return nil
}
```

- [ ] **Step 4: Commit**

```bash
git add server/infrastructure/queue/task_queue.go
git commit -m "feat: implement HandleDocumentProcess in task worker"
```

---

## Task 6: Handler 新增 API

**Files:**
- Modify: `server/app/wiki/handlers/document.go`
- Modify: `server/app/wiki/router/document.go`

- [ ] **Step 1: 在 document.go 新增 GetChunks 和 DeleteChunks**

```go
// handlers/document.go

// GetChunks 获取文档的 chunks
func (h *DocumentHandler) GetChunks(c *fiber.Ctx) error {
    docID := c.Params("id")

    // 验证文档存在
    _, err := h.documentRepository.FindOne(c.Context(), docID)
    if err != nil {
        return response.NotFoundCtx(c, "文档不存在")
    }

    // 从 Qdrant 获取 chunks
    chunks, err := h.chunkDAO.GetChunksByDocumentID(c.Context(), docID)
    if err != nil {
        log.Printf("❌ 获取 chunks 失败: %v", err)
        return response.InternalServerCtx(c, "获取 chunks 失败")
    }

    return response.SuccessCtx(c, fiber.Map{
        "document_id":  docID,
        "chunk_count":  len(chunks),
        "chunks":       chunks,
    })
}

// DeleteChunks 删除文档的 chunks
func (h *DocumentHandler) DeleteChunks(c *fiber.Ctx) error {
    docID := c.Params("id")

    // 验证文档存在
    _, err := h.documentRepository.FindOne(c.Context(), docID)
    if err != nil {
        return response.NotFoundCtx(c, "文档不存在")
    }

    // 从 Qdrant 删除 chunks
    if err := h.chunkDAO.DeleteChunksByDocumentID(c.Context(), docID); err != nil {
        log.Printf("❌ 删除 chunks 失败: %v", err)
        return response.InternalServerCtx(c, "删除 chunks 失败")
    }

    // 重置文档状态
    if err := h.documentRepository.UpdateStatus(c.Context(), docID, wiki_db.DocumentStatusPending, 0, nil); err != nil {
        log.Printf("⚠️  重置文档状态失败: %v", err)
    }

    return response.SuccessMsgCtx(c, "Chunks 已删除")
}
```

- [ ] **Step 2: 在 router 注册新路由**

```go
// router/document.go

docs := router.Group("/wiki/documents")
{
    // ... existing routes ...

    // 新增
    // 获取文档 chunks
    docs.Get("/:id/chunks", docHandler.GetChunks)
    // 删除文档 chunks
    docs.Delete("/:id/chunks", docHandler.DeleteChunks)
}
```

- [ ] **Step 3: Commit**

```bash
git add server/app/wiki/handlers/document.go server/app/wiki/router/document.go
git commit -m "feat: add GetChunks and DeleteChunks API endpoints"
```

---

## Task 7: 前端 API 层

**Files:**
- Modify: `web/src/api/wiki/document.ts`

- [ ] **Step 1: 新增 Chunk 类型和 API**

```typescript
// api/wiki/document.ts

export interface Chunk {
  chunk_id: string;
  chunk_index: number;
  chunk_text: string;
  chunk_size: number;
  document_id: string;
  filename: string;
  folder_id: string;
  vector_dim: number;
}

export interface ChunksResponse {
  document_id: string;
  chunk_count: number;
  chunks: Chunk[];
}

// 在 documentApi 中新增
export const documentApi = {
  // ... existing methods ...

  // 获取文档 chunks
  getChunks: (id: string) =>
    request.get<ChunksResponse>(`${RESOURCE_PATH}/${id}/chunks`),

  // 删除文档 chunks
  deleteChunks: (id: string) =>
    request.delete<{ message: string }>(`${RESOURCE_PATH}/${id}/chunks`),
};
```

- [ ] **Step 2: Commit**

```bash
git add web/src/api/wiki/document.ts
git commit -m "feat: add getChunks and deleteChunks API to frontend"
```

---

## Task 8: 前端 Chunk 展示

**Files:**
- Modify: `web/src/pages/wiki/documents/_components/data-table.tsx`
- Modify: `web/src/pages/wiki/documents/index.tsx`

- [ ] **Step 1: 在 DataTable 添加 chunk 列表展示**

在 `DataTable` 中:
1. 每行添加展开功能显示 chunk 数量
2. 展开后显示 chunk 列表
3. 每个 chunk 可展开查看完整文本

```tsx
// 展开行组件
const ChunkDetail = ({ docId }: { docId: string }) => {
  const { data: chunksData, isLoading } = useQuery({
    queryKey: ['chunks', docId],
    queryFn: () => documentApi.getChunks(docId),
    enabled: !!docId,
  });

  if (isLoading) return <div>加载中...</div>;
  if (!chunksData?.chunks?.length) return <div>暂无 chunk 数据</div>;

  return (
    <div className="space-y-2 p-2">
      {chunksData.chunks.map((chunk) => (
        <div key={chunk.chunk_id} className="border rounded p-2 text-sm">
          <div className="flex justify-between items-start">
            <span className="font-medium">Chunk {chunk.chunk_index + 1}</span>
            <span className="text-muted-foreground text-xs">
              {chunk.chunk_size} 字符 | {chunk.vector_dim} 维
            </span>
          </div>
          <p className="mt-1 text-muted-foreground truncate">
            {chunk.chunk_text}
          </p>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="link" size="sm" className="mt-1 h-auto p-0">
                查看完整内容
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>Chunk {chunk.chunk_index + 1}</DialogTitle>
              </DialogHeader>
              <div className="mt-4">
                <p className="whitespace-pre-wrap text-sm">{chunk.chunk_text}</p>
                <div className="mt-4 text-xs text-muted-foreground">
                  <p>字符数: {chunk.chunk_size}</p>
                  <p>向量维度: {chunk.vector_dim}</p>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      ))}
    </div>
  );
};
```

- [ ] **Step 2: 在表格行中添加展开触发器**

在处理状态列添加展开按钮:
```tsx
{processingIds.has(row.id) ? (
  <Button variant="ghost" size="sm" onClick={() => toggleExpand(row.id)}>
    <IconDots />
  </Button>
) : (
  <Button variant="ghost" size="sm" onClick={() => toggleExpand(row.id)}>
    <IconChevronDown />
  </Button>
)}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/wiki/documents/_components/data-table.tsx web/src/pages/wiki/documents/index.tsx
git commit -m "feat: display chunk list with expand detail view in documents page"
```

---

## 实现顺序

1. **Task 1** - Embedding Service
2. **Task 2** - Chunker Service
3. **Task 3** - Qdrant Chunk DAO
4. **Task 4** - Document Processor
5. **Task 5** - Task Queue Worker (依赖 1,2,3,4)
6. **Task 6** - Handler API (依赖 3)
7. **Task 7** - 前端 API
8. **Task 8** - 前端展示
