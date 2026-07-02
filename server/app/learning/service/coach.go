package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
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
// 输出包含 6 个 Markdown 段:用户输入、Wiki 文件夹、当前文档、学习小节、学习画像、最近学习记录;
// 末尾给出回复约束(自然语言 + 可选 <ACTION> 标签),供 ParseCoachAction 在回复中抽取跳转动作。
func BuildCoachQuery(ctx CoachRuntimeContext, message string) string {
	var sb strings.Builder
	sb.WriteString("你正在 Verve 的费曼学习入口帮助用户继续学习。\n")
	sb.WriteString("学习者说:")
	sb.WriteString(strings.TrimSpace(message))
	sb.WriteString("\n\n")

	sb.WriteString("## 当前 Wiki 文件夹\n")
	if len(ctx.Folders) == 0 {
		sb.WriteString("- 暂无文件夹\n")
	} else {
		for _, folder := range ctx.Folders {
			sb.WriteString(fmt.Sprintf("- %s (%s)", folder.Name, folder.ID))
			if folder.Description != nil && strings.TrimSpace(*folder.Description) != "" {
				sb.WriteString(" - ")
				sb.WriteString(strings.TrimSpace(*folder.Description))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n## 当前文档\n")
	if len(ctx.Documents) == 0 {
		sb.WriteString("- 暂无文档\n")
	} else {
		for _, doc := range ctx.Documents {
			sb.WriteString(fmt.Sprintf("- %s (%s), folder=%s\n", doc.Filename, doc.ID, doc.FolderID))
		}
	}

	sb.WriteString("\n## 学习小节\n")
	if len(ctx.Objectives) == 0 {
		if len(ctx.Documents) > 0 {
			sb.WriteString("- 暂无学习小节。请先选择最适合当前继续学习的一篇文档,调用 create_learning_objectives 生成学习小节。生成成功后,用 first_objective_id 输出 navigate_to_practice action。\n")
		} else {
			sb.WriteString("- 暂无学习小节。可以建议用户先去 Wiki 补充资料。\n")
		}
	} else {
		for _, obj := range ctx.Objectives {
			sb.WriteString(fmt.Sprintf("- %s (%s), status=%s, mastery=%s", obj.Title, obj.ID, obj.Status, obj.MasteryLevel))
			if obj.SourceFolderID != nil {
				sb.WriteString(", folder=")
				sb.WriteString(*obj.SourceFolderID)
			}
			if obj.SourceDocumentID != nil {
				sb.WriteString(", document=")
				sb.WriteString(*obj.SourceDocumentID)
			}
			if obj.Detail != nil && strings.TrimSpace(*obj.Detail) != "" {
				sb.WriteString("\n  要点:")
				sb.WriteString(strings.TrimSpace(*obj.Detail))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n## 学习画像\n")
	if len(ctx.Profiles) == 0 {
		sb.WriteString("- 暂无画像\n")
	} else {
		for _, profile := range ctx.Profiles {
			sb.WriteString(fmt.Sprintf("- folder=%s", profile.FolderID))
			if profile.CurrentLevel != nil {
				sb.WriteString(", 当前水平:")
				sb.WriteString(*profile.CurrentLevel)
			}
			if len(profile.CompletedTopics) > 0 {
				sb.WriteString(", 已掌握内容:")
				sb.WriteString(strings.Join(profile.CompletedTopics, "、"))
			}
			if len(profile.WeakPoints) > 0 {
				sb.WriteString(", 薄弱点:")
				sb.WriteString(strings.Join(profile.WeakPoints, "、"))
			}
			if profile.NextGoal != nil && strings.TrimSpace(*profile.NextGoal) != "" {
				sb.WriteString(", 下一目标:")
				sb.WriteString(strings.TrimSpace(*profile.NextGoal))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n## 最近学习记录\n")
	if len(ctx.Journals) == 0 {
		sb.WriteString("- 暂无记录\n")
	} else {
		for _, journal := range ctx.Journals {
			sb.WriteString(fmt.Sprintf("- %s folder=%s", journal.Date.Format(time.DateOnly), journal.FolderID))
			if journal.Learned != nil {
				sb.WriteString(", 学了:")
				sb.WriteString(strings.TrimSpace(*journal.Learned))
			}
			if journal.WeakPoints != nil && strings.TrimSpace(*journal.WeakPoints) != "" {
				sb.WriteString(", 薄弱点:")
				sb.WriteString(strings.TrimSpace(*journal.WeakPoints))
			}
			if journal.NextStep != nil && strings.TrimSpace(*journal.NextStep) != "" {
				sb.WriteString(", 下一步:")
				sb.WriteString(strings.TrimSpace(*journal.NextStep))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n请基于这些真实上下文决定下一步。")
	sb.WriteString("如果没有学习小节但有当前文档,优先调用 create_learning_objectives,不要只停留在口头建议。")
	sb.WriteString("如果已经能确定要进入某个小节,在自然语言回复后追加一段 <ACTION>{\"type\":\"navigate_to_practice\",\"objective_id\":\"...\",\"label\":\"进入练习\"}</ACTION>。")
	sb.WriteString("如果不能确定,只问用户一个选择题。")
	return sb.String()
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
