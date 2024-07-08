package openai

import (
	"testing"

	"github.com/jieliu2000/anyi/message"
	impl "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestChatWithNoClient(t *testing.T) {
	client := OpenAIClient{}
	_, err := client.Chat(nil)
	if err == nil {
		t.Error("Expected error when no client is set")
	}
}

func TestConvertToOpenAIChatMessages(t *testing.T) {
	messages := []message.Message{
		{Role: "system", Content: "You are an assisstant"},
		{Role: "user", Content: "Hello"},
	}

	openAIMessages := convertToOpenAIChatMessages(messages)

	if len(openAIMessages) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(openAIMessages))
	}

	assert.Equal(t, impl.ChatMessageRoleSystem, openAIMessages[0].Role, "Expected role %s, got %s", impl.ChatMessageRoleSystem, openAIMessages[0].Role)
	assert.Equal(t, "You are an assisstant", openAIMessages[0].Content, "Expected content %s, got %s", "You are an assisstant", openAIMessages[0].Content)
	assert.Equal(t, impl.ChatMessageRoleUser, openAIMessages[1].Role, "Expected role %s, got %s", impl.ChatMessageRoleUser, openAIMessages[1].Role)
	assert.Equal(t, "Hello", openAIMessages[1].Content, "Expected content %s, got %s", "Hello", openAIMessages[1].Content)
}

func TestNewConfig(t *testing.T) {
	config := NewConfig("token", "model", "URL")

	assert.Equal(t, "token", config.APIKey, "Expected token %s, got %s", "token", config.APIKey)
	assert.Equal(t, "model", config.Model, "Expected model %s, got %s", "model", config.Model)
	assert.Equal(t, "URL", config.BaseURL, "Expected URL %s, got %s", "URL", config.BaseURL)
}
