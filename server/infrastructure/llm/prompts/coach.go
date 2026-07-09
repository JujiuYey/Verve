package prompts

import (
	"fmt"
	"strings"
	"time"
)

func CoachPrompt(input Input) string {
	input = normalizeInput(input)
	switch input.PresetKey {
	case DefaultPresetKey:
		return coachInstruction
	default:
		return coachInstruction
	}
}

type CoachQueryInput struct {
	Message    string
	Folders    []CoachFolder
	Documents  []CoachDocument
	Objectives []CoachObjective
	Profiles   []CoachProfile
	Journals   []CoachJournal
}

type CoachFolder struct {
	ID          string
	Name        string
	Description string
}

type CoachDocument struct {
	ID       string
	FolderID string
	Filename string
}

type CoachObjective struct {
	ID               string
	Title            string
	Detail           string
	SourceFolderID   string
	SourceDocumentID string
	Status           string
	MasteryLevel     string
}

type CoachProfile struct {
	FolderID        string
	CurrentLevel    string
	CompletedTopics []string
	WeakPoints      []string
	NextGoal        string
}

type CoachJournal struct {
	FolderID   string
	Date       time.Time
	Learned    string
	WeakPoints string
	NextStep   string
}

func CoachQueryPrompt(input CoachQueryInput) string {
	var sb strings.Builder
	sb.WriteString("你正在 Verve 的费曼学习入口帮助用户继续学习。\n")
	sb.WriteString("学习者说:")
	sb.WriteString(strings.TrimSpace(input.Message))
	sb.WriteString("\n\n")

	renderCoachFolders(&sb, input.Folders)
	renderCoachDocuments(&sb, input.Documents)
	renderCoachObjectives(&sb, input.Objectives, len(input.Documents) > 0)
	renderCoachProfiles(&sb, input.Profiles)
	renderCoachJournals(&sb, input.Journals)
	renderCoachReplyContract(&sb)
	return sb.String()
}

func renderCoachFolders(sb *strings.Builder, folders []CoachFolder) {
	sb.WriteString("## 当前 Wiki 文件夹\n")
	if len(folders) == 0 {
		sb.WriteString("- 暂无文件夹\n")
		return
	}
	for _, folder := range folders {
		sb.WriteString(fmt.Sprintf("- %s (%s)", folder.Name, folder.ID))
		if strings.TrimSpace(folder.Description) != "" {
			sb.WriteString(" - ")
			sb.WriteString(strings.TrimSpace(folder.Description))
		}
		sb.WriteString("\n")
	}
}

func renderCoachDocuments(sb *strings.Builder, documents []CoachDocument) {
	sb.WriteString("\n## 当前文档\n")
	if len(documents) == 0 {
		sb.WriteString("- 暂无文档\n")
		return
	}
	for _, doc := range documents {
		sb.WriteString(fmt.Sprintf("- %s (%s), folder=%s\n", doc.Filename, doc.ID, doc.FolderID))
	}
}

func renderCoachObjectives(sb *strings.Builder, objectives []CoachObjective, hasDocuments bool) {
	sb.WriteString("\n## 学习小节\n")
	if len(objectives) == 0 {
		if hasDocuments {
			sb.WriteString("- 暂无学习小节。请先选择最适合当前继续学习的一篇文档,调用 create_learning_objectives 生成学习小节。生成成功后,用 first_objective_id 输出 navigate_to_practice action。\n")
		} else {
			sb.WriteString("- 暂无学习小节。可以建议用户先去 Wiki 补充资料。\n")
		}
		return
	}
	for _, obj := range objectives {
		sb.WriteString(fmt.Sprintf("- %s (%s), status=%s, mastery=%s", obj.Title, obj.ID, obj.Status, obj.MasteryLevel))
		if obj.SourceFolderID != "" {
			sb.WriteString(", folder=")
			sb.WriteString(obj.SourceFolderID)
		}
		if obj.SourceDocumentID != "" {
			sb.WriteString(", document=")
			sb.WriteString(obj.SourceDocumentID)
		}
		if strings.TrimSpace(obj.Detail) != "" {
			sb.WriteString("\n  要点:")
			sb.WriteString(strings.TrimSpace(obj.Detail))
		}
		sb.WriteString("\n")
	}
}

