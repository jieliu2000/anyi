package agent

import (
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

}

func (job *AgentJob) PlanTasks() []string {

	return []string{}
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
