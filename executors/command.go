package executors

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/shello"
)

// RunCommandExecutor is an executor that runs system commands.
// It executes the command specified in the flow context's Text field.
type RunCommandExecutor struct {
	Silent          bool   `json:"silent" yaml:"silent" mapstructure:"silent"`
	OutputToContext bool   `json:"outputToContext" yaml:"outputToContext" mapstructure:"outputToContext"`
	Path            string `json:"path" yaml:"path" mapstructure:"path"`
}

// Init initializes the RunCommandExecutor.
// This implementation has no initialization requirements.
func (executor *RunCommandExecutor) Init() error {
	return nil
}

// Run executes the command specified in the flow context's Text field.
//
// Parameters:
//   - flowContext: The flow context containing the command to execute in its Text field
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context (with command output in Text field if OutputToContext is true)
//   - Any error encountered during command execution
func (executor *RunCommandExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	commandText := flowContext.Text
	if commandText == "" {
		return &flowContext, errors.New("no command provided")
	}
	if !executor.Silent {
		log.Infof("Running command: %s", commandText)
	}

	outputString, _, err := shello.RunOutputWithDir(commandText, executor.Path)

	if err != nil {
		return &flowContext, err
	}
	if !executor.Silent {
		log.Infof("%s\n", outputString)
	}
	if executor.OutputToContext {
		flowContext.Text = outputString
	}
	return &flowContext, nil
}