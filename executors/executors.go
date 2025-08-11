package executors

import (
	"errors"
	"fmt"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/registry"
	"github.com/mitchellh/mapstructure"
)

// NewExecutorFromConfig creates a new executor from an executor configuration.
// It instantiates the appropriate executor type based on the configuration,
// decodes the configuration parameters, and initializes the executor.
//
// Parameters:
//   - executorConfig: Executor configuration containing type and parameters
//
// Returns:
//   - A new step executor
//   - Any error encountered during executor creation
func NewExecutorFromConfig(executorConfig *ExecutorConfig) (flow.StepExecutor, error) {
	if executorConfig == nil {
		return nil, errors.New("executor config is nil")
	}

	if executorConfig.Type == "" {
		return nil, errors.New("executor type is not set")
	}

	metaExecutor, err := registry.GetExecutor(executorConfig.Type)
	if err != nil {
		return nil, err
	}

	executor := metaExecutor

	if executor == nil {
		return nil, fmt.Errorf("executor type %s is not found", executorConfig.Type)
	}

	mapstructure.Decode(executorConfig.WithConfig, executor)
	executor.Init()
	return executor, nil
}
