package chat

import (
	"encoding/json"
	"errors"
)

type ChatOptions struct {
	Format string `json:"format"`
}

func NewChatOptions(format string) ChatOptions {
	return ChatOptions{Format: format}
}

func SetChatOptions[T any](options *ChatOptions, target *T) error {
	if options == nil {
		return nil
	}
	if target == nil {
		return errors.New("target is nil")
	}

	bytes, err := json.Marshal(options)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, target)
}
