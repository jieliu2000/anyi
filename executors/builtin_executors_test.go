package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

// ... existing code ...

func TestSetVariablesExecutor_Run(t *testing.T) {
	// Test Case 1: Setting variables with an empty variable name
	t.Run("Skip empty variable names", func(t *testing.T) {
		executor := SetVariablesExecutor{
			Variables: map[string]any{
				"":      "emptyName",
				"valid": "validValue",
			},
		}
		flowContext := flow.FlowContext{
			Variables: make(map[string]any),
		}
		step := &flow.Step{}
		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Variables))
		assert.Equal(t, "validValue", result.Variables["valid"])
	})

	// Test Case 2: Setting multiple variables
	t.Run("Set multiple variables", func(t *testing.T) {
		executor := SetVariablesExecutor{
			Variables: map[string]any{
				"var1": "value1",
				"var2": 42,
				"var3": true,
			},
		}
		flowContext := flow.FlowContext{
			Variables: make(map[string]any),
		}
		step := &flow.Step{}
		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Variables))
		assert.Equal(t, "value1", result.Variables["var1"])
		assert.Equal(t, 42, result.Variables["var2"])
		assert.Equal(t, true, result.Variables["var3"])
	})

	// Test Case 3: Setting a variable when Variables is nil
	t.Run("Initialize Variables if nil", func(t *testing.T) {
		executor := SetVariablesExecutor{
			Variables: map[string]any{
				"testVar": "testValue",
			},
		}
		flowContext := flow.FlowContext{
			Variables: nil,
		}
		step := &flow.Step{}
		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.NotNil(t, result.Variables)
		assert.Equal(t, "testValue", result.Variables["testVar"])
	})

	// Test Case 4: Always overwrite existing variables
	t.Run("Always overwrite existing variables", func(t *testing.T) {
		executor := SetVariablesExecutor{
			Variables: map[string]any{
				"existingVar": "newValue",
				"newVar":      "value",
			},
		}
		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"existingVar": "existingValue",
			},
		}
		step := &flow.Step{}
		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.Equal(t, "newValue", result.Variables["existingVar"])
		assert.Equal(t, "value", result.Variables["newVar"])
	})

	// Test Case 5: Complex variable values with different types
	t.Run("Complex variable values with different types", func(t *testing.T) {
		// Create a complex nested structure with different variable types
		nestedMap := map[string]any{
			"setting1": "enabled",
			"setting2": 42,
			"nested": map[string]bool{
				"feature1": true,
				"feature2": false,
			},
		}

		// Create an array/slice
		arrayData := []any{"item1", 2, true}

		executor := SetVariablesExecutor{
			Variables: map[string]any{
				"stringVar": "Hello World",                       // String
				"intVar":    42,                                  // Integer
				"floatVar":  3.14159,                             // Float
				"boolVar":   true,                                // Boolean
				"mapVar":    nestedMap,                           // Nested map
				"arrayVar":  arrayData,                           // Array/slice
				"nullVar":   nil,                                 // Nil value
				"structVar": struct{ Name string }{Name: "Test"}, // Struct
			},
		}

		flowContext := flow.FlowContext{
			Variables: make(map[string]any),
		}
		step := &flow.Step{}

		// Run the executor
		result, err := executor.Run(flowContext, step)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, 8, len(result.Variables))

		// Check individual variables
		assert.Equal(t, "Hello World", result.Variables["stringVar"])
		assert.Equal(t, 42, result.Variables["intVar"])
		assert.Equal(t, 3.14159, result.Variables["floatVar"])
		assert.Equal(t, true, result.Variables["boolVar"])
		assert.Equal(t, nil, result.Variables["nullVar"])

		// Check nested map
		resultMap, ok := result.Variables["mapVar"].(map[string]any)
		assert.True(t, ok, "mapVar should be a map")
		assert.Equal(t, "enabled", resultMap["setting1"])
		assert.Equal(t, 42, resultMap["setting2"])

		// Check array
		resultArray, ok := result.Variables["arrayVar"].([]any)
		assert.True(t, ok, "arrayVar should be an array")
		assert.Equal(t, 3, len(resultArray))
		assert.Equal(t, "item1", resultArray[0])
		assert.Equal(t, 2, resultArray[1])
		assert.Equal(t, true, resultArray[2])

		// Check struct
		resultStruct, ok := result.Variables["structVar"].(struct{ Name string })
		assert.True(t, ok, "structVar should be a struct")
		assert.Equal(t, "Test", resultStruct.Name)
	})

	// Test Case 6: Respect VarsImmutable flag
	t.Run("Respect VarsImmutable flag", func(t *testing.T) {
		executor := SetVariablesExecutor{
			Variables: map[string]any{
				"testVar": "testValue",
			},
		}

		// Create a flow context with existing variables
		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"existingVar": "existingValue",
			},
		}

		// Create a step with VarsImmutable set to true
		step := &flow.Step{
			VarsImmutable: true,
		}

		// Run the executor
		result, err := executor.Run(flowContext, step)

		// Verify no errors
		assert.NoError(t, err)

		// Verify variables were not modified
		assert.Equal(t, 1, len(result.Variables))
		assert.Equal(t, "existingValue", result.Variables["existingVar"])
		_, exists := result.Variables["testVar"]
		assert.False(t, exists, "New variable should not have been added when VarsImmutable is true")
	})
}

func TestDeepSeekStyleResponseFilter_Think(t *testing.T) {
	// Create a new filter
	filter := &DeepSeekStyleResponseFilter{}
	err := filter.Init()
	assert.NoError(t, err)

	// Test case 1: Text with <think> tags
	testText := "Let me think about this. <think>This is my thinking process. I need to consider several factors.</think> Based on my analysis, the answer is 42."
	flowContext := flow.FlowContext{
		Text: testText,
	}

	// Run the filter
	result, err := filter.Run(flowContext, nil)
	assert.NoError(t, err)

	// Check that Think field contains the extracted thinking content
	assert.Equal(t, "<think>This is my thinking process. I need to consider several factors.</think>", result.Think)

	// Check that Text field has thinking content removed
	assert.Equal(t, "Let me think about this.  Based on my analysis, the answer is 42.", result.Text)

	// Test case 2: Test with OutputJSON=true
	filter.OutputJSON = true
	result, err = filter.Run(flowContext, nil)
	assert.NoError(t, err)

	// Check that Think field still contains the extracted thinking content
	assert.Equal(t, "<think>This is my thinking process. I need to consider several factors.</think>", result.Think)

	// Check that Text field has JSON format with both thinking and result
	expectedJSON := `{"think": "<think>This is my thinking process. I need to consider several factors.</think>", "result": "Let me think about this.  Based on my analysis, the answer is 42."}`
	assert.Equal(t, expectedJSON, result.Text)
}

func TestDeepSeekStyleResponseFilter_Init(t *testing.T) {
	// ... existing code ...
}

// MCP Executor tests have been moved to mcp_executor_test.go