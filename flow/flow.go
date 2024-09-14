package flow

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/llm"
)

const (
	DefaultMaxRetryTimes = 3
)

type Flow struct {
	Name string

	Steps []Step
	// The default ClientImpl for the flow
	ClientImpl llm.Client
}
type StepExecutor interface {
	Init() error
	Run(flowContext FlowContext, Step *Step) (*FlowContext, error)
}

// StepValidator is the interface for validators of step output.
// In a flow if a step validator is set, the output of the step will be checked against the validator's Validate method.
type StepValidator interface {
	Init() error
	Validate(stepOutput string, Step *Step) bool
}

func NewFlow(client llm.Client, name string, steps ...Step) (*Flow, error) {

	if name == "" {
		return nil, errors.New("flow name cannot be empty")
	}

	flow := &Flow{Steps: steps, Name: name, ClientImpl: client}

	return flow, nil
}

type Step struct {
	ClientImpl         llm.Client
	validateClientImpl llm.Client

	Executor StepExecutor

	Validator     StepValidator
	runTimes      int
	MaxRetryTimes int
	Name          string
}

// GetClient function returns the client of the Step.
// If the clientImpl of the Step is not nil, it returns the clientImpl.
// Otherwise, it returns nil.
func (step *Step) GetClient() llm.Client {
	if step.ClientImpl != nil {
		return step.ClientImpl
	}
	return nil
}

type ShortTermMemory any

// FlowContext is the flowContext for a flow. It will be passed to each flow step.
// The Text field provides the input and output string for a step. For example, before the step runs, the "Text" field is the input string (which might be formatted by the template formatter). After the step runs, the "Text" field is the output string.
// The Data field could be any data you want to pass between steps.
type FlowContext struct {
	Text   string
	Memory ShortTermMemory
	Flow   *Flow
}

func (fc *FlowContext) UnmarshalJsonText(entity any) error {
	return json.Unmarshal([]byte(fc.Text), entity)
}

func NewFlowContext(flowContext string, data any) *FlowContext {
	return &FlowContext{Text: flowContext, Memory: data}
}

func (flow *Flow) NewFlowContext(flowContext string, data any) *FlowContext {
	return &FlowContext{Text: flowContext, Memory: data, Flow: flow}
}

func NewStep(executor StepExecutor, validator StepValidator, client llm.Client) *Step {
	return &Step{Executor: executor, Validator: validator, ClientImpl: client, MaxRetryTimes: DefaultMaxRetryTimes}
}

// Create a new flow step with executor and validator.
// Parameters:
//
//   - stepConfig: the configuration of the step. This parameter is used to pass parameters to the executor and validator. This can be a flexible type object which will be used by the executor and validator. We provide a default StepConfig in Anyi, see [LLMFlowStepConfig] for more details.
//   - executor: the executor of the step. See [StepExecutor] for more details.
//   - validator: the validator of the step. Set this parameter to nil if you don't want to validate the step output. See [StepValidator] for more details.
//   - client: the default client of the step. If set to nil, the default client of the flow will be used.
//   - validateClient: the client used to validate the step output. If set to nil, the default client of the step will be used. If the step doesn't have a default client, the default client of the flow will be used.
func NewStepWithValidator(stepConfig any, executor StepExecutor, validator StepValidator, client llm.Client, validateClient llm.Client) *Step {
	return &Step{
		Executor:           executor,
		Validator:          validator,
		runTimes:           0,
		MaxRetryTimes:      DefaultMaxRetryTimes,
		ClientImpl:         client,
		validateClientImpl: validateClient,
	}
}

func tryStep(step *Step, flowContext FlowContext) (*FlowContext, error) {
	var err error

	log.Debug("Running step ", step, " with flowContext:", flowContext, ".")
	// Run the step and get the updated flowContext
	result, err := step.Executor.Run(flowContext, step)
	step.runTimes++
	if err != nil {
		return result, err
	}
	if step.runTimes > step.MaxRetryTimes+1 {
		log.Error("Step retry times exceeded, returning error.")
		return result, errors.New("step retry times exceeded")
	}
	if step.Validator != nil {
		// Validate the step output
		if step.Validator.Validate(result.Text, step) {
			// If the step output is valid, update flowContext and continue to the next step
			return result, nil
		} else {
			// Otherwise, try again
			return tryStep(step, *result)
		}
	}
	// If no validator is set, simply return the updated context.
	return result, nil
}

func (flow *Flow) RunWithInput(input string) (*FlowContext, error) {
	// Create a new flowContext with the input
	flowContext := FlowContext{
		Text: input,
	}

	return flow.Run(flowContext)
}

func (flow *Flow) Run(initialFlowContext FlowContext) (*FlowContext, error) {

	flowContext := &initialFlowContext
	flowContext.Flow = flow

	log.Debug("Starting run flow ", flow.Name, " with initial context: ", flowContext, ".")
	// For each step in the flow
	for _, step := range flow.Steps {
		// Run the step and get the updated flowContext

		result, err := tryStep(&step, *flowContext)

		log.Debug("Step running finished. Result: ", result, ". Error:", err, ".")
		if err != nil {
			return nil, err
		}

		// Update the flowContext
		flowContext = result
	}

	// Return the flowContext content
	return flowContext, nil
}
