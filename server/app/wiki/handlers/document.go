package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"

	wiki_db "sag-wiki/app/wiki/models/db"
	wiki_payload "sag-wiki/app/wiki/models/payload"
	wiki_repo "sag-wiki/app/wiki/repository"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
	"sag-wiki/infrastructure/storage"
)

type documentRepository interface {
	Page(ctx context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error)
	List(ctx context.Context, name, folderID string) ([]*wiki_db.Document, error)
	FindOne(ctx context.Context, id string) (*wiki_db.Document, error)
	Create(ctx context.Context, folderID string, filename string, fileSize int64, filePath string) (*wiki_db.Document, error)
	UpdateFileSize(ctx context.Context, docID string, fileSize int64) error
	Delete(ctx context.Context, id string) error
}

// 文档处理器
type DocumentHandler struct {
	documentRepository documentRepository
	minioService       *storage.MinIOService
}

// 创建文档处理器
func NewDocumentHandler(dbService *database.DatabaseService, minioService *storage.MinIOService) *DocumentHandler {
	return &DocumentHandler{
		documentRepository: wiki_repo.NewDocumentRepository(dbService.GetDB()),
		minioService:       minioService,
	}
}

// 获取文档列表
func (h *DocumentHandler) FindPage(c *fiber.Ctx) error {
	var req wiki_payload.PageDocumentsRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c, "参数错误: "+err.Error())
	}

	// 验证并修正分页参数
	req.Validate()

	docs, total, err := h.documentRepository.Page(
		c.Context(),
		req.PageSize,    // 使用 PageSize 替代 Limit
		req.GetOffset(), // 使用 GetOffset() 替代 Offset
		req.Name,
		req.FolderID,
	)
	if err != nil {
		return response.InternalServerCtx(c, "获取文档列表失败")
	}

	// 返回统一分页响应
	return response.PaginateCtx(c, docs, total, req.Page, req.PageSize)
}

// 获取文档列表（不分页）
func (h *DocumentHandler) FindList(c *fiber.Ctx) error {
	var req wiki_payload.DocumentListRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c, "参数错误: "+err.Error())
	}

	docs, err := h.documentRepository.List(
		c.Context(),
		req.Name,
		req.FolderID,
	)
	if err != nil {
		return response.InternalServerCtx(c, "获取文档列表失败")
	}

	return response.SuccessCtx(c, docs)
}

// 获取文档详情
func (h *DocumentHandler) FindOne(c *fiber.Ctx) error {
	docID := c.Params("id")
	doc, err := h.documentRepository.FindOne(c.Context(), docID)
	if err != nil {
		return response.NotFoundCtx(c, "文档不存在")
	}
	return response.SuccessCtx(c, doc)
}

// 上传文档接口（仅上传，不处理）
func (h *DocumentHandler) Upload(c *fiber.Ctx) error {
	log.Println("📤 收到文件上传请求")

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("❌ 文件上传失败: %v", err)
		return response.BadRequestCtx(c, "文件上传失败: "+err.Error())
	}
	log.Printf("📄 接收到文件: %s (大小: %d bytes)", file.Filename, file.Size)

	folderID := c.FormValue("folder_id")
	if folderID == "" {
		log.Printf("❌ 缺少文件夹ID")
		return response.BadRequestCtx(c, "缺少文件夹ID")
	}

	// 生成唯一的文件路径：documents/{uuid}/{filename}
	objectName := fmt.Sprintf("documents/%s/%s", uuid.New().String(), file.Filename)

	// 打开文件
	f, err := file.Open()
	if err != nil {
		log.Printf("❌ 无法打开文件: %v", err)
		return response.InternalServerCtx(c, "无法打开文件: "+err.Error())
	}
	defer f.Close()

	// 1. 上传到 MinIO
	log.Println("☁️  正在上传文件到 MinIO...")
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	err = h.minioService.UploadFile(c.Context(), objectName, f, file.Size, contentType)
	if err != nil {
		log.Printf("❌ 上传到 MinIO 失败: %v", err)
		return response.InternalServerCtx(c, "上传文件失败: "+err.Error())
	}

	// 2. 创建数据库记录（使用预生成的 UUID）
	doc, err := h.documentRepository.Create(c.Context(), folderID, file.Filename, file.Size, objectName)
	if err != nil {
		log.Printf("❌ 创建文档记录失败: %v", err)
		// 回滚：删除已上传的文件
		h.minioService.DeleteFile(c.Context(), objectName)
		return response.InternalServerCtx(c, "创建文档记录失败: "+err.Error())
	}
	log.Printf("✅ 文档上传成功，ID: %s", doc.ID)

	return response.SuccessCtx(c, fiber.Map{
		"message":     "文档上传成功",
		"filename":    file.Filename,
		"document_id": doc.ID,
		"file_path":   objectName,
	})
}

