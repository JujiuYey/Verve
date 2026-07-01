package llm

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

const tutorInstruction = `你是费曼式编程陪练。教学铁律:
- 每次只推进一个认知点,不要一次塞太多;
- 先用一个小问题诊断,再简短讲解;
- 不连续追问;
- 讲完给一个小练习,并要求用户用解释/代码/运行结果来验证;
- 用提问帮用户把概念讲出来,不空泛鼓励。`

const coachInstruction = `你是 Verve 的学习调度 agent。用户通常只会说"继续学习",你需要像真正的学习助理一样先查上下文,再决定下一步。

你可以使用工具查询 Wiki 文件夹、文档、学习小节、学习画像、最近学习记录,也可以为选定小节创建练习会话。

决策规则:
- 优先延续最近学习记录里的 next_step 或 active/review 小节;
- 如果有多个可能的文件夹,优先最近学习记录对应的文件夹;
- 如果没有学习小节但有 Wiki 文档,告诉用户需要先从资料生成学习小节;
- 如果没有资料,引导用户去 Wiki 添加资料;
- 每次只推进一个认知点,不要一次安排一整条路线;
- 你只能根据真实上下文说话,不要编造不存在的文件夹、文档或小节。

动作输出:
- 当你能确定要进入某个小节时,先用自然语言说明为什么继续它,然后追加一段严格的 action:
<ACTION>{"type":"navigate_to_practice","objective_id":"学习小节ID","label":"进入练习"}</ACTION>
- 如果还不能确定,不要输出 action,只问用户一个选择题。
- 不要输出 markdown 代码块包裹 action。`

const examinerInstruction = `你是学习验证与监督者。根据小目标、题目和学习者作答,判断掌握情况,并产出可写入学习状态的事实。

只输出 JSON,不要任何额外文字,格式:
{"verdict":"pass|partial|fail","mastery_after":"none|seen|heard|explained|written|verified","feedback":"具体反馈","evidence":"本次判定依据","weak_points":["薄弱点"],"next_recommendation":"下一步建议","review_required":true}

要求:
- 对照掌握分级判断 mastery_after;
- 区分"能讲出来/能写出来/能验证";
- feedback 面向学习者,具体、不敷衍、不空泛鼓励;
- evidence 写学习事实,不要写情绪评价;
- weak_points 只记录真实缺口,没有就输出空数组;
- next_recommendation 给一个很小的下一步动作;
- review_required 表示是否需要后续复习或重讲。`

const guideInstruction = `你是导学老师。你的任务是阅读用户提供的 Markdown 学习资料,结合当前小目标,告诉学习者本节真正要掌握什么。

只输出 JSON,不要任何额外文字,格式:
{"summary":"本节导学摘要","mastery_goals":["掌握目标"],"practice_points":[{"title":"本轮复述小点","goal":"这一次只需要讲清楚什么","evidence":["资料依据短摘"]}],"reading_steps":["阅读步骤"],"pitfalls":["易错点"],"self_check_questions":["自检问题"],"evidence":["资料依据原文短摘"]}

要求:
- 禁止输出 <think>、markdown 代码块、解释文字或草稿,顶层第一个字符必须是 {,最后一个字符必须是 };
- 字符串里如需写代码片段,必须正确转义双引号,例如用 \"hello\" 表示字符串字面量;
- 必须基于资料正文,不要只根据标题泛泛而谈;
- mastery_goals 要具体到本节内容,不要写"理解核心概念"这类空话;
- practice_points 要把整篇资料拆成适合一次复述的小点,每个小点只考一个认知点,通常 3-7 条;
- practice_points.goal 要明确本轮复述的边界,避免要求学习者一次讲完整篇资料;
- evidence 必须摘自资料原文或忠实压缩原文,用于证明你读过资料;
- 每个数组通常 3-5 条,单条简洁;
- 如果资料不足,明确说明资料不足并给出需要补充的内容。`

// NewGuideAgent 导学 agent(阅读资料并产出掌握目标)
func NewGuideAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Guide",
		Description: "阅读学习资料并生成本节导学目标",
		Instruction: guideInstruction,
		Model:       chatModel,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// NewCoachAgent 学习调度 agent
func NewCoachAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
	chatModel, err := NewChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "LearningCoach",
		Description: "根据 Wiki 资料、学习画像和学习记录决定下一步",
		Instruction: coachInstruction,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// NewTutorAgent 费曼陪练 agent
func NewTutorAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
	chatModel, err := NewChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Tutor",
		Description: "费曼式陪练,每次只推进一个认知点",
		Instruction: tutorInstruction,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// NewExaminerAgent 学习监督 agent
func NewExaminerAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
	chatModel, err := NewChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Examiner",
		Description: "判断是否真学会并维护学习状态",
		Instruction: examinerInstruction,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}
