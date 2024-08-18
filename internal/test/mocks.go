package test

import "github.com/jieliu2000/anyi/chat"

type MockClient struct {
	ChatOutput string
	Err        error
}

func (m *MockClient) Chat(messages []chat.Message, options chat.ChatOptions) (*chat.Message, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.ChatOutput != "" {
		m := chat.NewAssistantMessage(m.ChatOutput)
		return &m, nil
	}
	if len(messages) > 0 {
		m := chat.NewAssistantMessage(messages[0].Content)
		return &m, nil
	}
	return nil, nil
}

func NewMockClient() *MockClient {
	return &MockClient{}
}
