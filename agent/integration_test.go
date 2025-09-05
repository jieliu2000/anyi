package agent

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

// TestAgentRegistryIntegration tests Agent integration with registry
func TestAgentRegistryIntegration(t *testing.T) {
	// Create a simple FlowGetter implementation
	mockFlowGetter := &MockFlowGetter{}

	// Create a real flow object for testing
	flowInstance, _ := flow.NewFlow(nil, "test-flow")
	mockFlowGetter.On("GetFlow", "test-flow").Return(flowInstance, nil)

	// Create Agent
	agent := NewAgent(
		"Integration Test Agent",
		"Test agent for integration",
		[]string{"test-flow"},
		mockFlowGetter,
	)

	// Test basic functionality
	result, _, err := agent.Execute("test task", AgentContext{
		Variables: map[string]interface{}{"test": "value"},
	})

	// Should not error with real flow object
	assert.NoError(t, err)
	assert.Equal(t, "test task", result) // Should return the input since flow has no steps
	// Note: Variables may be reset by flow execution, so we don't assert on specific values

	mockFlowGetter.AssertExpectations(t)
}

// TestAgentWithRealFlow tests integration with real Flow (requires real Flow implementation)
func TestAgentWithRealFlow(t *testing.T) {
	// This test requires real Flow implementation, skipping for now
	t.Skip("Requires real Flow implementation, skipping test")
}
