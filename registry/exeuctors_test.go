package registry

import (
	"testing"

	"github.com/jieliu2000/anyi/executors"
	"github.com/jieliu2000/anyi/flow"

	"github.com/stretchr/testify/assert"
)

type MockExecutor struct {
	Param1 string
	Param2 int
}

func (m *MockExecutor) Run(flowContext flow.FlowContext, Step *flow.Step) (*flow.FlowContext, error) {

	return &flowContext, nil
}

func (m *MockExecutor) Init() error {

	return nil
}

func TestNewExecutorFromConfig(t *testing.T) {

	t.Run("Invalid type", func(t *testing.T) {

		executorConfig := &executors.ExecutorConfig{
			Type: "invalid-executor",
		}

		executor, err := NewExecutorFromConfig(executorConfig)

		assert.Error(t, err)
		assert.Nil(t, executor)
	})

	t.Run("Success path with param", func(t *testing.T) {

		executor1 := &MockExecutor{}
		RegisterExecutor("valid-executor", executor1)

		executorConfig := &executors.ExecutorConfig{
			Type: "valid-executor",
			WithConfig: map[string]interface{}{
				"param1": "value1",
				"param2": 10,
			},
		}

		result, err := NewExecutorFromConfig(executorConfig)
		executor := result.(*MockExecutor)

		assert.NoError(t, err)
		assert.NotNil(t, executor)

		assert.Equal(t, "value1", executor.Param1)
		assert.Equal(t, 10, executor.Param2)

	})
}
