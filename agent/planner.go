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

	// If LLM client is available, create AI planning flow lazily and use flow-based planning
	if a.Client != nil {
		// Create the AI planning flow on-demand
		a.createAIPlanningFlow()
		if a.aiPlanningFlow != nil {
			return a.flowBasedPlanExecution(task, ctx)
		}
		// If creation failed, fall back to traditional AI planning
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
	// Build detailed flow information with descriptions
	var flowDetails strings.Builder
	for _, flowName := range a.AvailableFlows {
		flow, err := a.getFlow.GetFlow(flowName)
		if err != nil {
			// If we can't get the flow, just include the name
			flowDetails.WriteString(fmt.Sprintf("- %s (description unavailable)\n", flowName))
		} else {
			// Include flow name and description
			if flow.Description != "" {
				flowDetails.WriteString(fmt.Sprintf("- %s: %s\n", flowName, flow.Description))
			} else {
				flowDetails.WriteString(fmt.Sprintf("- %s (no description)\n", flowName))
			}
		}
	}

	prompt := fmt.Sprintf(`# TASK PLANNING INSTRUCTION

## TASK OVERVIEW
Task: %s

## AGENT CONTEXT
Role: %s
Background: %s
Context Variables: %v

## AVAILABLE FLOWS
%s

## PLANNING REQUIREMENTS
1. Analyze the task requirements and match them with the most appropriate flows
2. Consider the logical sequence and dependencies between flows
3. Leverage your role expertise (%s) to make informed decisions
4. Use context variables to personalize the execution plan
5. Only use flows that are actually available in the list above

## THINKING PROCESS (internal reasoning)
- What is the core objective of this task?
- Which flows are most relevant based on their descriptions?
- What execution order makes logical sense?
- Are there any dependencies between flows?
- How can my role expertise inform this plan?

## RESPONSE FORMAT
You MUST respond with ONLY a valid JSON array containing the flow names in execution order.

## EXAMPLES
Valid response: ["data_processing", "analysis", "report_generation"]
Invalid response: "I think we should use data_processing first, then..."

## CONSTRAINTS
- Do NOT include any explanations, thoughts, or additional text
- Do NOT use flows that are not in the available list
- Do NOT modify the flow names in any way
- If no suitable flows are found, return an empty array: []

## OUTPUT
Provide ONLY the JSON array:`,
		task, a.Role, a.BackStory, ctx.Variables, flowDetails.String(), a.Role)

	return prompt
}

// parseAIPlanResponse parses the AI response to create an execution plan
func (a *Agent) parseAIPlanResponse(aiResponse, task string) (*ExecutionPlan, error) {
	// Robust JSON parsing for AI response
	steps := make([]ExecutionStep, 0)

	// Clean up the response
	response := strings.TrimSpace(aiResponse)

	// Try to parse as JSON array first
	var flowNames []string
	err := json.Unmarshal([]byte(response), &flowNames)
	if err != nil {
		// If JSON parsing fails, try to extract from markdown or other formats
		// Look for JSON array pattern
		if strings.Contains(response, "[") && strings.Contains(response, "]") {
			start := strings.Index(response, "[")
			end := strings.LastIndex(response, "]")
			if start >= 0 && end > start {
				jsonPart := response[start : end+1]
				err = json.Unmarshal([]byte(jsonPart), &flowNames)
			}
		}

		// If still can't parse, fallback to simple string extraction
		if err != nil {
			// Remove brackets and split by commas
			cleaned := strings.Trim(response, "[]")
			if cleaned == "" {
				// If AI response is empty, fallback to simple planning
				return a.simplePlanExecution(task, AgentContext{Variables: make(map[string]interface{})})
			}

			// Split and clean flow names
			rawNames := strings.Split(cleaned, ",")
			for _, name := range rawNames {
				name = strings.TrimSpace(name)
				name = strings.Trim(name, "\"'`") // Remove quotes
				if name != "" {
					flowNames = append(flowNames, name)
				}
			}
		}
	}

	// Validate and create execution steps
	for _, flowName := range flowNames {
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
