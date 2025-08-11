package agent

import (
	"testing"
	"time"

	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

func TestAgent_StartJob(t *testing.T) {
	// Test case 1: Create an agent without flows - should return error
	agent := &Agent{
		Role:              "Test Agent",
		Client:            &mockClient{},
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{},
	}

	// Create a context
	context := &AgentContext{
		Goal:            "Test the StartJob function",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without flows - should return error
	job, err := agent.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Test case 2: Create an agent without client - should return error
	agentWithoutClient := &Agent{
		Role:              "Test Agent",
		Client:            nil,
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{},
	}

	// Try to start a job without client - should return error
	job, err = agentWithoutClient.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Test case 3: Create an agent with flows but without client - should return error
	client := &mockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	agentWithFlowNoClient := &Agent{
		Role:              "Test Agent",
		Client:            nil,
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{mockFlow},
	}

	// Try to start a job with flows but without client - should return error
	job, err = agentWithFlowNoClient.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Test case 4: Create an agent with flows and client - should succeed
	agentWithFlowAndClient := &Agent{
		Role:              "Test Agent",
		Client:            &mockClient{},
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{mockFlow},
	}

	// Start a job with flows and client - should succeed
	job, err = agentWithFlowAndClient.StartJob(context)
	assert.NoError(t, err)
	assert.NotNil(t, job)

	// Give some time for the goroutine to execute
	time.Sleep(10 * time.Millisecond)

	// Check that we got a job back
	assert.NotNil(t, job)
	assert.Equal(t, agentWithFlowAndClient, job.Agent)
	assert.Equal(t, context, job.Context)
	
	// Job should be completed since PlanTasks returns an empty slice
	assert.Equal(t, "completed", job.Status)
}

func TestAgentContext_Structure(t *testing.T) {
	// Test that we can create an AgentContext
	context := &AgentContext{
		Goal:            "Test goal",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	assert.Equal(t, "Test goal", context.Goal)
	assert.Empty(t, context.ShortTermMemory)
	assert.Empty(t, context.ExecuteLog)

	// Test that we can add items to the context
	context.ShortTermMemory["key"] = "value"
	context.ExecuteLog = append(context.ExecuteLog, "log entry")

	assert.Equal(t, "value", context.ShortTermMemory["key"])
	assert.Contains(t, context.ExecuteLog, "log entry")
}

func TestAgentConfig_Structure(t *testing.T) {
	// Test that we can create an AgentConfig
	config := &AgentConfig{
		Role:              "Test Agent",
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		ClientName:        "test-client",
		Flows:             []string{"flow1", "flow2"},
	}

	assert.Equal(t, "Test Agent", config.Role)
	assert.Equal(t, "English", config.PreferredLanguage)
	assert.Equal(t, "A test agent", config.BackStory)
	assert.Equal(t, "test-client", config.ClientName)
	assert.Equal(t, []string{"flow1", "flow2"}, config.Flows)
}

func TestAgent_Structure(t *testing.T) {
	// Test that we can create an Agent
	client := &mockClient{}
	flows := []*flow.Flow{}

	agent := &Agent{
		Role:              "Test Agent",
		Client:            client,
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             flows,
	}

	assert.Equal(t, "Test Agent", agent.Role)
	assert.Equal(t, client, agent.Client)
	assert.Equal(t, "English", agent.PreferredLanguage)
	assert.Equal(t, "A test agent", agent.BackStory)
	assert.Equal(t, flows, agent.Flows)
}
