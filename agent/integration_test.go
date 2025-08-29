package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAgentRegistryIntegration tests Agent integration with registry
func TestAgentRegistryIntegration(t *testing.T) {
	// Create a simple FlowGetter implementation
	mockFlowGetter := &MockFlowGetter{}

	// Create a simple Flow implementation
	type simpleFlow struct{}

	flowInstance := &simpleFlow{}
	mockFlowGetter.On("GetFlow", "test-flow").Return(flowInstance, nil)

	// Create Agent
	agent := NewAgent(
		"Integration Test Agent",
		"Test agent for integration",
		[]string{"test-flow"},
		mockFlowGetter,
	)

	// Test basic functionality
	result, ctx, err := agent.Execute("test task", AgentContext{
		Variables: map[string]interface{}{"test": "value"},
	})

	// Since simpleFlow does not implement Execute method, it should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement Execute method")
	assert.Equal(t, "value", ctx.Variables["test"])
	assert.Equal(t, "", result) // Should return empty result when flow execution fails

	mockFlowGetter.AssertExpectations(t)
}

// TestAgentWithRealFlow tests integration with real Flow (requires real Flow implementation)
func TestAgentWithRealFlow(t *testing.T) {
	// This test requires real Flow implementation, skipping for now
	t.Skip("Requires real Flow implementation, skipping test")
}
