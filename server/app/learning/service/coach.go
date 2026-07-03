package service

import (
	"encoding/json"
	"strings"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
	"verve/infrastructure/llm/prompts"
)

type CoachRuntimeContext struct {
	UserID     string
	Folders    []*wiki_db.Folder
	Documents  []*wiki_db.Document
	Objectives []*learning_db.LearningObjective
	Profiles   []*learning_db.LearningProfile
	Journals   []*learning_db.LearningJournal
}

type CoachAction struct {
	Type        string `json:"type"`
	ObjectiveID string `json:"objective_id,omitempty"`
	FolderID    string `json:"folder_id,omitempty"`
	Label       string `json:"label,omitempty"`
}

// BuildCoachQuery 把 runtime context + 用户消息拼成一段 prompt 喂给陪练 agent。
//
// Service 层只负责把数据库模型映射成 prompt input; Markdown 渲染和回复约束由 prompts 包维护。
func BuildCoachQuery(ctx CoachRuntimeContext, message string) string {
	return prompts.CoachQueryPrompt(prompts.CoachQueryInput{
		Message:    message,
		Folders:    mapCoachFolders(ctx.Folders),
		Documents:  mapCoachDocuments(ctx.Documents),
		Objectives: mapCoachObjectives(ctx.Objectives),
		Profiles:   mapCoachProfiles(ctx.Profiles),
		Journals:   mapCoachJournals(ctx.Journals),
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

func mapCoachObjectives(objectives []*learning_db.LearningObjective) []prompts.CoachObjective {
	result := make([]prompts.CoachObjective, 0, len(objectives))
	for _, obj := range objectives {
		if obj == nil {
			continue
		}
		result = append(result, prompts.CoachObjective{
			ID:               obj.ID,
			Title:            obj.Title,
			Detail:           trimStringPtr(obj.Detail),
			SourceFolderID:   trimStringPtr(obj.SourceFolderID),
			SourceDocumentID: trimStringPtr(obj.SourceDocumentID),
			Status:           obj.Status,
			MasteryLevel:     obj.MasteryLevel,
		})
	}
	return result
}

func mapCoachProfiles(profiles []*learning_db.LearningProfile) []prompts.CoachProfile {
	result := make([]prompts.CoachProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profile == nil {
			continue
		}
		result = append(result, prompts.CoachProfile{
			FolderID:        profile.FolderID,
			CurrentLevel:    trimStringPtr(profile.CurrentLevel),
			CompletedTopics: profile.CompletedTopics,
			WeakPoints:      profile.WeakPoints,
			NextGoal:        trimStringPtr(profile.NextGoal),
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
	return &action
}
