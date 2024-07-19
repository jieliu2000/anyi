package openai

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/message"
	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-3.5-turbo"
)

type OpenAIModelConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

type OpenAIClient struct {
	Config     *OpenAIModelConfig
	clientImpl *impl.Client
}

func DefaultConfig(apiKey string) *OpenAIModelConfig {
	return NewConfig(apiKey, "", "")
}

func NewConfigWithModel(apiKey string, model string) *OpenAIModelConfig {
	return NewConfig(apiKey, model, "")
}

// Create a new config with the given API, model, and baseURL
// If you don't know the baseURL or model, you can leave them as blank string. The function will use default values if they are not provided.
func NewConfig(apiKey string, model string, baseURL string) *OpenAIModelConfig {
	if model == "" {
		model = DefaultModel
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &OpenAIModelConfig{APIKey: apiKey, Model: model, BaseURL: baseURL}
}

func NewClient(config *OpenAIModelConfig) (*OpenAIClient, error) {

	if config == nil {
		return nil, errors.New("config cannot be null")
	}

	configImpl := impl.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		configImpl.BaseURL = config.BaseURL
	}

	client := &OpenAIClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	return client, nil
}

func (c *OpenAIClient) Chat(messages []message.Message) (*message.Message, error) {
	if c.clientImpl == nil {
		return nil, errors.New("client not initialized")
	}

	messagesInput := ConvertToOpenAIChatMessages(messages)

	resp, err := c.clientImpl.CreateChatCompletion(
		context.Background(),
		impl.ChatCompletionRequest{
			Model:    c.Config.Model,
			Messages: messagesInput,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no chat completion choices returned")
	}

	return &message.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}, nil
}
