package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

type Flow struct {
	Name         string
	Steps        []FlowStep
	ClientConfig llm.ClientConfig
	// The default clientImpl for the flow
	clientImpl llm.Client
}

type StepExecutor func(context FlowContext, Step *FlowStep) (*FlowContext, error)

type StepValidator func(stepOutput string, Step *FlowStep) error

type FlowStep struct {
	clientImpl llm.Client

	StepConfig     any
	Run            StepExecutor
	Validate       StepValidator
	repeatTimes    int
	MaxRepeatTimes int
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
		StepConfig:     stepConfig,
		Run:            executor,
		Validate:       validator,
		repeatTimes:    0,
		MaxRepeatTimes: 1,
		clientImpl:     client,
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

// Run the flow step.
func RunForLLMStep(context FlowContext, step *FlowStep) (*FlowContext, error) {

	if step == nil {
		return nil, errors.New("no step provided")
	}

	llmStepConfig := step.StepConfig.(LLMFlowStepConfig)

	// Check if the client is set for the flow step
	if step.clientImpl == nil {
		step.clientImpl = context.flow.clientImpl
	}
	// Check if the client is set for the flow step
	if step.clientImpl == nil {
		return nil, errors.New("no client set for flow step")
	}

	// Get the template formatter for the step
	formatter := llmStepConfig.TemplateFormatter

	// Get the input from the flow context
	input := context.Context

	// Apply the formatter if it's not nil
	var err error
	if formatter != nil {
		input, err = formatter.Format(context)
		if err != nil {
			return nil, err
		}
	}

	// Initialize the message slice
	messages := []message.Message{}

	// If there is a system message set for the step, append it to the messages
	if llmStepConfig.SystemMessage != "" {
		messages = append(messages, message.NewSystemMessage(llmStepConfig.SystemMessage))
	}

	// Append a user message with the input to the messages
	messages = append(messages, message.NewUserMessage(input))

	// Send the messages to the chat client
	output, err := step.clientImpl.Chat(messages)

	// If there is an error during the chat, return it
	if err != nil {
		return nil, err
	}

	// Update the context with the chat output content
	context.Context = output.Content

	// Return the updated context and nil error
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
