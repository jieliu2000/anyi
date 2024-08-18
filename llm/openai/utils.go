package openai

import (
	"github.com/jieliu2000/anyi/message"
	impl "github.com/sashabaranov/go-openai"
)

func ConvertToOpenAIChatMessages(messages []message.Message) []impl.ChatCompletionMessage {
	result := []impl.ChatCompletionMessage{}
	for _, msg := range messages {
		openaiMessage := convertToOpenAIChatMessage(msg)
		result = append(result, openaiMessage)
	}
	return result
}

func convertToOpenAIChatMessage(msg message.Message) impl.ChatCompletionMessage {
	result := impl.ChatCompletionMessage{
		Role: msg.Role,
	}
	if msg.Content != "" {
		result.Content = msg.Content
		return result
	}
	if len(msg.ContentParts) > 0 {
		messageParts := []impl.ChatMessagePart{}
		for _, img := range msg.ContentParts {

			if img.Text != "" {
				textPart := impl.ChatMessagePart{
					Type: impl.ChatMessagePartTypeText,
					Text: img.Text,
				}

				messageParts = append(messageParts, textPart)
			} else if img.ImageUrl != "" {
				imagePart := impl.ChatMessagePart{
					Type: impl.ChatMessagePartTypeImageURL,
					ImageURL: &impl.ChatMessageImageURL{
						URL: img.ImageUrl,
					},
				}
				messageParts = append(messageParts, imagePart)
			}
		}
		result.MultiContent = messageParts
	}
	return result
}
