package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jieliu2000/anyi/message"
)

const (
	DefaultOllamaUrl = "http://localhost:11434/api"
)

type OllamaModelConfig struct {
	//The url of the ollama server. Note that don't add "/chat" to the end of this url. In [Chat] function it will be added automatically.
	OllamaApiURL string `json:"ollama_api_url"`

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
		Model:        model,
		OllamaApiURL: DefaultOllamaUrl,
	}
}

// NewConfig creates and returns a pointer to OllamaModelConfig.
// Parameters:
// - model string: The model name used by ollama. See [Ollama's documentation] for more information on the available models.
// - ollamaApiURL string: The Ollama API URL. Leaving this blank will use the default Ollama API URL. See [DefaultOllamaUrl]. Note that don't add "/chat" to the end of this url. In [Chat] function it will be added automatically.
//
// [Ollama's documentation]: https://github.com/ollama/ollama/blob/main/README.md
func NewConfig(model string, ollamaApiURL string) *OllamaModelConfig {
	if len(ollamaApiURL) == 0 {
		ollamaApiURL = DefaultOllamaUrl
	}
	return &OllamaModelConfig{
		Model:        model,
		OllamaApiURL: ollamaApiURL,
	}
}

// NewClient creates a new OllamaClient instance based on the provided OllamaModelConfig.
// If the config is nil, it will return an error. If the model in the config is empty, it will return an error as well because ollama chat cannot be called without a model.
// the OllamaApiURL in the config can be left blank. The default Ollama API URL will be used in that case.
func NewClient(config *OllamaModelConfig) (*OllamaClient, error) {

	// Check if the config is nil to prevent panic or unexpected behavior
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	if config.OllamaApiURL == "" {
		config.OllamaApiURL = DefaultOllamaUrl
	}

	if config.Model == "" {
		return nil, errors.New("model cannot be empty")
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
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
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

	if httpClient == nil {
		return nil, errors.New("http client cannot be nil, maybe you didn't initiatialize the client. Considering using NewClient function")
	}

	ollamaMessages, err := ConvertToOllamaMessages(messages)
	if err != nil {
		return nil, err
	}
	requestJson, err := json.Marshal(OllamaRequest{
		Model:    c.Config.Model,
		Messages: ollamaMessages,
	})

	if err != nil {
		return nil, err
	}

	response, err := httpClient.Post(c.Config.OllamaApiURL+"/chat", "application/json", bytes.NewBuffer(requestJson))

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response status from ollama chat api: %d", response.StatusCode)
	}

	defer response.Body.Close()

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
