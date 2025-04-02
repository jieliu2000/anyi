package flow

import (
	"errors"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/stretchr/testify/assert"
)

type MockStepExecutor struct {
	RunWithError  bool
	InitCompleted bool
	RunCompleted  bool
}

func (executor *MockStepExecutor) Init() error {
	executor.InitCompleted = true
	return nil
}

func (executor *MockStepExecutor) Run(flowContext FlowContext, step *Step) (*FlowContext, error) {
	executor.RunCompleted = true
	if executor.RunWithError {
		return nil, errors.New("error")
	}
	return &flowContext, nil
}

func TestNewFlow(t *testing.T) {

	client := test.MockClient{}
	flow, err := NewFlow(&client, "flow1")
	assert.NoError(t, err)

	assert.NotNil(t, flow)
	assert.Equal(t, "flow1", flow.Name)
	assert.Equal(t, &client, flow.ClientImpl)

}

func TestNewStep(t *testing.T) {
	executor := &MockStepExecutor{}
	validator := &MockValidator{}
	client := &test.MockClient{}

	step := NewStep(executor, validator, client)

	assert.NotNil(t, step)
	assert.Equal(t, executor, step.Executor)
	assert.Equal(t, validator, step.Validator)
	assert.Equal(t, client, step.ClientImpl)
	assert.Equal(t, DefaultMaxRetryTimes, step.MaxRetryTimes)
}

func TestNewStepWithDefaultMaxRetry(t *testing.T) {
	executor := &MockStepExecutor{}
	validator := &MockValidator{}
	client := &test.MockClient{}

	step := NewStep(executor, validator, client)

	assert.NotNil(t, step)
	assert.Equal(t, executor, step.Executor)
	assert.Equal(t, validator, step.Validator)
	assert.Equal(t, client, step.ClientImpl)
	assert.Equal(t, DefaultMaxRetryTimes, step.MaxRetryTimes)
}

func TestNewStepWithCustomMaxRetry(t *testing.T) {
	executor := &MockStepExecutor{}
	validator := &MockValidator{}
	client := &test.MockClient{}
	customMaxRetry := 5

	step := NewStep(executor, validator, client)
	step.MaxRetryTimes = customMaxRetry

	assert.NotNil(t, step)
	assert.Equal(t, executor, step.Executor)
	assert.Equal(t, validator, step.Validator)
	assert.Equal(t, client, step.ClientImpl)
	assert.Equal(t, customMaxRetry, step.MaxRetryTimes)
}

type MockExecutor struct {
	Mock func(flowContext FlowContext, step *Step) (*FlowContext, error)
}

type MockValidator struct {
	Mock func(output string, step *Step) bool
}

func (m MockExecutor) Run(flowContext FlowContext, step *Step) (*FlowContext, error) {
	return m.Mock(flowContext, step)
}

func (m MockExecutor) Init() error {
	return nil
}

func (m MockValidator) Init() error {
	return nil
}

func (m MockValidator) Validate(output string, step *Step) bool {
	if m.Mock == nil {
		return true
	}
	return m.Mock(output, step)
}

func NewStepWithValidatorAndExectorFunction(name string, runFunc func(flowContext FlowContext, step *Step) (*FlowContext, error), validateFunc func(output string, step *Step) bool, client llm.Client) *Step {

	return NewStep(
		MockExecutor{
			Mock: runFunc,
		},
		MockValidator{
			Mock: validateFunc,
		}, client)
}

