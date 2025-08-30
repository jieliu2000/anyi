package executors

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/registry"
)

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
		_, err := registry.GetFlow(value)
		if err != nil {
			return errors.Join(err, errors.New("failed to get flow "+value))
		}
	}

	// Validate default flow if provided
	if executor.Default != "" {
		_, err := registry.GetFlow(executor.Default)
		if err != nil {
			return errors.Join(err, errors.New("failed to get default flow "+executor.Default))
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

	flow, err := registry.GetFlow(flowName)
	if err != nil {
		return &flowContext, err
	}

	return flow.Run(flowContext)
}
