package qdrant

import (
	"context"
	"fmt"

	qdrantpb "github.com/qdrant/go-client/qdrant"
)

// CollectionInfo collection 信息
type CollectionInfo struct {
	Name             string `json:"name"`
	VectorsCount     uint64 `json:"vectors_count"`
	PointsCount      uint64 `json:"points_count"`
	SegmentsCount    uint64 `json:"segments_count"`
	VectorSize       uint64 `json:"vector_size"`
	Status           string `json:"status"`
	DistanceFunction string `json:"distance_function"`
	CreatedAt        string `json:"created_at,omitempty"`
}

// PointInfo point 信息
type PointInfo struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score,omitempty"`
	Payload map[string]interface{} `json:"payload"`
	Vector  []float32              `json:"vector,omitempty"`
}

// CollectionService collection 服务
type CollectionService struct {
	client *QdrantClient
}

// NewCollectionService 创建 collection 服务
func NewCollectionService() (*CollectionService, error) {
	client, err := NewQdrantClient()
	if err != nil {
		return nil, err
	}
	return &CollectionService{client: client}, nil
}

// ListCollections 获取所有 collection
func (s *CollectionService) ListCollections(ctx context.Context) ([]*CollectionInfo, error) {
	// 使用 Client 的便捷方法获取 collection 名称列表
	collectionNames, err := s.client.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 collection 列表失败: %w", err)
	}

	result := make([]*CollectionInfo, 0, len(collectionNames))
	for _, collectionName := range collectionNames {
		info, err := s.GetCollectionInfo(ctx, collectionName)
		if err != nil {
			result = append(result, &CollectionInfo{
				Name:   collectionName,
				Status: "unknown",
			})
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// GetCollectionInfo 获取 collection 详情
func (s *CollectionService) GetCollectionInfo(ctx context.Context, collectionName string) (*CollectionInfo, error) {
	collectionsClient := s.client.client.GetCollectionsClient()

	resp, err := collectionsClient.Get(ctx, &qdrantpb.GetCollectionInfoRequest{
		CollectionName: collectionName,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 collection 详情失败: %w", err)
	}

	info := resp.GetResult()
	if info == nil {
		return nil, fmt.Errorf("collection 不存在: %s", collectionName)
	}

	// 从 Config 中获取 Params，再从 Params 中获取 VectorsConfig
	var vectorSize uint64
	var distance string = "Cosine"

	config := info.GetConfig()
	if config != nil {
		params := config.GetParams()
		if params != nil {
			vectorsConfig := params.GetVectorsConfig()
			if vectorsConfig != nil {
				vectorParams := vectorsConfig.GetParams()
				if vectorParams != nil {
					vectorSize = vectorParams.GetSize()
					switch vectorParams.GetDistance() {
					case qdrantpb.Distance_Cosine:
						distance = "Cosine"
					case qdrantpb.Distance_Euclid:
						distance = "Euclidean"
					case qdrantpb.Distance_Dot:
						distance = "Dot"
					}
				}
			}
		}
	}

	return &CollectionInfo{
		Name:             collectionName,
		VectorsCount:     info.GetVectorsCount(),
		PointsCount:      info.GetPointsCount(),
		SegmentsCount:    info.GetSegmentsCount(),
		VectorSize:       vectorSize,
		Status:           info.GetStatus().String(),
		DistanceFunction: distance,
	}, nil
}

// CreateCollection 创建 collection
func (s *CollectionService) CreateCollection(ctx context.Context, name string, vectorSize uint64, distance string) error {
	var distanceType qdrantpb.Distance
	switch distance {
	case "Cosine":
		distanceType = qdrantpb.Distance_Cosine
	case "Euclidean":
		distanceType = qdrantpb.Distance_Euclid
	case "Dot":
		distanceType = qdrantpb.Distance_Dot
	default:
		distanceType = qdrantpb.Distance_Cosine
	}

	collectionsClient := s.client.client.GetCollectionsClient()

	_, err := collectionsClient.Create(ctx, &qdrantpb.CreateCollection{
		CollectionName: name,
		VectorsConfig: &qdrantpb.VectorsConfig{
			Config: &qdrantpb.VectorsConfig_Params{
				Params: &qdrantpb.VectorParams{
					Size:     vectorSize,
					Distance: distanceType,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("创建 collection 失败: %w", err)
	}

	return nil
}

// DeleteCollection 删除 collection
func (s *CollectionService) DeleteCollection(ctx context.Context, name string) error {
	collectionsClient := s.client.client.GetCollectionsClient()

	_, err := collectionsClient.Delete(ctx, &qdrantpb.DeleteCollection{
		CollectionName: name,
	})
	if err != nil {
		return fmt.Errorf("删除 collection 失败: %w", err)
	}

	return nil
}

// GetPoints 获取 collection 中的 points（分页）
func (s *CollectionService) GetPoints(ctx context.Context, collectionName string, offset *qdrantpb.PointId, limit uint64) ([]*PointInfo, *qdrantpb.PointId, error) {
	pointsClient := s.client.client.GetPointsClient()

	limitVal := uint32(limit)

	scrollReq := &qdrantpb.ScrollPoints{
		CollectionName: collectionName,
		Limit:         &limitVal,
		WithPayload:   &qdrantpb.WithPayloadSelector{SelectorOptions: &qdrantpb.WithPayloadSelector_Enable{Enable: true}},
	}

	if offset != nil {
		scrollReq.Offset = offset
	}

	resp, err := pointsClient.Scroll(ctx, scrollReq)
	if err != nil {
		return nil, nil, fmt.Errorf("获取 points 失败: %w", err)
	}

	result := make([]*PointInfo, 0, len(resp.GetResult()))
	for _, point := range resp.GetResult() {
		pointInfo := &PointInfo{
			Payload: make(map[string]interface{}),
		}

		if point.Id != nil {
			if uuid, ok := point.Id.PointIdOptions.(*qdrantpb.PointId_Uuid); ok {
				pointInfo.ID = uuid.Uuid
			}
		}

		if point.Payload != nil {
			for k, v := range point.Payload {
				pointInfo.Payload[k] = s.client.extractValue(v)
			}
		}

		result = append(result, pointInfo)
	}

	return result, resp.GetNextPageOffset(), nil
}

// GetCollectionStats 获取 collection 统计信息
func (s *CollectionService) GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	info, err := s.GetCollectionInfo(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":              info.Name,
		"vectors_count":     info.VectorsCount,
		"points_count":      info.PointsCount,
		"segments_count":    info.SegmentsCount,
		"vector_size":       info.VectorSize,
		"status":            info.Status,
		"distance_function": info.DistanceFunction,
	}, nil
}
