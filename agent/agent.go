package agent

import (
	"github.com/jieliu2000/anyi/flow"
)

// Agent represents an AI agent with specific role and capabilities
type Agent struct {
	// Role defines the role name of the agent, e.g. "Senior Data Researcher"
	Role string `json:"role" yaml:"role" mapstructure:"role"`
	
	// PreferredLanguage is the default language for the agent, e.g. "English"
	PreferredLanguage string `json:"preferredLanguage" yaml:"preferredLanguage" mapstructure:"preferredLanguage"`
	
	// BackStory provides detailed description of the agent's identity
	BackStory string `json:"backStory" yaml:"backStory" mapstructure:"backStory"`
	
	// Flows is a list of flows that this agent can use
	Flows []*flow.Flow `json:"flows" yaml:"flows" mapstructure:"flows"`
}

// AgentContext contains the context for agent execution
type AgentContext struct {
	// Goal is the target that the agent needs to achieve
	Goal string `json:"goal" yaml:"goal" mapstructure:"goal"`
	
	// ShortTermMemory stores information generated during agent execution
	ShortTermMemory map[string]interface{} `json:"shortTermMemory" yaml:"shortTermMemory" mapstructure:"shortTermMemory"`
	
	// ExecuteLog contains all human-AI conversation logs during agent execution
	ExecuteLog []string `json:"executeLog" yaml:"executeLog" mapstructure:"executeLog"`
}

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
}

// StartJob starts a new job for the agent with the given context
// It returns an AgentJob reference immediately while the job runs asynchronously
func (a *Agent) StartJob(context *AgentContext) *AgentJob {
	job := &AgentJob{
		Agent:   a,
		Context: context,
		Status:  "running",
	}
	
	// Run the job asynchronously
	go job.execute()
	
	return job
}

// execute runs the agent job asynchronously
func (job *AgentJob) execute() {
	// First, call the built-in AgentPlanningFlow to plan the execution
	// based on the goal in AgentContext
	planningContext := &flow.FlowContext{
		Text:      job.Context.Goal,
		Memory:    job.Context.ShortTermMemory,
		Variables: make(map[string]interface{}),
	}
	
	// Add agent information to the context
	planningContext.Variables["agent_role"] = job.Agent.Role
	planningContext.Variables["agent_backstory"] = job.Agent.BackStory
	planningContext.Variables["agent_language"] = job.Agent.PreferredLanguage
	
	// Execute the planning flow
	// TODO: Implement actual flow execution
	// planningResult, err := AgentPlanningFlow.Execute(planningContext)
	// if err != nil {
	// 	job.Status = "failed"
	// 	return
	// }
	
	// Based on the planning result, build the flow execution plan
	// job.FlowExecutionPlan = buildFlowExecutionPlan(planningResult)
	
	// Execute each flow in the plan
	for _, flowItem := range job.FlowExecutionPlan {
		// Create a context for this flow execution with agent information
		flowContext := &flow.FlowContext{
			Text:      job.Context.Goal, // Or a more specific goal for this flow
			Memory:    job.Context.ShortTermMemory,
			Variables: make(map[string]interface{}),
			Flow:      flowItem,
		}
		
		// Add agent context to the flow
		flowContext.Variables["agent_role"] = job.Agent.Role
		flowContext.Variables["agent_backstory"] = job.Agent.BackStory
		flowContext.Variables["agent_language"] = job.Agent.PreferredLanguage
		
		// Add execution log summary
		// flowContext.Variables["execution_log_summary"] = summarizeExecutionLog(job.Context.ExecuteLog)
		
		// Execute the flow
		// _, err := flowItem.Execute(flowContext)
		// if err != nil {
		// 	job.Status = "failed"
		// 	return
		// }
		
		// After each flow execution, call the reflection flow to check
		// if the goal has been achieved
		// reflectionContext := &flow.FlowContext{
		// 	Text:      job.Context.Goal,
		// 	Memory:    job.Context.ShortTermMemory,
		// 	Variables: flowContext.Variables,
		// }
		
		// _, err = AgentReflectionFlow.Execute(reflectionContext)
		// if err != nil {
		// 	job.Status = "failed"
		// 	return
		// }
		
		// TODO: Check reflection result to determine if we need to replan
		// if !goalAchieved(reflectionContext) {
		// 	// Replan and continue execution
		// 	job.execute()
		// 	return
		// }
	}
	
	job.Status = "completed"
}

// Resume continues a paused job
func (job *AgentJob) Resume() error {
	// When resuming, we replan based on the existing context
	job.Status = "running"
	go job.execute()
	return nil
}

// Stop pauses the job execution
func (job *AgentJob) Stop() error {
	// TODO: Implement proper job stopping logic
	job.Status = "paused"
	return nil
}

// buildFlowExecutionPlan creates a flow execution plan based on planning result
func buildFlowExecutionPlan(planningResult string) []*flow.Flow {
	// TODO: Parse the planning result and build actual flow execution plan
	// This would involve looking up flows by name from the registry
	var plan []*flow.Flow
	
	// For now, we just return an empty plan
	return plan
}

// summarizeExecutionLog creates a summary of the execution log
func summarizeExecutionLog(logs []string) string {
	// TODO: Implement proper log summarization
	if len(logs) == 0 {
		return "No previous execution logs"
	}
	
	// For now, just return a simple summary
	return "Previous execution logs exist"
}

// goalAchieved checks if the agent's goal has been achieved
func goalAchieved(context *flow.FlowContext) bool {
	// TODO: Implement proper goal achievement checking
	return true
}