package flow

import (
	"errors"
	"log"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/stretchr/testify/assert"
)

func TestNewFlow(t *testing.T) {

	client := test.MockClient{}
	flow, err := NewFlow(&client, "flow1")
	assert.NoError(t, err)

	assert.NotNil(t, flow)
	assert.Equal(t, "flow1", flow.Name)
	assert.Equal(t, &client, flow.clientImpl)

}

func TestNewStep(t *testing.T) {
	executor := &LLMStepExecutor{}
	validator := &StringValidator{}
	client := &test.MockClient{}

	step := NewStep(executor, validator, client)

	assert.NotNil(t, step)
	assert.Equal(t, executor, step.Executor)
	assert.Equal(t, validator, step.Validator)
	assert.Equal(t, client, step.clientImpl)
	assert.Equal(t, DefaultMaxRetryTimes, step.MaxRetryTimes)
}

func TestNewStepWithDefaultMaxRetry(t *testing.T) {
	executor := &LLMStepExecutor{}
	validator := &StringValidator{}
	client := &test.MockClient{}

	step := NewStep(executor, validator, client)

	assert.NotNil(t, step)
	assert.Equal(t, executor, step.Executor)
	assert.Equal(t, validator, step.Validator)
	assert.Equal(t, client, step.clientImpl)
	assert.Equal(t, DefaultMaxRetryTimes, step.MaxRetryTimes)
}

func TestNewStepWithCustomMaxRetry(t *testing.T) {
	executor := &LLMStepExecutor{}
	validator := &StringValidator{}
	client := &test.MockClient{}
	customMaxRetry := 5

	step := NewStep(executor, validator, client)
	step.MaxRetryTimes = customMaxRetry

	assert.NotNil(t, step)
	assert.Equal(t, executor, step.Executor)
	assert.Equal(t, validator, step.Validator)
	assert.Equal(t, client, step.clientImpl)
	assert.Equal(t, customMaxRetry, step.MaxRetryTimes)
}

func TestNewLLMStepWithTemplateFile(t *testing.T) {

	step, err := NewLLMStepWithTemplateFile("../internal/test/test_prompt1.tmpl", "system_message", nil)

	assert.NoError(t, err)
	assert.NotNil(t, step)

	executor := step.Executor.(*LLMStepExecutor)
	assert.Equal(t, "system_message", executor.SystemMessage)

	formatter := executor.TemplateFormatter
	assert.NotNil(t, formatter)
	assert.Equal(t, "../internal/test/test_prompt1.tmpl", formatter.File)

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

type MockExecutor struct {
	Mock func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error)
}

type MockValidator struct {
	Mock func(output string, step *Step) bool
}

func (m MockExecutor) Run(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
	return m.Mock(memory, step)
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

func NewStepWithValidatorAndExectorFunction(name string, runFunc func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error), validateFunc func(output string, step *Step) bool, client llm.Client) *Step {

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
		func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return nil, errors.New("run error")
		},
		func(output string, step *Step) bool {
			return true
		},
		test.NewMockClient(),
	)
	_, err := tryStep(step, ShortTermMemory{})
	assert.EqualError(t, err, "run error")
}
func Test_tryStep_RetryExceeded(t *testing.T) {

	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return &ShortTermMemory{}, nil
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
	_, err := tryStep(step, ShortTermMemory{})
	assert.EqualError(t, err, "step retry times exceeded")

	step.runTimes = 0
	step.MaxRetryTimes = 1
	_, err = tryStep(step, ShortTermMemory{})
	assert.EqualError(t, err, "step retry times exceeded")

	step.runTimes = 0
	step.MaxRetryTimes = 3
	_, err = tryStep(step, ShortTermMemory{})
	assert.NoError(t, err, "step should be success because retry times doesn't exceed MaxRetryTimes")
}
func Test_tryStep_ValidatorError(t *testing.T) {
	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return &ShortTermMemory{}, nil
		},
		func(output string, step *Step) bool {
			return false
		},
		test.NewMockClient(),
	)
	_, err := tryStep(step, ShortTermMemory{})
	assert.Error(t, err)
}
func Test_tryStep_ValidatorSuccess(t *testing.T) {
	step := NewStepWithValidatorAndExectorFunction(
		"",
		func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return &ShortTermMemory{}, nil
		},
		func(output string, step *Step) bool {
			return true
		},
		test.NewMockClient(),
	)
	result, err := tryStep(step, ShortTermMemory{})
	assert.Nil(t, err)
	assert.Equal(t, result, &ShortTermMemory{})
}

