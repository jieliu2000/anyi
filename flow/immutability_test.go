package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVarsImmutable tests that variables cannot be modified when VarsImmutable is true
func TestVarsImmutable(t *testing.T) {
	// Create a mock executor that tries to modify variables
	executor := &MockExecutor{
		Mock: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Try to modify existing variable
			flowContext.SetVariable("existingVar", "newValue")
			// Try to add new variable
			flowContext.SetVariable("newVar", "value")
			return &flowContext, nil
		},
	}

	// Create a step with VarsImmutable set to true
	step := NewStep(executor, nil, nil)
	step.VarsImmutable = true

	// Create initial context with variables
	initialContext := FlowContext{
		Variables: map[string]any{
			"existingVar": "originalValue",
		},
	}

	// Run the step
	result, err := tryStep(step, initialContext)

	// Verify no errors
	assert.NoError(t, err)

	// Verify variables were not modified
	assert.Equal(t, "originalValue", result.Variables["existingVar"])
	_, exists := result.Variables["newVar"]
	assert.False(t, exists, "New variable should not have been added")
}

// TestTextImmutable tests that text cannot be modified when TextImmutable is true
func TestTextImmutable(t *testing.T) {
	// Create a mock executor that tries to modify text
	executor := &MockExecutor{
		Mock: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Try to modify text
			flowContext.Text = "Modified Text"
			return &flowContext, nil
		},
	}

	// Create a step with TextImmutable set to true
	step := NewStep(executor, nil, nil)
	step.TextImmutable = true

	// Create initial context with text
	initialContext := FlowContext{
		Text: "Original Text",
	}

	// Run the step
	result, err := tryStep(step, initialContext)

	// Verify no errors
	assert.NoError(t, err)

	// Verify text was not modified
	assert.Equal(t, "Original Text", result.Text)
}

// TestMemoryImmutable tests that memory cannot be modified when MemoryImmutable is true
func TestMemoryImmutable(t *testing.T) {
	// Create a mock executor that tries to modify memory
	executor := &MockExecutor{
		Mock: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Try to modify memory
			flowContext.Memory = map[string]string{
				"key": "modified",
			}
			return &flowContext, nil
		},
	}

	// Create a step with MemoryImmutable set to true
	step := NewStep(executor, nil, nil)
	step.MemoryImmutable = true

	// Create initial context with memory
	initialMemory := map[string]string{
		"key": "original",
	}
	initialContext := FlowContext{
		Memory: initialMemory,
	}

	// Run the step
	result, err := tryStep(step, initialContext)

	// Verify no errors
	assert.NoError(t, err)

	// Verify memory was not modified
	memory := result.Memory.(map[string]string)
	assert.Equal(t, "original", memory["key"])
}

// TestMutableProperties tests that properties can be modified when immutability flags are false
func TestMutableProperties(t *testing.T) {
	// Create a mock executor that modifies all properties
	executor := &MockExecutor{
		Mock: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Modify variables
			flowContext.SetVariable("existingVar", "newValue")
			flowContext.SetVariable("newVar", "value")

			// Modify text
			flowContext.Text = "Modified Text"

			// Modify memory
			flowContext.Memory = map[string]string{
				"key": "modified",
			}

			return &flowContext, nil
		},
	}

	// Create a step with all immutability flags set to false
	step := NewStep(executor, nil, nil)
	step.VarsImmutable = false
	step.TextImmutable = false
	step.MemoryImmutable = false

	// Create initial context
	initialContext := FlowContext{
		Variables: map[string]any{
			"existingVar": "originalValue",
		},
		Text: "Original Text",
		Memory: map[string]string{
			"key": "original",
		},
	}

	// Run the step
	result, err := tryStep(step, initialContext)

	// Verify no errors
	assert.NoError(t, err)

	// Verify all properties were modified
	assert.Equal(t, "newValue", result.Variables["existingVar"])
	assert.Equal(t, "value", result.Variables["newVar"])
	assert.Equal(t, "Modified Text", result.Text)
	memory := result.Memory.(map[string]string)
	assert.Equal(t, "modified", memory["key"])
}
