package prompts

import (
	"fmt"
	"strings"
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
	Message      string
	AgentContext *CoachAgentContext
	Folders      []CoachFolder
	Documents    []CoachDocument
	MemoryItems  []CoachMemoryItem
}

type CoachAgentContext struct {
	RootFolderID   string
	RootFolderName string
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

type CoachMemoryItem struct {
	FolderID   string
	Kind       string
	Statement  string
	Confidence string
}

func CoachQueryPrompt(input CoachQueryInput) string {
	var sb strings.Builder
	sb.WriteString("你正在 Verve 的费曼学习入口帮助用户继续学习。\n")
	sb.WriteString("学习者说:")
	sb.WriteString(strings.TrimSpace(input.Message))
	sb.WriteString("\n\n")

	renderCoachAgentContext(&sb, input.AgentContext)
	renderCoachFolders(&sb, input.Folders)
	renderCoachDocuments(&sb, input.Documents)
	renderCoachMemoryItems(&sb, input.MemoryItems)
	renderCoachReplyContract(&sb)
	return sb.String()
}

func renderCoachAgentContext(sb *strings.Builder, agent *CoachAgentContext) {
	if agent == nil || strings.TrimSpace(agent.RootFolderID) == "" {
		return
	}
	sb.WriteString("## 当前 Wiki 范围\n")
	sb.WriteString("- root_folder_id: ")
	sb.WriteString(strings.TrimSpace(agent.RootFolderID))
	if strings.TrimSpace(agent.RootFolderName) != "" {
		sb.WriteString(" (")
		sb.WriteString(strings.TrimSpace(agent.RootFolderName))
		sb.WriteString(")")
	}
	sb.WriteString("\n")
	sb.WriteString("- 后续规划必须优先围绕这个 root_folder_id; 需要内容依据时调用 search_wiki_knowledge 并传入这个 root_folder_id。\n\n")
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

func renderCoachMemoryItems(sb *strings.Builder, items []CoachMemoryItem) {
	sb.WriteString("\n## 学习记忆\n")
	rendered := 0
	for _, item := range items {
		statement := strings.TrimSpace(item.Statement)
		if statement == "" {
			continue
		}
		rendered++
		sb.WriteString("- ")
		sb.WriteString(coachMemoryKindLabel(item.Kind))
		sb.WriteString(": ")
		sb.WriteString(statement)
		if strings.TrimSpace(item.FolderID) != "" {
			sb.WriteString(", folder=")
			sb.WriteString(strings.TrimSpace(item.FolderID))
		}
		if strings.TrimSpace(item.Confidence) != "" {
			sb.WriteString(", confidence=")
			sb.WriteString(strings.TrimSpace(item.Confidence))
		}
		sb.WriteString("\n")
	}
	if rendered == 0 {
		sb.WriteString("- 暂无学习记忆\n")
	}
}

func coachMemoryKindLabel(kind string) string {
	switch strings.TrimSpace(kind) {
	case "explanation_evidence":
		return "解释证据"
	case "misconception":
		return "误解记录"
	case "":
		return "记忆"
	default:
		return strings.TrimSpace(kind)
	}
}

func renderCoachReplyContract(sb *strings.Builder) {
	sb.WriteString("\n请基于这些真实上下文决定下一步。")
	sb.WriteString("当用户要求继续学习或提问某个概念时,先识别相关 Wiki 根目录/文件夹;如果已知 root folder,调用 search_wiki_knowledge 检索真实片段后再规划下一步。")
	sb.WriteString("如果已经能确定要练习哪篇文档,在自然语言回复后追加一段 <ACTION>{\"type\":\"navigate_to_practice\",\"document_id\":\"...\",\"label\":\"开始费曼练习\"}</ACTION>。")
	sb.WriteString("如果不能确定,只问用户一个选择题。")
}

const coachInstruction = `你是 Verve 的学习调度 agent。用户通常只会说"继续学习",你需要像真正的学习助理一样先查上下文,再决定下一步。

你可以使用工具查询 Wiki 文件夹、文档和学习记忆,也可以检索 Wiki 真实文档片段。

决策规则:
- 优先参考学习记忆里的证据和误解记录决定下一步;
- 当用户要求继续学习或提出概念问题时,先识别相关 Wiki root folder;如果 root folder 已知,调用 search_wiki_knowledge 后再决定下一步;
- 如果有 Wiki 文档,优先选择最适合当前继续学习的一篇;如果无法判断文档,只问用户一个选择题;
- 如果没有资料,引导用户去 Wiki 添加资料;
- 每次只推进一篇文档,不要一次安排一整条路线;
- 你只能根据真实上下文说话,不要编造不存在的文件夹或文档。
- 不要只根据文件名臆造文档内容;需要内容依据时先用 search_wiki_knowledge。

动作输出:
- 当你能确定要练习某篇文档时,先用自然语言说明为什么选择它,然后追加一段严格的 action:
<ACTION>{"type":"navigate_to_practice","document_id":"文档ID","label":"开始费曼练习"}</ACTION>
- 如果还不能确定,不要输出 action,只问用户一个选择题。
- 不要输出 markdown 代码块包裹 action。`