func TestFlow_Run(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			memory.NonTextData = "Data from Step 1"
			return &memory, nil
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			memory.NonTextData = memory.NonTextData.(string) + " and Step 2"
			return &memory, nil
		}, nil, nil),
	)
	assert.Nil(t, err)

	shortTermMemory, err := flow.Run(ShortTermMemory{Text: "Initial"})
	assert.Nil(t, err)
	assert.Equal(t, "Initial", shortTermMemory.Text)
	assert.Equal(t, "Data from Step 1 and Step 2", shortTermMemory.NonTextData)
}
func TestFlow_Run_WithInvalidStep(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			memory.NonTextData = "Data from Step 1"
			return &memory, nil
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return nil, errors.New("Error in Step 2")
		}, nil, nil),
	)

	assert.Nil(t, err)
	_, err = flow.Run(ShortTermMemory{Text: "Initial"})
	assert.NotNil(t, err)
}
func TestFlow_Run_WithMaxRetryTimes(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return nil, errors.New("Error in Step 1")
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return nil, errors.New("Error in Step 2")
		}, nil, nil),
		*NewStepWithValidatorAndExectorFunction("Step 3", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			return nil, errors.New("Error in Step 3")
		}, nil, nil),
	)

	assert.NoError(t, err)
	_, err = flow.Run(ShortTermMemory{Text: "Initial"})

	assert.NotNil(t, err)
}

func TestFlow_Run_WithValidator_And_Validation_Passed(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			assert.Equal(t, "Initial", memory.Text)
			memory.Text = "Data from Step 1"
			return &memory, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 1"
		}, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			assert.Equal(t, "Data from Step 1", memory.Text)
			memory.Text = "Data from Step 2"
			return &memory, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 2"
		}, nil),
	)

	assert.NoError(t, err)
	shortTermMemory, err := flow.Run(ShortTermMemory{Text: "Initial"})
	assert.NoError(t, err)
	assert.Equal(t, "Data from Step 2", shortTermMemory.Text)
}
func TestFlow_Run_WithValidator_And_Validation_Failed(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {

			return &memory, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 1"
		}, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2", func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
			memory.NonTextData = "Data from Step 1"
			return &memory, nil
		}, func(stepOutput string, step *Step) bool {
			return stepOutput == "Data from Step 1 and Step 2"
		}, nil),
	)

	assert.NoError(t, err)
	_, err = flow.Run(ShortTermMemory{Text: "Initial"})
	assert.Error(t, err)
}
func TestFlow_Run_WithValidatorAndInvalidStep(t *testing.T) {
	flow, err := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStepWithValidatorAndExectorFunction("Step 1",
			func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
				return &memory, nil
			},
			func(stepOutput string, step *Step) bool {
				return stepOutput == "Data from Step 1"
			}, nil),
		*NewStepWithValidatorAndExectorFunction("Step 2",
			func(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
				return nil, errors.New("Error in Step 2")
			},
			func(stepOutput string, step *Step) bool {
				return stepOutput == "Data from Step 1 and Step 2"
			}, nil),
	)

	assert.NoError(t, err)
	_, err = flow.Run(ShortTermMemory{Text: "Initial"})
	assert.NotNil(t, err)
}

func TestRunForLLMStep(t *testing.T) {
	t.Run("nil step", func(t *testing.T) {
		ctx := ShortTermMemory{
			Text: "input",
			flow: &Flow{

				clientImpl: &test.MockClient{},
			},
		}
		_, err := (&LLMStepExecutor{}).Run(ctx, nil)
		assert.Error(t, err, "no step provided")
	})

	t.Run("no client set for flow step", func(t *testing.T) {
		step := Step{}
		ctx := ShortTermMemory{
			Text: "input",
			flow: &Flow{

				clientImpl: nil,
			},
		}
		_, err := (&LLMStepExecutor{}).Run(ctx, &step)
		assert.Error(t, err, "no client set for flow step")
	})
	t.Run("client chat error", func(t *testing.T) {

		step := Step{
			clientImpl: &test.MockClient{
				Err: errors.New("client chat error"),
			},
		}
		ctx := ShortTermMemory{}
		_, err := (&LLMStepExecutor{}).Run(ctx, &step)
		assert.Error(t, err, "client chat error")
	})
	t.Run("success", func(t *testing.T) {

		step := Step{
			clientImpl: &test.MockClient{
				ChatOutput: "output",
			},
		}
		ctx := ShortTermMemory{}
		newCtx, err := (&LLMStepExecutor{}).Run(ctx, &step)
		assert.Nil(t, err)
		assert.Equal(t, "output", newCtx.Text)
	})

	t.Run("template formatter success", func(t *testing.T) {
		templateFromatter, _ := chat.NewPromptTemplateFormatter("Hello, {{.}}")
		step := Step{

			clientImpl: &test.MockClient{},
		}
		ctx := ShortTermMemory{
			NonTextData: "world",
			flow: &Flow{
				clientImpl: &test.MockClient{},
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
		step := Step{
			clientImpl: &test.MockClient{},
		}
		ctx := ShortTermMemory{
			NonTextData: "world",
		}
		executor := &LLMStepExecutor{
			TemplateFormatter: templateFromatter,
		}
		output, err := executor.Run(ctx, &step)
		assert.Error(t, err)
		assert.Nil(t, output)

	})
}
