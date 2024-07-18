package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

type Flow struct {
	Steps        []FlowStep
	ClientConfig llm.ClientConfig
	// The default clientImpl for the flow
	clientImpl llm.Client
}

type FlowStep interface {
	Run(context *FlowContext) (*FlowContext, error)
}

// FlowContext is the context for a flow. It will be passed to each flow step.
// The @Context field provides the input and output string for a step. For example, before the step runs, the "Context" field is the input string (which might be formatted by the template formatter). After the step runs, the "Context" field is the output string.
// The @Data field could be any data you want to pass between steps.
type FlowContext struct {
	Context string
	Data    any
	flow    *Flow
}

type LLMFlowStep struct {
	// The client of the flow step. If not set, the default client of the flow will be used.
	Client            llm.Client
	TemplateFormatter *message.MessageTemplateFormatter
	SystemMessage     string
}

func NewFlow(client llm.Client, steps ...FlowStep) *Flow {
	return &Flow{Steps: steps, clientImpl: client}
}

func (flow *Flow) Run(initialContext *FlowContext) (*FlowContext, error) {

	context := initialContext
	context.flow = flow

	// For each step in the flow
	for _, step := range flow.Steps {
		// Run the step and get the updated context

		var err error
		context, err = step.Run(context)
		if err != nil {
			return context, err
		}
	}

	// Return the context content
	return context, nil
}

// Run the flow step.
func (step *LLMFlowStep) Run(context *FlowContext) (*FlowContext, error) {
	// Check if the client is set for the flow step
	if step.Client == nil {
		step.Client = context.flow.clientImpl
	}

	// Check if the client is set for the flow step
	if step.Client == nil {
		return nil, errors.New("no client set for flow step")
	}

	// Get the template formatter for the step
	formatter := step.TemplateFormatter

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
	if step.SystemMessage != "" {
		messages = append(messages, message.NewSystemMessage(step.SystemMessage))
	}

	// Append a user message with the input to the messages
	messages = append(messages, message.NewUserMessage(input))

	// Send the messages to the chat client
	output, err := step.Client.Chat(messages)

	// If there is an error during the chat, return it
	if err != nil {
		return nil, err
	}

	// Update the context with the chat output content
	context.Context = output.Content

	// Return the updated context and nil error
	return context, nil
}
