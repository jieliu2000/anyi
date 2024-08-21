package chat_test

import (
	"testing"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/stretchr/testify/assert"
)

type chatOptions struct {
	Name string `json:"name"`
}

func TestSetChatOptions(t *testing.T) {
	options := chatOptions{
		Name: "TestName",
	}
	var target chatOptions
	err := chat.SetChatOptions(options, &target)
	assert.NoError(t, err)
	assert.Equal(t, options.Name, target.Name)
}
func TestSetChatOptionsWithNilOptions(t *testing.T) {
	var target chatOptions
	err := chat.SetChatOptions(nil, &target)
	assert.NoError(t, err)
	assert.Zero(t, target)
}
func TestSetChatOptionsWithNilTarget(t *testing.T) {
	options := chatOptions{
		Name: "TestName",
	}
	err := chat.SetChatOptions[string](options, nil)
	assert.Error(t, err)
}
