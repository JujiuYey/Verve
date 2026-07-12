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

// NewLearningTeacherAgent answers learning questions from grounded Wiki evidence.
func NewLearningTeacherAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "LearningTeacher", Description: "根据当前 Wiki 证据回答学习问题",
		Instruction: prompts.LearningTeacherPrompt(prompts.Input{}), Model: chatModel,
	})
}

// NewWikiCuratorAgent proposes complete Markdown changes without write tools.
func NewWikiCuratorAgent(ctx context.Context) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "WikiCurator", Description: "根据用户要求提出完整 Wiki Markdown 修改建议",
		Instruction: prompts.WikiCuratorPrompt(prompts.Input{}), Model: chatModel,
	})
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
