package service

import (
	"encoding/json"
	"strings"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
	"verve/infrastructure/llm/prompts"
)

type CoachRuntimeContext struct {
	UserID          string
	AgentInstanceID string
	AgentName       string
	RootFolderID    string
	RootFolderName  string
	Folders         []*wiki_db.Folder
	Documents       []*wiki_db.Document
	MemoryItems     []*learning_db.LearningMemoryItem
	Journals        []*learning_db.LearningJournal
}

type CoachAction struct {
	Type       string `json:"type"`
	DocumentID string `json:"document_id,omitempty"`
	FolderID   string `json:"folder_id,omitempty"`
	Label      string `json:"label,omitempty"`
}

// BuildCoachQuery 把 runtime context + 用户消息拼成一段 prompt 喂给陪练 agent。
//
// Service 层只负责把数据库模型映射成 prompt input; Markdown 渲染和回复约束由 prompts 包维护。
func BuildCoachQuery(ctx CoachRuntimeContext, message string) string {
	return prompts.CoachQueryPrompt(prompts.CoachQueryInput{
		Message: message,
		AgentContext: &prompts.CoachAgentContext{
			AgentInstanceID: ctx.AgentInstanceID,
			AgentName:       ctx.AgentName,
			RootFolderID:    ctx.RootFolderID,
			RootFolderName:  ctx.RootFolderName,
		},
		Folders:     mapCoachFolders(ctx.Folders),
		Documents:   mapCoachDocuments(ctx.Documents),
		MemoryItems: mapCoachMemoryItems(ctx.MemoryItems),
		Journals:    mapCoachJournals(ctx.Journals),
	})
}

func mapCoachFolders(folders []*wiki_db.Folder) []prompts.CoachFolder {
	result := make([]prompts.CoachFolder, 0, len(folders))
	for _, folder := range folders {
		if folder == nil {
			continue
		}
		result = append(result, prompts.CoachFolder{
			ID:          folder.ID,
			Name:        folder.Name,
			Description: trimStringPtr(folder.Description),
		})
	}
	return result
}

func mapCoachDocuments(documents []*wiki_db.Document) []prompts.CoachDocument {
	result := make([]prompts.CoachDocument, 0, len(documents))
	for _, doc := range documents {
		if doc == nil {
			continue
		}
		result = append(result, prompts.CoachDocument{
			ID:       doc.ID,
			FolderID: doc.FolderID,
			Filename: doc.Filename,
		})
	}
	return result
}

// FilterCoachFoldersForUser returns folders owned by the user plus public folders.
func FilterCoachFoldersForUser(folders []*wiki_db.Folder, userID string) []*wiki_db.Folder {
	result := make([]*wiki_db.Folder, 0, len(folders))
	for _, folder := range folders {
		if folder == nil {
			continue
		}
		if folder.UserID != nil && *folder.UserID != "" && *folder.UserID != userID {
			continue
		}
		result = append(result, folder)
	}
	return result
}

func CoachFolderIsAccessible(folders []*wiki_db.Folder, folderID string) bool {
	folderID = strings.TrimSpace(folderID)
	if folderID == "" {
		return false
	}
	for _, folder := range folders {
		if folder != nil && folder.ID == folderID {
			return true
		}
	}
	return false
}

// FilterCoachDocumentsByFolders excludes documents whose folder is not visible.
func FilterCoachDocumentsByFolders(documents []*wiki_db.Document, folders []*wiki_db.Folder) []*wiki_db.Document {
	allowed := make(map[string]struct{}, len(folders))
	for _, folder := range folders {
		if folder != nil {
			allowed[folder.ID] = struct{}{}
		}
	}
	result := make([]*wiki_db.Document, 0, len(documents))
	for _, document := range documents {
		if document == nil {
			continue
		}
		if _, ok := allowed[document.FolderID]; ok {
			result = append(result, document)
		}
	}
	return result
}

func mapCoachMemoryItems(items []*learning_db.LearningMemoryItem) []prompts.CoachMemoryItem {
	result := make([]prompts.CoachMemoryItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, prompts.CoachMemoryItem{
			FolderID:   trimStringPtr(item.FolderID),
			Kind:       item.Kind,
			Statement:  item.Statement,
			Confidence: item.Confidence,
		})
	}
	return result
}

func mapCoachJournals(journals []*learning_db.LearningJournal) []prompts.CoachJournal {
	result := make([]prompts.CoachJournal, 0, len(journals))
	for _, journal := range journals {
		if journal == nil {
			continue
		}
		result = append(result, prompts.CoachJournal{
			FolderID:   journal.FolderID,
			Date:       journal.Date,
			Learned:    trimStringPtr(journal.Learned),
			WeakPoints: trimStringPtr(journal.WeakPoints),
			NextStep:   trimStringPtr(journal.NextStep),
		})
	}
	return result
}

func trimStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func ParseCoachAction(content string) *CoachAction {
	start := strings.Index(content, "<ACTION>")
	end := strings.Index(content, "</ACTION>")
	if start < 0 || end <= start {
		return nil
	}
	raw := strings.TrimSpace(content[start+len("<ACTION>") : end])
	var action CoachAction
	if err := json.Unmarshal([]byte(raw), &action); err != nil {
		return nil
	}
	if strings.TrimSpace(action.Type) == "" {
		return nil
	}
	if action.Type == "navigate_to_practice" && strings.TrimSpace(action.DocumentID) == "" {
		return nil
	}
	return &action
}

func ParseCoachActionForDocuments(content string, documents []*wiki_db.Document) *CoachAction {
	action := ParseCoachAction(content)
	if action == nil || action.Type != "navigate_to_practice" {
		return nil
	}
	for _, document := range documents {
		if document != nil && document.ID == action.DocumentID {
			return action
		}
	}
	return nil
}
