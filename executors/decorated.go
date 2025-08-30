package executors

import (
	"errors"

	"github.com/jieliu2000/anyi/config"
	"github.com/jieliu2000/anyi/flow"
)

// DecoratedExecutor is an executor that wraps another executor with pre-run and post-run functions.
// This allows for adding behavior before and after the execution of the wrapped executor.
type DecoratedExecutor struct {
	ExecutorImpl flow.StepExecutor                                                              `json:"-" yaml:"-" mapstructure:"-"`
	PreRun       func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	PostRun      func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	With         *config.ExecutorConfig                                                                `json:"with" yaml:"with" mapstructure:"with"`
}

// Init initializes the DecoratedExecutor.
// It checks if an executor is provided and if pre or post run functions are set.
// If a configuration is provided but no executor, it creates one from the configuration.
//
// Returns:
//   - An error if no executor is provided or if neither pre nor post run functions are set
func (executor *DecoratedExecutor) Init() error {
	if executor.With != nil && executor.ExecutorImpl == nil {
		impl, err := config.NewExecutorFromConfig(executor.With)
		if err != nil {
			return err
		}
		executor.ExecutorImpl = impl
	}
	if executor.ExecutorImpl == nil {
		return errors.New("no executor provided")
	}

	if executor.PreRun == nil && executor.PostRun == nil {
		return errors.New("no pre or post run function provided")
	}
	return executor.ExecutorImpl.Init()
}

// Run executes the step within the provided flow context.
// It applies the pre-run function (if set), then the wrapped executor, then the post-run function (if set).
//
// Parameters:
//   - flowContext: The current flow context
//   - step: The step to be executed
//
// Returns:
//   - Updated flow context after execution
//   - Any error encountered during execution
func (executor *DecoratedExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	context := &flowContext
	if executor.ExecutorImpl == nil {
		return context, errors.New("no executor provided")
	}
	if executor.PreRun != nil {
		var err error
		context, err := executor.PreRun(*context, step)
		if err != nil {
			return context, err
		}
	}
	context, err := executor.ExecutorImpl.Run(*context, step)
	if executor.PostRun != nil {

		context, err = executor.PostRun(*context, step)
	}
	return context, err
}