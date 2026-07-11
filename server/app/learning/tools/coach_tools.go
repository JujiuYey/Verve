package tools

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	learning_service "verve/app/learning/service"
	rag_payload "verve/app/rag/models/payload"
	rag_service "verve/app/rag/service"
	wiki_db "verve/app/wiki/models/db"
	"verve/infrastructure/database"
)

type ListFoldersInput struct {
	Limit int `json:"limit" jsonschema_description:"最多返回多少个文件夹,默认 50"`
}

type ListDocumentsInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID,可留空"`
	Limit    int    `json:"limit" jsonschema_description:"最多返回多少个文档,默认 50"`
}

type coachFolderLister interface {
	List(context.Context, map[string]interface{}) ([]*wiki_db.Folder, error)
}

type coachDocumentLister interface {
	List(context.Context, string, string) ([]*wiki_db.Document, error)
}

type coachKnowledgeSearcher interface {
	Search(context.Context, string, string, int) ([]rag_payload.SearchResult, error)
}

type SearchLearningMemoryInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID,可留空表示搜索最近的全局学习记忆"`
	Query    string `json:"query" jsonschema_description:"要搜索的学习记忆关键词,可留空"`
	Limit    int    `json:"limit" jsonschema_description:"最多返回多少条学习记忆,默认 10"`
}

type SearchLearningMemoryOutput struct {
	Results []map[string]interface{} `json:"results"`
}

type SearchWikiKnowledgeInput struct {
	RootFolderID string `json:"root_folder_id" jsonschema_description:"限定检索的 Wiki 根目录 ID"`
	Query        string `json:"query" jsonschema_description:"要检索的学习问题或概念"`
	Limit        int    `json:"limit" jsonschema_description:"最多返回多少个片段,默认 6"`
}

type SearchWikiKnowledgeOutput struct {
	Results []map[string]interface{} `json:"results"`
}

type CompleteTaskInput struct {
	Summary string `json:"summary" jsonschema_description:"本轮 agent 已完成或被阻塞的简短说明"`
	Status  string `json:"status" jsonschema_description:"success/partial/blocked"`
}

