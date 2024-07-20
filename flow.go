package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
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

type StepValidator func(stepOutput string, Step *FlowStep) error

type FlowStep struct {
	clientImpl llm.Client

	ClientName    string
	StepConfig    any
	Run           StepExecutor
	Validate      StepValidator
	retryTimes    int
	MaxRetryTimes int
}

// FlowContext is the context for a flow. It will be passed to each flow step.
// The @Context field provides the input and output string for a step. For example, before the step runs, the "Context" field is the input string (which might be formatted by the template formatter). After the step runs, the "Context" field is the output string.
// The @Data field could be any data you want to pass between steps.
type FlowContext struct {
	Context string
	Data    any
	flow    *Flow
}

func NewStep(stepConfig any, executor StepExecutor, validator StepValidator, client llm.Client) *FlowStep {
	return &FlowStep{
		StepConfig:    stepConfig,
		Run:           executor,
		Validate:      validator,
		retryTimes:    0,
		MaxRetryTimes: 1,
		clientImpl:    client,
	}
}

func (flow *Flow) Run(initialContext FlowContext) (*FlowContext, error) {

	context := initialContext
	context.flow = flow

	// For each step in the flow
	for _, step := range flow.Steps {
		// Run the step and get the updated context

		var err error
		result, err := step.Run(context, &step)
		if err != nil {
			return result, err
		}
	}

	// Return the context content
	return &context, nil
}

type LLMFlowStepConfig struct {
	// The client of the flow step. If not set, the default client of the flow will be used.
	TemplateFormatter *message.MessageTemplateFormatter
	SystemMessage     string
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
	formatter, err := message.NewMessageTemplateFormatterFromFile(templateFilePath)
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
