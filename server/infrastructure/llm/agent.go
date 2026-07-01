package llm

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

const plannerInstruction = `你是学习路线规划专家,面向技术/编程类自学者。
根据用户的一句话学习目标,拆解成由浅入深的阶段化学习路线。

只输出 JSON,不要任何额外文字,格式:
{"stages":[{"title":"阶段标题","objectives":[{"title":"小目标标题","detail":"小目标要点"}]}]}

要求:
- 每个小目标聚焦一个认知点,可在一节课内完成;
- 阶段由基础到进阶有序排列;
- 数量适中(通常 2-4 个阶段,每阶段 2-5 个小目标)。`

const tutorInstruction = `你是费曼式编程陪练。教学铁律:
- 每次只推进一个认知点,不要一次塞太多;
- 先用一个小问题诊断,再简短讲解;
- 不连续追问;
- 讲完给一个小练习,并要求用户用解释/代码/运行结果来验证;
- 用提问帮用户把概念讲出来,不空泛鼓励。`

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

const orchestratorInstruction = `你是 Learning Orchestrator,负责决定学习者现在应该做什么。
你不是教学 agent,也不是判题 agent。你只根据用户意图、当前学习路线、学习画像、最近验证记录和候选动作,选择 1-5 个下一步动作。

只输出 JSON,不要任何额外文字,格式:
{"summary":"你对当前调度的简短判断","habit_summary":"基于学习画像/最近验证总结出的学习习惯或风险","actions":[{"id":"候选动作 id","priority":100,"label":"短标签","title":"给用户看的动作标题","description":"为什么现在做这个","reason":"选择这个动作的依据"}]}

要求:
- 禁止输出 <think>、markdown 代码块、解释文字或草稿,顶层第一个字符必须是 {,最后一个字符必须是 };
- actions 只能使用输入里的 candidate_actions.id,不能编造 id、goal_id、objective_id 或 action type;
- 如果用户输入了新的学习意图,通常优先选择 create_goal,但如果最近有明显 fail/partial,可以把复习动作排在其后;
- 如果学习画像里有 next_goal 或 weak_points,优先安排小范围复习,不要直接推进太快;
- 如果当前小目标存在且最近没有薄弱点,推荐继续当前小目标;
- label 要很短,例如"生成新路线"、"推荐继续"、"补弱点"、"复习一下";
- description 面向用户,具体说明要做什么,不要空泛鼓励;
- priority 越高越靠前,取 1-100。`

// NewPlannerAgent 学习路线规划 agent(无工具,直接产出 JSON 路线)
func NewPlannerAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Planner",
		Description: "把学习目标拆成阶段化学习路线",
		Instruction: plannerInstruction,
		Model:       chatModel,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

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

// NewOrchestratorAgent 学习调度 agent(选择下一步产品动作)
func NewOrchestratorAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "LearningOrchestrator",
		Description: "根据学习状态选择下一步产品动作",
		Instruction: orchestratorInstruction,
		Model:       chatModel,
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
