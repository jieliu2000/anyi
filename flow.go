package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

const (
	DefaultMaxRetryTimes = 3
)

type Flow struct {
	Name string

	ClientName   string
	Steps        []FlowStep
	ClientConfig llm.ClientConfig
	// The default clientImpl for the flow
	clientImpl llm.Client
}

func NewFlow(client llm.Client, name string, steps ...FlowStep) *Flow {
	return &Flow{Steps: steps, Name: name, clientImpl: client}
}

type StepExecutor func(context FlowContext, Step *FlowStep) (*FlowContext, error)

type StepValidator func(stepOutput string, Step *FlowStep) bool

type FlowStep struct {
	clientImpl         llm.Client
	validateClientImpl llm.Client

	StepConfig any
	ClientName string

	Run StepExecutor

	// The client name which will be used to validate the step output. If not set, validator will use the default client of the step (which is identified by the ClientName field). If the step doesn't have a default client, the validator will use the default client of the flow.
	ValidatorClientName string
	Validate            StepValidator
	runTimes            int
	MaxRetryTimes       int
}

// FlowContext is the context for a flow. It will be passed to each flow step.
// The Context field provides the input and output string for a step. For example, before the step runs, the "Context" field is the input string (which might be formatted by the template formatter). After the step runs, the "Context" field is the output string.
// The Data field could be any data you want to pass between steps.
type FlowContext struct {
	Context string
	Data    any
	flow    *Flow
}

func NewStep(stepConfig any, executor StepExecutor, validator StepValidator, client llm.Client) *FlowStep {
	return &FlowStep{StepConfig: stepConfig, Run: executor, Validate: validator, clientImpl: client}
}

// Create a new flow step with executor and validator.
// Parameters:
//
//   - stepConfig: the configuration of the step. This parameter is used to pass parameters to the executor and validator. This can be a flexible type object which will be used by the executor and validator. We provide a default StepConfig in Anyi, see [LLMFlowStepConfig] for more details.
//   - executor: the executor of the step. See [StepExecutor] for more details.
//   - validator: the validator of the step. Set this parameter to nil if you don't want to validate the step output. See [StepValidator] for more details.
//   - client: the default client of the step. If set to nil, the default client of the flow will be used.
//   - validateClient: the client used to validate the step output. If set to nil, the default client of the step will be used. If the step doesn't have a default client, the default client of the flow will be used.
func NewStepWithValidator(stepConfig any, executor StepExecutor, validator StepValidator, client llm.Client, validateClient llm.Client) *FlowStep {
	return &FlowStep{
		StepConfig:         stepConfig,
		Run:                executor,
		Validate:           validator,
		runTimes:           0,
		MaxRetryTimes:      DefaultMaxRetryTimes,
		clientImpl:         client,
		validateClientImpl: validateClient,
	}
}

func tryStep(step *FlowStep, context FlowContext) (*FlowContext, error) {
	var err error

	// Run the step and get the updated context
	result, err := step.Run(context, step)
	step.runTimes++
	if err != nil {
		return result, err
	}
	if step.runTimes > step.MaxRetryTimes {
		return result, errors.New("step retry times exceeded")
	}
	if step.Validate != nil {
		// Validate the step output
		if step.Validate(result.Context, step) {
			// If the step output is valid, update context and continue to the next step
			return result, nil
		} else {
			// Otherwise, try again
			return tryStep(step, *result)
		}
	}
	// If no validator is set, simply return the updated context.
	return result, nil
}

func (flow *Flow) Run(initialContext FlowContext) (*FlowContext, error) {

	context := &initialContext
	context.flow = flow

	// For each step in the flow
	for _, step := range flow.Steps {
		// Run the step and get the updated context

		result, err := tryStep(&step, *context)

		if err != nil {
			return nil, err
		}

		// Update the context
		context = result
	}

	// Return the context content
	return context, nil
}

type LLMFlowStepConfig struct {
	TemplateFormatter       *message.PromptyTemplateFormatter
	SystemMessage           string
	ValidatePromptFormatter *message.PromptyTemplateFormatter
	ValidateSystemMessage   string
}

func RunForLLMStep(context FlowContext, step *FlowStep) (*FlowContext, error) {
	if step == nil {
		return nil, errors.New("no step provided")
	}

	llmStepConfig, ok := step.StepConfig.(LLMFlowStepConfig)
	if !ok {
		return nil, errors.New("invalid step config type")
	}

	if step.clientImpl == nil {
		step.clientImpl = context.flow.clientImpl
	}
	if step.clientImpl == nil {
		return nil, errors.New("no client set for flow step")
	}

	var input string
	if llmStepConfig.TemplateFormatter != nil {
		var err error
		input, err = llmStepConfig.TemplateFormatter.Format(context)
		if err != nil {
			return nil, err
		}
	} else {
		input = context.Context
	}

	messages := make([]message.Message, 0, 2)
	if llmStepConfig.SystemMessage != "" {
		messages = append(messages, message.NewSystemMessage(llmStepConfig.SystemMessage))
	}
	messages = append(messages, message.NewUserMessage(input))

	output, err := step.clientImpl.Chat(messages)
	if err != nil {
		return nil, err
	}

	context.Context = output.Content
	return &context, nil
}

func NewLLMStepWithTemplateFile(templateFilePath string, systemMessage string, client llm.Client) (*FlowStep, error) {

	// Create a new formatter with the template
	formatter, err := message.NewPromptTemplateFormatterFromFile(templateFilePath)
	if err != nil {
		return nil, err
	}

	// Return the step config
	stepConfig := LLMFlowStepConfig{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}

	step := NewStep(stepConfig, RunForLLMStep, nil, client)

	return step, nil
}
