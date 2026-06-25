package handlers

import (
	"log"
	"strconv"

	"sag-wiki/common/response"
	"sag-wiki/infrastructure/qdrant"

	"github.com/gofiber/fiber/v2"
)

// CollectionHandler collection 处理器
type CollectionHandler struct {
	service *qdrant.CollectionService
}

// NewCollectionHandler 创建 collection 处理器
func NewCollectionHandler() (*CollectionHandler, error) {
	service, err := qdrant.NewCollectionService()
	if err != nil {
		return nil, err
	}
	return &CollectionHandler{service: service}, nil
}

// List 获取所有 collection
func (h *CollectionHandler) List(c *fiber.Ctx) error {
	collections, err := h.service.ListCollections(c.Context())
	if err != nil {
		log.Printf("List collections error: %v", err)
		return response.InternalServerCtx(c, "获取 collection 列表失败")
	}

	return response.SuccessCtx(c, collections)
}

// Get 获取单个 collection 详情
func (h *CollectionHandler) Get(c *fiber.Ctx) error {
	name := c.Params("name")

	info, err := h.service.GetCollectionInfo(c.Context(), name)
	if err != nil {
		log.Printf("Get collection info error: %v", err)
		return response.InternalServerCtx(c, "获取 collection 详情失败")
	}

	return response.SuccessCtx(c, info)
}

// Create 创建 collection
func (h *CollectionHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name       string `json:"name"`
		VectorSize uint64 `json:"vector_size"`
		Distance   string `json:"distance"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, "无效的请求参数")
	}

	if req.Name == "" {
		return response.BadRequestCtx(c, "collection 名称不能为空")
	}

	if req.VectorSize == 0 {
		req.VectorSize = 1024 // 默认 1024 维
	}

	if req.Distance == "" {
		req.Distance = "Cosine"
	}

	if err := h.service.CreateCollection(c.Context(), req.Name, req.VectorSize, req.Distance); err != nil {
		log.Printf("Create collection error: %v", err)
		return response.InternalServerCtx(c, "创建 collection 失败")
	}

	return response.SuccessCtx(c, map[string]string{
		"message": "Collection 创建成功",
	})
}

// Delete 删除 collection
func (h *CollectionHandler) Delete(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := h.service.DeleteCollection(c.Context(), name); err != nil {
		log.Printf("Delete collection error: %v", err)
		return response.InternalServerCtx(c, "删除 collection 失败")
	}

	return response.SuccessCtx(c, map[string]string{
		"message": "Collection 删除成功",
	})
}

// GetPoints 获取 collection 中的 points
func (h *CollectionHandler) GetPoints(c *fiber.Ctx) error {
	name := c.Params("name")

	// 获取分页参数
	limitStr := c.Query("limit", "20")
	limit, err := strconv.ParseUint(limitStr, 10, 32)
	if err != nil || limit == 0 {
		limit = 20
	}

	// 调用服务获取 points（暂时不支持 offset 分页）
	points, _, err := h.service.GetPoints(c.Context(), name, nil, limit)
	if err != nil {
		log.Printf("Get points error: %v", err)
		return response.InternalServerCtx(c, "获取 points 失败")
	}

	return response.SuccessCtx(c, fiber.Map{
		"points": points,
	})
}

// GetStats 获取 collection 统计信息
func (h *CollectionHandler) GetStats(c *fiber.Ctx) error {
	name := c.Params("name")

	stats, err := h.service.GetCollectionStats(c.Context(), name)
	if err != nil {
		log.Printf("Get collection stats error: %v", err)
		return response.InternalServerCtx(c, "获取 collection 统计信息失败")
	}

	return response.SuccessCtx(c, stats)
}
