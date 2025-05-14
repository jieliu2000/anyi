package anyi

import (
	"errors"
	"log"
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/stretchr/testify/assert"
)

type MockStepExecutor struct {
	RunWithError   bool
	InitCompleted  bool
	RunCompleted   bool
	ExpectedOutput string
}

func (executor *MockStepExecutor) Init() error {
	executor.InitCompleted = true
	return nil
}

func (executor *MockStepExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	executor.RunCompleted = true
	if executor.RunWithError {
		return nil, errors.New("error")
	}
	if executor.ExpectedOutput != "" {
		flowContext.Text = executor.ExpectedOutput
	}
	return &flowContext, nil
}

func TestLLMStepExecutor_Init(t *testing.T) {
	executor := &LLMExecutor{}
	err := executor.Init()
	assert.Error(t, err)
	executor = &LLMExecutor{
		Template: "Hello, {{.name}}!",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
	assert.Equal(t, "Hello, {{.name}}!", executor.TemplateFormatter.TemplateString)

	executor = &LLMExecutor{
		TemplateFile: "./internal/test/test_prompt2.tmpl",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
}

func TestDecratedStepExecutor_Init_NoExecutorProvided(t *testing.T) {
	executor := DecoratedExecutor{}
	err := executor.Init()
	assert.Error(t, err)
}
func TestDecratedStepExecutor_Init_NoPreOrPostRunProvided(t *testing.T) {
	executor := DecoratedExecutor{
		ExecutorImpl: &DecoratedExecutor{},
	}
	err := executor.Init()
	assert.Error(t, err)
	assert.EqualError(t, err, "no pre or post run function provided")
}

func TestDecratedStepExecutor_Init_NoExecutor(t *testing.T) {
	executor := DecoratedExecutor{}
	err := executor.Init()
	assert.Error(t, err)
	assert.EqualError(t, err, "no executor provided")
}
func TestDecratedStepExecutor_Init(t *testing.T) {
	mockExecutor := &MockStepExecutor{}
	executor := DecoratedExecutor{
		PreRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
			return &flowContext, nil
		},
		ExecutorImpl: mockExecutor,
	}
	err := executor.Init()
	assert.NoError(t, err)
	assert.True(t, mockExecutor.InitCompleted)

}

func TestLLMStepExecutor_Init_NoTemplateAndNoTemplateFileProvided(t *testing.T) {
	executor := LLMExecutor{}
	err := executor.Init()
	assert.Error(t, err)
}

func TestDecratedStepExecutor_Run(t *testing.T) {
	t.Run("pre-run and post-run are not called when no executor is provided", func(t *testing.T) {
		preRunExecuted := false
		postRunExecuted := false

		executor := &DecoratedExecutor{
			PreRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
				preRunExecuted = true
				return &flowContext, nil
			},
			PostRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
				postRunExecuted = true
				return &flowContext, nil
			},
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
		}
		step := &flow.Step{
			Executor: &MockStepExecutor{},
		}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err, "no executor provided")
		assert.False(t, preRunExecuted)
		assert.False(t, postRunExecuted)
	})
	t.Run("pre-run and post-run are called when executor is provided", func(t *testing.T) {
		preRunCalled := false
		postRunCalled := false
		executor := &DecoratedExecutor{
			ExecutorImpl: &MockStepExecutor{},
			PreRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
				preRunCalled = true
				return &flowContext, nil
			},
			PostRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
				postRunCalled = true
				return &flowContext, nil
			},
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
		}
		step := &flow.Step{
			Executor: &MockStepExecutor{},
		}
		_, err := executor.Run(flowContext, step)
		assert.Nil(t, err)
		assert.True(t, preRunCalled)
		assert.True(t, postRunCalled)
	})
}
func TestDecratedStepExecutor_Run_WithErrors(t *testing.T) {
	t.Run("pre-run returns an error", func(t *testing.T) {
		executor := &DecoratedExecutor{
			ExecutorImpl: &MockStepExecutor{},
			PreRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
				return nil, errors.New("error")
			},
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
		}
		step := &flow.Step{
			Executor: executor,
		}
		_, err := executor.Run(flowContext, step)
		assert.Equal(t, errors.New("error"), err)
	})
	t.Run("post-run returns an error", func(t *testing.T) {
		executor := &DecoratedExecutor{
			ExecutorImpl: &MockStepExecutor{},
			PostRun: func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
				return nil, errors.New("error")
			},
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
		}
		step := &flow.Step{
			Executor: executor,
		}
		_, err := executor.Run(flowContext, step)
		assert.Equal(t, errors.New("error"), err)
	})
	t.Run("executor.WithExecutor.Run returns an error", func(t *testing.T) {
		executor := &DecoratedExecutor{
			ExecutorImpl: &MockStepExecutor{
				RunWithError: true,
			},
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
		}
		step := &flow.Step{
			Executor: executor,
		}
		_, err := executor.Run(flowContext, step)
		assert.Equal(t, errors.New("error"), err)
	})
}
func TestLLMStepExecutor_Run(t *testing.T) {
	t.Run("step is nil", func(t *testing.T) {
		executor := &LLMExecutor{}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
			Flow: &flow.Flow{},
		}
		_, err := executor.Run(flowContext, nil)
		assert.Error(t, err)
	})
	t.Run("step.clientImpl is nil", func(t *testing.T) {
		executor := &LLMExecutor{}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
			Flow: &flow.Flow{},
		}
		step := &flow.Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})
	t.Run("template formatter is nil and template is provided", func(t *testing.T) {
		executor := &LLMExecutor{
			Template: "Hello, {{.name}}!",
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
			Flow: &flow.Flow{},
		}
		step := &flow.Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})
	t.Run("template formatter is nil and template file is provided", func(t *testing.T) {
		executor := &LLMExecutor{
			TemplateFile: "testdata/template.tmpl",
		}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
			Flow: &flow.Flow{},
		}
		step := &flow.Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})

}

