package registry

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/executors"
	"github.com/jieliu2000/anyi/flow"
)

// DecoratedExecutor is an executor that wraps another executor with pre-run and post-run functions.
// This allows for adding behavior before and after the execution of the wrapped executor.
type DecoratedExecutor struct {
	ExecutorImpl flow.StepExecutor                                                              `json:"-" yaml:"-" mapstructure:"-"`
	PreRun       func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	PostRun      func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	With         *executors.ExecutorConfig                                                      `json:"with" yaml:"with" mapstructure:"with"`
}

// Init initializes the DecoratedExecutor.
// It checks if an executor is provided and if pre or post run functions are set.
// If a configuration is provided but no executor, it creates one from the configuration.
//
// Returns:
//   - An error if no executor is provided or if neither pre nor post run functions are set
func (executor *DecoratedExecutor) Init() error {
	if executor.With != nil && executor.ExecutorImpl == nil {
		impl, err := NewExecutorFromConfig(executor.With)
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

// ConditionalFlowExecutor is an executor that routes flow execution based on conditions.
// It uses the text in the flow context to determine which flow to execute next.
// If no condition matches and a Default flow is specified, it will execute the default flow.
type ConditionalFlowExecutor struct {
	Switch  map[string]string `json:"switch" yaml:"switch" mapstructure:"switch"`
	Default string            `json:"default" yaml:"default" mapstructure:"default"`
	Trim    string            `json:"trim" yaml:"trim" mapstructure:"trim"`
}

// Init initializes the ConditionalFlowExecutor.
// It checks the provided switches and default flow, and retrieves the corresponding flows.
//
// Returns:
//   - An error if no switches are provided or if any referenced flow cannot be found
func (executor *ConditionalFlowExecutor) Init() error {
	if len(executor.Switch) == 0 {
		return errors.New("no switch provided")
	}

	// Validate switch flows
	for _, value := range executor.Switch {
		flow, err := GetFlow(value)
		if err != nil {
			return errors.Join(err, errors.New("failed to get flow "+value))
		}
		if flow == nil {
			return errors.New("flow " + value + " not found")
		}
	}

	// Validate default flow if provided
	if executor.Default != "" {
		flow, err := GetFlow(executor.Default)
		if err != nil {
			return errors.Join(err, errors.New("failed to get default flow "+executor.Default))
		}
		if flow == nil {
			return errors.New("default flow " + executor.Default + " not found")
		}
	}

	return nil
}

// Run executes the flow based on the condition in the flow context.
// The text in the flow context is used as a key to find the next flow to execute.
// If no matching condition is found, it will execute the default flow if specified.
//
// Parameters:
//   - flowContext: The current flow context containing the condition text
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context after the selected flow executes
//   - An error if no matching flow is found and no default flow is specified, or if the flow execution fails
func (executor *ConditionalFlowExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	condition := flowContext.Text
	if executor.Trim != "" {
		condition = strings.Trim(condition, executor.Trim)
	}

	var flowName string
	var found bool

	// Try to find a matching condition in the switch
	flowName, found = executor.Switch[condition]

	// If no match found, use default flow if available
	if !found || flowName == "" {
		if executor.Default != "" {
			flowName = executor.Default
			log.Infof("No matching condition found for '%s', using default flow: %s", condition, flowName)
		} else {
			return &flowContext, fmt.Errorf("no matching flow found for condition '%s' and no default flow specified", condition)
		}
	}

	flow, err := GetFlow(flowName)
	if err != nil {
		return &flowContext, err
	}
	if flow == nil {
		return &flowContext, fmt.Errorf("flow %s not found", flowName)
	}

	return flow.Run(flowContext)
}
