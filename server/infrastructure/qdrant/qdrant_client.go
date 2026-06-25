package qdrant

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"sag-wiki/config"

	"github.com/google/uuid"
	qdrantpb "github.com/qdrant/go-client/qdrant"
)

const (
	CollectionName = "documents"
	VectorSize     = 1024
)

// 导出 Qdrant PointStruct 类型别名
type PointStruct = qdrantpb.PointStruct

// 搜索结果
type SearchResult struct {
	ID      string
	Score   float32
	Payload map[string]interface{}
}

// Qdrant 客户端包装器（使用官方 go-client）
type QdrantClient struct {
	client     *qdrantpb.Client
	collection string
}

// 创建 Qdrant 客户端
func NewQdrantClient() (*QdrantClient, error) {
	qdrantConfig := config.GetQdrantConfig()

	// 解析 URL
	u, err := url.Parse(qdrantConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("解析 Qdrant URL 失败: %w", err)
	}

	// 解析端口
	port := 6334 // 默认 gRPC 端口
	if u.Port() != "" {
		if p, err := strconv.Atoi(u.Port()); err == nil {
			if p == 6333 {
				port = 6334 // HTTP 端口转 gRPC 端口
			} else {
				port = p
			}
		}
	}

	// 创建客户端
	client, err := qdrantpb.NewClient(&qdrantpb.Config{
		Host: u.Hostname(),
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Qdrant 客户端失败: %w", err)
	}

	qdrantClient := &QdrantClient{
		client:     client,
		collection: CollectionName,
	}

	// 自动创建 collection
	if err := qdrantClient.EnsureCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("初始化 Qdrant collection 失败: %w", err)
	}

	return qdrantClient, nil
}

