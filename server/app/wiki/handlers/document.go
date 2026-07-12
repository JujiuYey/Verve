package handlers

import (
	"context"
	"io"
	"log"

	"github.com/gofiber/fiber/v2"

	rag_db "verve/app/rag/models/db"
	rag_service "verve/app/rag/service"
	wiki_db "verve/app/wiki/models/db"
	wiki_payload "verve/app/wiki/models/payload"
	wiki_repo "verve/app/wiki/repository"
	wiki_service "verve/app/wiki/service"
	"verve/common/response"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
)

type documentRepository interface {
	Page(ctx context.Context, pageSize int, offset int, name, folderID string) ([]*wiki_db.Document, int, error)
	List(ctx context.Context, name, folderID string) ([]*wiki_db.Document, error)
	FindOne(ctx context.Context, id string) (*wiki_db.Document, error)
	Create(ctx context.Context, folderID string, filename string, fileSize int64, filePath string) (*wiki_db.Document, error)
	UpdateFileSize(ctx context.Context, docID string, fileSize int64) error
	Delete(ctx context.Context, id string) error
	DeleteWithChunks(ctx context.Context, id string) error
}

type documentFileStore interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, objectName string) (string, error)
	GetFileContent(ctx context.Context, objectName string) (string, error)
	PutFileContent(ctx context.Context, objectName string, content string) error
	DeleteFile(ctx context.Context, objectName string) error
}

type documentIndexer interface {
	ProcessJob(ctx context.Context, jobID string) error
	DeleteDocumentVectors(ctx context.Context, documentID string) error
}

type documentVersionService interface {
	CreateInitial(ctx context.Context, input wiki_service.InitialDocumentInput) (*wiki_db.Document, *rag_db.IndexJob, error)
	SaveDirectEdit(ctx context.Context, userID, documentID, content string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error)
}

type revisionPathRepository interface {
	ListRevisionObjectPaths(ctx context.Context, documentID string) ([]string, error)
}

// 文档处理器
type DocumentHandler struct {
	documentRepository documentRepository
	minioService       documentFileStore
	indexer            documentIndexer
	versions           documentVersionService
	revisions          revisionPathRepository
}

// 创建文档处理器
func NewDocumentHandler(dbService *database.DatabaseService, minioService *storage.MinIOService, indexer *rag_service.Indexer) *DocumentHandler {
	return NewDocumentHandlerWithDependencies(
		wiki_repo.NewDocumentRepository(dbService.GetDB()),
		minioService,
		indexer,
		wiki_service.NewDocumentVersionService(dbService.Revisions, dbService.Versions, dbService.Documents, minioService, dbService.ChangeRequests),
		dbService.Revisions,
	)
}

func NewDocumentHandlerWithDependencies(repo documentRepository, minioService documentFileStore, indexer documentIndexer, dependencies ...any) *DocumentHandler {
	handler := &DocumentHandler{
		documentRepository: repo,
		minioService:       minioService,
		indexer:            indexer,
	}
	for _, dependency := range dependencies {
		switch dependency := dependency.(type) {
		case documentVersionService:
			handler.versions = dependency
		case revisionPathRepository:
			handler.revisions = dependency
		}
	}
	return handler
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

	// 打开文件
	f, err := file.Open()
	if err != nil {
		log.Printf("❌ 无法打开文件: %v", err)
		return response.InternalServerCtx(c, "无法打开文件: "+err.Error())
	}
	defer f.Close()

	if h.versions == nil {
		return response.InternalServerCtx(c, "文档版本服务未初始化")
	}
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return response.InternalServerCtx(c, "读取上传文件失败")
	}
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/markdown"
	}
	doc, job, err := h.versions.CreateInitial(c.Context(), wiki_service.InitialDocumentInput{
		UserID: userID, FolderID: folderID, Filename: file.Filename, Content: content, ContentType: contentType,
	})
	if err != nil {
		return response.InternalServerCtx(c, "创建文档版本失败")
	}
	log.Printf("✅ 文档上传成功，ID: %s", doc.ID)
	h.processJobAsync(job.ID)

	return response.SuccessCtx(c, doc)
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

	if h.versions == nil {
		return response.InternalServerCtx(c, "文档版本服务未初始化")
	}
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未登录或登录已过期")
	}
	_, job, err := h.versions.SaveDirectEdit(c.Context(), userID, docID, req.Content)
	if err != nil {
		return response.InternalServerCtx(c, "保存文件内容失败")
	}

	log.Printf("✅ 文档内容已更新: %s", docID)
	h.processJobAsync(job.ID)
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

	if h.indexer != nil {
		if err := h.indexer.DeleteDocumentVectors(c.Context(), docID); err != nil {
			return response.InternalServerCtx(c, "删除文档索引失败: "+err.Error())
		}
	}

	paths := []string{doc.FilePath}
	if h.revisions != nil {
		revisionPaths, err := h.revisions.ListRevisionObjectPaths(c.Context(), docID)
		if err != nil {
			return response.InternalServerCtx(c, "获取文档修订失败")
		}
		if len(revisionPaths) > 0 {
			paths = revisionPaths
		}
	}
	for _, objectPath := range paths {
		if err := h.minioService.DeleteFile(c.Context(), objectPath); err != nil {
			return response.InternalServerCtx(c, "删除文档文件失败: "+err.Error())
		}
	}

	if err := h.documentRepository.DeleteWithChunks(c.Context(), docID); err != nil {
		return response.InternalServerCtx(c, "删除文档失败: "+err.Error())
	}

	return response.SuccessMsgCtx(c, "文档删除成功")
}

func (h *DocumentHandler) processJobAsync(jobID string) {
	if h.indexer == nil {
		return
	}
	go func() {
		if err := h.indexer.ProcessJob(context.Background(), jobID); err != nil {
			log.Printf("⚠️  文档索引失败: job_id=%s err=%v", jobID, err)
		}
	}()
}
