package model

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	ai_model "sag-wiki/app/ai/model"
	system_repo "sag-wiki/app/system/repository"
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

const examinerInstruction = `你是学习验证与监督者。根据小目标、题目和学习者作答,判断掌握情况。

只输出 JSON,不要任何额外文字,格式:
{"verdict":"pass|partial|fail","mastery_after":"none|seen|heard|explained|written|verified","feedback":"具体反馈"}

要求:
- 对照掌握分级判断 mastery_after;
- 区分"能讲出来/能写出来/能验证";
- 反馈具体、不敷衍、不空泛鼓励。`

// NewPlannerAgent 学习路线规划 agent(无工具,直接产出 JSON 路线)
func NewPlannerAgent(ctx context.Context, modelRepo system_repo.ModelConfigRepository) (adk.Agent, error) {
	chatModel, err := ai_model.NewPlannerChatModel(ctx, modelRepo)
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

// NewTutorAgent 费曼陪练 agent
func NewTutorAgent(ctx context.Context, modelRepo system_repo.ModelConfigRepository, tools []tool.BaseTool) (adk.Agent, error) {
	chatModel, err := ai_model.NewChatModel(ctx, modelRepo)
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
func NewExaminerAgent(ctx context.Context, modelRepo system_repo.ModelConfigRepository, tools []tool.BaseTool) (adk.Agent, error) {
	chatModel, err := ai_model.NewChatModel(ctx, modelRepo)
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
