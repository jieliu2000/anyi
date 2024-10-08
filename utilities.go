package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
)

func SimpleChat(input string) (string, error) {
	if input == "" {
		return "", errors.New("empty input")
	}
	client, err := GetDefaultClient()
	if err != nil {
		return "", err
	}

	result, _, err := client.Chat([]chat.Message{
		chat.NewUserMessage(input),
	}, nil)

	if err != nil {
		return "", err
	}
	return result.Content, nil
}
