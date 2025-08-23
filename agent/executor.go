package agent

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jieliu2000/anyi/flow"

	log "github.com/sirupsen/logrus"
)

// AgentExecutor is responsible for executing the generated execution plans.
type AgentExecutor struct {
	registry AgentRegistry
	agent    *Agent
	memory   AgentMemory
	mutex    sync.RWMutex
}

// NewAgentExecutor creates a new AgentExecutor instance.
func NewAgentExecutor(registry AgentRegistry, agent *Agent, memory AgentMemory) (*AgentExecutor, error) {
	if registry == nil {
		return nil, errors.New("registry cannot be nil")
	}
	if agent == nil {
		return nil, errors.New("agent cannot be nil")
	}
	if memory == nil {
		memory = NewSimpleMemory() // Use default memory if none provided
	}

	return &AgentExecutor{
		registry: registry,
		agent:    agent,
		memory:   memory,
	}, nil
}

// Execute executes the given execution plan and returns the task result.
func (e *AgentExecutor) Execute(plan *ExecutionPlan) (*TaskResult, error) {
	if plan == nil {
		return nil, errors.New("execution plan cannot be nil")
	}

	log.Infof("Starting execution of plan: %s", plan.Objective)

	startTime := time.Now()
	result := &TaskResult{
		Objective:   plan.Objective,
		Status:      "running",
		StartTime:   startTime,
		StepResults: make([]StepResult, 0, len(plan.Steps)),
	}

	// Store the task in memory
	e.memory.StoreTask(result)

	// Sort steps by order
	sortedSteps := make([]ExecutionStep, len(plan.Steps))
	copy(sortedSteps, plan.Steps)
	sort.Slice(sortedSteps, func(i, j int) bool {
		return sortedSteps[i].Order < sortedSteps[j].Order
	})

	// Execute each step
	for i, step := range sortedSteps {
		log.Debugf("Executing step %d: %s", i+1, step.FlowName)

		stepResult, err := e.executeStep(step, result)
		if err != nil {
			log.Errorf("Step %d failed: %v", i+1, err)
			result.Status = "failed"
			result.Error = err.Error()
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(startTime)

			// Update memory with failed result
			e.memory.StoreTask(result)
			return result, fmt.Errorf("step %d (%s) failed: %w", i+1, step.FlowName, err)
		}

		result.StepResults = append(result.StepResults, *stepResult)

		// Update memory after each step
		e.memory.StoreTask(result)

		log.Debugf("Step %d completed successfully", i+1)
	}

	// Mark as completed
	result.Status = "completed"
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	// Store final result in memory
	e.memory.StoreTask(result)

	log.Infof("Plan execution completed successfully in %v", result.Duration)
	return result, nil
}

// executeStep executes a single execution step.
func (e *AgentExecutor) executeStep(step ExecutionStep, taskResult *TaskResult) (*StepResult, error) {
	stepStartTime := time.Now()

	stepResult := &StepResult{
		FlowName:    step.FlowName,
		Input:       step.Input,
		Variables:   step.Variables,
		Description: step.Description,
		Status:      "running",
		StartTime:   stepStartTime,
	}

	// Get the flow from registry
	flowInstance, err := e.getFlow(step.FlowName)
	if err != nil {
		stepResult.Status = "failed"
		stepResult.Error = err.Error()
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepStartTime)
		return stepResult, err
	}

	// Prepare flow input
	flowInput := e.prepareFlowInput(step, taskResult)

	// Execute the flow
	log.Debugf("Executing flow '%s' with input: %s", step.FlowName, flowInput)

	flowResult, err := e.executeFlow(flowInstance, flowInput, step.Variables)
	if err != nil {
		stepResult.Status = "failed"
		stepResult.Error = err.Error()
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepStartTime)
		return stepResult, fmt.Errorf("flow execution failed: %w", err)
	}

	// Update step result
	stepResult.Output = flowResult
	stepResult.Status = "completed"
	stepResult.EndTime = time.Now()
	stepResult.Duration = stepResult.EndTime.Sub(stepStartTime)

	return stepResult, nil
}

// getFlow retrieves a flow from the registry.
func (e *AgentExecutor) getFlow(flowName string) (*flow.Flow, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	flows, err := e.registry.GetFlows(e.agent)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows for agent: %w", err)
	}

	for _, f := range flows {
		if f.Name == flowName {
			return f, nil
		}
	}

	return nil, fmt.Errorf("flow '%s' not found in agent's available flows", flowName)
}

// prepareFlowInput prepares the input for flow execution, potentially incorporating previous step results.
func (e *AgentExecutor) prepareFlowInput(step ExecutionStep, taskResult *TaskResult) string {
	input := step.Input

	// If input contains placeholders, try to replace them with previous step results
	if strings.Contains(input, "{{") {
		input = e.replacePlaceholders(input, taskResult)
	}

	return input
}

// replacePlaceholders replaces placeholders in the input with values from previous step results.
func (e *AgentExecutor) replacePlaceholders(input string, taskResult *TaskResult) string {
	result := input

	// Replace {{previous_output}} with the output of the previous step
	if len(taskResult.StepResults) > 0 {
		lastStep := taskResult.StepResults[len(taskResult.StepResults)-1]
		result = strings.ReplaceAll(result, "{{previous_output}}", lastStep.Output)
	}

	// Replace {{step_N_output}} with the output of step N
	for i, stepResult := range taskResult.StepResults {
		placeholder := fmt.Sprintf("{{step_%d_output}}", i+1)
		result = strings.ReplaceAll(result, placeholder, stepResult.Output)
	}

	// Replace {{objective}} with the task objective
	result = strings.ReplaceAll(result, "{{objective}}", taskResult.Objective)

	return result
}

// executeFlow executes a flow with the given input and variables.
func (e *AgentExecutor) executeFlow(flowInstance *flow.Flow, input string, variables map[string]interface{}) (string, error) {
	if flowInstance == nil {
		return "", errors.New("flow instance cannot be nil")
	}

	// Create flow context with variables
	context := flow.FlowContext{
		Text:      input,
		Variables: make(map[string]interface{}),
	}

	for k, v := range variables {
		context.Variables[k] = v
	}

	// Execute the flow
	result, err := flowInstance.Run(context)
	if err != nil {
		return "", fmt.Errorf("flow '%s' execution failed: %w", flowInstance.Name, err)
	}

	// Return the text result
	if result == nil {
		return "", nil
	}

	return result.Text, nil
}

// GetExecutionHistory returns the execution history from memory.
func (e *AgentExecutor) GetExecutionHistory() []*TaskResult {
	return e.memory.GetTaskHistory()
}

// GetTaskResult retrieves a specific task result by objective.
func (e *AgentExecutor) GetTaskResult(objective string) (*TaskResult, error) {
	history := e.memory.GetTaskHistory()
	for _, task := range history {
		if task.Objective == objective {
			return task, nil
		}
	}
	return nil, fmt.Errorf("task with objective '%s' not found in history", objective)
}