// 下载文档接口
func (h *DocumentHandler) Download(c *fiber.Ctx) error {
	docID := c.Params("id")

	// 获取文档记录
	doc, err := h.documentRepository.FindOne(c.Context(), docID)
	if err != nil {
		return response.NotFoundCtx(c, "文档不存在")
	}

	// 生成预签名 URL
	url, err := h.minioService.GetPresignedURL(c.Context(), doc.FilePath)
	if err != nil {
		return response.InternalServerCtx(c, "生成下载链接失败")
	}

	return response.SuccessCtx(c, fiber.Map{
		"download_url": url,
		"filename":     doc.Filename,
		"expires_in":   "1 hour",
	})
}

// 获取文档内容
func (h *DocumentHandler) GetContent(c *fiber.Ctx) error {
	docID := c.Params("id")

	// 获取文档记录
	doc, err := h.documentRepository.FindOne(c.Context(), docID)
	if err != nil {
		return response.NotFoundCtx(c, "文档不存在")
	}

	// 从 MinIO 读取文件内容
	content, err := h.minioService.GetFileContent(c.Context(), doc.FilePath)
	if err != nil {
		log.Printf("❌ 读取文件内容失败: %v", err)
		return response.InternalServerCtx(c, "读取文件内容失败")
	}

	return response.SuccessCtx(c, fiber.Map{
		"content":  content,
		"filename": doc.Filename,
	})
}

// 更新文档内容
func (h *DocumentHandler) UpdateContent(c *fiber.Ctx) error {
	docID := c.Params("id")

	// 解析请求体
	var req wiki_payload.UpdateContentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, "参数错误: "+err.Error())
	}

	// 获取文档记录
	doc, err := h.documentRepository.FindOne(c.Context(), docID)
	if err != nil {
		return response.NotFoundCtx(c, "文档不存在")
	}

	// 写入 MinIO
	if err := h.minioService.PutFileContent(c.Context(), doc.FilePath, req.Content); err != nil {
		log.Printf("❌ 写入文件内容失败: %v", err)
		return response.InternalServerCtx(c, "保存文件内容失败")
	}

	// 更新文件大小
	fileSize := int64(len(req.Content))
	if err := h.documentRepository.UpdateFileSize(c.Context(), docID, fileSize); err != nil {
		log.Printf("⚠️  更新文件大小失败: %v", err)
	}

	log.Printf("✅ 文档内容已更新: %s", docID)
	return response.SuccessMsgCtx(c, "文档内容已保存")
}

// 删除文档接口
func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	docID := c.Params("id")

	// 获取文档记录
	doc, err := h.documentRepository.FindOne(c.Context(), docID)
	if err != nil {
		return response.NotFoundCtx(c, "文档不存在")
	}

	// 删除 MinIO 中的文件
	if err := h.minioService.DeleteFile(c.Context(), doc.FilePath); err != nil {
		log.Printf("⚠️  删除 MinIO 文件失败: %v", err)
	}

	// 软删除数据库记录
	if err := h.documentRepository.Delete(c.Context(), docID); err != nil {
		return response.InternalServerCtx(c, "删除文档失败")
	}

	return response.SuccessMsgCtx(c, "文档删除成功")
}
