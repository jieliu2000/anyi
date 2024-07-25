package anyi

import (
	"errors"
	"log"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
	"github.com/stretchr/testify/assert"
)

func TestNewFlow(t *testing.T) {

	client := test.MockClient{}
	flow := NewFlow(&client, "flow1")

	assert.NotNil(t, flow)
	assert.Equal(t, "flow1", flow.Name)
	assert.Equal(t, &client, flow.clientImpl)

}

func TestNewLLMStepWithTemplateFile(t *testing.T) {

	step, err := NewLLMStepWithTemplateFile("internal/test/test_prompt1.tmpl", "system_message", nil)

	assert.NoError(t, err)
	assert.NotNil(t, step)

	stepConfig := step.StepConfig.(LLMFlowStepConfig)
	assert.Equal(t, "system_message", stepConfig.SystemMessage)

	formatter := stepConfig.TemplateFormatter
	assert.NotNil(t, formatter)
	assert.Equal(t, "internal/test/test_prompt1.tmpl", formatter.File)

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

	stepConfig := step.StepConfig.(LLMFlowStepConfig)

	assert.Equal(t, "system_message", stepConfig.SystemMessage)

	formatter := stepConfig.TemplateFormatter

	assert.NotNil(t, formatter)

	output, err := formatter.Format("Build an AI operating system")
	assert.Nil(t, err)

	assert.Equal(t, "Analyze this target and break it into action plans: Build an AI operating system", output)

}

func Test_tryStep_RunError(t *testing.T) {
	step := NewStep(
		"",
		func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return nil, errors.New("run error")
		},
		func(output string, step *FlowStep) bool {
			return true
		},
		test.NewMockClient(),
	)
	_, err := tryStep(step, FlowContext{})
	assert.EqualError(t, err, "run error")
}
func Test_tryStep_RetryExceeded(t *testing.T) {

	step := NewStep(
		"",
		func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return &FlowContext{}, nil
		},
		func(output string, step *FlowStep) bool {
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
	step := NewStep(
		"",
		func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return &FlowContext{}, nil
		},
		func(output string, step *FlowStep) bool {
			return false
		},
		test.NewMockClient(),
	)
	_, err := tryStep(step, FlowContext{})
	assert.Error(t, err)
}
func Test_tryStep_ValidatorSuccess(t *testing.T) {
	step := NewStep(
		"",
		func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return &FlowContext{}, nil
		},
		func(output string, step *FlowStep) bool {
			return true
		},
		test.NewMockClient(),
	)
	result, err := tryStep(step, FlowContext{})
	assert.Nil(t, err)
	assert.Equal(t, result, &FlowContext{})
}

func TestFlow_Run(t *testing.T) {
	flow := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStep("Step 1", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			context.Data = "Data from Step 1"
			return &context, nil
		}, nil, nil),
		*NewStep("Step 2", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			context.Data = context.Data.(string) + " and Step 2"
			return &context, nil
		}, nil, nil),
	)
	flowContext, err := flow.Run(FlowContext{Context: "Initial"})
	assert.Nil(t, err)
	assert.Equal(t, "Initial", flowContext.Context)
	assert.Equal(t, "Data from Step 1 and Step 2", flowContext.Data)
}
func TestFlow_Run_WithInvalidStep(t *testing.T) {
	flow := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStep("Step 1", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			context.Data = "Data from Step 1"
			return &context, nil
		}, nil, nil),
		*NewStep("Step 2", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return nil, errors.New("Error in Step 2")
		}, nil, nil),
	)
	_, err := flow.Run(FlowContext{Context: "Initial"})
	assert.NotNil(t, err)
}
func TestFlow_Run_WithMaxRetryTimes(t *testing.T) {
	flow := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStep("Step 1", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return nil, errors.New("Error in Step 1")
		}, nil, nil),
		*NewStep("Step 2", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return nil, errors.New("Error in Step 2")
		}, nil, nil),
		*NewStep("Step 3", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			return nil, errors.New("Error in Step 3")
		}, nil, nil),
	)
	_, err := flow.Run(FlowContext{Context: "Initial"})
	assert.NotNil(t, err)
}

