package agent_test

import (
	"testing"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

// TestAgentRegistryIntegration demonstrates the usage according to design
func TestAgentRegistryIntegration(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create Agent with dependency injection to resolve circular references
	researchAgent := agent.NewAgent(
		"Research Assistant",
		"Expert at information analysis and report generation",
		[]string{},      // Empty flows for this test
		registry.Global, // Inject registry implementing FlowGetter
	)

	// Register Agent
	err := registry.RegisterAgent("researcher", researchAgent)
	assert.NoError(t, err)

	// Verify agent registration
	retrievedAgent, err := registry.GetAgent("researcher")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedAgent)
	assert.Equal(t, "Research Assistant", retrievedAgent.Role)
	assert.Equal(t, "Expert at information analysis and report generation", retrievedAgent.BackStory)
	assert.Equal(t, []string{}, retrievedAgent.AvailableFlows)

	// Create initial context
	initialCtx := agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "detailed",
			"sources": 10,
			"format":  "markdown",
		},
	}

	// Execute task - uses value type, safe from modification
	result, updatedCtx, err := researchAgent.Execute(
		"Research AI applications in healthcare and write a comprehensive report",
		initialCtx, // Value passed, won't modify originalCtx
	)

	// Should succeed even without registered flows (will use simple planning)
	assert.NoError(t, err)
	assert.Equal(t, "Research AI applications in healthcare and write a comprehensive report", result)
	assert.NotNil(t, updatedCtx.Variables)
	assert.Equal(t, "detailed", updatedCtx.Variables["depth"])
	assert.Equal(t, 10, updatedCtx.Variables["sources"])
	assert.Equal(t, "markdown", updatedCtx.Variables["format"])

	// Verify original context is unchanged (value type safety)
	assert.Equal(t, "detailed", initialCtx.Variables["depth"])
}

// TestRegistryOperations tests basic registry operations
func TestRegistryOperations(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Test agent registration and retrieval
	agent := agent.NewAgent("Test Agent", "Test BackStory", []string{"flow1"}, registry.Global)

	err := registry.RegisterAgent("test", agent)
	assert.NoError(t, err)

	retrievedAgent, err := registry.GetAgent("test")
	assert.NoError(t, err)
	assert.Equal(t, agent, retrievedAgent)

	// Test duplicate registration
	err = registry.RegisterAgent("test", agent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test non-existent agent
	_, err = registry.GetAgent("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test listing agents
	agents := registry.ListAgents()
	assert.Len(t, agents, 1)
	assert.Contains(t, agents, "test")
}
