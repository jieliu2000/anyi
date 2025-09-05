package agent_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

// MockLLMClientForPlanning implements llm.Client interface for testing flow-based planning
type MockLLMClientForPlanning struct {
	response string
}

func (m *MockLLMClientForPlanning) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Return a predefined response for testing
	return &chat.Message{
		Role:    "assistant",
		Content: m.response,
	}, chat.ResponseInfo{}, nil
}

func (m *MockLLMClientForPlanning) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Not used in this test
	return &chat.Message{}, chat.ResponseInfo{}, nil
}



// registerMockFlows registers mock flows in the registry for testing
// MockFlowGetter implements FlowGetter interface for testing
type MockFlowGetter struct {
	flows map[string]*flow.Flow
}

func NewMockFlowGetter() *MockFlowGetter {
	return &MockFlowGetter{
		flows: make(map[string]*flow.Flow),
	}
}

func (m *MockFlowGetter) GetFlow(name string) (*flow.Flow, error) {
	flow, exists := m.flows[name]
	if !exists {
		return nil, fmt.Errorf("flow %s not found", name)
	}
	return flow, nil
}

func (m *MockFlowGetter) RegisterFlow(name string, flow *flow.Flow) {
	m.flows[name] = flow
}

func registerMockFlows(flowGetter *MockFlowGetter) {
	// Create real flow objects for testing with proper steps
	researchFlow, _ := flow.NewFlow(nil, "research")
	analyzeFlow, _ := flow.NewFlow(nil, "analyze")
	summarizeFlow, _ := flow.NewFlow(nil, "summarize")

	// Register flows in the mock flow getter
	flowGetter.RegisterFlow("research", researchFlow)
	flowGetter.RegisterFlow("analyze", analyzeFlow)
	flowGetter.RegisterFlow("summarize", summarizeFlow)
}

// TestPlanParserExecutor tests the PlanParserExecutor functionality
func TestPlanParserExecutor(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create PlanParserExecutor with available flows
	availableFlows := []string{"research", "analyze", "summarize"}
	executor := agent.NewPlanParserExecutor(availableFlows)

	// Test initialization
	err := executor.Init()
	assert.NoError(t, err)

	// Test JSON response parsing
	jsonResponse := `["research", "analyze", "summarize"]`
	flowContext := flow.NewFlowContext(jsonResponse, nil)
	step := flow.NewStep(executor, nil, nil)

	resultContext, err := executor.Run(*flowContext, step)
	assert.NoError(t, err)
	assert.NotNil(t, resultContext)

	// Parse the result to verify execution plan
	var steps []agent.ExecutionStep
	err = json.Unmarshal([]byte(resultContext.Text), &steps)
	assert.NoError(t, err)
	assert.Len(t, steps, 3)
	assert.Equal(t, "research", steps[0].FlowName)
	assert.Equal(t, "analyze", steps[1].FlowName)
	assert.Equal(t, "summarize", steps[2].FlowName)
	assert.True(t, steps[0].Retryable)
}

// TestPlanParserExecutorWithSimpleResponse tests the PlanParserExecutor with simple (non-JSON) response
func TestPlanParserExecutorWithSimpleResponse(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create PlanParserExecutor with available flows
	availableFlows := []string{"research", "analyze", "summarize"}
	executor := agent.NewPlanParserExecutor(availableFlows)

	// Test initialization
	err := executor.Init()
	assert.NoError(t, err)

	// Test simple response parsing
	simpleResponse := "research, analyze, summarize"
	flowContext := flow.NewFlowContext(simpleResponse, nil)
	step := flow.NewStep(executor, nil, nil)

	resultContext, err := executor.Run(*flowContext, step)
	assert.NoError(t, err)
	assert.NotNil(t, resultContext)

	// Parse the result to verify execution plan
	var steps []agent.ExecutionStep
	err = json.Unmarshal([]byte(resultContext.Text), &steps)
	assert.NoError(t, err)
	assert.Len(t, steps, 3)
	assert.Equal(t, "research", steps[0].FlowName)
	assert.Equal(t, "analyze", steps[1].FlowName)
	assert.Equal(t, "summarize", steps[2].FlowName)
}

