package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/uuid"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
)

type versionRevisionWriter interface {
	CreateInitial(ctx context.Context, document *wiki_db.Document, revision *wiki_db.DocumentRevision, job *rag_db.IndexJob) error
	EnsureLegacyRevision(ctx context.Context, snapshot *wiki_repo.LegacyRevisionSnapshot) error
}

type versionPublisher interface {
	ApplyDirectEdit(ctx context.Context, input wiki_repo.ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error)
	ApplyChangeRequest(ctx context.Context, input wiki_repo.ApplyVersionInput) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error)
	FindAppliedChangeRequest(ctx context.Context, changeRequestID string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error)
}

type versionDocumentFinder interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Document, error)
}

type versionFileStore interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	GetFileContent(ctx context.Context, objectName string) (string, error)
}

type versionChangeRequestFinder interface {
	FindChangeRequest(ctx context.Context, id string) (*wiki_db.DocumentChangeRequest, error)
}

// InitialDocumentInput 是首次上传所需的文档内容和归属信息。
type InitialDocumentInput struct {
	UserID      string
	FolderID    string
	Filename    string
	ContentType string
	Content     []byte
}

// DocumentVersionService 编排对象存储和版本事务，所有内容写入均生成不可变版本。
type DocumentVersionService struct {
	revisions versionRevisionWriter
	publisher versionPublisher
	documents versionDocumentFinder
	files     versionFileStore
	requests  versionChangeRequestFinder
}

func NewDocumentVersionService(
	revisions versionRevisionWriter,
	publisher versionPublisher,
	documents versionDocumentFinder,
	files versionFileStore,
	requests ...versionChangeRequestFinder,
) *DocumentVersionService {
	service := &DocumentVersionService{
		revisions: revisions,
		publisher: publisher,
		documents: documents,
		files:     files,
	}
	if len(requests) > 0 {
		service.requests = requests[0]
	}
	return service
}

// CreateInitial 写入版本一对象后，原子创建文档、修订和索引任务。
func (s *DocumentVersionService) CreateInitial(ctx context.Context, input InitialDocumentInput) (*wiki_db.Document, *rag_db.IndexJob, error) {
	if s.revisions == nil || s.files == nil {
		return nil, nil, fmt.Errorf("文档版本服务未初始化")
	}
	if strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.FolderID) == "" || strings.TrimSpace(input.Filename) == "" {
		return nil, nil, fmt.Errorf("文档归属和文件名不能为空")
	}

	documentID := newDocumentID()
	objectPath := revisionObjectPath(documentID, 1, input.Filename)
	contentHash := contentSHA256(input.Content)
	contentType := input.ContentType
	if contentType == "" {
		contentType = "text/markdown"
	}
	if err := s.files.UploadFile(ctx, objectPath, bytes.NewReader(input.Content), int64(len(input.Content)), contentType); err != nil {
		return nil, nil, fmt.Errorf("写入初始文档对象失败: %w", err)
	}

	document := &wiki_db.Document{
		ID:             documentID,
		Filename:       input.Filename,
		FileSize:       int64(len(input.Content)),
		ContentType:    contentType,
		FolderID:       input.FolderID,
		FilePath:       objectPath,
		CurrentVersion: 1,
		ContentHash:    &contentHash,
	}
	revision := &wiki_db.DocumentRevision{
		ID:            newDocumentID(),
		DocumentID:    documentID,
		Version:       1,
		ObjectPath:    objectPath,
		ContentHash:   contentHash,
		FileSize:      document.FileSize,
		ChangedBy:     input.UserID,
		ChangeSummary: "初始上传",
	}
	job := &rag_db.IndexJob{
		ID:              newDocumentID(),
		DocumentID:      documentID,
		DocumentVersion: 1,
		ObjectPath:      objectPath,
		Status:          "pending",
		MaxAttempts:     3,
	}
	if err := s.revisions.CreateInitial(ctx, document, revision, job); err != nil {
		return nil, nil, fmt.Errorf("提交初始文档版本失败: %w", err)
	}
	return document, job, nil
}

