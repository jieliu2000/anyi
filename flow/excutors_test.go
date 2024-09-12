package flow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockStepExecutor struct {
	RunWithError bool
}

func (executor *MockStepExecutor) Init() error {
	return nil
}

func (executor *MockStepExecutor) Run(flowContext FlowContext, step *Step) (*FlowContext, error) {
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
		TemplateFile: "../internal/test/test_prompt2.tmpl",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
}

func TestDecratedStepExecutor_Init_NoExecutorProvided(t *testing.T) {
	executor := DecratedStepExecutor{}
	err := executor.Init()
	assert.Error(t, err)
}
func TestDecratedStepExecutor_Init_NoPreOrPostRunProvided(t *testing.T) {
	executor := DecratedStepExecutor{
		WithExecutor: nil,
	}
	err := executor.Init()
	assert.Error(t, err)
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

		executor := &DecratedStepExecutor{
			PreRun: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				preRunExecuted = true
				return &flowContext, nil
			},
			PostRun: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				postRunExecuted = true
				return &flowContext, nil
			},
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
		}
		step := &Step{
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
		executor := &DecratedStepExecutor{
			WithExecutor: &MockStepExecutor{},
			PreRun: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				preRunCalled = true
				return &flowContext, nil
			},
			PostRun: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				postRunCalled = true
				return &flowContext, nil
			},
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
		}
		step := &Step{
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
		executor := &DecratedStepExecutor{
			WithExecutor: &MockStepExecutor{},
			PreRun: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				return nil, errors.New("error")
			},
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
		}
		step := &Step{
			Executor: executor,
		}
		_, err := executor.Run(flowContext, step)
		assert.Equal(t, errors.New("error"), err)
	})
	t.Run("post-run returns an error", func(t *testing.T) {
		executor := &DecratedStepExecutor{
			WithExecutor: &MockStepExecutor{},
			PostRun: func(flowContext FlowContext, step *Step) (*FlowContext, error) {
				return nil, errors.New("error")
			},
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
		}
		step := &Step{
			Executor: executor,
		}
		_, err := executor.Run(flowContext, step)
		assert.Equal(t, errors.New("error"), err)
	})
	t.Run("executor.WithExecutor.Run returns an error", func(t *testing.T) {
		executor := &DecratedStepExecutor{
			WithExecutor: &MockStepExecutor{
				RunWithError: true,
			},
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
		}
		step := &Step{
			Executor: executor,
		}
		_, err := executor.Run(flowContext, step)
		assert.Equal(t, errors.New("error"), err)
	})
}
func TestLLMStepExecutor_Run(t *testing.T) {
	t.Run("step is nil", func(t *testing.T) {
		executor := &LLMStepExecutor{}
		flowContext := FlowContext{
			Text: "Hello, World!",
			flow: &Flow{},
		}
		_, err := executor.Run(flowContext, nil)
		assert.Error(t, err)
	})
	t.Run("step.clientImpl is nil", func(t *testing.T) {
		executor := &LLMStepExecutor{}
		flowContext := FlowContext{
			Text: "Hello, World!",
			flow: &Flow{},
		}
		step := &Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})
	t.Run("template formatter is nil and template is provided", func(t *testing.T) {
		executor := &LLMStepExecutor{
			Template: "Hello, {{.name}}!",
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
			flow: &Flow{},
		}
		step := &Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})
	t.Run("template formatter is nil and template file is provided", func(t *testing.T) {
		executor := &LLMStepExecutor{
			TemplateFile: "testdata/template.tmpl",
		}
		flowContext := FlowContext{
			Text: "Hello, World!",
			flow: &Flow{},
		}
		step := &Step{}
		_, err := executor.Run(flowContext, step)
		assert.Error(t, err)
	})

}
