package agent

import (
	"fmt"
	"github.com/jieliu2000/anyi/flow"
)

// AgentJob represents an asynchronous job that an agent executes
type AgentJob struct {
	// Agent is the agent that executes the job
	Agent *Agent

	// Context is the context for the job execution
	Context *AgentContext

	// Status indicates the current status of the job
	Status string // "running", "paused", "completed", "failed"

	// FlowExecutionPlan is the planned flows to execute
	FlowExecutionPlan []*flow.Flow
	
	// stopChan is used to signal the job to stop execution
	stopChan chan struct{}
}

// Execute runs the agent job asynchronously
func (job *AgentJob) Execute() {
	// Initialize stop channel if not already done
	if job.stopChan == nil {
		job.stopChan = make(chan struct{})
	}

	var taskPlan = job.PlanTasks()

	for _, task := range taskPlan {
		// Check if stop was requested
		select {
		case <-job.stopChan:
			job.Status = "paused"
			return
		default:
			// Continue execution
		}
		
		// Execute each task in the plan
		job.RunTask(task)
	}

	job.Status = "completed"
}

func (job *AgentJob) RunTask(task string) {
	// Find the flow by name
	var targetFlow *flow.Flow
	for _, flow := range job.Agent.Flows {
		if flow.Name == task {
			targetFlow = flow
			break
		}
	}

	if targetFlow == nil {
		// Log error: flow not found
		job.Context.ExecuteLog = append(job.Context.ExecuteLog, fmt.Sprintf("Error: Flow '%s' not found", task))
		return
	}

	// Execute the flow
	// For now, we just log the execution
	job.Context.ExecuteLog = append(job.Context.ExecuteLog, fmt.Sprintf("Executing flow: %s - %s", targetFlow.Name, targetFlow.Description))

	// In a real implementation, we would execute the flow steps here
	// flowContext := flow.NewFlowContext()
	// result, err := targetFlow.Run(flowContext)
}

func (job *AgentJob) PlanTasks() []string {
	// Use AgentPlanningFlow to generate the task plan
	// For now, we'll create a simple plan based on the goal

	// Log that we're planning tasks
	job.Context.ExecuteLog = append(job.Context.ExecuteLog, "Planning tasks to achieve goal: " + job.Context.Goal)

	// In a real implementation, we would use the LLM client here
	// to generate a plan based on the agent's backstory, role, goal, and available flows

	// Create a prompt for the LLM
	prompt := job.createPlanningPrompt()

	// For demonstration, return a hardcoded plan
	// In reality, this would come from the LLM response
	return []string{"Task1", "Task2", "Task3"}
}

func (job *AgentJob) createPlanningPrompt() string {
	// Build the prompt with agent information and available flows
	prompt := fmt.Sprintf("Agent Role: %s\n", job.Agent.Role)
	prompt += fmt.Sprintf("Backstory: %s\n", job.Agent.BackStory)
	prompt += fmt.Sprintf("Goal: %s\n\n", job.Context.Goal)
	prompt += "Available Flows:\n"

	for _, flow := range job.Agent.Flows {
		prompt += fmt.Sprintf("- %s: %s\n", flow.Name, flow.Description)
	}

	prompt += "\nPlease generate a sequence of flow names to execute to achieve the goal. Only return the flow names separated by commas."

	return prompt
}

// Resume continues a paused job
func (job *AgentJob) Resume() error {
	// When resuming, we replan based on the existing context
	job.Status = "running"
	
	// Re-initialize stop channel for resumed execution
	job.stopChan = make(chan struct{})
	
	go job.Execute()
	return nil
}

// Stop pauses the job execution
func (job *AgentJob) Stop() error {
	job.Status = "paused"

	// Initialize and close stop channel to signal stop
	if job.stopChan == nil {
		job.stopChan = make(chan struct{})
	}
	close(job.stopChan)

	return nil
}