// TestPlanParserExecutorWithInvalidFlows tests the PlanParserExecutor with invalid flow names
func TestPlanParserExecutorWithInvalidFlows(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create PlanParserExecutor with available flows
	availableFlows := []string{"research", "analyze"}
	executor := agent.NewPlanParserExecutor(availableFlows)

	// Test initialization
	err := executor.Init()
	assert.NoError(t, err)

	// Test response with invalid flow names
	response := `["research", "invalid_flow", "analyze"]`
	flowContext := flow.NewFlowContext(response, nil)
	step := flow.NewStep(executor, nil, nil)

	resultContext, err := executor.Run(*flowContext, step)
	assert.NoError(t, err)
	assert.NotNil(t, resultContext)

	// Parse the result to verify execution plan (should only contain valid flows)
	var steps []agent.ExecutionStep
	err = json.Unmarshal([]byte(resultContext.Text), &steps)
	assert.NoError(t, err)
	assert.Len(t, steps, 2)
	assert.Equal(t, "research", steps[0].FlowName)
	assert.Equal(t, "analyze", steps[1].FlowName)
}

// TestPlanParserExecutorWithEmptyResponse tests the PlanParserExecutor with empty response
func TestPlanParserExecutorWithEmptyResponse(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()

	// Create PlanParserExecutor with available flows
	availableFlows := []string{"research", "analyze"}
	executor := agent.NewPlanParserExecutor(availableFlows)

	// Test initialization
	err := executor.Init()
	assert.NoError(t, err)

	// Test with empty response
	flowContext := flow.NewFlowContext("", nil)
	step := flow.NewStep(executor, nil, nil)

	resultContext, err := executor.Run(*flowContext, step)
	assert.Error(t, err)
	assert.Nil(t, resultContext)
}

// TestFlowBasedPlanExecution tests the flow-based planning functionality
func TestFlowBasedPlanExecution(t *testing.T) {
	// Create mock flow getter
	flowGetter := NewMockFlowGetter()
	// Register mock flows
	registerMockFlows(flowGetter)

	// Create mock LLM client with predefined response
	mockClient := &MockLLMClientForPlanning{
		response: `["research", "analyze", "summarize"]`,
	}

	// Create Agent with LLM client
	availableFlows := []string{"research", "analyze", "summarize"}
	aiAgent := agent.NewAgentWithClient(
		"AI Research Assistant",
		"Expert at intelligent task planning and execution using AI",
		availableFlows,
		flowGetter, // Inject mock flow getter
		mockClient, // LLM client for intelligent planning
	)

	// Verify AI planning flow is created
	assert.NotNil(t, aiAgent)

	// Create context
	ctx := agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "detailed",
			"sources": 10,
			"format":  "markdown",
		},
	}

	// Test flow-based planning
	result, _, err := aiAgent.Execute("Research AI applications in healthcare", ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Since we're using real flow objects (no steps), result should be the input
	assert.Equal(t, "Research AI applications in healthcare", result)
}

// TestFlowBasedPlanExecutionFallback tests the fallback behavior when flow-based planning fails
func TestFlowBasedPlanExecutionFallback(t *testing.T) {
	// Create mock flow getter
	flowGetter := NewMockFlowGetter()
	// Register mock flows
	registerMockFlows(flowGetter)

	// Create mock LLM client with empty response (will cause fallback)
	mockClient := &MockLLMClientForPlanning{
		response: ``,
	}

	// Create Agent with LLM client
	availableFlows := []string{"research", "analyze", "summarize"}
	aiAgent := agent.NewAgentWithClient(
		"AI Research Assistant",
		"Expert at intelligent task planning and execution using AI",
		availableFlows,
		flowGetter, // Inject mock flow getter
		mockClient, // LLM client for intelligent planning
	)

	// Verify AI planning flow is created
	assert.NotNil(t, aiAgent)

	// Create context
	ctx := agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "detailed",
			"sources": 10,
			"format":  "markdown",
		},
	}

	// Test flow-based planning with fallback
	result, _, err := aiAgent.Execute("Research AI applications in healthcare", ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Since we're using real flow objects (no steps), result should be the input
	assert.Equal(t, "Research AI applications in healthcare", result)
}

// TestAgentWithoutLLMClient tests that Agent without LLM client uses simple planning
func TestAgentWithoutLLMClient(t *testing.T) {
	// Create mock flow getter
	flowGetter := NewMockFlowGetter()
	// Register mock flows
	registerMockFlows(flowGetter)

	// Create Agent without LLM client
	availableFlows := []string{"research", "analyze", "summarize"}
	simpleAgent := agent.NewAgent(
		"Simple Research Assistant",
		"Expert at basic task planning and execution",
		availableFlows,
		flowGetter, // Inject mock flow getter
	)

	// Verify AI planning flow is not created
	assert.NotNil(t, simpleAgent)

	// Create context
	ctx := agent.AgentContext{
		Variables: map[string]interface{}{
			"depth":   "basic",
			"sources": 5,
			"format":  "text",
		},
	}

	// Test simple planning
	result, _, err := simpleAgent.Execute("Research AI applications in healthcare", ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Since we're using real flow objects (no steps), result should be the input
	assert.Equal(t, "Research AI applications in healthcare", result)
}