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
    Score       float32 // 搜索相似度分数
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
                    if f, ok := val.(int64); ok {
                        chunkInfo.ChunkIndex = int(f)
                    }
                case "chunk_text":
                    chunkInfo.ChunkText, _ = val.(string)
                case "chunk_size":
                    if f, ok := val.(int64); ok {
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

// SearchChunksByVector 根据向量搜索最相似的 chunks
// queryVector: 查询向量
// filter: 可选的 Qdrant filter 条件（用于限定 folder_id 或 document_id）
// limit: 返回结果数量限制
func (d *ChunkDAO) SearchChunksByVector(ctx context.Context, queryVector []float32, filter *qdrantpb.Filter, limit uint64) ([]*ChunkInfo, error) {
    searchReq := &qdrantpb.SearchPoints{
        CollectionName: d.collection,
        Vector:         queryVector,
        Limit:          limit,
        WithPayload:    &qdrantpb.WithPayloadSelector{SelectorOptions: &qdrantpb.WithPayloadSelector_Enable{Enable: true}},
        Filter:         filter,
    }

    pointsClient := d.client.client.GetPointsClient()
    searchResp, err := pointsClient.Search(ctx, searchReq)
    if err != nil {
        return nil, fmt.Errorf("向量搜索 chunks 失败: %w", err)
    }

    chunks := make([]*ChunkInfo, 0, len(searchResp.Result))
    for _, point := range searchResp.Result {
        chunkInfo := &ChunkInfo{
            VectorDim: VectorSize,
            Score:     point.Score,
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
                    if f, ok := val.(int64); ok {
                        chunkInfo.ChunkIndex = int(f)
                    }
                case "chunk_text":
                    chunkInfo.ChunkText, _ = val.(string)
                case "chunk_size":
                    if f, ok := val.(int64); ok {
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
        Points:        qdrantpb.NewPointsSelectorIDs(pointIDs),
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
