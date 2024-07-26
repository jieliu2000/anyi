package ollama

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultOllamaUrl = "'http://localhost:11434/v1/'"
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
	clientImpl *impl.Client
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

	// Ollama's openAI compatible API use "ollama" as the API key
	configImpl := impl.DefaultConfig("ollama")

	// Set the BaseURL from the provided config
	configImpl.BaseURL = config.OllamaUrl

	// Create a new OllamaClient using the provided config and the configured client implementation
	client := &OllamaClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	// Return the newly created OllamaClient and nil error
	return client, nil
}

func (c *OllamaClient) Chat(messages []message.Message) (*message.Message, error) {

	// Check if the client implementation is initialized
	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
	}

	// Convert the messages to OpenAI ChatMessages format
	messagesInput := openai.ConvertToOpenAIChatMessages(messages)

	// Create a ChatCompletion request using the client and the converted messages
	resp, err := client.CreateChatCompletion(
		context.Background(),
		impl.ChatCompletionRequest{
			Model:    c.Config.Model,
			Messages: messagesInput,
		},
	)

	// Check if there was an error in creating the ChatCompletion
	if err != nil {
		return nil, err
	}

	// Check if there are no choices in the response
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices found in the response")
	}

	// Extract the first choice from the response and create a new message object
	result := message.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}

	// Return the new message object and nil error
	return &result, nil
}