// 确保集合存在
func (c *QdrantClient) EnsureCollection(ctx context.Context) error {
	// 检查集合是否存在
	exists, err := c.client.CollectionExists(ctx, c.collection)
	if err != nil {
		return fmt.Errorf("检查集合失败: %w", err)
	}

	if !exists {
		// 创建集合
		err = c.client.CreateCollection(ctx, &qdrantpb.CreateCollection{
			CollectionName: c.collection,
			VectorsConfig: &qdrantpb.VectorsConfig{
				Config: &qdrantpb.VectorsConfig_Params{
					Params: &qdrantpb.VectorParams{
						Size:     VectorSize,
						Distance: qdrantpb.Distance_Cosine,
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}
	}

	return nil
}

// 使用向量进行搜索
func (c *QdrantClient) SearchByVector(ctx context.Context, vector []float32, limit uint64, filters map[string]interface{}) ([]SearchResult, error) {
	// 构建搜索请求
	searchReq := &qdrantpb.SearchPoints{
		CollectionName: c.collection,
		Vector:         vector,
		Limit:          limit,
		WithPayload:    &qdrantpb.WithPayloadSelector{SelectorOptions: &qdrantpb.WithPayloadSelector_Enable{Enable: true}},
	}

	// 构建过滤条件
	if filters != nil {
		searchReq.Filter = c.buildFilter(filters)
	}

	// 获取 PointsClient 并执行搜索
	pointsClient := c.client.GetPointsClient()
	searchResp, err := pointsClient.Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换结果
	searchResult := searchResp.GetResult()
	results := make([]SearchResult, 0, len(searchResult))
	for _, point := range searchResult {
		payload := make(map[string]interface{})
		if point.Payload != nil {
			for k, v := range point.Payload {
				payload[k] = c.extractValue(v)
			}
		}

		// 提取 ID
		var id string
		if point.Id != nil {
			switch v := point.Id.PointIdOptions.(type) {
			case *qdrantpb.PointId_Uuid:
				id = v.Uuid
			case *qdrantpb.PointId_Num:
				id = fmt.Sprintf("%d", v.Num)
			default:
				id = fmt.Sprintf("%v", point.Id)
			}
		}

		results = append(results, SearchResult{
			ID:      id,
			Score:   point.Score,
			Payload: payload,
		})
	}

	return results, nil
}

// 插入或更新向量点
func (c *QdrantClient) UpsertPoints(ctx context.Context, points []*PointStruct) error {
	pointsClient := c.client.GetPointsClient()
	wait := true
	_, err := pointsClient.Upsert(ctx, &qdrantpb.UpsertPoints{
		CollectionName: c.collection,
		Points:         points,
		Wait:           &wait,
	})
	if err != nil {
		return fmt.Errorf("插入向量点失败: %w", err)
	}
	return nil
}

// 创建向量点
func CreatePoint(id string, vector []float32, payload map[string]interface{}) *PointStruct {
	// 构建 payload
	qdrantPayload := make(map[string]*qdrantpb.Value)
	for k, v := range payload {
		qdrantPayload[k] = toQdrantValue(v)
	}

	pointID := &qdrantpb.PointId{
		PointIdOptions: &qdrantpb.PointId_Uuid{Uuid: id},
	}

	return &PointStruct{
		Id:      pointID,
		Vectors: &qdrantpb.Vectors{VectorsOptions: &qdrantpb.Vectors_Vector{Vector: &qdrantpb.Vector{Data: vector}}},
		Payload: qdrantPayload,
	}
}

// buildFilter 构建过滤条件
func (c *QdrantClient) buildFilter(filters map[string]interface{}) *qdrantpb.Filter {
	conditions := make([]*qdrantpb.Condition, 0)

	if must, ok := filters["must"].([]map[string]interface{}); ok {
		for _, cond := range must {
			if key, ok := cond["key"].(string); ok {
				if match, ok := cond["match"].(map[string]interface{}); ok {
					if value, ok := match["value"]; ok {
						// 目前只支持字符串匹配（使用 Keyword）
						if strValue, ok := value.(string); ok {
							matchValue := &qdrantpb.Match{
								MatchValue: &qdrantpb.Match_Keyword{Keyword: strValue},
							}
							conditions = append(conditions, &qdrantpb.Condition{
								ConditionOneOf: &qdrantpb.Condition_Field{
									Field: &qdrantpb.FieldCondition{
										Key:   key,
										Match: matchValue,
									},
								},
							})
						}
					}
				}
			}
		}
	}

	if len(conditions) == 0 {
		return nil
	}

	return &qdrantpb.Filter{
		Must: conditions,
	}
}

// extractValue 提取 Qdrant Value 的值
func (c *QdrantClient) extractValue(v *qdrantpb.Value) interface{} {
	switch val := v.Kind.(type) {
	case *qdrantpb.Value_StringValue:
		return val.StringValue
	case *qdrantpb.Value_IntegerValue:
		return val.IntegerValue
	case *qdrantpb.Value_DoubleValue:
		return val.DoubleValue
	case *qdrantpb.Value_BoolValue:
		return val.BoolValue
	case *qdrantpb.Value_ListValue:
		list := make([]interface{}, len(val.ListValue.Values))
		for i, item := range val.ListValue.Values {
			list[i] = c.extractValue(item)
		}
		return list
	default:
		return nil
	}
}

// toQdrantValue 转换为 Qdrant Value
func toQdrantValue(v interface{}) *qdrantpb.Value {
	switch val := v.(type) {
	case string:
		// 清理无效 UTF-8 字符
		sanitized := strings.ToValidUTF8(val, "")
		return &qdrantpb.Value{Kind: &qdrantpb.Value_StringValue{StringValue: sanitized}}
	case int:
		return &qdrantpb.Value{Kind: &qdrantpb.Value_IntegerValue{IntegerValue: int64(val)}}
	case int64:
		return &qdrantpb.Value{Kind: &qdrantpb.Value_IntegerValue{IntegerValue: val}}
	case float32:
		return &qdrantpb.Value{Kind: &qdrantpb.Value_DoubleValue{DoubleValue: float64(val)}}
	case float64:
		return &qdrantpb.Value{Kind: &qdrantpb.Value_DoubleValue{DoubleValue: val}}
	case bool:
		return &qdrantpb.Value{Kind: &qdrantpb.Value_BoolValue{BoolValue: val}}
	default:
		return &qdrantpb.Value{Kind: &qdrantpb.Value_StringValue{StringValue: fmt.Sprintf("%v", val)}}
	}
}

// 生成新的 ID
func GenerateID() string {
	return uuid.New().String()
}
