package ollama

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChatWithNoClientImplmentation(t *testing.T) {
	client := OllamaClient{}
	_, err := client.Chat(nil)

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
