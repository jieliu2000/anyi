package flow

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"

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

type StepValidator interface {
	Init() error
	Validate(stepOutput string, Step *Step) bool
}

type StepExecutor interface {
	Init() error

	Run(context ShortTermMemory, Step *Step) (*ShortTermMemory, error)
}

type Step struct {
	clientImpl         llm.Client
	validateClientImpl llm.Client

	Executor StepExecutor

	Validator     StepValidator
	runTimes      int
	MaxRetryTimes int
}

// ShortTermMemory is the context for a flow. It will be passed to each flow step.
// The Context field provides the input and output string for a step. For example, before the step runs, the "Context" field is the input string (which might be formatted by the template formatter). After the step runs, the "Context" field is the output string.
// The Data field could be any data you want to pass between steps.
type ShortTermMemory struct {
	Context string
	Data    any
	flow    *Flow
}

func NewContext(context string, data any) *ShortTermMemory {
	return &ShortTermMemory{Context: context, Data: data}
}

func (flow *Flow) NewContext(context string, data any) *ShortTermMemory {
	return &ShortTermMemory{Context: context, Data: data, flow: flow}
}

func NewStep(executor StepExecutor, validator StepValidator, client llm.Client) *Step {
	return &Step{Executor: executor, Validator: validator, clientImpl: client}
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

func tryStep(step *Step, context ShortTermMemory) (*ShortTermMemory, error) {
	var err error

	// Run the step and get the updated context
	result, err := step.Executor.Run(context, step)
	step.runTimes++
	if err != nil {
		return result, err
	}
	if step.runTimes > step.MaxRetryTimes+1 {
		return result, errors.New("step retry times exceeded")
	}
	if step.Validator != nil {
		// Validate the step output
		if step.Validator.Validate(result.Context, step) {
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

func (flow *Flow) RunWithInput(input string) (*ShortTermMemory, error) {
	// Create a new context with the input
	context := ShortTermMemory{
		Context: input,
	}

	return flow.Run(context)
}

func (flow *Flow) Run(initialContext ShortTermMemory) (*ShortTermMemory, error) {

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

type LLMStepExecutor struct {
	Template          string `json:"template" yaml:"template" mapstructure:"template"`
	TemplateFile      string `json:"templateFile" yaml:"templateFile" mapstructure:"templateFile"`
	TemplateFormatter *chat.PromptyTemplateFormatter
	SystemMessage     string `json:"systemMessage" yaml:"systemMessage" mapstructure:"systemMessage"`
}

type LLMStepValidator struct {
	Template          string `json:"template" yaml:"template" mapstructure:"template"`
	TemplateFile      string `json:"templateFile" yaml:"templateFile" mapstructure:"templateFile"`
	TemplateFormatter *chat.PromptyTemplateFormatter
	SystemMessage     string `json:"systemMessage" yaml:"systemMessage" mapstructure:"systemMessage"`
}

func (executor *LLMStepExecutor) Init() error {
	if executor.TemplateFormatter == nil && executor.Template != "" {
		formatter, err := chat.NewPromptTemplateFormatter(executor.Template)
		if err != nil {
			return err
		}
		executor.TemplateFormatter = formatter
	}
	if executor.TemplateFormatter == nil && executor.TemplateFile != "" {
		formatter, err := chat.NewPromptTemplateFormatterFromFile(executor.TemplateFile)
		if err != nil {
			return err
		}
		executor.TemplateFormatter = formatter
	}
	return nil
}

func (executor *LLMStepExecutor) Run(context ShortTermMemory, step *Step) (*ShortTermMemory, error) {
	if step == nil {
		return nil, errors.New("no step provided")
	}

	if step.clientImpl == nil {
		step.clientImpl = context.flow.clientImpl
	}
	if step.clientImpl == nil {
		return nil, errors.New("no client set for flow step")
	}

	if executor.TemplateFormatter == nil && executor.Template != "" {
		var err error
		executor.TemplateFormatter, err = chat.NewPromptTemplateFormatter(executor.Template)
		if err != nil {
			return nil, err
		}
	}

	if executor.TemplateFormatter == nil && executor.TemplateFile != "" {
		var err error
		executor.TemplateFormatter, err = chat.NewPromptTemplateFormatterFromFile(executor.TemplateFile)
		if err != nil {
			return nil, err
		}
	}

	var input string
	if executor.TemplateFormatter != nil {
		var err error
		input, err = executor.TemplateFormatter.Format(context)
		if err != nil {
			return nil, err
		}
	} else {
		input = context.Context
	}

	messages := make([]chat.Message, 0, 2)
	if executor.SystemMessage != "" {
		messages = append(messages, chat.NewSystemMessage(executor.SystemMessage))
	}
	messages = append(messages, chat.NewUserMessage(input))

	output, _, err := step.clientImpl.Chat(messages, nil)
	if err != nil {
		return nil, err
	}

	context.Context = output.Content
	return &context, nil
}

func NewLLMStepWithTemplateFile(templateFilePath string, systemMessage string, client llm.Client) (*Step, error) {

	// Create a new formatter with the template
	formatter, err := chat.NewPromptTemplateFormatterFromFile(templateFilePath)
	if err != nil {
		return nil, err
	}
	executor := &LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := NewStep(executor, nil, client)

	return step, nil
}

func NewLLMStepWithTemplate(tmplate string, systemMessage string, client llm.Client) (*Step, error) {
	// Create a new formatter with the template
	formatter, err := chat.NewPromptTemplateFormatter(tmplate)
	if err != nil {
		return nil, err
	}

	executor := &LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := NewStep(executor, nil, client)
	return step, nil
}
