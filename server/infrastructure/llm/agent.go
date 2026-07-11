package llm

import (
	"context"

	"verve/infrastructure/llm/prompts"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

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
