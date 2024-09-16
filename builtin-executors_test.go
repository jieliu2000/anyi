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
	RunWithError  bool
	InitCompleted bool
	RunCompleted  bool
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
	return &flowContext, nil
}

func TestLLMStepExecutor_Init(t *testing.T) {
	executor := &LLMStepExecutor{}
	err := executor.Init()
	assert.Error(t, err)
	executor = &LLMStepExecutor{
		Template: "Hello, {{.name}}!",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
	assert.Equal(t, "Hello, {{.name}}!", executor.TemplateFormatter.TemplateString)

	executor = &LLMStepExecutor{
		TemplateFile: "./internal/test/test_prompt2.tmpl",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
}

func TestDecratedStepExecutor_Init_NoExecutorProvided(t *testing.T) {
	executor := DecoratedStepExecutor{}
	err := executor.Init()
	assert.Error(t, err)
}
func TestDecratedStepExecutor_Init_NoPreOrPostRunProvided(t *testing.T) {
	executor := DecoratedStepExecutor{
		ExecutorImpl: &DecoratedStepExecutor{},
	}
	err := executor.Init()
	assert.Error(t, err)
	assert.EqualError(t, err, "no pre or post run function provided")
}

func TestDecratedStepExecutor_Init_NoExecutor(t *testing.T) {
	executor := DecoratedStepExecutor{}
	err := executor.Init()
	assert.Error(t, err)
	assert.EqualError(t, err, "no executor provided")
}
func TestDecratedStepExecutor_Init(t *testing.T) {
	mockExecutor := &MockStepExecutor{}
	executor := DecoratedStepExecutor{
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
	executor := LLMStepExecutor{}
	err := executor.Init()
	assert.Error(t, err)
}

func TestDecratedStepExecutor_Run(t *testing.T) {
	t.Run("pre-run and post-run are not called when no executor is provided", func(t *testing.T) {
		preRunExecuted := false
		postRunExecuted := false

		executor := &DecoratedStepExecutor{
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
		executor := &DecoratedStepExecutor{
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
		executor := &DecoratedStepExecutor{
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
		executor := &DecoratedStepExecutor{
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
		executor := &DecoratedStepExecutor{
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
		executor := &LLMStepExecutor{}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
			Flow: &flow.Flow{},
		}
		_, err := executor.Run(flowContext, nil)
		assert.Error(t, err)
	})
	t.Run("step.clientImpl is nil", func(t *testing.T) {
		executor := &LLMStepExecutor{}
		flowContext := flow.FlowContext{
			Text: "Hello, World!",
			Flow: &flow.Flow{},
		}
		step := &flow.Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})
	t.Run("template formatter is nil and template is provided", func(t *testing.T) {
		executor := &LLMStepExecutor{
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
		executor := &LLMStepExecutor{
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

	executor := step.Executor.(*LLMStepExecutor)

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
		_, err := (&LLMStepExecutor{}).Run(ctx, nil)
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
		_, err := (&LLMStepExecutor{}).Run(ctx, &step)
		assert.Error(t, err, "no client set for flow step")
	})
	t.Run("client chat error", func(t *testing.T) {

		step := flow.Step{
			ClientImpl: &test.MockClient{
				Err: errors.New("client chat error"),
			},
		}
		ctx := flow.FlowContext{}
		_, err := (&LLMStepExecutor{}).Run(ctx, &step)
		assert.Error(t, err, "client chat error")
	})
	t.Run("success", func(t *testing.T) {

		step := flow.Step{
			ClientImpl: &test.MockClient{
				ChatOutput: "output",
			},
		}
		ctx := flow.FlowContext{}
		newCtx, err := (&LLMStepExecutor{}).Run(ctx, &step)
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
		executor := &LLMStepExecutor{
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
		executor := &LLMStepExecutor{
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
		executor := &LLMStepExecutor{
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

	executor := step.Executor.(*LLMStepExecutor)
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
