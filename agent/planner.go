package agent

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

// replanExecution re-plans execution (reserved interface)
func (a *Agent) replanExecution(currentResult string, ctx AgentContext) (*ExecutionPlan, error) {
	// Simple implementation: return original plan
	return a.planExecution(currentResult, ctx)
}