func Test_tryStep_RunError(t *testing.T) {
	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return nil, errors.New("run error")
		},
		func(output string, step *Step) bool {
			return true
		},
		test.NewMockClient(),
	)
	_, err := tryStep(step, FlowContext{})
	assert.EqualError(t, err, "run error")
}
func Test_tryStep_RetryExceeded(t *testing.T) {

	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return &FlowContext{}, nil
		},
		func(output string, step *Step) bool {
			if step.runTimes <= 2 {
				return false
			} else {
				return true
			}
		},
		test.NewMockClient(),
	)
	step.MaxRetryTimes = 0
	_, err := tryStep(step, FlowContext{})
	assert.EqualError(t, err, "step retry times exceeded")

	step.runTimes = 0
	step.MaxRetryTimes = 1
	_, err = tryStep(step, FlowContext{})
	assert.EqualError(t, err, "step retry times exceeded")

	step.runTimes = 0
	step.MaxRetryTimes = 3
	_, err = tryStep(step, FlowContext{})
	assert.NoError(t, err, "step should be success because retry times doesn't exceed MaxRetryTimes")
}
func Test_tryStep_ValidatorError(t *testing.T) {
	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return &FlowContext{}, nil
		},
		func(output string, step *Step) bool {
			return false
		},
		test.NewMockClient(),
	)
	_, err := tryStep(step, FlowContext{})
	assert.Error(t, err)
}
func Test_tryStep_ValidatorSuccess(t *testing.T) {
	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return &FlowContext{}, nil
		},
		func(output string, step *Step) bool {
			return true
		},
		test.NewMockClient(),
	)
	result, err := tryStep(step, FlowContext{})
	assert.Nil(t, err)
	assert.Equal(t, result, &FlowContext{})
}

func TestFlow_Run(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			flowContext.Memory = "Data from Step 1"
			return &flowContext, nil
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			flowContext.Memory = flowContext.Memory.(string) + " and Step 2"
			return &flowContext, nil
		}, nil, nil),
	)
	assert.Nil(t, err)

	flowContext, err := flow.Run(FlowContext{Text: "Initial"})
	assert.Nil(t, err)
	assert.Equal(t, "Initial", flowContext.Text)
	assert.Equal(t, "Data from Step 1 and Step 2", flowContext.Memory)
}
func TestFlow_Run_WithInvalidStep(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			flowContext.Memory = "Data from Step 1"
			return &flowContext, nil
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return nil, errors.New("Error in Step 2")
		}, nil, nil),
	)

	assert.Nil(t, err)
	_, err = flow.Run(FlowContext{Text: "Initial"})
	assert.NotNil(t, err)
}
func TestFlow_Run_WithMaxRetryTimes(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return nil, errors.New("Error in Step 1")
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return nil, errors.New("Error in Step 2")
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 3", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			return nil, errors.New("Error in Step 3")
		}, nil, nil),
	)

	assert.NoError(t, err)
	_, err = flow.Run(FlowContext{Text: "Initial"})

	assert.NotNil(t, err)
}

func TestFlow_Run_WithValidator_And_Validation_Passed(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			assert.Equal(t, "Initial", flowContext.Text)
			flowContext.Text = "Data from Step 1"
			return &flowContext, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 1"
		}, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			assert.Equal(t, "Data from Step 1", flowContext.Text)
			flowContext.Text = "Data from Step 2"
			return &flowContext, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 2"
		}, nil),
	)

	assert.NoError(t, err)
	flowContext, err := flow.Run(FlowContext{Text: "Initial"})
	assert.NoError(t, err)
	assert.Equal(t, "Data from Step 2", flowContext.Text)
}
func TestFlow_Run_WithValidator_And_Validation_Failed(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {

			return &flowContext, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 1"
		}, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			flowContext.Memory = "Data from Step 1"
			return &flowContext, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 1 and Step 2"
		}, nil),
	)

	assert.NoError(t, err)
	_, err = flow.Run(FlowContext{Text: "Initial"})
	assert.Error(t, err)
}
func TestFlow_Run_WithValidatorAndInvalidStep(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1",
			func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				return &flowContext, nil
			},
			func(stepOutput string, step *Step) bool {
				return stepOutput == "Data from Step 1"
			}, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2",
			func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				return nil, errors.New("Error in Step 2")
			},
			func(stepOutput string, step *Step) bool {
				return stepOutput == "Data from Step 1 and Step 2"
			}, nil),
	)

	assert.NoError(t, err)
	_, err = flow.Run(FlowContext{Text: "Initial"})
	assert.NotNil(t, err)
}

