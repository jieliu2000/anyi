package agent

import (
	"errors"

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

// AgentConfig defines the configuration structure for agents.
// Agents are autonomous entities that can plan and execute workflows.
type AgentConfig struct {
	Role              string   `mapstructure:"role" json:"role" yaml:"role"`
	PreferredLanguage string   `mapstructure:"preferredLanguage" json:"preferredLanguage" yaml:"preferredLanguage"`
	BackStory         string   `mapstructure:"backStory" json:"backStory" yaml:"backStory"`
	ClientName        string   `mapstructure:"clientName" json:"clientName" yaml:"clientName"`
	Flows             []string `mapstructure:"flows" json:"flows" yaml:"flows"`
}

// StartJob starts a new job for the agent with the given context
// It returns an AgentJob reference immediately while the job runs asynchronously
func (a *Agent) StartJob(context *AgentContext) (*AgentJob, error) {
	// Check if agent has at least one flow
	if len(a.Flows) == 0 {
		return nil, errors.New("agent must have at least one flow to start a job")
	}

	if a.Client == nil {
		return nil, errors.New("agent must have a valid client to start a job")
	}

	job := &AgentJob{
		Agent:    a,
		Context:  context,
		Status:   "running",
		stopChan: make(chan struct{}),
	}

	// Run the job asynchronously
	go job.Execute()

	return job, nil
}
