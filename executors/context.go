package executors

import (
	"github.com/jieliu2000/anyi/flow"
)

// SetContextExecutor is an executor that sets values in the flow context.
// It can modify the Text field and Memory object in the flow context.
type SetContextExecutor struct {
	Text   string               `json:"text" yaml:"text" mapstructure:"text"`
	Memory flow.ShortTermMemory `json:"memory" yaml:"memory" mapstructure:"memory"`

	Force bool `json:"force" yaml:"force" mapstructure:"force"`
}

// Init initializes the SetContextExecutor.
// This implementation has no initialization requirements.
func (executor *SetContextExecutor) Init() error {
	return nil
}

// Run sets the text and memory of the flow context.
// If the Force flag is set to true, it will override the existing text and memory.
// Otherwise, it will only set the text and memory if they are not empty.
//
// Parameters:
//   - flowContext: The current flow context to modify
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context with modified text and memory
//   - Any error encountered during execution
func (executor *SetContextExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {

	if executor.Text != "" || executor.Force {
		flowContext.Text = executor.Text
	}

	if executor.Memory != nil || executor.Force {
		flowContext.Memory = executor.Memory
	}
	return &flowContext, nil
}

// SetVariablesExecutor is an executor that sets multiple variables in the flow context at once
type SetVariablesExecutor struct {
	// Variables to set, map of variable names to their corresponding values
	// Example: { "var1": "value1", "var2": 123, "var3": true }
	Variables map[string]any `json:"variables" yaml:"variables" mapstructure:"variables"`
}

// Run executes the variable setting operation for multiple variables at once
//
// Parameters:
//   - flowContext: The current flow context
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context with the new variables set
//   - Any error encountered during execution
//
// Example usage in configuration:
//
//	{
//	  "type": "setVariables",
//	  "variables": {
//	    "username": "john_doe",
//	    "age": 30,
//	    "isActive": true,
//	    "preferences": { "theme": "dark", "notifications": false }
//	  }
//	}
func (executor *SetVariablesExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// Check if variables are immutable for this step
	if step != nil && step.VarsImmutable {
		// If variables are immutable, return the context unchanged
		return &flowContext, nil
	}

	// Ensure Variables is initialized in flowContext
	if flowContext.Variables == nil {
		flowContext.Variables = make(map[string]any)
	}

	// Set multiple variables simultaneously. Each key in the map is a variable name,
	// and its corresponding value will be assigned to that variable.
	// For example, with Variables = {"name": "John", "age": 30, "active": true},
	// this will create/update three different variables with their respective values.
	if executor.Variables != nil {
		for name, value := range executor.Variables {
			if name == "" {
				continue // Skip empty variable names
			}

			// Set the variable (always overwrite existing values)
			flowContext.Variables[name] = value
		}
	}

	return &flowContext, nil
}

// Init initializes SetVariablesExecutor
func (executor *SetVariablesExecutor) Init() error {
	return nil
}