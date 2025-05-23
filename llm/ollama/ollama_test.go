package ollama

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/stretchr/testify/assert"
)

func TestChatWithNoClientImplmentation(t *testing.T) {
	client := OllamaClient{}
	_, _, err := client.Chat(nil, nil)

	assert.Error(t, err, "Chat should fail with no client implementation")
}

func TestNewConfigWithURL(t *testing.T) {
	model := "test_model"
	ollamaApiURL := "http://localhost:11434/api"
	config := NewConfig(model, ollamaApiURL)

	assert.Equal(t, config.Model, model, "NewConfig created a config with model %s, want %s", config.Model, model)

	assert.Equal(t, config.OllamaApiURL, ollamaApiURL, "NewConfig created a config with OllamaApiURL %s, want %s", config.OllamaApiURL, ollamaApiURL)

}

func TestNewConfigWithEmptyOllamaUrl(t *testing.T) {
	model := "test_model"
	config := NewConfig(model, "")

	assert.Equal(t, config.Model, model, "NewConfig created a config with model %s, want %s", config.Model, model)

	assert.Equal(t, config.OllamaApiURL, DefaultOllamaUrl, "NewConfig created a config with OllamaApiURL %s, want %s", config.OllamaApiURL, DefaultOllamaUrl)

}

func TestNewClient(t *testing.T) {
	t.Run("Returns error when config is nil", func(t *testing.T) {
		_, err := NewClient(nil)
		assert.Error(t, err)
	})
	t.Run("Fills in default values when some are missing", func(t *testing.T) {
		config := &OllamaModelConfig{
			Model: "test-model",
		}
		client, err := NewClient(config)
		assert.NoError(t, err)
		assert.Equal(t, DefaultOllamaUrl, client.Config.OllamaApiURL)
	})
	t.Run("Returns error when model is empty", func(t *testing.T) {
		config := &OllamaModelConfig{}
		_, err := NewClient(config)
		assert.Error(t, err)
	})

	t.Run("Should return valid client", func(t *testing.T) {
		config := &OllamaModelConfig{
			OllamaApiURL: "http://localhost:11434/api",
			Model:        "test-model",
		}
		client, err := NewClient(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.clientImpl)
		assert.Equal(t, config, client.Config)
	})
}

