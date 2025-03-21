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

	t.Run("should return success if all switch values are existing flows", func(t *testing.T) {

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