func TestNewLLMStepWithTemplateString(t *testing.T) {
	step, err := NewLLMStepWithTemplate("Analyze this target and break it into action plans: {{.}}", "system_message", nil)

	assert.NoError(t, err)
	assert.NotNil(t, step)

	executor := step.Executor.(*LLMExecutor)

	assert.Equal(t, "system_message", executor.SystemMessage)

	formatter := executor.TemplateFormatter

	assert.NotNil(t, formatter)

	output, err := formatter.Format("Build an AI operating system")
	assert.Nil(t, err)

	assert.Equal(t, "Analyze this target and break it into action plans: Build an AI operating system", output)

}

func TestRunForLLMStep(t *testing.T) {
	t.Run("nil step", func(t *testing.T) {
		ctx := flow.FlowContext{
			Text: "input",
			Flow: &flow.Flow{

				ClientImpl: &test.MockClient{},
			},
		}
		_, err := (&LLMExecutor{}).Run(ctx, nil)
		assert.Error(t, err, "no step provided")
	})

	t.Run("no client set for flow step", func(t *testing.T) {
		step := flow.Step{}
		ctx := flow.FlowContext{
			Text: "input",
			Flow: &flow.Flow{

				ClientImpl: nil,
			},
		}
		_, err := (&LLMExecutor{}).Run(ctx, &step)
		assert.Error(t, err, "no client set for flow step")
	})
	t.Run("client chat error", func(t *testing.T) {

		step := flow.Step{
			ClientImpl: &test.MockClient{
				Err: errors.New("client chat error"),
			},
		}
		ctx := flow.FlowContext{}
		_, err := (&LLMExecutor{}).Run(ctx, &step)
		assert.Error(t, err, "client chat error")
	})
	t.Run("success", func(t *testing.T) {

		step := flow.Step{
			ClientImpl: &test.MockClient{
				ChatOutput: "output",
			},
		}
		ctx := flow.FlowContext{}
		newCtx, err := (&LLMExecutor{}).Run(ctx, &step)
		assert.Nil(t, err)
		assert.Equal(t, "output", newCtx.Text)
	})

	t.Run("template formatter success - with trim settings", func(t *testing.T) {
		templateFromatter, _ := chat.NewPromptTemplateFormatter(" Hello, {{.Memory}} \"")
		step := flow.Step{

			ClientImpl: &test.MockClient{},
		}
		ctx := flow.FlowContext{
			Memory: "world",
			Flow: &flow.Flow{
				ClientImpl: &test.MockClient{},
			},
		}
		executor := &LLMExecutor{
			TemplateFormatter: templateFromatter,
			Trim:              " \"",
		}
		output, err := executor.Run(ctx, &step)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, world", output.Text)
	})

	t.Run("template formatter success", func(t *testing.T) {
		templateFromatter, _ := chat.NewPromptTemplateFormatter("Hello, {{.Memory}}")
		step := flow.Step{

			ClientImpl: &test.MockClient{},
		}
		ctx := flow.FlowContext{
			Memory: "world",
			Flow: &flow.Flow{
				ClientImpl: &test.MockClient{},
			},
		}
		executor := &LLMExecutor{
			TemplateFormatter: templateFromatter,
		}
		output, err := executor.Run(ctx, &step)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, world", output.Text)
	})

	t.Run("template formatter with Variables", func(t *testing.T) {
		templateFromatter, _ := chat.NewPromptTemplateFormatter("Name: {{.Variables.name}}, Age: {{.Variables.age}}, Active: {{.Variables.active}}")
		step := flow.Step{
			ClientImpl: &test.MockClient{},
		}
		// Create FlowContext with variables
		ctx := flow.FlowContext{
			Variables: map[string]any{
				"name":   "John Doe",
				"age":    30,
				"active": true,
			},
			Flow: &flow.Flow{
				ClientImpl: &test.MockClient{},
			},
		}
		executor := &LLMExecutor{
			TemplateFormatter: templateFromatter,
		}
		output, err := executor.Run(ctx, &step)

		assert.NoError(t, err)
		assert.Equal(t, "Name: John Doe, Age: 30, Active: true", output.Text)
	})

	t.Run("template formatter with nested Variables", func(t *testing.T) {
		templateFromatter, _ := chat.NewPromptTemplateFormatter("User: {{.Variables.user.name}}, Theme: {{.Variables.settings.theme}}, Notifications: {{.Variables.settings.notifications}}")
		step := flow.Step{
			ClientImpl: &test.MockClient{},
		}
		// Create FlowContext with nested variables
		ctx := flow.FlowContext{
			Variables: map[string]any{
				"user": map[string]any{
					"name": "Alice",
					"id":   12345,
				},
				"settings": map[string]any{
					"theme":         "dark",
					"notifications": false,
				},
			},
			Flow: &flow.Flow{
				ClientImpl: &test.MockClient{},
			},
		}
		executor := &LLMExecutor{
			TemplateFormatter: templateFromatter,
		}
		output, err := executor.Run(ctx, &step)

		assert.NoError(t, err)
		assert.Equal(t, "User: Alice, Theme: dark, Notifications: false", output.Text)
	})

	t.Run("variables set by SetVariablesExecutor used in template", func(t *testing.T) {
		// Create a flow with SetVariablesExecutor followed by LLMExecutor

		// Step 1: SetVariablesExecutor to set variables
		setVarExecutor := &SetVariablesExecutor{
			Variables: map[string]any{
				"product": "Laptop",
				"specs": map[string]any{
					"cpu":  "Intel i7",
					"ram":  "16GB",
					"disk": "512GB SSD",
				},
				"price": 1299.99,
			},
		}

		// Step 2: LLMExecutor to use those variables in template
		templateFormatter, _ := chat.NewPromptTemplateFormatter(
			"Product: {{.Variables.product}}\n" +
				"CPU: {{.Variables.specs.cpu}}\n" +
				"RAM: {{.Variables.specs.ram}}\n" +
				"Disk: {{.Variables.specs.disk}}\n" +
				"Price: ${{.Variables.price}}")

		llmExecutor := &LLMExecutor{
			TemplateFormatter: templateFormatter,
		}

		// Create steps
		step1 := flow.Step{
			Name:     "Set Variables",
			Executor: setVarExecutor,
		}

		step2 := flow.Step{
			Name:       "Use Variables",
			Executor:   llmExecutor,
			ClientImpl: &test.MockClient{},
		}

		// Create flow
		testFlow, _ := flow.NewFlow(&test.MockClient{}, "Test Flow", step1, step2)

		// Run flow
		result, err := testFlow.RunWithInput("This input will be ignored")

		// Verify results
		assert.NoError(t, err)
		expectedOutput := "Product: Laptop\n" +
			"CPU: Intel i7\n" +
			"RAM: 16GB\n" +
			"Disk: 512GB SSD\n" +
			"Price: $1299.99"
		assert.Equal(t, expectedOutput, result.Text)
	})

	t.Run("template formatter error", func(t *testing.T) {
		templateFromatter, _ := chat.NewPromptTemplateFormatter("Hello, {{.None}}")
		step := flow.Step{
			ClientImpl: &test.MockClient{},
		}
		ctx := flow.FlowContext{
			Memory: "world",
		}
		executor := &LLMExecutor{
			TemplateFormatter: templateFromatter,
		}
		output, err := executor.Run(ctx, &step)
		assert.Error(t, err)
		assert.Nil(t, output)

	})
}

