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
