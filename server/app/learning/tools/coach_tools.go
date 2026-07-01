package tools

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/database"
)

type ListFoldersInput struct {
	Limit int `json:"limit" jsonschema_description:"最多返回多少个文件夹,默认 50"`
}

type ListDocumentsInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID,可留空"`
	Limit    int    `json:"limit" jsonschema_description:"最多返回多少个文档,默认 50"`
}

type ListObjectivesInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID,可留空表示返回最近的小节"`
	Limit    int    `json:"limit" jsonschema_description:"最多返回多少个小节,默认 50"`
}

type GetLearningProfileInput struct {
	FolderID string `json:"folder_id" jsonschema_description:"文件夹 ID"`
}

type ListJournalsInput struct {
	Limit int `json:"limit" jsonschema_description:"最多返回多少条最近学习记录,默认 10"`
}

type CreatePracticeSessionInput struct {
	ObjectiveID string `json:"objective_id" jsonschema_description:"要进入练习的小节 ID"`
}

type CreatePracticeSessionOutput struct {
	SessionID   string `json:"session_id"`
	ObjectiveID string `json:"objective_id"`
}

type CompleteTaskInput struct {
	Summary string `json:"summary" jsonschema_description:"本轮 agent 已完成或被阻塞的简短说明"`
	Status  string `json:"status" jsonschema_description:"success/partial/blocked"`
}

type CompleteTaskOutput struct {
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

func NewCoachTools(db *database.DatabaseService, userID string) []tool.BaseTool {
	return []tool.BaseTool{
		newListFoldersTool(db, userID),
		newListDocumentsTool(db),
		newListObjectivesTool(db, userID),
		newGetLearningProfileTool(db),
		newListJournalsTool(db, userID),
		newCreatePracticeSessionTool(db, userID),
		newCompleteTaskTool(),
	}
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

func newListObjectivesTool(db *database.DatabaseService, userID string) tool.InvokableTool {
	t, err := utils.InferTool("list_objectives", "列出学习小节,用于选择继续学习或复习的小点",
		func(ctx context.Context, input *ListObjectivesInput) ([]map[string]interface{}, error) {
			var objectives []*learning_db.LearningObjective
			var err error
			if strings.TrimSpace(input.FolderID) != "" {
				objectives, err = db.Objectives.FindByFolder(ctx, input.FolderID)
			} else {
				objectives, err = db.Objectives.FindRecentByUser(ctx, userID, normalizeLimit(input.Limit, 50))
			}
			if err != nil {
				return nil, err
			}
			limit := normalizeLimit(input.Limit, 50)
			result := make([]map[string]interface{}, 0, min(len(objectives), limit))
			for _, obj := range objectives {
				result = append(result, map[string]interface{}{
					"id":                 obj.ID,
					"title":              obj.Title,
					"detail":             stringValue(obj.Detail),
					"status":             obj.Status,
					"mastery_level":      obj.MasteryLevel,
					"source_folder_id":   stringValue(obj.SourceFolderID),
					"source_document_id": stringValue(obj.SourceDocumentID),
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

func newGetLearningProfileTool(db *database.DatabaseService) tool.InvokableTool {
	t, err := utils.InferTool("get_learning_profile", "读取某个 Wiki 文件夹的学习画像",
		func(ctx context.Context, input *GetLearningProfileInput) (map[string]interface{}, error) {
			profile, err := db.Profiles.FindByFolder(ctx, input.FolderID)
			if err == sql.ErrNoRows {
				return map[string]interface{}{"folder_id": input.FolderID, "empty": true}, nil
			}
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"folder_id":           profile.FolderID,
				"current_level":       stringValue(profile.CurrentLevel),
				"completed_topics":    profile.CompletedTopics,
				"weak_points":         profile.WeakPoints,
				"verification_habits": stringValue(profile.VerificationHabits),
				"next_goal":           stringValue(profile.NextGoal),
			}, nil
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

func newCreatePracticeSessionTool(db *database.DatabaseService, userID string) tool.InvokableTool {
	t, err := utils.InferTool("create_practice_session", "为指定学习小节创建练习会话",
		func(ctx context.Context, input *CreatePracticeSessionInput) (*CreatePracticeSessionOutput, error) {
			obj, err := db.Objectives.FindOne(ctx, input.ObjectiveID)
			if err != nil {
				return nil, err
			}
			if obj.UserID != userID {
				return nil, sql.ErrNoRows
			}
			session := &learning_db.LearningSession{
				UserID:      userID,
				ObjectiveID: obj.ID,
				Status:      "active",
			}
			if err := db.Sessions.Create(ctx, session); err != nil {
				return nil, err
			}
			return &CreatePracticeSessionOutput{SessionID: session.ID, ObjectiveID: obj.ID}, nil
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