func TestFlow_Run_WithValidator_And_Validation_Passed(t *testing.T) {
	flow := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStep("Step 1", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			assert.Equal(t, "Initial", context.Context)
			context.Context = "Data from Step 1"
			return &context, nil
		}, func(stepOutput string, step *FlowStep) bool {
			return stepOutput == "Data from Step 1"
		}, nil),
		*NewStep("Step 2", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			assert.Equal(t, "Data from Step 1", context.Context)
			context.Context = "Data from Step 2"
			return &context, nil
		}, func(stepOutput string, step *FlowStep) bool {
			return stepOutput == "Data from Step 2"
		}, nil),
	)
	flowContext, err := flow.Run(FlowContext{Context: "Initial"})
	assert.NoError(t, err)
	assert.Equal(t, "Data from Step 2", flowContext.Context)
}
func TestFlow_Run_WithValidator_And_Validation_Failed(t *testing.T) {
	flow := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStep("Step 1", func(context FlowContext, step *FlowStep) (*FlowContext, error) {

			return &context, nil
		}, func(stepOutput string, step *FlowStep) bool {
			return stepOutput == "Data from Step 1"
		}, nil),
		*NewStep("Step 2", func(context FlowContext, step *FlowStep) (*FlowContext, error) {
			context.Data = "Data from Step 1"
			return &context, nil
		}, func(stepOutput string, step *FlowStep) bool {
			return stepOutput == "Data from Step 1 and Step 2"
		}, nil),
	)
	_, err := flow.Run(FlowContext{Context: "Initial"})
	assert.Error(t, err)
}
func TestFlow_Run_WithValidatorAndInvalidStep(t *testing.T) {
	flow := NewFlow(&test.MockClient{}, "Test Flow",
		*NewStep("Step 1",
			func(context FlowContext, step *FlowStep) (*FlowContext, error) {
				return &context, nil
			},
			func(stepOutput string, step *FlowStep) bool {
				return stepOutput == "Data from Step 1"
			}, nil),
		*NewStep("Step 2",
			func(context FlowContext, step *FlowStep) (*FlowContext, error) {
				return nil, errors.New("Error in Step 2")
			},
			func(stepOutput string, step *FlowStep) bool {
				return stepOutput == "Data from Step 1 and Step 2"
			}, nil),
	)
	_, err := flow.Run(FlowContext{Context: "Initial"})
	assert.NotNil(t, err)
}

func TestRunForLLMStep(t *testing.T) {
	t.Run("nil step", func(t *testing.T) {
		ctx := FlowContext{
			Context: "input",
			flow: &Flow{
				ClientConfig: llm.ClientConfig{
					Name: "mock",
				},
				clientImpl: &test.MockClient{},
			},
		}
		_, err := RunForLLMStep(ctx, nil)
		assert.Error(t, err, "no step provided")
	})
	t.Run("invalid step config type", func(t *testing.T) {
		step := FlowStep{
			StepConfig: struct{}{},
		}
		ctx := FlowContext{
			Context: "input",
			flow: &Flow{
				ClientConfig: llm.ClientConfig{
					Name: "mock",
				},
				clientImpl: &test.MockClient{},
			},
		}
		_, err := RunForLLMStep(ctx, &step)
		assert.Error(t, err, "invalid step config type")
	})
	t.Run("no client set for flow step", func(t *testing.T) {
		step := FlowStep{
			StepConfig: LLMFlowStepConfig{},
		}
		ctx := FlowContext{
			Context: "input",
			flow: &Flow{
				ClientConfig: llm.ClientConfig{
					Name: "mock",
				},
				clientImpl: nil,
			},
		}
		_, err := RunForLLMStep(ctx, &step)
		assert.Error(t, err, "no client set for flow step")
	})
	t.Run("client chat error", func(t *testing.T) {

		step := FlowStep{
			StepConfig: LLMFlowStepConfig{},
			clientImpl: &test.MockClient{
				Err: errors.New("client chat error"),
			},
		}
		ctx := FlowContext{}
		_, err := RunForLLMStep(ctx, &step)
		assert.Error(t, err, "client chat error")
	})
	t.Run("success", func(t *testing.T) {

		step := FlowStep{
			StepConfig: LLMFlowStepConfig{},
			clientImpl: &test.MockClient{
				ChatOutput: "output",
			},
		}
		ctx := FlowContext{}
		newCtx, err := RunForLLMStep(ctx, &step)
		assert.Nil(t, err)
		assert.Equal(t, "output", newCtx.Context)
	})

	t.Run("template formatter success", func(t *testing.T) {
		templateFromatter, _ := message.NewPromptTemplateFormatter("Hello, {{.Data}}")
		step := FlowStep{
			StepConfig: LLMFlowStepConfig{
				TemplateFormatter: templateFromatter,
			},
			clientImpl: &test.MockClient{},
		}
		ctx := FlowContext{
			Data: "world",
			flow: &Flow{
				ClientConfig: llm.ClientConfig{
					Name: "mock",
				},
				clientImpl: &test.MockClient{},
			},
		}
		output, err := RunForLLMStep(ctx, &step)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, world", output.Context)
	})

	t.Run("template formatter error", func(t *testing.T) {
		templateFromatter, _ := message.NewPromptTemplateFormatter("Hello, {{.None}}")
		step := FlowStep{
			StepConfig: LLMFlowStepConfig{
				TemplateFormatter: templateFromatter,
			},
			clientImpl: &test.MockClient{},
		}
		ctx := FlowContext{
			flow: &Flow{
				ClientConfig: llm.ClientConfig{
					Name: "mock",
				},
				clientImpl: &test.MockClient{},
			},
		}
		output, err := RunForLLMStep(ctx, &step)
		assert.Error(t, err)
		assert.Nil(t, output)

	})
}
