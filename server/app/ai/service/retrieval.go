// retrieval.go
package service

import (
	"context"

	qdrantpb "github.com/qdrant/go-client/qdrant"
	qdrantdao "sag-wiki/infrastructure/qdrant"
	wiki_repo "sag-wiki/app/wiki/repository"
)

// RetrievalService RAG 检索服务
type RetrievalService struct {
	embeddingService *EmbeddingService
	chunkDAO         *qdrantdao.ChunkDAO
	folderExpander   *wiki_repo.FolderExpander
	documentRepo     wiki_repo.DocumentRepository
}

// NewRetrievalService 创建检索服务
func NewRetrievalService(
	embeddingService *EmbeddingService,
	chunkDAO *qdrantdao.ChunkDAO,
	folderExpander *wiki_repo.FolderExpander,
	documentRepo wiki_repo.DocumentRepository,
) *RetrievalService {
	return &RetrievalService{
		embeddingService: embeddingService,
		chunkDAO:         chunkDAO,
		folderExpander:   folderExpander,
		documentRepo:    documentRepo,
	}
}

// SearchResult 检索结果
type SearchResult struct {
	ChunkInfo *qdrantdao.ChunkInfo
	Score     float32
}

// SearchRequest 检索请求
type SearchRequest struct {
	Query      string // 用户问题
	FolderID   string // 可选：限定文件夹
	DocumentID string // 可选：限定文档
	UserID     string // 用户ID（用于权限过滤）
	Limit      int    // 返回 chunks 数量限制
}

// Search 执行 RAG 检索
func (s *RetrievalService) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
	// 1. 调用 EmbeddingService.GetEmbedding 将 query 转为向量
	queryVector, err := s.embeddingService.GetEmbedding(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	// 2. 构建 Qdrant filter
	var filter *qdrantpb.Filter

	if req.DocumentID != "" {
		// 3. 如果指定了 DocumentID，直接使用该 document
		filter = &qdrantpb.Filter{
			Must: []*qdrantpb.Condition{
				{
					ConditionOneOf: &qdrantpb.Condition_Field{
						Field: &qdrantpb.FieldCondition{
							Key: "document_id",
							Match: &qdrantpb.Match{
								MatchValue: &qdrantpb.Match_Keyword{Keyword: req.DocumentID},
							},
						},
					},
				},
			},
		}
	} else if req.FolderID != "" {
		// 4. 如果指定了 FolderID，调用 folderExpander.ExpandAndFilter 获取有权限的文件夹列表
		folderIDs, err := s.folderExpander.ExpandAndFilter(ctx, req.FolderID, req.UserID)
		if err != nil {
			return nil, err
		}

		if len(folderIDs) == 0 {
			// 没有权限访问任何文件夹，返回空结果
			return []*SearchResult{}, nil
		}

		// 构建 folder_id IN (...) filter - 使用 Should + 多个 Must 条件
		conditions := make([]*qdrantpb.Condition, 0, len(folderIDs))
		for _, fid := range folderIDs {
			conditions = append(conditions, &qdrantpb.Condition{
				ConditionOneOf: &qdrantpb.Condition_Field{
					Field: &qdrantpb.FieldCondition{
						Key: "folder_id",
						Match: &qdrantpb.Match{
							MatchValue: &qdrantpb.Match_Keyword{Keyword: fid},
						},
					},
				},
			})
		}

		filter = &qdrantpb.Filter{
			Should: conditions,
		}
	}

	// 5. 调用 chunkDAO.SearchChunksByVector 执行向量检索
	limit := uint64(req.Limit)
	if limit == 0 {
		limit = 10 // 默认返回 10 条
	}

	chunks, err := s.chunkDAO.SearchChunksByVector(ctx, queryVector, filter, limit)
	if err != nil {
		return nil, err
	}

	// 6. 构建返回结果（按 score 排序）
	results := make([]*SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		results = append(results, &SearchResult{
			ChunkInfo: chunk,
			Score:     chunk.Score,
		})
	}

	return results, nil
}
