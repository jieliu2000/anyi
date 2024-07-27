package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jieliu2000/anyi/message"
)

const (
	DefaultOllamaUrl = "http://localhost:11434/api"
)

type OllamaModelConfig struct {
	OllamaUrl string `json:"base_url"`

	//The model name used by ollama. See [Ollama's documentation] for more information on the available models.
	//
	//[Ollama's documentation]: https://github.com/ollama/ollama/blob/main/README.md#quickstart
	Model string `json:"model"`
}

type OllamaClient struct {
	Config     *OllamaModelConfig
	clientImpl *http.Client
}

// Creats a default Ollama model config.
func DefaultConfig(model string) *OllamaModelConfig {
	return &OllamaModelConfig{
		Model:     model,
		OllamaUrl: DefaultOllamaUrl,
	}
}

func NewConfig(apiKey string, model string, ollamaUrl string) *OllamaModelConfig {
	if len(ollamaUrl) == 0 {
		ollamaUrl = DefaultOllamaUrl
	}
	return &OllamaModelConfig{
		Model:     model,
		OllamaUrl: ollamaUrl,
	}
}

func NewClient(config *OllamaModelConfig) (*OllamaClient, error) {

	// Check if the config is nil to prevent panic or unexpected behavior
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Create a new OllamaClient using the provided config and the configured client implementation
	client := &OllamaClient{
		Config: config,
	}

	client.clientImpl = &http.Client{}

	// Return the newly created OllamaClient and nil error
	return client, nil
}

type OllamaRequest struct {
	Model    string            `json:"model"`
	Messages []message.Message `json:"messages"`
	Stream   bool              `json:"stream"`
}

type OllamaResponse struct {
	Message       message.Message `json:"message"`
	CreatedAt     time.Time       `json:"created_at"`
	Done          bool            `json:"done"`
	TotalDuration int             `json:"total_duration"`
	LoadDuration  int             `json:"load_duration"`
}

func (c *OllamaClient) Chat(messages []message.Message) (*message.Message, error) {

	httpClient := c.clientImpl

	requestJson, err := json.Marshal(OllamaRequest{
		Model:    c.Config.Model,
		Messages: messages,
	})

	if err != nil {
		return nil, err
	}

	response, err := httpClient.Post(c.Config.OllamaUrl+"/chat", "application/json", bytes.NewBuffer(requestJson))

	if err != nil {
		return nil, err
	}
	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	ollamaResponse := OllamaResponse{}
	err = json.Unmarshal(responseBody, &ollamaResponse)
	if err != nil {
		return nil, err
	}

	return &ollamaResponse.Message, nil
}
