package llm

import (
	"context"

	"verve/infrastructure/llm/prompts"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewObjectiveGeneratorAgent 读取 Markdown 并生成学习小节。
func NewObjectiveGeneratorAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ObjectiveGenerator",
		Description: "从 Wiki Markdown 资料生成费曼学习小节",
		Instruction: prompts.ObjectiveGeneratorPrompt(prompts.Input{}),
		Model:       chatModel,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// NewFeynmanReviewerAgent listens to a learner explanation and returns structured feedback.
func NewFeynmanReviewerAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx)
	if err != nil {
		return nil, err
	}
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "FeynmanReviewer",
		Description: "根据完整文档或检索证据倾听并审阅学习者解释",
		Instruction: prompts.FeynmanReviewerPrompt(prompts.Input{}),
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
		Instruction: prompts.CoachPrompt(prompts.Input{}),
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
		Instruction: prompts.TutorPrompt(prompts.Input{}),
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
		Instruction: prompts.ExaminerPrompt(prompts.Input{}),
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