func TestNewLLMStepWithTemplateFile(t *testing.T) {

	step, err := NewLLMStepWithTemplateFile("./internal/test/test_prompt1.tmpl", "system_message", nil)

	assert.NoError(t, err)
	assert.NotNil(t, step)

	executor := step.Executor.(*LLMExecutor)
	assert.Equal(t, "system_message", executor.SystemMessage)

	formatter := executor.TemplateFormatter
	assert.NotNil(t, formatter)
	assert.Equal(t, "./internal/test/test_prompt1.tmpl", formatter.File)

	type AgentTasks struct {
		Tasks     []string
		Objective string
	}

	tasks := AgentTasks{
		Tasks:     []string{"task1", "task2"},
		Objective: "objective",
	}

	output, err := formatter.Format(tasks)
	assert.Nil(t, err)
	log.Printf("output: %s", output)

	assert.Greater(t, len(output), 10)

}

func TestConditionalFlowExecutor_Init(t *testing.T) {
	t.Run("should return error if switch is nil", func(t *testing.T) {
		executor := &ConditionalFlowExecutor{
			Switch: nil,
		}
		err := executor.Init()
		assert.Error(t, err)
	})
	t.Run("should return error if switch is an empty map", func(t *testing.T) {
		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{},
		}
		err := executor.Init()
		assert.Error(t, err)
	})
	t.Run("should return error if switch value is not existing flow", func(t *testing.T) {

		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{
				"foo": "bar",
			},
		}
		err := executor.Init()
		assert.Error(t, err)
	})
	t.Run("should return success if all switch values are existing flows", func(t *testing.T) {
		GlobalRegistry = &anyiRegistry{
			Flows: make(map[string]*flow.Flow),
		}
		RegisterFlow("bar", &flow.Flow{})
		RegisterFlow("qux", &flow.Flow{})

		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		}
		err := executor.Init()
		assert.NoError(t, err)
	})

	t.Run("should return success if all switch values contains nil values", func(t *testing.T) {

		GlobalRegistry = &anyiRegistry{
			Flows: make(map[string]*flow.Flow),
		}
		RegisterFlow("bar", &flow.Flow{})
		RegisterFlow("qux", nil)

		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		}
		err := executor.Init()
		assert.Error(t, err)
	})

	t.Run("should return error if some switch values are not valid", func(t *testing.T) {
		GlobalRegistry = &anyiRegistry{
			Flows: make(map[string]*flow.Flow),
		}
		RegisterFlow("bar", &flow.Flow{})

		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		}
		err := executor.Init()
		assert.Error(t, err)
	})
}