// Test GeneralLLMConfig configuration options
func TestGeneralLLMConfig(t *testing.T) {
	// Create configuration
	config := &OllamaModelConfig{
		GeneralLLMConfig: config.GeneralLLMConfig{
			Temperature:      0.7,
			TopP:             0.9,
			MaxTokens:        100,
			PresencePenalty:  0.5,
			FrequencyPenalty: 0.5,
			Stop:             []string{"stop1", "stop2"},
		},
		Model:        "test-model",
		OllamaApiURL: "http://localhost:11434/api",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Check if the configuration is correctly saved
	assert.Equal(t, float32(0.7), client.Config.Temperature)
	assert.Equal(t, float32(0.9), client.Config.TopP)
	assert.Equal(t, 100, client.Config.MaxTokens)
	assert.Equal(t, float32(0.5), client.Config.PresencePenalty)
	assert.Equal(t, float32(0.5), client.Config.FrequencyPenalty)
	assert.Equal(t, []string{"stop1", "stop2"}, client.Config.Stop)
}

// Test request applying GeneralLLMConfig
func TestChatWithGeneralLLMConfig(t *testing.T) {
	mockServer := test.NewTestServer()

	mockServer.RequestHandler = func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected POST method")
		assert.Equal(t, "/chat", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		requestMap := make(map[string]interface{})
		err = json.Unmarshal(body, &requestMap)

		assert.NoError(t, err)
		assert.Equal(t, "test-model", requestMap["model"])

		// Verify that parameters from GeneralLLMConfig are correctly passed
		assert.Equal(t, float64(0.7), requestMap["temperature"])
		assert.Equal(t, float64(0.9), requestMap["top_p"])
		assert.Equal(t, float64(100), requestMap["num_predict"])
		assert.Equal(t, float64(0.5), requestMap["presence_penalty"])
		assert.Equal(t, float64(0.5), requestMap["frequency_penalty"])

		// Verify stop word list
		stopWords, ok := requestMap["stop"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(stopWords))
		assert.Equal(t, "stop1", stopWords[0])
		assert.Equal(t, "stop2", stopWords[1])

		io.WriteString(w, `{
    "model": "mistral",
    "created_at": "2024-07-27T00:21:55.1718475Z",
    "message": {
        "role": "assistant",
        "content": "Reply to your input"
    },
    "done_reason": "stop",
    "done": true,
    "total_duration": 8369719900,
    "load_duration": 5773330000,
    "prompt_eval_count": 11,
    "prompt_eval_duration": 32476000,
    "eval_count": 134,
    "eval_duration": 2548614000
}`)
	}

	defer mockServer.Close()
	mockServer.Start()

	// Create client with GeneralLLMConfig configuration
	config := &OllamaModelConfig{
		GeneralLLMConfig: config.GeneralLLMConfig{
			Temperature:      0.7,
			TopP:             0.9,
			MaxTokens:        100,
			PresencePenalty:  0.5,
			FrequencyPenalty: 0.5,
			Stop:             []string{"stop1", "stop2"},
		},
		Model:        "test-model",
		OllamaApiURL: mockServer.URL(),
	}

	client, err := NewClient(config)
	assert.NoError(t, err)

	messages := []chat.Message{
		{Role: "user", Content: "Hello"},
	}

	// Call Chat method, verify that GeneralLLMConfig parameters are applied
	response, _, err := client.Chat(messages, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Reply to your input", response.Content)
}

// Test overriding format setting
func TestChatWithFormatOption(t *testing.T) {
	mockServer := test.NewTestServer()

	mockServer.RequestHandler = func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected POST method")
		assert.Equal(t, "/chat", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		requestMap := make(map[string]interface{})
		err = json.Unmarshal(body, &requestMap)

		assert.NoError(t, err)
		assert.Equal(t, "test-model", requestMap["model"])

		// Verify format option is correctly set
		assert.Equal(t, "json", requestMap["format"])

		io.WriteString(w, `{
    "model": "mistral",
    "created_at": "2024-07-27T00:21:55.1718475Z",
    "message": {
        "role": "assistant",
        "content": "Reply to your input"
    },
    "done_reason": "stop",
    "done": true,
    "total_duration": 8369719900,
    "load_duration": 5773330000,
    "prompt_eval_count": 11,
    "prompt_eval_duration": 32476000,
    "eval_count": 134,
    "eval_duration": 2548614000
}`)
	}

	defer mockServer.Close()
	mockServer.Start()

	config := &OllamaModelConfig{
		Model:        "test-model",
		OllamaApiURL: mockServer.URL(),
	}

	client, err := NewClient(config)
	assert.NoError(t, err)

	messages := []chat.Message{
		{Role: "user", Content: "Hello"},
	}

	options := &chat.ChatOptions{
		Format: "json",
	}

	// Call Chat method, verify that format option is applied
	response, _, err := client.Chat(messages, options)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Reply to your input", response.Content)
}

func TestChat(t *testing.T) {
	mockServer := test.NewTestServer()

	mockServer.RequestHandler = func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected POST method")
		assert.Equal(t, "/chat", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		requestMap := make(map[string]interface{})
		err = json.Unmarshal(body, &requestMap)

		assert.NoError(t, err)
		assert.Equal(t, "test-model", requestMap["model"])

		messages := requestMap["messages"].([]interface{})
		assert.Equal(t, 2, len(messages))

		assert.Equal(t, "system", messages[0].(map[string]interface{})["role"])
		assert.Equal(t, "You are an assisstant", messages[0].(map[string]interface{})["content"])
		assert.Equal(t, "user", messages[1].(map[string]interface{})["role"])
		assert.Equal(t, "Hello", messages[1].(map[string]interface{})["content"])

		io.WriteString(w, `{
    "model": "mistral",
    "created_at": "2024-07-27T00:21:55.1718475Z",
    "message": {
        "role": "assistant",
        "content": "Reply to your input"
    },
    "done_reason": "stop",
    "done": true,
    "total_duration": 8369719900,
    "load_duration": 5773330000,
    "prompt_eval_count": 11,
    "prompt_eval_duration": 32476000,
    "eval_count": 134,
    "eval_duration": 2548614000
}`)
	}

	defer mockServer.Close()
	mockServer.Start()

	config := NewConfig("test-model", mockServer.URL())

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