func renderCoachProfiles(sb *strings.Builder, profiles []CoachProfile) {
	sb.WriteString("\n## 学习画像\n")
	if len(profiles) == 0 {
		sb.WriteString("- 暂无画像\n")
		return
	}
	for _, profile := range profiles {
		sb.WriteString(fmt.Sprintf("- folder=%s", profile.FolderID))
		if profile.CurrentLevel != "" {
			sb.WriteString(", 当前水平:")
			sb.WriteString(profile.CurrentLevel)
		}
		if len(profile.CompletedTopics) > 0 {
			sb.WriteString(", 已掌握内容:")
			sb.WriteString(strings.Join(profile.CompletedTopics, "、"))
		}
		if len(profile.WeakPoints) > 0 {
			sb.WriteString(", 薄弱点:")
			sb.WriteString(strings.Join(profile.WeakPoints, "、"))
		}
		if strings.TrimSpace(profile.NextGoal) != "" {
			sb.WriteString(", 下一目标:")
			sb.WriteString(strings.TrimSpace(profile.NextGoal))
		}
		sb.WriteString("\n")
	}
}

func renderCoachJournals(sb *strings.Builder, journals []CoachJournal) {
	sb.WriteString("\n## 最近学习记录\n")
	if len(journals) == 0 {
		sb.WriteString("- 暂无记录\n")
		return
	}
	for _, journal := range journals {
		sb.WriteString(fmt.Sprintf("- %s folder=%s", journal.Date.Format(time.DateOnly), journal.FolderID))
		if journal.Learned != "" {
			sb.WriteString(", 学了:")
			sb.WriteString(strings.TrimSpace(journal.Learned))
		}
		if strings.TrimSpace(journal.WeakPoints) != "" {
			sb.WriteString(", 薄弱点:")
			sb.WriteString(strings.TrimSpace(journal.WeakPoints))
		}
		if strings.TrimSpace(journal.NextStep) != "" {
			sb.WriteString(", 下一步:")
			sb.WriteString(strings.TrimSpace(journal.NextStep))
		}
		sb.WriteString("\n")
	}
}

func renderCoachReplyContract(sb *strings.Builder) {
	sb.WriteString("\n请基于这些真实上下文决定下一步。")
	sb.WriteString("当用户要求继续学习或提问某个概念时,先识别相关 Wiki 根目录/文件夹;如果已知 root folder,调用 search_wiki_knowledge 检索真实片段后再规划下一步。")
	sb.WriteString("如果没有学习小节但有当前文档,优先调用 create_learning_objectives,不要只停留在口头建议。")
	sb.WriteString("如果已经能确定要进入某个小节,在自然语言回复后追加一段 <ACTION>{\"type\":\"navigate_to_practice\",\"objective_id\":\"...\",\"label\":\"进入练习\"}</ACTION>。")
	sb.WriteString("如果不能确定,只问用户一个选择题。")
}

const coachInstruction = `你是 Verve 的学习调度 agent。用户通常只会说"继续学习",你需要像真正的学习助理一样先查上下文,再决定下一步。

你可以使用工具查询 Wiki 文件夹、文档、学习小节、学习画像、最近学习记录,也可以检索 Wiki 真实文档片段、从 Wiki 文档生成学习小节,并为选定小节创建练习会话。

决策规则:
- 优先延续最近学习记录里的 next_step 或 active/review 小节;
- 如果有多个可能的文件夹,优先最近学习记录对应的文件夹;
- 当用户要求继续学习或提出概念问题时,先识别相关 Wiki root folder;如果 root folder 已知,调用 search_wiki_knowledge 后再决定下一步;
- 如果没有学习小节但有 Wiki 文档,优先选择最适合当前继续学习的文档并调用 create_learning_objectives 生成学习小节;如果无法判断文档,只问用户一个选择题;
- 如果没有资料,引导用户去 Wiki 添加资料;
- 每次只推进一个认知点,不要一次安排一整条路线;
- 你只能根据真实上下文说话,不要编造不存在的文件夹、文档或小节。
- 不要只根据文件名臆造文档内容;需要内容依据时先用 search_wiki_knowledge。

动作输出:
- 当你能确定要进入某个小节时,先用自然语言说明为什么继续它,然后追加一段严格的 action:
<ACTION>{"type":"navigate_to_practice","objective_id":"学习小节ID","label":"进入练习"}</ACTION>
- 当 create_learning_objectives 成功返回 first_objective_id 时,说明已经从文档生成学习小节,然后追加 navigate_to_practice action 指向 first_objective_id。
- 如果还不能确定,不要输出 action,只问用户一个选择题。
- 不要输出 markdown 代码块包裹 action。`
