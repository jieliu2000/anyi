package openai

import (
	"github.com/jieliu2000/anyi/llm/chat"
	impl "github.com/sashabaranov/go-openai"
)

type OpenAIModelConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

type OpenAIClient struct {
	Config     *OpenAIModelConfig
	clientImpl *impl.Client
}

func (c *OpenAIClient) Init() error {

	return nil
}

func (c *OpenAIClient) Chat(messages []*chat.Message, history []*chat.Message) (*chat.Message, error) {

	return nil, nil
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