type CompleteTaskOutput struct {
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

func NewCoachTools(db *database.DatabaseService, retriever *rag_service.Retriever, userID string) []tool.BaseTool {
	tools := []tool.BaseTool{
		newListFoldersTool(db, userID),
		newListDocumentsTool(db, userID),
		newSearchLearningMemoryTool(db, userID),
		newCompleteTaskTool(),
	}
	if retriever != nil {
		tools = append(tools, newSearchWikiKnowledgeTool(db.Folders, db.Documents, retriever, userID))
	}
	return tools
}

func newSearchWikiKnowledgeTool(folders coachFolderLister, documents coachDocumentLister, retriever coachKnowledgeSearcher, userID string) tool.InvokableTool {
	t, err := utils.InferTool("search_wiki_knowledge", "按 Wiki 根目录检索真实文档片段,用于回答概念问题或决定下一步学习内容",
		func(ctx context.Context, input *SearchWikiKnowledgeInput) (*SearchWikiKnowledgeOutput, error) {
			return searchAccessibleWikiKnowledge(ctx, folders, documents, retriever, userID, input)
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func searchAccessibleWikiKnowledge(
	ctx context.Context,
	folders coachFolderLister,
	documents coachDocumentLister,
	retriever coachKnowledgeSearcher,
	userID string,
	input *SearchWikiKnowledgeInput,
) (*SearchWikiKnowledgeOutput, error) {
	if input == nil {
		return nil, sql.ErrNoRows
	}
	allFolders, err := folders.List(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	accessibleFolders := learning_service.FilterCoachFoldersForUser(allFolders, userID)
	rootFolderID := strings.TrimSpace(input.RootFolderID)
	if !learning_service.CoachFolderIsAccessible(accessibleFolders, rootFolderID) {
		return nil, sql.ErrNoRows
	}
	allDocuments, err := documents.List(ctx, "", "")
	if err != nil {
		return nil, err
	}
	accessibleDocuments := learning_service.FilterCoachDocumentsByFolders(allDocuments, accessibleFolders)
	accessibleDocumentIDs := make(map[string]struct{}, len(accessibleDocuments))
	for _, document := range accessibleDocuments {
		accessibleDocumentIDs[document.ID] = struct{}{}
	}
	results, err := retriever.Search(ctx, rootFolderID, input.Query, input.Limit)
	if err != nil {
		return nil, err
	}
	out := &SearchWikiKnowledgeOutput{Results: make([]map[string]interface{}, 0, len(results))}
	for _, result := range results {
		if _, ok := accessibleDocumentIDs[result.DocumentID]; !ok {
			continue
		}
		out.Results = append(out.Results, map[string]interface{}{
			"document_title": result.DocumentTitle,
			"folder_path":    result.FolderPath,
			"heading_path":   result.HeadingPath,
			"content":        result.Content,
			"score":          result.Score,
		})
	}
	return out, nil
}

func newSearchLearningMemoryTool(db *database.DatabaseService, userID string) tool.InvokableTool {
	t, err := utils.InferTool("search_learning_memory", "搜索学习记忆",
		func(ctx context.Context, input *SearchLearningMemoryInput) (*SearchLearningMemoryOutput, error) {
			out := &SearchLearningMemoryOutput{Results: []map[string]interface{}{}}
			if db == nil || db.Memories == nil {
				return out, nil
			}
			if input == nil {
				input = &SearchLearningMemoryInput{}
			}

			limit := normalizeLimit(input.Limit, 10)
			query := strings.ToLower(strings.TrimSpace(input.Query))
			candidateLimit := limit
			if query != "" {
				candidateLimit = max(limit*5, 50)
			}
			items, err := learning_service.NewMemoryService(db).FindCoachItems(ctx, userID, input.FolderID, candidateLimit)
			if err != nil {
				return nil, err
			}

			for _, item := range items {
				if item == nil {
					continue
				}
				if query != "" {
					statement := strings.ToLower(item.Statement)
					kind := strings.ToLower(item.Kind)
					if !strings.Contains(statement, query) && !strings.Contains(kind, query) {
						continue
					}
				}
				if len(out.Results) >= limit {
					break
				}
				out.Results = append(out.Results, map[string]interface{}{
					"id":           item.ID,
					"folder_id":    stringValue(item.FolderID),
					"document_id":  stringValue(item.DocumentID),
					"kind":         item.Kind,
					"statement":    item.Statement,
					"confidence":   item.Confidence,
					"last_seen_at": item.LastSeenAt,
				})
			}
			return out, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newListFoldersTool(db *database.DatabaseService, userID string) tool.InvokableTool {
	t, err := utils.InferTool("list_folders", "列出 Wiki 文件夹,用于判断用户有哪些学习范围",
		func(ctx context.Context, input *ListFoldersInput) ([]map[string]interface{}, error) {
			folders, err := db.Folders.List(ctx, map[string]interface{}{})
			if err != nil {
				return nil, err
			}
			limit := normalizeLimit(input.Limit, 50)
			result := make([]map[string]interface{}, 0, min(len(folders), limit))
			for _, folder := range folders {
				if folder.UserID != nil && *folder.UserID != "" && *folder.UserID != userID {
					continue
				}
				result = append(result, map[string]interface{}{
					"id":          folder.ID,
					"name":        folder.Name,
					"description": stringValue(folder.Description),
					"parent_id":   stringValue(folder.ParentID),
				})
				if len(result) >= limit {
					break
				}
			}
			return result, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newListDocumentsTool(db *database.DatabaseService, userID string) tool.InvokableTool {
	t, err := utils.InferTool("list_documents", "列出 Wiki 文档,可按文件夹过滤",
		func(ctx context.Context, input *ListDocumentsInput) ([]map[string]interface{}, error) {
			if input == nil {
				input = &ListDocumentsInput{}
			}
			docs, err := loadAccessibleCoachDocuments(ctx, db.Folders, db.Documents, userID, input.FolderID)
			if err != nil {
				return nil, err
			}
			limit := normalizeLimit(input.Limit, 50)
			result := make([]map[string]interface{}, 0, min(len(docs), limit))
			for _, doc := range docs {
				result = append(result, map[string]interface{}{
					"id":        doc.ID,
					"filename":  doc.Filename,
					"folder_id": doc.FolderID,
					"file_size": doc.FileSize,
				})
				if len(result) >= limit {
					break
				}
			}
			return result, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func loadAccessibleCoachDocuments(
	ctx context.Context,
	folders coachFolderLister,
	documents coachDocumentLister,
	userID string,
	folderID string,
) ([]*wiki_db.Document, error) {
	allFolders, err := folders.List(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	accessibleFolders := learning_service.FilterCoachFoldersForUser(allFolders, userID)
	folderID = strings.TrimSpace(folderID)
	if folderID != "" && !learning_service.CoachFolderIsAccessible(accessibleFolders, folderID) {
		return nil, sql.ErrNoRows
	}
	docs, err := documents.List(ctx, "", folderID)
	if err != nil {
		return nil, err
	}
	return learning_service.FilterCoachDocumentsByFolders(docs, accessibleFolders), nil
}

func newCompleteTaskTool() tool.InvokableTool {
	t, err := utils.InferTool("complete_task", "结束本轮学习调度",
		func(ctx context.Context, input *CompleteTaskInput) (*CompleteTaskOutput, error) {
			status := input.Status
			if status == "" {
				status = "success"
			}
			return &CompleteTaskOutput{Summary: input.Summary, Status: status}, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func normalizeLimit(limit, fallback int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