func TestConditionalFlowExecutor_Run(t *testing.T) {
	t.Run("WithMatchingCondition", func(t *testing.T) {
		RegisterFlow("flow1", &flow.Flow{
			Steps: []flow.Step{
				{
					Executor: &MockStepExecutor{
						ExpectedOutput: "bar",
					},
				},
			},
		})

		executor := ConditionalFlowExecutor{
			Switch: map[string]string{
				"foo": "flow1",
			},
		}
		flowContext := flow.FlowContext{
			Text: "foo",
		}
		step := &flow.Step{}
		context, err := executor.Run(flowContext, step)
		assert.Nil(t, err)
		assert.Equal(t, "bar", context.Text)
	})
	t.Run("WithNonMatchingCondition", func(t *testing.T) {
		RegisterFlow("flow1", &flow.Flow{
			Steps: []flow.Step{
				{
					Executor: &MockStepExecutor{
						ExpectedOutput: "bar",
					},
				},
			},
		})

		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{
				"goodbye": "flow1",
			},
		}
		flowContext := flow.FlowContext{
			Text: "foo",
		}
		step := &flow.Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no matching flow found for condition")
	})
	t.Run("WithNonExistingFlow", func(t *testing.T) {
		flowName := "invalid_flow"
		condition := "hello"
		executor := &ConditionalFlowExecutor{
			Switch: map[string]string{
				condition: flowName,
			},
		}
		flowContext := flow.FlowContext{
			Text: condition,
		}
		step := &flow.Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no flow found with the given name: invalid_flow")
	})
}

