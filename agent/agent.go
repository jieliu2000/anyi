package agent

import (
	"errors"
	"fmt"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"

	log "github.com/sirupsen/logrus"
)

// GetName returns the agent's name.
func (a Agent) GetName() string {
	return a.Name
}

// GetFlows returns the list of flows available to the agent.
func (a Agent) GetFlows() []string {
	return a.Flows
}

// GetClientName returns the name of the LLM client used for planning.
func (a Agent) GetClientName() string {
	return a.ClientName
}

// GetClient retrieves the LLM client for this agent from the registry.
func (a *Agent) GetClient(registry AgentRegistry) (llm.Client, error) {
	if a.ClientName == "" {
		return nil, errors.New("agent has no client configured")
	}
	return registry.GetClient(a.ClientName)
}

// AgentRegistry defines the interface for agent registry operations needed by agents.
type AgentRegistry interface {
	GetFlows(agent *Agent) ([]*flow.Flow, error)
	GetClient(clientName string) (llm.Client, error)
}

// Execute executes the given objective using the provided registry functions.
func (a *Agent) ExecuteWithRegistry(objective string, getFlow func(string) (*flow.Flow, error), getClient func(string) (llm.Client, error)) (*TaskResult, error) {
	if objective == "" {
		return nil, errors.New("objective cannot be empty")
	}

	log.Infof("Agent '%s' executing objective: %s", a.Name, objective)

	// Initialize memory if not set
	if a.Memory == nil {
		a.Memory = NewSimpleMemory()
	}

	// Create registry adapter with function pointers
	registryAdapter := &FunctionalRegistryAdapter{
		Functions: RegistryFunctions{
			GetFlow:   getFlow,
			GetClient: getClient,
		},
	}

	// Create executor
	executor, err := NewAgentExecutor(registryAdapter, a, a.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create planner
	planner, err := NewTaskPlanner(registryAdapter, a)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner: %w", err)
	}

	// Generate execution plan
	plan, err := planner.PlanExecution(objective)
	if err != nil {
		return nil, fmt.Errorf("failed to plan execution: %w", err)
	}

	// Execute the plan
	result, err := executor.Execute(plan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute plan: %w", err)
	}

	return result, nil
}

// Execute executes the given objective using the global anyi registry.
// This method will be called from the main anyi package.
func (a *Agent) Execute(objective string) (*TaskResult, error) {
	// This will be implemented by the main anyi package via method injection
	return nil, fmt.Errorf("Execute method requires registry access - use ExecuteWithRegistry instead")
}

// GetExecutionHistory returns the execution history for this agent.
func (a *Agent) GetExecutionHistory() []*TaskResult {
	if a.Memory == nil {
		return []*TaskResult{}
	}
	return a.Memory.GetTaskHistory()
}

// ClearMemory clears the agent's memory.
func (a *Agent) ClearMemory() {
	if a.Memory != nil {
		a.Memory.Clear()
	}
}

// IsExecuting checks if the agent is currently executing a task.
func (a *Agent) IsExecuting() bool {
	// This is a simple implementation - in a real scenario,
	// you might want to track execution state
	return false
}

// NewAgent creates a new Agent instance with the given parameters.
func NewAgent(name, description, clientName string, flows []string) *Agent {
	return &Agent{
		Name:        name,
		Description: description,
		Flows:       flows,
		ClientName:  clientName,
		Memory:      NewSimpleMemory(),
		Config:      make(map[string]any),
	}
}