// SaveDirectEdit 先保存旧文档快照，再发布编辑后的新版本。
func (s *DocumentVersionService) SaveDirectEdit(ctx context.Context, userID, documentID, content string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	if s.publisher == nil || s.documents == nil || s.files == nil {
		return nil, nil, fmt.Errorf("文档版本服务未初始化")
	}
	document, err := s.documents.FindOne(ctx, documentID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.ensureLegacyRevision(ctx, userID, document); err != nil {
		return nil, nil, err
	}

	objectPath := revisionObjectPath(document.ID, document.CurrentVersion+1, document.Filename)
	contentBytes := []byte(content)
	contentHash := contentSHA256(contentBytes)
	if err := s.files.UploadFile(ctx, objectPath, bytes.NewReader(contentBytes), int64(len(contentBytes)), "text/markdown"); err != nil {
		return nil, nil, fmt.Errorf("写入文档新版本失败: %w", err)
	}
	return s.publisher.ApplyDirectEdit(ctx, wiki_repo.ApplyVersionInput{
		DocumentID:      document.ID,
		ExpectedVersion: document.CurrentVersion,
		ObjectPath:      objectPath,
		ContentHash:     contentHash,
		FileSize:        int64(len(contentBytes)),
		ChangedBy:       userID,
		ChangeSummary:   "手动编辑",
	})
}

// ApplyChangeRequest 应用用户自己的待确认申请，并为其发布一个新版本。
func (s *DocumentVersionService) ApplyChangeRequest(ctx context.Context, userID, requestID string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	if s.publisher == nil || s.documents == nil || s.files == nil || s.requests == nil {
		return nil, nil, fmt.Errorf("文档版本服务未初始化")
	}
	request, err := s.requests.FindChangeRequest(ctx, requestID)
	if err != nil {
		return nil, nil, err
	}
	if request.RequestedBy != userID {
		return nil, nil, wiki_repo.ErrChangeRequestForbidden
	}
	if request.Status == wiki_db.ChangeRequestStatusApplied {
		return s.publisher.FindAppliedChangeRequest(ctx, request.ID)
	}
	if request.Status != wiki_db.ChangeRequestStatusProposed {
		return nil, nil, wiki_repo.ErrChangeRequestNotProposed
	}

	document, err := s.documents.FindOne(ctx, request.DocumentID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.ensureLegacyRevision(ctx, userID, document); err != nil {
		return nil, nil, err
	}
	contentBytes := []byte(request.ProposedContent)
	objectPath := revisionObjectPath(document.ID, request.BaseVersion+1, document.Filename)
	contentHash := contentSHA256(contentBytes)
	if err := s.files.UploadFile(ctx, objectPath, bytes.NewReader(contentBytes), int64(len(contentBytes)), "text/markdown"); err != nil {
		return nil, nil, fmt.Errorf("写入申请文档版本失败: %w", err)
	}
	return s.publisher.ApplyChangeRequest(ctx, wiki_repo.ApplyVersionInput{
		ChangeRequestID: &request.ID,
		DocumentID:      document.ID,
		ExpectedVersion: request.BaseVersion,
		ObjectPath:      objectPath,
		ContentHash:     contentHash,
		FileSize:        int64(len(contentBytes)),
		ChangedBy:       userID,
		ChangeSummary:   request.ChangeSummary,
	})
}

func (s *DocumentVersionService) CancelChangeRequest(ctx context.Context, userID, requestID string) error {
	if s.requests == nil {
		return fmt.Errorf("文档版本服务未初始化")
	}
	if canceller, ok := s.requests.(interface {
		CancelChangeRequest(context.Context, string, string) error
	}); ok {
		return canceller.CancelChangeRequest(ctx, requestID, userID)
	}
	return fmt.Errorf("变更申请仓储不支持取消")
}

func (s *DocumentVersionService) ensureLegacyRevision(ctx context.Context, userID string, document *wiki_db.Document) error {
	if document.ContentHash != nil {
		return nil
	}
	if s.revisions == nil {
		return fmt.Errorf("修订仓储未初始化")
	}
	if document.CurrentVersion != 1 {
		return wiki_repo.ErrVersionConflict
	}
	content, err := s.files.GetFileContent(ctx, document.FilePath)
	if err != nil {
		return fmt.Errorf("读取旧文档内容失败: %w", err)
	}
	contentBytes := []byte(content)
	objectPath := revisionObjectPath(document.ID, 1, document.Filename)
	if err := s.files.UploadFile(ctx, objectPath, bytes.NewReader(contentBytes), int64(len(contentBytes)), "text/markdown"); err != nil {
		return fmt.Errorf("写入旧文档快照失败: %w", err)
	}
	return s.revisions.EnsureLegacyRevision(ctx, &wiki_repo.LegacyRevisionSnapshot{
		DocumentID:    document.ID,
		ObjectPath:    objectPath,
		ContentHash:   contentSHA256(contentBytes),
		FileSize:      int64(len(contentBytes)),
		ChangedBy:     userID,
		ChangeSummary: "历史文档快照",
	})
}

func revisionObjectPath(documentID string, version int64, filename string) string {
	return fmt.Sprintf("documents/%s/revisions/%d/%s", documentID, version, path.Base(filename))
}

func contentSHA256(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

func newDocumentID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
