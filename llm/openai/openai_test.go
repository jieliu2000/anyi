package openai

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm/chat"
	impl "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestChatWithNoClientImplmentation(t *testing.T) {
	client := OpenAIClient{}
	_, _, err := client.Chat(nil, nil)
	if err == nil {
		t.Error("Expected error when no client is set")
	}
}

func TestConvertToOpenAIChatMessages(t *testing.T) {
	messages := []chat.Message{
		{Role: "system", Content: "You are an assisstant"},
		{Role: "user", Content: "Hello"},
	}

	openAIMessages := ConvertToOpenAIChatMessages(messages)

	if len(openAIMessages) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(openAIMessages))
	}

	assert.Equal(t, impl.ChatMessageRoleSystem, openAIMessages[0].Role, "Expected role %s, got %s", impl.ChatMessageRoleSystem, openAIMessages[0].Role)
	assert.Equal(t, "You are an assisstant", openAIMessages[0].Content, "Expected content %s, got %s", "You are an assisstant", openAIMessages[0].Content)
	assert.Equal(t, impl.ChatMessageRoleUser, openAIMessages[1].Role, "Expected role %s, got %s", impl.ChatMessageRoleUser, openAIMessages[1].Role)
	assert.Equal(t, "Hello", openAIMessages[1].Content, "Expected content %s, got %s", "Hello", openAIMessages[1].Content)
}

func TestNewClient(t *testing.T) {
	apiKey := "your-api-key"

	// Test with only API key provided
	config1 := NewConfig(apiKey, "", "")
	client1, err := NewClient(config1)
	assert.NoError(t, err)
	assert.NotNil(t, client1)
	assert.Equal(t, apiKey, client1.Config.APIKey)
	assert.Equal(t, DefaultModel, client1.Config.Model)
	assert.Equal(t, DefaultBaseURL, client1.Config.BaseURL)
	assert.NotNil(t, client1.clientImpl)

	// Test with all parameters provided
	config2 := NewConfig(apiKey, "some-model", "some-base-url")
	client2, err := NewClient(config2)
	assert.NoError(t, err)
	assert.NotNil(t, client2)
	assert.Equal(t, apiKey, client2.Config.APIKey)
	assert.Equal(t, "some-model", client2.Config.Model)
	assert.Equal(t, "some-base-url", client2.Config.BaseURL)
	assert.NotNil(t, client2.clientImpl)

	// Test with nil config
	_, err = NewClient(nil)
	assert.Error(t, err)
}

func TestNewConfig(t *testing.T) {
	config := NewConfig("token", "model", "URL")

	assert.Equal(t, "token", config.APIKey, "Expected token %s, got %s", "token", config.APIKey)
	assert.Equal(t, "model", config.Model, "Expected model %s, got %s", "model", config.Model)
	assert.Equal(t, "URL", config.BaseURL, "Expected URL %s, got %s", "URL", config.BaseURL)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-api-key")
	assert.Equal(t, "test-api-key", config.APIKey)
	assert.Equal(t, DefaultModel, config.Model)
	assert.Equal(t, DefaultBaseURL, config.BaseURL)
}

func TestChat(t *testing.T) {
	mockServer := test.NewTestServer()
	mockServer.RequestHandler = func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected POST method")
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		requestMap := make(map[string]interface{})
		err = json.Unmarshal(body, &requestMap)

		assert.NoError(t, err)
		assert.Equal(t, "gpt-3.5-turbo", requestMap["model"])

		messages := requestMap["messages"].([]interface{})
		assert.Equal(t, 2, len(messages))

		assert.Equal(t, "system", messages[0].(map[string]interface{})["role"])
		assert.Equal(t, "You are an assisstant", messages[0].(map[string]interface{})["content"])
		assert.Equal(t, "user", messages[1].(map[string]interface{})["role"])
		assert.Equal(t, "Hello", messages[1].(map[string]interface{})["content"])

		io.WriteString(w, `{
		"id":"chat-123",
		"object":"chat.completion",
		"choices":[
			{
			"message":{
				"role":"assistant",
				"content":"Reply to your input"
				},
			"finish_reason":"stop"
			}
		],
		"usage":{
			"prompt_tokens":10,
			"completion_tokens":25,
			"total_tokens":35
			},
		"model":"text-davinci-002",
		"created":1624850937,
		"model_version":"2021-06-25"
		}`)
	}

	defer mockServer.Close()
	mockServer.Start()

	config := NewConfig("test-api-key", "", mockServer.URL())

	client, err := NewClient(config)
	assert.NoError(t, err)

	messages := []chat.Message{
		{Role: "system", Content: "You are an assisstant"},
		{Role: "user", Content: "Hello"},
	}

	response, _, err := client.Chat(messages, nil)

	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, "Reply to your input", response.Content)
}
