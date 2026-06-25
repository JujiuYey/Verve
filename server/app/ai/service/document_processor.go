// document_processor.go
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	qdrantdao "sag-wiki/infrastructure/qdrant"
	"sag-wiki/infrastructure/storage"
	wiki_db "sag-wiki/app/wiki/models/db"
	wiki_repo "sag-wiki/app/wiki/repository"
)

type DocumentProcessor struct {
	embeddingService *EmbeddingService
	chunker          *Chunker
	chunkDAO         *qdrantdao.ChunkDAO
	minioService     *storage.MinIOService
	documentRepo     *wiki_repo.DocumentRepository
}

func NewDocumentProcessor(
	embeddingService *EmbeddingService,
	chunkDAO *qdrantdao.ChunkDAO,
	minioService *storage.MinIOService,
	documentRepo *wiki_repo.DocumentRepository,
) *DocumentProcessor {
	return &DocumentProcessor{
		embeddingService: embeddingService,
		// 使用语义分块器，替代原来的简单分块
		chunker:      NewSemanticChunker(DefaultChunkerConfig()),
		chunkDAO:     chunkDAO,
		minioService: minioService,
		documentRepo: documentRepo,
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

	// 2. 文本分块（语义分块）
	chunks := p.chunker.ChunkText(content)
	if len(chunks) == 0 {
		log.Printf("⚠️  文档内容为空: %s", docID)
		return &ProcessResult{Error: fmt.Errorf("文档内容为空")}
	}
	log.Printf("📝 分块完成，共 %d 个 chunks（语义分块）", len(chunks))

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

// DocumentRepo returns the document repository (needed by TaskWorker for status updates)
func (p *DocumentProcessor) DocumentRepo() *wiki_repo.DocumentRepository {
	return p.documentRepo
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
