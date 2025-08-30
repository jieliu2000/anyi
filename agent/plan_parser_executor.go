package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jieliu2000/anyi/flow"
)

// PlanParserExecutor is an executor that parses LLM responses to create execution plans
// It implements the StepExecutor interface
// This executor is specifically designed for agent internal use and doesn't need to be registered globally
type PlanParserExecutor struct {
	AvailableFlows []string
}

// Init initializes the PlanParserExecutor
func (e *PlanParserExecutor) Init() error {
	// No initialization needed for now
	return nil
}

// Run parses the LLM response to create an execution plan
// The LLM response should be a JSON array of flow names
func (e *PlanParserExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	if flowContext.Text == "" {
		return nil, fmt.Errorf("empty LLM response")
	}

	// Try to parse as JSON array first
	var flowNames []string
	err := json.Unmarshal([]byte(flowContext.Text), &flowNames)
	if err != nil {
		// Fallback to simple parsing if JSON parsing fails
		flowNames = e.parseSimpleResponse(flowContext.Text)
	}

	// Create execution plan
	steps := make([]ExecutionStep, 0)
	for _, flowName := range flowNames {
		flowName = strings.TrimSpace(flowName)
		
		// Only add flows that are actually available
		for _, availableFlow := range e.AvailableFlows {
			if availableFlow == flowName {
				steps = append(steps, ExecutionStep{
					FlowName:  flowName,
					Retryable: true,
				})
				break
			}
		}
	}

	// If no valid flows were found, create a plan with all available flows
	if len(steps) == 0 {
		for _, flowName := range e.AvailableFlows {
			steps = append(steps, ExecutionStep{
				FlowName:  flowName,
				Retryable: true,
			})
		}
	}

	// Convert execution plan to JSON for storage in flow context
	planJSON, err := json.Marshal(steps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execution plan: %w", err)
	}

	flowContext.Text = string(planJSON)
	return &flowContext, nil
}

// parseSimpleResponse provides a fallback parsing method for non-JSON responses
func (e *PlanParserExecutor) parseSimpleResponse(response string) []string {
	// Simple parsing: extract flow names from response
	response = strings.TrimSpace(response)
	response = strings.Trim(response, "[]")
	
	if response == "" {
		return []string{}
	}
	
	flowNames := strings.Split(response, ",")
	for i, flowName := range flowNames {
		flowNames[i] = strings.TrimSpace(flowName)
		flowNames[i] = strings.Trim(flowName, "\"") // Remove quotes
	}
	
	return flowNames
}

// NewPlanParserExecutor creates a new PlanParserExecutor with the available flows
func NewPlanParserExecutor(availableFlows []string) *PlanParserExecutor {
	return &PlanParserExecutor{
		AvailableFlows: availableFlows,
	}
}