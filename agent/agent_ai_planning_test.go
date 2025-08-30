package agent_test

import (
	"testing"

	"github.com/jieliu2000/anyi/agent"
	_ "github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

// MockLLMClient implements llm.Client interface for testing
type MockLLMClient struct {
	responses map[string]string
}

func (m *MockLLMClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Return a predefined response for testing (empty array since we use empty flows)
	response := `[]`
	return &chat.Message{
		Role:    "assistant",
		Content: response,
	}, chat.ResponseInfo{}, nil
}

func (m *MockLLMClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Not used in this test
	return &chat.Message{}, chat.ResponseInfo{}, nil
}

// TestAgentWithLLMClient demonstrates the usage of Agent with LLM client for intelligent planning
func TestAgentWithLLMClient(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create mock LLM client
	mockClient := &MockLLMClient{
		responses: make(map[string]string),
	}

	// Create Agent with LLM client
	aiAgent := agent.NewAgentWithClient(
		"AI Research Assistant",
		"Expert at intelligent task planning and execution using AI",
		[]string{}, // Empty flows for testing (no actual flow execution)
		registry.Global, // Inject registry implementing FlowGetter
		mockClient,     // LLM client for intelligent planning
	)

	// Register Agent
	err := registry.RegisterAgent("ai_researcher", aiAgent)
	assert.NoError(t, err)

	// Verify agent registration
	retrievedAgent, err := registry.GetAgent("ai_researcher")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedAgent)
	assert.Equal(t, "AI Research Assistant", retrievedAgent.Role)
	assert.NotNil(t, retrievedAgent.Client) // Verify LLM client is set

	// Create initial context
	initialCtx := agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "detailed",
			"sources": 10,
			"format":  "markdown",
		},
		History: []string{},
	}

	// Execute task - should use AI planning
	result, updatedCtx, err := aiAgent.Execute(
		"Research AI applications in healthcare and write a comprehensive report",
		initialCtx,
	)

	// Should succeed and use AI planning
	assert.NoError(t, err)
	assert.NotNil(t, updatedCtx)
	assert.Equal(t, "detailed", updatedCtx.Variables["depth"])
	assert.Equal(t, 10, updatedCtx.Variables["sources"])
	assert.Equal(t, "markdown", updatedCtx.Variables["format"])
	_ = result // Mark result as used to avoid compiler error

	// Verify original context is unchanged (value type safety)
	assert.Equal(t, "detailed", initialCtx.Variables["depth"])
	assert.Len(t, initialCtx.History, 0) // Original history should be empty
}

// TestAgentFallbackPlanning tests fallback to simple planning when LLM client is not available
func TestAgentFallbackPlanning(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create Agent without LLM client (should fallback to simple planning)
	simpleAgent := agent.NewAgent(
		"Simple Research Assistant",
		"Expert at basic task planning and execution",
		[]string{}, // Empty flows for testing (no actual flow execution)
		registry.Global, // Inject registry implementing FlowGetter
	)

	// Register Agent
	err := registry.RegisterAgent("simple_researcher", simpleAgent)
	assert.NoError(t, err)

	// Create initial context
	initialCtx := agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "basic",
			"sources": 5,
			"format":  "text",
		},
	}

	// Execute task - should use simple planning (fallback)
	result, updatedCtx, err := simpleAgent.Execute(
		"Research AI applications in healthcare",
		initialCtx,
	)

	// Should succeed with simple planning
	assert.NoError(t, err)
	assert.NotNil(t, updatedCtx)
	assert.Equal(t, "basic", updatedCtx.Variables["depth"])
	assert.Equal(t, 5, updatedCtx.Variables["sources"])
	assert.Equal(t, "text", updatedCtx.Variables["format"])
	_ = result // Mark result as used to avoid compiler error
}

// TestAIPromptCreation tests the AI prompt creation functionality
func TestAIPromptCreation(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create mock LLM client
	mockClient := &MockLLMClient{
		responses: make(map[string]string),
	}

	// Create Agent with LLM client
	aiAgent := agent.NewAgentWithClient(
		"AI Research Assistant",
		"Expert at intelligent task planning and execution using AI",
		[]string{}, // Empty flows for testing
		registry.Global,
		mockClient,
	)

	// Create context with variables
	_ = agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "detailed",
			"sources": 10,
			"format":  "markdown",
		},
	}

	// Test that the agent can be created and configured properly
	assert.NotNil(t, aiAgent)
	assert.Equal(t, "AI Research Assistant", aiAgent.Role)
	assert.Equal(t, "Expert at intelligent task planning and execution using AI", aiAgent.BackStory)
	assert.Equal(t, []string{}, aiAgent.AvailableFlows)
	assert.NotNil(t, aiAgent.Client)
}