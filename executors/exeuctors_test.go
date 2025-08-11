package executors

import (
	"testing"

	"github.com/jieliu2000/anyi/registry"

	"github.com/stretchr/testify/assert"
)

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
		registry.RegisterExecutor("valid-executor", executor1)

		executorConfig := &ExecutorConfig{
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
