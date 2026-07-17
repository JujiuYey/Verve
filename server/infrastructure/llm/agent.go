package llm

import (
	"context"

	"verve/infrastructure/llm/prompts"

	"github.com/cloudwego/eino/adk"
)

// agent_key / scene_key 映射,与前端 Agent 配置页对齐。
const (
	// AgentKeyKnowledgeQA retains the persisted "coach" key for model-config compatibility.
	AgentKeyKnowledgeQA     = "coach"
	AgentKeyLearningTeacher = "learning_teacher"
	AgentKeyFeynmanReviewer = "feynman_reviewer"
	AgentKeyWikiCurator     = "wiki_curator"

	SceneKeyDefault = "default"
)

// NewFeynmanReviewerAgent listens to a learner explanation and returns structured feedback.
func NewFeynmanReviewerAgent(ctx context.Context, resolver AgentModelResolver) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx, resolver, AgentKeyFeynmanReviewer, SceneKeyDefault)
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
func NewLearningTeacherAgent(ctx context.Context, resolver AgentModelResolver) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx, resolver, AgentKeyLearningTeacher, SceneKeyDefault)
	if err != nil {
		return nil, err
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "LearningTeacher", Description: "根据当前 Wiki 证据回答学习问题",
		Instruction: prompts.LearningTeacherPrompt(prompts.Input{}), Model: chatModel,
	})
}

// NewWikiCuratorAgent proposes complete Markdown changes without write tools.
func NewWikiCuratorAgent(ctx context.Context, resolver AgentModelResolver) (adk.Agent, error) {
	chatModel, err := NewStructuredChatModel(ctx, resolver, AgentKeyWikiCurator, SceneKeyDefault)
	if err != nil {
		return nil, err
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "WikiCurator", Description: "根据用户要求提出完整 Wiki Markdown 修改建议",
		Instruction: prompts.WikiCuratorPrompt(prompts.Input{}), Model: chatModel,
	})
}
