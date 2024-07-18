package anyi

import (
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

type Flow struct {
	Steps        []FlowStep
	ClientConfig llm.ClientConfig
	// The default client for the flow
	client llm.Client
}

type FlowStep interface {
	Execute(context any) (any, error)
}

type LLMFlowStep struct {
	// The client of the flow step. If not set, the default client of the flow will be used.
	Client            llm.Client
	TemplateFormatter message.MessageTemplateFormatter
}

func (step LLMFlowStep) Execute(context any) (any, error) {

	return nil, nil
}
