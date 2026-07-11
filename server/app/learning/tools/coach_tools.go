package tools

import (
	"context"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	rag_service "verve/app/rag/service"
	"verve/infrastructure/database"
)

type ListFoldersInput struct {
	Limit int `json:"limit" jsonschema_description:"最多返回多少个文件夹,默认 50"`
}

type ListDocumentsInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID,可留空"`
	Limit    int    `json:"limit" jsonschema_description:"最多返回多少个文档,默认 50"`
}

type SearchLearningMemoryInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID,可留空表示搜索最近的全局学习记忆"`
	Query    string `json:"query" jsonschema_description:"要搜索的学习记忆关键词,可留空"`
	Limit    int    `json:"limit" jsonschema_description:"最多返回多少条学习记忆,默认 10"`
}

type SearchLearningMemoryOutput struct {
	Results []map[string]interface{} `json:"results"`
}

type ListJournalsInput struct {
	Limit int `json:"limit" jsonschema_description:"最多返回多少条最近学习记录,默认 10"`
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
		newListDocumentsTool(db),
		newSearchLearningMemoryTool(db, userID),
		newListJournalsTool(db, userID),
		newCompleteTaskTool(),
	}
	if retriever != nil {
		tools = append(tools, newSearchWikiKnowledgeTool(retriever))
	}
	return tools
}

func newSearchWikiKnowledgeTool(retriever *rag_service.Retriever) tool.InvokableTool {
	t, err := utils.InferTool("search_wiki_knowledge", "按 Wiki 根目录检索真实文档片段,用于回答概念问题或决定下一步学习内容",
		func(ctx context.Context, input *SearchWikiKnowledgeInput) (*SearchWikiKnowledgeOutput, error) {
			results, err := retriever.Search(ctx, input.RootFolderID, input.Query, input.Limit)
			if err != nil {
				return nil, err
			}
			out := &SearchWikiKnowledgeOutput{Results: make([]map[string]interface{}, 0, len(results))}
			for _, result := range results {
				out.Results = append(out.Results, map[string]interface{}{
					"document_title": result.DocumentTitle,
					"folder_path":    result.FolderPath,
					"heading_path":   result.HeadingPath,
					"content":        result.Content,
					"score":          result.Score,
				})
			}
			return out, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
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
			items, err := db.Memories.FindItemsByUser(ctx, userID, input.FolderID, candidateLimit)
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

func newListDocumentsTool(db *database.DatabaseService) tool.InvokableTool {
	t, err := utils.InferTool("list_documents", "列出 Wiki 文档,可按文件夹过滤",
		func(ctx context.Context, input *ListDocumentsInput) ([]map[string]interface{}, error) {
			docs, err := db.Documents.List(ctx, "", input.FolderID)
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

func newListJournalsTool(db *database.DatabaseService, userID string) tool.InvokableTool {
	t, err := utils.InferTool("list_learning_journals", "列出最近学习记录",
		func(ctx context.Context, input *ListJournalsInput) ([]map[string]interface{}, error) {
			limit := normalizeLimit(input.Limit, 10)
			journals, _, err := db.Journals.FindByUser(ctx, userID, 0, limit)
			if err != nil {
				return nil, err
			}
			result := make([]map[string]interface{}, 0, len(journals))
			for _, journal := range journals {
				result = append(result, map[string]interface{}{
					"id":          journal.ID,
					"folder_id":   journal.FolderID,
					"date":        journal.Date,
					"learned":     stringValue(journal.Learned),
					"evidence":    stringValue(journal.Evidence),
					"weak_points": stringValue(journal.WeakPoints),
					"next_step":   stringValue(journal.NextStep),
				})
			}
			return result, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
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
