package azureopenai

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("key", "deploy", "endpoint")

	assert.Equal(t, "key", config.APIKey, "Expected token %s, got %s", "key", config.APIKey)
	assert.Equal(t, "deploy", config.ModelDeploymentId, "Expected model %s, got %s", "deploy", config.ModelDeploymentId)
	assert.Equal(t, "endpoint", config.Endpoint, "Expected URL %s, got %s", "endpoint", config.Endpoint)
}

func TestNewClient(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		APIKey:            "test_api_key",
		ModelDeploymentId: "test_model_deployment_id",
		Endpoint:          "test_endpoint",
	}
	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config)
	assert.NotNil(t, client.clientImpl)
}

func TestNewClient_NilConfig(t *testing.T) {
	client, err := NewClient(nil)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyConfig(t *testing.T) {
	config := &AzureOpenAIModelConfig{}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyAPIKey(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		ModelDeploymentId: "test_model_deployment_id",
		Endpoint:          "test_endpoint",
	}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyModelDeploymentId(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		APIKey:   "test_api_key",
		Endpoint: "test_endpoint",
	}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyEndpoint(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		APIKey:            "test_api_key",
		ModelDeploymentId: "test_model_deployment_id",
	}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestChat(t *testing.T) {
	mockServer := test.NewTestServer()
	mockServer.RequestHandler = func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected POST method")
		assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		requestMap := make(map[string]interface{})
		err = json.Unmarshal(body, &requestMap)

		assert.NoError(t, err)
		assert.Equal(t, "test-deploy", requestMap["model"])

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

	config := NewConfig("test-api-key", "test-deploy", mockServer.URL())
	config.AllowInsecureHttp = true

	client, err := NewClient(config)
	assert.NoError(t, err)

	messages := []chat.Message{
		{Role: "system", Content: "You are an assisstant"},
		{Role: "user", Content: "Hello"},
	}

	response, err := client.Chat(messages, nil)

	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, "Reply to your input", response.Content)
}
