package registry

import (
	"testing"

	"github.com/jieliu2000/anyi/agent"
	"github.com/stretchr/testify/assert"
)

func TestRegistryOperations(t *testing.T) {
	// Clear registry for clean test
	Clear()

	// Test agent registration and retrieval
	testAgent := agent.NewAgent("Test Agent", "Test BackStory", []string{"flow1"}, Global)

	err := RegisterAgent("test", testAgent)
	assert.NoError(t, err)

	retrievedAgent, err := GetAgent("test")
	assert.NoError(t, err)
	assert.Equal(t, testAgent, retrievedAgent)

	// Test duplicate registration
	err = RegisterAgent("test", testAgent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test non-existent agent
	_, err = GetAgent("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test listing agents
	agents := ListAgents()
	assert.Len(t, agents, 1)
	assert.Contains(t, agents, "test")

	// Test clearing registry
	Clear()
	agents = ListAgents()
	assert.Len(t, agents, 0)
}

func TestFlowGetterInterface(t *testing.T) {
	// Clear registry for clean test
	Clear()

	// Test GetFlow method (implements agent.FlowGetter)
	_, err := Global.GetFlow("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flow nonexistent not found")
}