func TestRunCommandExecutor_Run_Success(t *testing.T) {
	executor := &RunCommandExecutor{}
	flowContext := flow.FlowContext{
		Text: "echo 'Hello, World!'",
	}
	step := &flow.Step{}
	result, err := executor.Run(flowContext, step)
	assert.Nil(t, err)
	assert.Equal(t, flowContext, *result)
}
func TestRunCommandExecutor_Run_EmptyCommand(t *testing.T) {
	executor := &RunCommandExecutor{}
	flowContext := flow.FlowContext{
		Text: "",
	}
	step := &flow.Step{}
	_, err := executor.Run(flowContext, step)
	assert.EqualError(t, err, "no command provided")
}
func TestRunCommandExecutor_Run_CommandError(t *testing.T) {
	executor := &RunCommandExecutor{}
	flowContext := flow.FlowContext{
		Text: "some-non-existing-command",
	}
	step := &flow.Step{}
	_, err := executor.Run(flowContext, step)
	assert.Error(t, err)
}
func TestRunCommandExecutor_Run_Silent(t *testing.T) {
	executor := &RunCommandExecutor{
		Silent: true,
	}
	flowContext := flow.FlowContext{
		Text: "echo 'Hello, World!'",
	}
	step := &flow.Step{}
	_, err := executor.Run(flowContext, step)
	assert.Nil(t, err)
}

func TestRunCommandExecutor_Run_OutputToContext(t *testing.T) {
	executor := RunCommandExecutor{OutputToContext: true}
	flowContext := flow.FlowContext{Text: "echo \"Hello, world!\""}
	step := &flow.Step{}
	resultFlowContext, err := executor.Run(flowContext, step)
	assert.Nil(t, err)

	assert.Equal(t, "Hello, world!", resultFlowContext.Text)
}

func TestSetContextExecutor_Run(t *testing.T) {
	// Test Case 1: Force is true, both Text and Memory are set
	t.Run("Force true, sets Text and Memory", func(t *testing.T) {
		executor := SetContextExecutor{
			Text: "Hello, World!",
			Memory: map[string]string{
				"key": "value",
			},

			Force: true,
		}
		flowContext := flow.FlowContext{
			Text: "Original Text",
			Memory: map[string]string{
				"foo": "bar",
			},
		}
		step := &flow.Step{}
		updatedContext, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.Equal(t, executor.Text, updatedContext.Text)
		memory := updatedContext.Memory.(map[string]string)
		assert.Equal(t, "value", memory["key"])
	})

	// Test Case 2: Force is false, Text is empty, Memory is set
	t.Run("Force false, Text empty, sets Memory", func(t *testing.T) {
		executor := SetContextExecutor{
			Text: "",
			Memory: map[string]string{
				"key": "value",
			},
			Force: false,
		}
		flowContext := flow.FlowContext{
			Text: "Original Text",
			Memory: map[string]string{
				"foo": "bar",
			},
		}
		step := &flow.Step{}
		updatedContext, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.Equal(t, "Original Text", updatedContext.Text)
		memory := updatedContext.Memory.(map[string]string)
		assert.Equal(t, "value", memory["key"])
	})

	// Test Case 3: Force is false, Memory is nil
	t.Run("Force false, Memory nil", func(t *testing.T) {
		executor := SetContextExecutor{
			Memory: nil,
			Force:  false,
		}
		flowContext := flow.FlowContext{
			Text: "Original Text",
			Memory: map[string]string{
				"foo": "bar",
			},
		}
		step := &flow.Step{}
		updatedContext, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.Equal(t, "Original Text", updatedContext.Text)
		memory := updatedContext.Memory.(map[string]string)
		assert.Equal(t, "bar", memory["foo"])
	})
}

func TestSetVariablesExecutor_Init(t *testing.T) {
	executor := &SetVariablesExecutor{}
	err := executor.Init()
	assert.NoError(t, err)
}

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
