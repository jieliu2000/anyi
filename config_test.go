package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockExecutor struct {
}

func (m *MockExecutor) Run(context flow.FlowContext, Step *flow.Step) (*flow.FlowContext, error) {

	return &context, nil
}

type MockValidator struct {
}

func (m *MockValidator) Validate(stepOutput string, Step *flow.Step) bool {

	return true
}

func TestNewFlowFromConfig_Success(t *testing.T) {
	// Setup

	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Executor:  "test-executor",
				Validator: "test-validator",
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, flowInstance)
	assert.Equal(t, flowConfig.Name, flowInstance.Name)
	assert.Equal(t, 1, len(flowInstance.Steps))
}
func TestNewFlowFromConfig_WithNil(t *testing.T) {
	// Execute
	flowInstance, err := NewFlowFromConfig(nil)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
func TestNewFlowFromConfig_WithInvalidClientName(t *testing.T) {
	// Setup
	flowConfig := &FlowConfig{
		ClientName: "invalid-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Executor:  "test-executor",
				Validator: "test-validator",
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
func TestNewFlowFromConfig_WithInvalidStepConfig(t *testing.T) {
	// Setup
	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Executor:  "invalid-executor",
				Validator: "test-validator",
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
func TestNewFlowFromConfig_WithEmptyStepExecutor(t *testing.T) {
	// Setup
	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Validator: "test-validator",
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
