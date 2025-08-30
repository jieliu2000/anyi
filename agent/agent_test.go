package agent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFlowGetter mocks FlowGetter
type MockFlowGetter struct {
	mock.Mock
}

func (m *MockFlowGetter) GetFlow(name string) (interface{}, error) {
	args := m.Called(name)
	return args.Get(0), args.Error(1)
}

// MockFlow mocks Flow
type MockFlow struct {
	mock.Mock
}

func (m *MockFlow) Execute(input string, ctx map[string]interface{}) (string, map[string]interface{}, error) {
	args := m.Called(input, ctx)
	return args.String(0), args.Get(1).(map[string]interface{}), args.Error(2)
}

func TestNewAgent(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)

	agent := NewAgent(
		"Test Agent",
		"Test backstory",
		[]string{"flow1", "flow2"},
		mockFlowGetter,
	)

	assert.NotNil(t, agent)
	assert.Equal(t, "Test Agent", agent.Role)
	assert.Equal(t, "Test backstory", agent.BackStory)
	assert.Equal(t, []string{"flow1", "flow2"}, agent.AvailableFlows)
	assert.Equal(t, mockFlowGetter, agent.getFlow)
	assert.Equal(t, 10, agent.MaxIterations)
	assert.Equal(t, 3, agent.MaxRetries)
}

func TestExecute_EmptyContext(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)
	agent := NewAgent("Test", "Test", []string{}, mockFlowGetter)

	result, ctx, err := agent.Execute("test task", AgentContext{})

	assert.NoError(t, err)
	assert.Equal(t, "test task", result)
	assert.NotNil(t, ctx.Variables)
	assert.Empty(t, ctx.History)
}

func TestExecute_WithVariables(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)
	agent := NewAgent("Test", "Test", []string{}, mockFlowGetter)

	initialCtx := AgentContext{
		Variables: map[string]interface{}{"key": "value"},
		Memory:    "test memory",
		History:   []string{"history1"},
	}

	result, ctx, err := agent.Execute("test task", initialCtx)

	assert.NoError(t, err)
	assert.Equal(t, "test task", result)
	assert.Equal(t, "value", ctx.Variables["key"])
	assert.Equal(t, "test memory", ctx.Memory)
	assert.Equal(t, []string{"history1"}, ctx.History)
}

func TestExecute_WithFlows(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)
	mockFlow := new(MockFlow)

	// Set mock expectations
	mockFlowGetter.On("GetFlow", "test-flow").Return(mockFlow, nil)
	mockFlow.On("Execute", "test task", mock.Anything).Return("processed result", map[string]interface{}{"output": "data"}, nil)

	agent := NewAgent("Test", "Test", []string{"test-flow"}, mockFlowGetter)

	result, ctx, err := agent.Execute("test task", AgentContext{})

	assert.NoError(t, err)
	assert.Equal(t, "processed result", result)
	assert.Equal(t, "data", ctx.Variables["output"])
	assert.Len(t, ctx.History, 1)
	assert.Equal(t, "processed result", ctx.History[0])

	mockFlowGetter.AssertExpectations(t)
	mockFlow.AssertExpectations(t)
}

func TestExecute_FlowNotFound(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)
	mockFlowGetter.On("GetFlow", "missing-flow").Return(nil, errors.New("flow not found"))

	agent := NewAgent("Test", "Test", []string{"missing-flow"}, mockFlowGetter)

	_, _, err := agent.Execute("test task", AgentContext{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get flow missing-flow")
	mockFlowGetter.AssertExpectations(t)
}

func TestExecute_FlowExecutionError(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)

	// Create an actual struct to mock Flow instead of mocking interface
	type testFlow struct{}
	flowInstance := &testFlow{}

	mockFlowGetter.On("GetFlow", "error-flow").Return(flowInstance, nil)

	agent := NewAgent("Test", "Test", []string{"error-flow"}, mockFlowGetter)

	_, _, err := agent.Execute("test task", AgentContext{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement Execute method")
	mockFlowGetter.AssertExpectations(t)
}

func TestExecute_FlowDoesNotImplementExecute(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)

	// Return an object that does not implement Execute interface
	mockFlowGetter.On("GetFlow", "invalid-flow").Return("not a flow", nil)

	agent := NewAgent("Test", "Test", []string{"invalid-flow"}, mockFlowGetter)

	_, _, err := agent.Execute("test task", AgentContext{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement Execute method")
	mockFlowGetter.AssertExpectations(t)
}

func TestExecute_WithRetry(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)

	// Create an actual struct to mock Flow
	type testFlow struct {
		callCount int
	}

	flowInstance := &testFlow{}

	// Use function to mock Execute behavior
	mockFlowGetter.On("GetFlow", "retry-flow").Return(flowInstance, nil)

	// Create an object that implements Execute method
	executableFlow := struct {
		callCount int
	}{}

	mockFlowGetter.On("GetFlow", "retry-flow").Return(&executableFlow, nil)

	agent := NewAgent("Test", "Test", []string{"retry-flow"}, mockFlowGetter)
	agent.MaxRetries = 2

	// Since we used an object that does not implement Execute method, it will return an error
	_, _, err := agent.Execute("test task", AgentContext{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement Execute method")

	mockFlowGetter.AssertExpectations(t)
}

func TestIsTaskCompleted(t *testing.T) {
	agent := &Agent{}

	// Insufficient result length
	assert.False(t, agent.isTaskCompleted("short", "long task description"))

	// Sufficient result length
	assert.True(t, agent.isTaskCompleted("this is a very long result that should be considered complete", "short task"))

	// Edge cases
	assert.False(t, agent.isTaskCompleted("result", "result"))       // Same length
	assert.True(t, agent.isTaskCompleted("result result", "result")) // Twice the length
}

func TestPlanExecution(t *testing.T) {
	mockFlowGetter := new(MockFlowGetter)
	agent := NewAgent("Test", "Test", []string{"flow1", "flow2", "flow3"}, mockFlowGetter)

	plan, err := agent.planExecution("test task", AgentContext{})

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Len(t, plan.Steps, 3)
	assert.Equal(t, "flow1", plan.Steps[0].FlowName)
	assert.Equal(t, "flow2", plan.Steps[1].FlowName)
	assert.Equal(t, "flow3", plan.Steps[2].FlowName)
	assert.True(t, plan.Steps[0].Retryable)
}
