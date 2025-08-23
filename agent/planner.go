package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"

	log "github.com/sirupsen/logrus"
)

// TaskPlanner is responsible for planning task execution by generating execution plans.
type TaskPlanner struct {
	registry AgentRegistry
	client   llm.Client
	agent    *Agent
}

// NewTaskPlanner creates a new TaskPlanner instance.
func NewTaskPlanner(registry AgentRegistry, agent *Agent) (*TaskPlanner, error) {
	if registry == nil {
		return nil, errors.New("registry cannot be nil")
	}
	if agent == nil {
		return nil, errors.New("agent cannot be nil")
	}

	client, err := agent.GetClient(registry)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for agent: %w", err)
	}

	return &TaskPlanner{
		registry: registry,
		client:   client,
		agent:    agent,
	}, nil
}

// PlanExecution generates an execution plan for the given objective.
func (p *TaskPlanner) PlanExecution(objective string) (*ExecutionPlan, error) {
	if objective == "" {
		return nil, errors.New("objective cannot be empty")
	}

	log.Debugf("Planning execution for objective: %s", objective)

	// Get available flows for the agent
	flows, err := p.registry.GetFlows(p.agent)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows for agent: %w", err)
	}

	if len(flows) == 0 {
		return nil, errors.New("no flows available for agent")
	}

	// Build planning prompt
	prompt := p.buildPlanningPrompt(objective, flows)

	// Generate execution plan using LLM
	planJSON, err := p.generateExecutionPlan(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate execution plan: %w", err)
	}

	// Parse and validate the plan
	plan, err := p.parseExecutionPlan(planJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse execution plan: %w", err)
	}

	// Validate that all flows in the plan are available
	err = p.validatePlan(plan, flows)
	if err != nil {
		return nil, fmt.Errorf("invalid execution plan: %w", err)
	}

	log.Debugf("Generated execution plan with %d steps", len(plan.Steps))
	return plan, nil
}

// buildPlanningPrompt creates the planning prompt for the LLM.
func (p *TaskPlanner) buildPlanningPrompt(objective string, flows []*flow.Flow) string {
	flowsInfo := p.formatFlowsInfo(flows)

	return fmt.Sprintf(`You are an intelligent task planner. Please create a detailed execution plan based on the following information:

**Task Objective**: %s

**Agent Description**: %s

**Available Flows**:
%s

Please create an execution plan that breaks down the task into a series of flow invocation steps. Each step must use one of the available flows listed above.

Output format should be JSON:
{
  "objective": "%s",
  "description": "Overall description of the execution plan",
  "steps": [
    {
      "flowName": "name_of_flow",
      "input": "input_text_for_this_step",
      "variables": {"key": "value"},
      "description": "description_of_what_this_step_does",
      "order": 1
    }
  ]
}

Important rules:
1. Only use flows from the available flows list
2. Each step should have a clear purpose
3. Steps should be ordered logically
4. Input for each step should be specific and actionable
5. Variables should be used when the flow requires specific parameters`,
		objective, p.agent.Description, flowsInfo, objective)
}

// formatFlowsInfo formats the available flows information for the prompt.
func (p *TaskPlanner) formatFlowsInfo(flows []*flow.Flow) string {
	var info strings.Builder
	for i, flow := range flows {
		info.WriteString(fmt.Sprintf("%d. %s", i+1, flow.Name))
		if len(flow.Steps) > 0 {
			info.WriteString(fmt.Sprintf(" (%d steps)", len(flow.Steps)))
		}
		info.WriteString("\n")
	}
	return info.String()
}

// generateExecutionPlan calls the LLM to generate the execution plan.
func (p *TaskPlanner) generateExecutionPlan(prompt string) (string, error) {
	messages := []chat.Message{
		chat.NewSystemMessage("You are an expert task planning assistant. Always respond with valid JSON format."),
		chat.NewUserMessage(prompt),
	}

	response, _, err := p.client.Chat(messages, nil)
	if err != nil {
		return "", fmt.Errorf("LLM chat failed: %w", err)
	}

	return response.Content, nil
}

// parseExecutionPlan parses the JSON response from the LLM into an ExecutionPlan.
func (p *TaskPlanner) parseExecutionPlan(planJSON string) (*ExecutionPlan, error) {
	// Clean up the JSON response (remove markdown code blocks if present)
	planJSON = strings.TrimSpace(planJSON)
	if strings.HasPrefix(planJSON, "```json") {
		planJSON = strings.TrimPrefix(planJSON, "```json")
		planJSON = strings.TrimSuffix(planJSON, "```")
		planJSON = strings.TrimSpace(planJSON)
	} else if strings.HasPrefix(planJSON, "```") {
		planJSON = strings.TrimPrefix(planJSON, "```")
		planJSON = strings.TrimSuffix(planJSON, "```")
		planJSON = strings.TrimSpace(planJSON)
	}

	var plan ExecutionPlan
	err := json.Unmarshal([]byte(planJSON), &plan)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution plan JSON: %w", err)
	}

	return &plan, nil
}

// validatePlan validates that the execution plan uses only available flows.
func (p *TaskPlanner) validatePlan(plan *ExecutionPlan, availableFlows []*flow.Flow) error {
	if plan == nil {
		return errors.New("plan cannot be nil")
	}

	if len(plan.Steps) == 0 {
		return errors.New("plan must have at least one step")
	}

	// Create a map of available flow names for quick lookup
	flowNames := make(map[string]bool)
	for _, flow := range availableFlows {
		flowNames[flow.Name] = true
	}

	// Validate each step
	for i, step := range plan.Steps {
		if step.FlowName == "" {
			return fmt.Errorf("step %d has empty flow name", i+1)
		}

		if !flowNames[step.FlowName] {
			return fmt.Errorf("step %d uses unavailable flow: %s", i+1, step.FlowName)
		}

		if step.Input == "" {
			return fmt.Errorf("step %d has empty input", i+1)
		}
	}

	return nil
}
