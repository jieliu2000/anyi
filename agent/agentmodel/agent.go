package agentmodel

import (
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
)

// Agent represents an AI agent with specific role and capabilities
type Agent struct {
	// Role defines the role name of the agent, e.g. "Senior Data Researcher"
	Role string `json:"role" yaml:"role" mapstructure:"role"`

	// Client is the LLM default client used by this agent to interact with AI models.
	// If this property is not set, it will use Anyi's default client.
	// Note that flows may have their own LLM client, which will override this one.
	Client llm.Client

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

	// stopChan is used to signal the job to stop execution
	stopChan chan struct{}
}

// AgentConfig defines the configuration structure for agents.
// Agents are autonomous entities that can plan and execute workflows.
type AgentConfig struct {
	Role              string   `mapstructure:"role" json:"role" yaml:"role"`
	PreferredLanguage string   `mapstructure:"preferredLanguage" json:"preferredLanguage" yaml:"preferredLanguage"`
	BackStory         string   `mapstructure:"backStory" json:"backStory" yaml:"backStory"`
	ClientName        string   `mapstructure:"clientName" json:"clientName" yaml:"clientName"`
	Flows             []string `mapstructure:"flows" json:"flows" yaml:"flows"`
}
