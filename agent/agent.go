package agent

import (
	"fmt"
	"strings"
	"time"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

// FlowGetter dependency interface - resolves circular references
type FlowGetter interface {
	GetFlow(name string) (*flow.Flow, error)
}

// Agent concrete type
type Agent struct {
	Role           string
	BackStory      string
	AvailableFlows []string
	MaxIterations  int           `mapstructure:"maxIterations" json:"maxIterations" yaml:"maxIterations"`
	MaxRetries     int           `mapstructure:"maxRetries" json:"maxRetries" yaml:"maxRetries"`
	Timeout        time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	getFlow        FlowGetter
	Client         llm.Client
	aiPlanningFlow *flow.Flow
}

// NewAgent creates a new Agent
func NewAgent(role, backstory string, availableFlows []string, getFlow FlowGetter) *Agent {
	return &Agent{
		Role:           role,
		BackStory:      backstory,
		AvailableFlows: availableFlows,
		getFlow:        getFlow,
		MaxIterations:  10,
		MaxRetries:     3,
		Timeout:        30 * time.Minute,
	}
}

// NewAgentWithClient creates a new Agent with LLM client
func NewAgentWithClient(role, backstory string, availableFlows []string, getFlow FlowGetter, client llm.Client) *Agent {
	agent := &Agent{
		Role:           role,
		BackStory:      backstory,
		AvailableFlows: availableFlows,
		getFlow:        getFlow,
		MaxIterations:  10,
		MaxRetries:     3,
		Timeout:        30 * time.Minute,
		Client:         client,
	}

	// Note: AI planning flow will be created lazily during planning if needed
	// This avoids unnecessary initialization overhead when the agent might not use AI planning

	return agent
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

		// Execute Flow using Execute method (for backward compatibility)
		// Check if flow implements Execute method
		if executableFlow, ok := interface{}(flow).(interface {
			Execute(input string, ctx map[string]interface{}) (string, map[string]interface{}, error)
		}); ok {
			// Use Execute method
			result, resultCtx.Variables, err = executableFlow.Execute(result, resultCtx.Variables)
			if err != nil {
				if step.Retryable && i < a.MaxRetries {
					continue // Retry current step
				}
				return "", resultCtx, fmt.Errorf("execute flow %s: %w", step.FlowName, err)
			}
		} else {
			// Use Run method (new approach)
			flowContext := flow.NewFlowContext(result, resultCtx.Variables)
			resultFlowContext, err := flow.Run(*flowContext)
			if err != nil {
				if step.Retryable && i < a.MaxRetries {
					continue // Retry current step
				}
				return "", resultCtx, fmt.Errorf("execute flow %s: %w", step.FlowName, err)
			}
			result = resultFlowContext.Text
			resultCtx.Variables = resultFlowContext.Variables
		}

		resultCtx.History = append(resultCtx.History, result)

		// Check if the goal is completed
		if a.isTaskCompleted(result, task) {
			break
		}

		// Check if maximum iterations exceeded
		if i >= a.MaxIterations-1 {
			break
		}
	}

	return result, resultCtx, nil
}

// createAIPlanningFlow creates a flow for AI-based planning
// This flow consists of two steps: LLM step for generating the plan and PlanParser step for parsing the response
func (a *Agent) createAIPlanningFlow() {
	if a.Client == nil {
		return // No LLM client available, cannot create AI planning flow
	}

	// Create LLM step for generating the plan
	llmStep, err := a.createLLMStep()
	if err != nil {
		// Log error but don't fail agent creation
		return
	}
	llmStep.Name = "llm-planning"

	// Create PlanParser step for parsing the LLM response
	planParserExecutor := NewPlanParserExecutor(a.AvailableFlows)
	planParserStep := flow.NewStep(planParserExecutor, nil, nil)
	planParserStep.Name = "plan-parser"

	// Create the AI planning flow
	aiPlanningFlow, err := flow.NewFlow(a.Client, "ai-planning", *llmStep, *planParserStep)
	if err != nil {
		// Log error but don't fail agent creation
		// In a real implementation, you might want to log this error
		return
	}

	a.aiPlanningFlow = aiPlanningFlow
}

// createLLMStep creates an LLM step for the AI planning flow
func (a *Agent) createLLMStep() (*flow.Step, error) {
	// Create a template formatter
	formatter, err := chat.NewPromptTemplateFormatter(a.createPlanningTemplate())
	if err != nil {
		return nil, err
	}

	// Create a custom LLM executor that implements the StepExecutor interface
	llmExecutor := &agentLLMExecutor{
		templateFormatter: formatter,
		systemMessage:     "You are an intelligent task planner. Based on the task description and available flows, create an optimal execution plan.",
		outputJSON:        true,
	}

	// Initialize the executor
	if err := llmExecutor.Init(); err != nil {
		return nil, err
	}

	return flow.NewStep(llmExecutor, nil, a.Client), nil
}

// agentLLMExecutor is a custom LLM executor implementation that avoids importing executors package
type agentLLMExecutor struct {
	templateFormatter chat.PromptFormatter
	systemMessage     string
	outputJSON        bool
}

// Init initializes the executor
func (e *agentLLMExecutor) Init() error {
	// The template formatter is already created, nothing else to initialize
	return nil
}

// Run executes the LLM step
func (e *agentLLMExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	if step == nil {
		return nil, fmt.Errorf("no step provided")
	}

	if step.GetClient() == nil {
		step.ClientImpl = flowContext.Flow.ClientImpl
	}
	if step.ClientImpl == nil {
		return nil, fmt.Errorf("no client set for flow step")
	}

	// Format the prompt using the template formatter
	input, err := e.templateFormatter.Format(flowContext)
	if err != nil {
		return nil, err
	}

	// Create messages
	messages := make([]chat.Message, 0, 2)
	if e.systemMessage != "" {
		messages = append(messages, chat.NewSystemMessage(e.systemMessage))
	}
	messages = append(messages, chat.NewUserMessage(input))

	// Set options if JSON output is requested
	var options *chat.ChatOptions
	if e.outputJSON {
		options = &chat.ChatOptions{
			Format: "json",
		}
	}

	// Call the LLM
	output, _, err := step.ClientImpl.Chat(messages, options)
	if err != nil {
		return nil, err
	}

	// Update the flow context with the response
	flowContext.Text = output.Content
	return &flowContext, nil
}

// createPlanningTemplate creates a template for AI planning
func (a *Agent) createPlanningTemplate() string {
	availableFlows := strings.Join(a.AvailableFlows, ", ")

	template := `Task: {{.Text}}

Agent Role: {{.Flow.Variables.AgentRole}}
Agent Background: {{.Flow.Variables.AgentBackground}}

Available Flows: ` + availableFlows + `

Context Variables: {{.Flow.Variables.ContextVariables}}

Please create an optimal execution plan by selecting and ordering the most appropriate flows from the available flows. 
Respond with a JSON array of flow names in the order they should be executed.
Example response: ["flow1", "flow2", "flow3"]

Only respond with the JSON array, nothing else.`

	return template
}

// isTaskCompleted checks if the task is completed (simplified implementation)
func (a *Agent) isTaskCompleted(result, task string) bool {
	// Simple implementation: consider completed if result length exceeds twice the task length
	return len(result) > len(task)*2
}