func TestFlow_Run_WithThinkTags(t *testing.T) {
	// Create a step that returns text with <think> tags
	step1 := NewStepWithValidatorAndExectorFunction(
		"Step with think tag",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Return text with <think> tags
			flowContext.Text = "This is the output result. <think>This is my thinking process, which should be extracted to the Think field</think> Final conclusion."
			return &flowContext, nil
		},
		nil,
		nil,
	)

	// Create a step that doesn't return <think> tags
	step2 := NewStepWithValidatorAndExectorFunction(
		"Step without think tag",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Add to the output of the previous step
			flowContext.Text = flowContext.Text + " This is the output of the second step, without any think tags."
			return &flowContext, nil
		},
		nil,
		nil,
	)

	// Create a test flow with these two steps
	flow, err := NewFlow(&test.MockClient{}, "Test Flow with Think Tags", *step1, *step2)
	assert.NoError(t, err)

	// Run the flow
	flowContext, err := flow.Run(FlowContext{Text: "Initial input"})
	assert.NoError(t, err)

	// Verify that the Think field contains the extracted thinking content
	assert.Contains(t, flowContext.Think, "This is my thinking process")
	assert.Equal(t, "<think>This is my thinking process, which should be extracted to the Think field</think>", flowContext.Think)

	// Verify that the Text field doesn't contain any <think> tag content
	assert.NotContains(t, flowContext.Text, "<think>")
	assert.NotContains(t, flowContext.Text, "</think>")
	assert.NotContains(t, flowContext.Text, "This is my thinking process")
}

// Test the case where multiple steps contain think tags
func TestFlow_Run_WithMultipleThinkTags(t *testing.T) {
	// Create the first step that returns text with <think> tags
	step1 := NewStepWithValidatorAndExectorFunction(
		"Step 1 with think tag",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			flowContext.Text = "First step result. <think>Thinking process of the first step</think> First step conclusion."
			return &flowContext, nil
		},
		nil,
		nil,
	)

	// Create the second step that returns text with <think> tags
	step2 := NewStepWithValidatorAndExectorFunction(
		"Step 2 with think tag",
		func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			flowContext.Text = "Second step processing.<think>The thinking process of the second step is more complex</think> Final decision."
			return &flowContext, nil
		},
		nil,
		nil,
	)

	// Create a test flow with these two steps
	flow, err := NewFlow(&test.MockClient{}, "Test Flow with Multiple Think Tags", *step1, *step2)
	assert.NoError(t, err)

	// Run the flow
	flowContext, err := flow.Run(FlowContext{Text: "Initial input"})
	assert.NoError(t, err)

	// Verify that the Think field contains the thinking content from the last step
	assert.Contains(t, flowContext.Think, "thinking process of the second step")
	assert.Equal(t, "<think>The thinking process of the second step is more complex</think>", flowContext.Think)

	// Verify that the Text field doesn't contain any <think> tag content
	assert.NotContains(t, flowContext.Text, "<think>")
	assert.NotContains(t, flowContext.Text, "</think>")
	assert.NotContains(t, flowContext.Text, "thinking process")

	// Verify that the Text field contains the cleaned text from the last step
	assert.Equal(t, "Second step processing. Final decision.", flowContext.Text)
}

func TestFlow_RunWithMemory(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Verify memory is correctly passed
			if flowContext.Memory != "test memory" {
				return nil, errors.New("memory not correctly passed")
			}
			return &flowContext, nil
		}, nil, nil),
	)

	assert.NoError(t, err)
	result, err := flow.RunWithMemory("test memory")

	assert.NoError(t, err)
	assert.Equal(t, "test memory", result.Memory)
}

