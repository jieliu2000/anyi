package openai

import (
	"github.com/jieliu2000/anyi/message"
	impl "github.com/sashabaranov/go-openai"
)

func ConvertToOpenAIChatMessages(messages []message.Message) []impl.ChatCompletionMessage {
	result := []impl.ChatCompletionMessage{}
	for _, msg := range messages {
		openaiMessage := impl.ChatCompletionMessage{
			Content: msg.Content,
			Role:    msg.Role,
		}
		result = append(result, openaiMessage)
	}
	return result
}
