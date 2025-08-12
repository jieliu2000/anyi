package agentmodel

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
)

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
	client := &test.MockClient{}
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
