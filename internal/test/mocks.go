package test

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

type MockClient struct {
	ChatOutput string
	Err        error
}

func (c *MockClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	return nil, chat.ResponseInfo{}, errors.New("not implemented")
}

func (m *MockClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	info := chat.ResponseInfo{}

	if m.Err != nil {
		return nil, info, m.Err
	}
	if m.ChatOutput != "" {
		m := chat.NewAssistantMessage(m.ChatOutput)
		return &m, info, nil
	}
	if len(messages) > 0 {
		m := chat.NewAssistantMessage(messages[0].Content)
		return &m, info, nil
	}
	return nil, info, nil
}

func NewMockClient() *MockClient {
	return &MockClient{}
}
