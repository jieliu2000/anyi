package agent

import (
	"fmt"
	"time"
	"github.com/jieliu2000/anyi/llm"
)

// FlowGetter dependency interface - resolves circular references
type FlowGetter interface {
	GetFlow(name string) (interface{}, error)
}

// Config Agent configuration
type Config struct {
	MaxIterations int           `mapstructure:"maxIterations" json:"maxIterations" yaml:"maxIterations"`
	MaxRetries    int           `mapstructure:"maxRetries" json:"maxRetries" yaml:"maxRetries"`
	Timeout       time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

// Agent concrete type
type Agent struct {
	Role           string
	BackStory      string
	AvailableFlows []string
	Config         Config
	getFlow        FlowGetter
	Client         llm.Client
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MaxIterations: 10,
		MaxRetries:    3,
		Timeout:       30 * time.Minute,
	}
}

// NewAgent creates a new Agent
func NewAgent(role, backstory string, availableFlows []string, getFlow FlowGetter) *Agent {
	return &Agent{
		Role:           role,
		BackStory:      backstory,
		AvailableFlows: availableFlows,
		getFlow:        getFlow,
		Config:         DefaultConfig(),
	}
}

// NewAgentWithClient creates a new Agent with LLM client
func NewAgentWithClient(role, backstory string, availableFlows []string, getFlow FlowGetter, client llm.Client) *Agent {
	return &Agent{
		Role:           role,
		BackStory:      backstory,
		AvailableFlows: availableFlows,
		getFlow:        getFlow,
		Config:         DefaultConfig(),
		Client:         client,
	}
}

// Execute executes a task - uses value type AgentContext
func (a *Agent) Execute(task string, ctx AgentContext) (string, AgentContext, error) {
	if ctx.Variables == nil {
		ctx.Variables = make(map[string]interface{})
	}

	// Use a copy for operations to avoid modifying the original context
	resultCtx := ctx
	result := task

	// 1. Intelligent planning
	plan, err := a.planExecution(task, resultCtx)
	if err != nil {
		return "", ctx, err
	}

	// 2. Execute plan
	for i, step := range plan.Steps {
		flow, err := a.getFlow.GetFlow(step.FlowName)
		if err != nil {
			return "", ctx, fmt.Errorf("get flow %s: %w", step.FlowName, err)
		}

		// Execute Flow
		if executable, ok := flow.(interface {
			Execute(input string, ctx map[string]interface{}) (string, map[string]interface{}, error)
		}); ok {
			result, resultCtx.Variables, err = executable.Execute(result, resultCtx.Variables)
			if err != nil {
				if step.Retryable && i < a.Config.MaxRetries {
					continue // Retry current step
				}
				return "", resultCtx, fmt.Errorf("execute flow %s: %w", step.FlowName, err)
			}
		} else {
			return "", resultCtx, fmt.Errorf("flow %s does not implement Execute method", step.FlowName)
		}

		resultCtx.History = append(resultCtx.History, result)

		// Check if the goal is completed
		if a.isTaskCompleted(result, task) {
			break
		}

		// Check if maximum iterations exceeded
		if i >= a.Config.MaxIterations-1 {
			break
		}
	}

	return result, resultCtx, nil
}

// isTaskCompleted checks if the task is completed (simplified implementation)
func (a *Agent) isTaskCompleted(result, task string) bool {
	// Simple implementation: consider completed if result length exceeds twice the task length
	return len(result) > len(task)*2
}
