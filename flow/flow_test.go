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
