package openai

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	impl "github.com/sashabaranov/go-openai"
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

func (c *OpenAIClient) Init() error {

	return nil
}

func NewClient(config *OpenAIModelConfig) (*OpenAIClient, error) {

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

func (c *OpenAIClient) Chat(messages []chat.Message) (*chat.Message, error) {

	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
	}

	messagesInput := convertToOpenAIChatMessages(messages)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		impl.ChatCompletionRequest{
			Model:    c.Config.Model,
			Messages: messagesInput,
		},
	)

	if err != nil {
		return nil, err
	}
	result := chat.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}
	return &result, nil
}

func convertToOpenAIChatMessages(messages []chat.Message) []impl.ChatCompletionMessage {
	result := []impl.ChatCompletionMessage{}
	for _, msg := range messages {
		openaiMessage := impl.ChatCompletionMessage{
			Content: msg.Content,
			Role:    msg.Role,
		}
		result = append(result, openaiMessage)
	}
	return result
}
