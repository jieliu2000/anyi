package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockExecutor struct {
	Param1 string
	Param2 int
}

func (m *MockExecutor) Run(context flow.ShortTermMemory, Step *flow.Step) (*flow.ShortTermMemory, error) {

	return &context, nil
}

func (m *MockExecutor) Init() error {

	return nil
}

type MockValidator struct {
}

func (m MockValidator) Init() error {

	return nil
}

func (m MockValidator) Validate(stepOutput string, Step *flow.Step) bool {

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

func TestNewExecutorFromConfig(t *testing.T) {

	t.Run("Invalid type", func(t *testing.T) {

		executorConfig := &ExecutorConfig{
			Type: "invalid-executor",
		}

		executor, err := NewExecutorFromConfig(executorConfig)

		assert.Error(t, err)
		assert.Nil(t, executor)
	})

	t.Run("Success path with param", func(t *testing.T) {

		executor1 := &MockExecutor{}
		DefineExecutorType("valid-executor", executor1)

		executorConfig := &ExecutorConfig{
			Type: "valid-executor",
			Name: "executor1",
			Config: map[string]interface{}{
				"param1": "value1",
				"param2": 10,
			},
		}

		executor, err := NewExecutorFromConfig(executorConfig)

		assert.NoError(t, err)
		assert.NotNil(t, executor)

		assert.Equal(t, executor1.Param1, "value1")
		assert.Equal(t, executor1.Param2, 10)

	})

}
