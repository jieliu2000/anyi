package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	
	"github.com/jieliu2000/anyi/llm/chat"
)

// ExecutionPlan execution plan
type ExecutionPlan struct {
	Steps []ExecutionStep
}

// ExecutionStep execution step
type ExecutionStep struct {
	FlowName  string
	Retryable bool
}

// planExecution plans execution steps
func (a *Agent) planExecution(task string, ctx AgentContext) (*ExecutionPlan, error) {
	// If AI planning flow is available, use flow-based planning
	if a.aiPlanningFlow != nil {
		return a.flowBasedPlanExecution(task, ctx)
	}
	
	// If LLM client is available, use traditional AI planning
	if a.Client != nil {
		return a.aiPlanExecution(task, ctx)
	}
	
	// Fallback to simple planning strategy
	return a.simplePlanExecution(task, ctx)
}

// flowBasedPlanExecution uses the AI planning flow to intelligently plan execution steps
func (a *Agent) flowBasedPlanExecution(task string, ctx AgentContext) (*ExecutionPlan, error) {
	if a.aiPlanningFlow == nil {
		// Fallback to simple planning if AI planning flow is not available
		return a.simplePlanExecution(task, ctx)
	}
	
	// Create flow context with planning variables
	flowContext := a.aiPlanningFlow.NewFlowContext(task, nil)
	flowContext.SetVariable("AgentRole", a.Role)
	flowContext.SetVariable("AgentBackground", a.BackStory)
	flowContext.SetVariable("AvailableFlows", strings.Join(a.AvailableFlows, ", "))
	flowContext.SetVariable("ContextVariables", ctx.Variables)
	
	// Execute the AI planning flow
	for _, step := range a.aiPlanningFlow.Steps {
		var err error
		flowContext, err = step.Executor.Run(*flowContext, &step)
		if err != nil {
			// Fallback to simple planning if flow execution fails
			return a.simplePlanExecution(task, ctx)
		}
	}
	
	// Parse the execution plan from the flow context
	var steps []ExecutionStep
	err := json.Unmarshal([]byte(flowContext.Text), &steps)
	if err != nil {
		// Fallback to simple planning if parsing fails
		return a.simplePlanExecution(task, ctx)
	}
	
	return &ExecutionPlan{Steps: steps}, nil
}

// aiPlanExecution uses AI to intelligently plan execution steps
func (a *Agent) aiPlanExecution(task string, ctx AgentContext) (*ExecutionPlan, error) {
	// Create planning prompt
	prompt := a.createPlanningPrompt(task, ctx)
	
	// Create message for LLM
	messages := []chat.Message{
		{
			Role:    "system",
			Content: "You are an intelligent task planner. Based on the task description and available flows, create an optimal execution plan.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	// Call LLM for planning
	response, _, err := a.Client.Chat(messages, nil)
	if err != nil {
		// Fallback to simple planning if AI planning fails
		return a.simplePlanExecution(task, ctx)
	}
	
	// Parse AI response to create execution plan
	return a.parseAIPlanResponse(response.Content, task)
}

// simplePlanExecution provides a simple fallback planning strategy
func (a *Agent) simplePlanExecution(task string, ctx AgentContext) (*ExecutionPlan, error) {
	// Simple planning strategy: use all available Flows in sequence
	steps := make([]ExecutionStep, 0, len(a.AvailableFlows))

	for _, flowName := range a.AvailableFlows {
		steps = append(steps, ExecutionStep{
			FlowName:  flowName,
			Retryable: true, // Default retryable
		})
	}

	return &ExecutionPlan{Steps: steps}, nil
}

// createPlanningPrompt creates a prompt for AI planning
func (a *Agent) createPlanningPrompt(task string, ctx AgentContext) string {
	availableFlows := strings.Join(a.AvailableFlows, ", ")
	
	prompt := fmt.Sprintf(`Task: %s

Agent Role: %s
Agent Background: %s

Available Flows: %s

Context Variables: %v

Please create an optimal execution plan by selecting and ordering the most appropriate flows from the available flows. 
Respond with a JSON array of flow names in the order they should be executed.
Example response: ["flow1", "flow2", "flow3"]

Only respond with the JSON array, nothing else.`,
		task, a.Role, a.BackStory, availableFlows, ctx.Variables)
	
	return prompt
}

// parseAIPlanResponse parses the AI response to create an execution plan
func (a *Agent) parseAIPlanResponse(aiResponse, task string) (*ExecutionPlan, error) {
	// Simple parsing: extract flow names from response
	// This is a basic implementation - in production, you'd want more robust JSON parsing
	steps := make([]ExecutionStep, 0)
	
	// For now, we'll use a simple approach: split by commas and clean up
	// In a real implementation, you'd parse the JSON properly
	response := strings.TrimSpace(aiResponse)
	response = strings.Trim(response, "[]")
	
	if response == "" {
		// If AI response is empty, fallback to simple planning
		return a.simplePlanExecution(task, AgentContext{Variables: make(map[string]interface{})})
	}
	
	flowNames := strings.Split(response, ",")
	for _, flowName := range flowNames {
		flowName = strings.TrimSpace(flowName)
		flowName = strings.Trim(flowName, "\"") // Remove quotes
		
		// Only add flows that are actually available
		for _, availableFlow := range a.AvailableFlows {
			if availableFlow == flowName {
				steps = append(steps, ExecutionStep{
					FlowName:  flowName,
					Retryable: true,
				})
				break
			}
		}
	}
	
	// If no valid flows were found, fallback to simple planning
	if len(steps) == 0 {
		return a.simplePlanExecution(task, AgentContext{Variables: make(map[string]interface{})})
	}
	
	return &ExecutionPlan{Steps: steps}, nil
}

// replanExecution re-plans execution (reserved interface)
func (a *Agent) replanExecution(currentResult string, ctx AgentContext) (*ExecutionPlan, error) {
	// Simple implementation: return original plan
	return a.planExecution(currentResult, ctx)
}
