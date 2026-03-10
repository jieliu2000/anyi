package builtin

import (
	"errors"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/registry"
)

// DecoratedExecutor is an executor that decorates another executor with pre and post steps.
type DecoratedExecutor struct {
	PreSteps  []flow.Step       `json:"preSteps" yaml:"preSteps" mapstructure:"preSteps"`
	PostSteps []flow.Step       `json:"postSteps" yaml:"postSteps" mapstructure:"postSteps"`
	Executor  flow.StepExecutor `json:"executor" yaml:"executor" mapstructure:"executor"`
}

// Init initializes the DecoratedExecutor.
// It initializes all the pre and post steps, as well as the decorated executor.
func (executor *DecoratedExecutor) Init() error {
	for i := range executor.PreSteps {
		if initable, ok := executor.PreSteps[i].Executor.(flow.Initable); ok {
			if err := initable.Init(); err != nil {
				return err
			}
		}
	}

	for i := range executor.PostSteps {
		if initable, ok := executor.PostSteps[i].Executor.(flow.Initable); ok {
			if err := initable.Init(); err != nil {
				return err
			}
		}
	}

	if executor.Executor != nil {
		if initable, ok := executor.Executor.(flow.Initable); ok {
			if err := initable.Init(); err != nil {
				return err
			}
		}
	} else {
		return errors.New("executor is nil")
	}

	return nil
}

// Run executes the decorated executor with pre and post steps.
// First, it runs all the pre steps, then the decorated executor, and finally all the post steps.
func (executor *DecoratedExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// Run pre steps
	for i := range executor.PreSteps {
		_, err := executor.PreSteps[i].Executor.Run(flowContext, &executor.PreSteps[i])
		if err != nil {
			return nil, err
		}
	}

	// Run the decorated executor
	result, err := executor.Executor.Run(flowContext, step)
	if err != nil {
		return nil, err
	}

	// Run post steps
	for i := range executor.PostSteps {
		_, err := executor.PostSteps[i].Executor.Run(*result, &executor.PostSteps[i])
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}