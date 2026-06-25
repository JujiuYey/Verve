package handlers

import (
	"io"
	"log"

	"github.com/gofiber/fiber/v2"

	"sag-wiki/common/response"
	"sag-wiki/infrastructure/storage"
)

// 文件处理器
type FileHandler struct {
	minioService *storage.MinIOService
}

// 创建文件处理器
func NewFileHandler(minioService *storage.MinIOService) *FileHandler {
	return &FileHandler{
		minioService: minioService,
	}
}

// 获取文件（代理访问 MinIO）
func (h *FileHandler) GetFile(c *fiber.Ctx) error {
	// 获取文件路径参数（例如：avatars/xxx.jpg）
	filePath := c.Params("filepath")

	if filePath == "" {
		return response.BadRequestCtx(c, "文件路径不能为空")
	}

	log.Printf("📥 请求文件: %s", filePath)

	// 从 MinIO 获取文件
	object, err := h.minioService.GetFile(c.Context(), filePath)
	if err != nil {
		log.Printf("❌ 获取文件失败: %v", err)
		return response.NotFoundCtx(c, "文件不存在")
	}
	defer object.Close()

	// 获取文件信息
	stat, err := object.Stat()
	if err != nil {
		log.Printf("❌ 获取文件信息失败: %v", err)
		return response.InternalServerCtx(c, "获取文件信息失败")
	}

	// 设置响应头
	c.Set("Content-Type", stat.ContentType)
	c.Set("Cache-Control", "public, max-age=31536000") // 缓存 1 年

	// 流式传输文件内容
	_, err = io.Copy(c.Response().BodyWriter(), object)
	if err != nil {
		log.Printf("❌ 传输文件失败: %v", err)
		return nil
	}

	log.Printf("✅ 文件传输成功: %s", filePath)
	return nil
}

// 预览文件
func (h *FileHandler) PreviewFile(c *fiber.Ctx) error {
	return h.GetFile(c)
}
