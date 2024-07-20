package test

import "github.com/jieliu2000/anyi/message"

type MockClient struct {
}

func (m *MockClient) Chat(messages []message.Message) (*message.Message, error) {
	return nil, nil
}
