package anyi

import (
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

func NewClient(config llm.ModelConfig) (llm.Client, error) {

	return llm.NewClient(config)
}

func NewMessage(role string, content string) message.Message {
	return message.Message{
		Role:    role,
		Content: content,
	}
}
