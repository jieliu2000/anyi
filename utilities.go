package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/message"
)

func SimpleChat(input string) (string, error) {
	if input == "" {
		return "", errors.New("empty input")
	}
	client, err := GetDefaultClient()
	if err != nil {
		return "", err
	}

	result, err := client.Chat([]message.Message{
		message.NewUserMessage(input),
	})

	if err != nil {
		return "", err
	}
	return result.Content, nil
}