func TestFlow_RunWithVariables(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Verify variables are correctly passed
			if flowContext.Variables["key1"] != "value1" || flowContext.Variables["key2"] != 42 {
				return nil, errors.New("variables not correctly passed")
			}

			// Add another variable
			flowContext.Variables["key3"] = true
			return &flowContext, nil
		}, nil, nil),
	)

	assert.NoError(t, err)
	variables := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	result, err := flow.RunWithVariables(variables)

	assert.NoError(t, err)
	assert.Equal(t, "value1", result.Variables["key1"])
	assert.Equal(t, 42, result.Variables["key2"])
	assert.Equal(t, true, result.Variables["key3"])
}

func TestFlow_RunWithInputAndVariables(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Verify input text and variables are correctly passed
			if flowContext.Text != "test input" {
				return nil, errors.New("input text not correctly passed")
			}

			if flowContext.Variables["name"] != "John" || flowContext.Variables["age"] != 30 {
				return nil, errors.New("variables not correctly passed")
			}

			// Modify text and add a variable
			flowContext.Text = "modified text"
			flowContext.Variables["active"] = true

			return &flowContext, nil
		}, nil, nil),
	)

	assert.NoError(t, err)
	variables := map[string]any{
		"name": "John",
		"age":  30,
	}

	result, err := flow.RunWithInputAndVariables("test input", variables)

	assert.NoError(t, err)
	assert.Equal(t, "modified text", result.Text)
	assert.Equal(t, "John", result.Variables["name"])
	assert.Equal(t, 30, result.Variables["age"])
	assert.Equal(t, true, result.Variables["active"])
}

// Test with nil variables map
func TestFlow_RunWithVariables_NilMap(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
			// Verify variables map is initialized
			if flowContext.Variables == nil {
				return nil, errors.New("variables map not initialized")
			}

			// Verify it's empty
			if len(flowContext.Variables) != 0 {
				return nil, errors.New("variables map not empty")
			}

			return &flowContext, nil
		}, nil, nil),
	)

	assert.NoError(t, err)
	result, err := flow.RunWithVariables(nil)

	assert.NoError(t, err)
	assert.NotNil(t, result.Variables)
	assert.Equal(t, 0, len(result.Variables))
}

