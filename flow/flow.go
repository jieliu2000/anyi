package flow

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
)

const (
	DefaultMaxRetryTimes = 3
)

type Flow struct {
	Name string

	Steps []Step
	// The default clientImpl for the flow
	clientImpl llm.Client
}

func NewFlow(client llm.Client, name string, steps ...Step) (*Flow, error) {

	if name == "" {
		return nil, errors.New("flow name cannot be empty")
	}

	flow := &Flow{Steps: steps, Name: name, clientImpl: client}

	return flow, nil
}

type Step struct {
	clientImpl         llm.Client
	validateClientImpl llm.Client

	Executor StepExecutor

	Validator     StepValidator
	runTimes      int
	MaxRetryTimes int
}

// ShortTermMemory is the memory for a flow. It will be passed to each flow step.
// The ShortTermMemory field provides the input and output string for a step. For example, before the step runs, the "ShortTermMemory" field is the input string (which might be formatted by the template formatter). After the step runs, the "ShortTermMemory" field is the output string.
// The Data field could be any data you want to pass between steps.
type ShortTermMemory struct {
	Text        string
	NonTextData any
	flow        *Flow
}

func NewShortTermMemory(memory string, data any) *ShortTermMemory {
	return &ShortTermMemory{Text: memory, NonTextData: data}
}

func (flow *Flow) NewShortTermMemory(memory string, data any) *ShortTermMemory {
	return &ShortTermMemory{Text: memory, NonTextData: data, flow: flow}
}

func NewStep(executor StepExecutor, validator StepValidator, client llm.Client) *Step {
	return &Step{Executor: executor, Validator: validator, clientImpl: client, MaxRetryTimes: DefaultMaxRetryTimes}
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
		clientImpl:         client,
		validateClientImpl: validateClient,
	}
}

func tryStep(step *Step, memory ShortTermMemory) (*ShortTermMemory, error) {
	var err error

	// Run the step and get the updated memory
	result, err := step.Executor.Run(memory, step)
	step.runTimes++
	if err != nil {
		return result, err
	}
	if step.runTimes > step.MaxRetryTimes+1 {
		return result, errors.New("step retry times exceeded")
	}
	if step.Validator != nil {
		// Validate the step output
		if step.Validator.Validate(result.Text, step) {
			// If the step output is valid, update memory and continue to the next step
			return result, nil
		} else {
			// Otherwise, try again
			return tryStep(step, *result)
		}
	}
	// If no validator is set, simply return the updated context.
	return result, nil
}

func (flow *Flow) RunWithInput(input string) (*ShortTermMemory, error) {
	// Create a new memory with the input
	memory := ShortTermMemory{
		Text: input,
	}

	return flow.Run(memory)
}

func (flow *Flow) Run(initialShortTermMemory ShortTermMemory) (*ShortTermMemory, error) {

	memory := &initialShortTermMemory
	memory.flow = flow

	// For each step in the flow
	for _, step := range flow.Steps {
		// Run the step and get the updated memory

		result, err := tryStep(&step, *memory)

		if err != nil {
			return nil, err
		}

		// Update the memory
		memory = result
	}

	// Return the memory content
	return memory, nil
}
