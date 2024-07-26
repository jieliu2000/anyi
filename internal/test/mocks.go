package test

import (
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/message"
)

type MockClient struct {
	ChatOutput string
	Err        error
}

func (m *MockClient) Chat(messages []message.Message) (*message.Message, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.ChatOutput != "" {
		m := message.NewAssistantMessage(m.ChatOutput)
		return &m, nil
	}
	if len(messages) > 0 {
		m := message.NewAssistantMessage(messages[0].Content)
		return &m, nil
	}
	return nil, nil
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

type MockExecutor struct {
}

func (m *MockExecutor) Run(context flow.FlowContext, Step *flow.Step) (*flow.FlowContext, error) {

	return &context, nil
}

type MockValidator struct {
}

func (m *MockValidator) Validate(stepOutput string, Step *flow.Step) bool {

	return true
}