func TestFlowContext_Variables(t *testing.T) {
	// Test initialization of Variables in FlowContext constructors
	t.Run("NewFlowContext initializes Variables", func(t *testing.T) {
		fc := NewFlowContext("test", nil)
		assert.NotNil(t, fc.Variables)
		assert.Equal(t, 0, len(fc.Variables))
	})

	t.Run("Flow.NewFlowContext initializes Variables", func(t *testing.T) {
		flow := &Flow{}
		fc := flow.NewFlowContext("test", nil)
		assert.NotNil(t, fc.Variables)
		assert.Equal(t, 0, len(fc.Variables))
	})

	// Test GetVariable and SetVariable methods
	t.Run("GetVariable and SetVariable", func(t *testing.T) {
		fc := NewFlowContext("test", nil)

		// Initially variable should be nil
		assert.Nil(t, fc.GetVariable("key1"))

		// Set and get a string variable
		fc.SetVariable("key1", "value1")
		assert.Equal(t, "value1", fc.GetVariable("key1"))

		// Set and get an integer variable
		fc.SetVariable("key2", 123)
		assert.Equal(t, 123, fc.GetVariable("key2"))

		// Set and get a boolean variable
		fc.SetVariable("key3", true)
		assert.Equal(t, true, fc.GetVariable("key3"))

		// Override an existing variable
		fc.SetVariable("key1", "new_value")
		assert.Equal(t, "new_value", fc.GetVariable("key1"))
	})

	// Test typed getter methods
	t.Run("GetVariableString", func(t *testing.T) {
		fc := NewFlowContext("test", nil)

		// Default value for non-existent key
		assert.Equal(t, "default", fc.GetVariableString("missing", "default"))

		// String value
		fc.SetVariable("str_key", "value")
		assert.Equal(t, "value", fc.GetVariableString("str_key", "default"))

		// Non-string value should return default
		fc.SetVariable("int_key", 123)
		assert.Equal(t, "default", fc.GetVariableString("int_key", "default"))
	})

	t.Run("GetVariableInt", func(t *testing.T) {
		fc := NewFlowContext("test", nil)

		// Default value for non-existent key
		assert.Equal(t, 42, fc.GetVariableInt("missing", 42))

		// Integer value
		fc.SetVariable("int_key", 123)
		assert.Equal(t, 123, fc.GetVariableInt("int_key", 42))

		// Non-integer value should return default
		fc.SetVariable("str_key", "value")
		assert.Equal(t, 42, fc.GetVariableInt("str_key", 42))
	})

	t.Run("GetVariableBool", func(t *testing.T) {
		fc := NewFlowContext("test", nil)

		// Default value for non-existent key
		assert.Equal(t, true, fc.GetVariableBool("missing", true))

		// Boolean value
		fc.SetVariable("bool_key", false)
		assert.Equal(t, false, fc.GetVariableBool("bool_key", true))

		// Non-boolean value should return default
		fc.SetVariable("str_key", "value")
		assert.Equal(t, true, fc.GetVariableBool("str_key", true))
	})

	// Test WithVariable method
	t.Run("WithVariable creates new context", func(t *testing.T) {
		fc := NewFlowContext("test", nil)
		fc.SetVariable("key1", "value1")

		// Create new context with additional variable
		newFc := fc.WithVariable("key2", "value2")

		// Original context should be unchanged
		assert.Equal(t, 1, len(fc.Variables))
		assert.Equal(t, "value1", fc.GetVariable("key1"))
		assert.Nil(t, fc.GetVariable("key2"))

		// New context should have both variables
		assert.Equal(t, 2, len(newFc.Variables))
		assert.Equal(t, "value1", newFc.GetVariable("key1"))
		assert.Equal(t, "value2", newFc.GetVariable("key2"))

		// Modify new context should not affect original
		newFc.SetVariable("key1", "modified")
		assert.Equal(t, "value1", fc.GetVariable("key1"))
		assert.Equal(t, "modified", newFc.GetVariable("key1"))
	})

	// Test variables passing through flow steps
	t.Run("Variables are preserved in flow execution", func(t *testing.T) {
		flow, err := NewFlow(&test.MockClient{}, "Test Flow",
			*NewStepWithValidatorAndExectorFunction("Step 1", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				// Verify variable from initial context
				if flowContext.GetVariableString("initial", "") != "value" {
					return nil, errors.New("initial variable not found or incorrect")
				}

				// Set a new variable
				flowContext.SetVariable("step1", "executed")
				return &flowContext, nil
			}, nil, nil),
			*NewStepWithValidatorAndExectorFunction("Step 2", func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				// Verify variables from initial context and step 1
				if flowContext.GetVariableString("initial", "") != "value" {
					return nil, errors.New("initial variable not found or incorrect")
				}
				if flowContext.GetVariableString("step1", "") != "executed" {
					return nil, errors.New("step1 variable not found or incorrect")
				}

				// Set another variable
				flowContext.SetVariable("step2", "completed")
				return &flowContext, nil
			}, nil, nil),
		)

		assert.NoError(t, err)

		// Create context with initial variable
		initialContext := FlowContext{Text: "Initial"}
		initialContext.Variables = make(map[string]any)
		initialContext.Variables["initial"] = "value"

		// Run the flow
		resultContext, err := flow.Run(initialContext)

		// Verify no errors
		assert.NoError(t, err)

		// Verify all variables are present in the result
		assert.Equal(t, "value", resultContext.GetVariable("initial"))
		assert.Equal(t, "executed", resultContext.GetVariable("step1"))
		assert.Equal(t, "completed", resultContext.GetVariable("step2"))
	})
}
