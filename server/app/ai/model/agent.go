package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"

	"sag-wiki/app/ai/tools"
	"sag-wiki/app/ai/repository"
	"sag-wiki/infrastructure/database"
)

// NewAgentWithSystemTools creates a system management agent with user/dept/role tools
func NewAgentWithSystemTools(ctx context.Context, dbService *database.DatabaseService, modelRepo repository.ModelConfigRepository) (adk.Agent, error) {
	// Create chat model
	chatModel, err := NewChatModel(ctx, modelRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// Create system tools
	systemTools := tools.NewSystemTools(
		dbService.Users,
		dbService.Departments,
		dbService.Roles,
	)

	// Create the agent
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SystemManager",
		Description: "A system management agent that can help with user, department and role management",
		Instruction: `You are a helpful system management assistant. You can help users manage:
- Users: search, create, update, delete user accounts
- Departments: manage organizational departments with hierarchy support
- Roles: manage user roles and permissions

When a user asks to perform an operation, use the appropriate tool to complete the task.
Always confirm the result of operations to the user.
For create operations, report back the created entity with its ID.`,
		Model: chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: systemTools,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return a, nil
}

// AgentRunner wraps adk.Runner for easier use
type AgentRunner struct {
	runner *adk.Runner
}

func NewAgentRunner(ctx context.Context, a adk.Agent) *AgentRunner {
	return &AgentRunner{
		runner: adk.NewRunner(ctx, adk.RunnerConfig{
			EnableStreaming: true,
			Agent:          a,
		}),
	}
}

// Query runs the agent with a query and returns the event iterator
func (r *AgentRunner) Query(ctx context.Context, query string) *adk.AsyncIterator[*adk.AgentEvent] {
	return r.runner.Query(ctx, query)
}

// GetRunner returns the underlying adk.Runner (for streaming SSE)
func (r *AgentRunner) GetRunner() *adk.Runner {
	return r.runner
}
