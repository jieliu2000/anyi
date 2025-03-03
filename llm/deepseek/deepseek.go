package deepseek

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://api.deepseek.com/v1"
)

type DeepSeekModelConfig struct {
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseUrl string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type DeepSeekClient struct {
	Config     *DeepSeekModelConfig
	clientImpl *impl.Client
}

func DefaultConfig(apiKey string, model string) *DeepSeekModelConfig {
	return &DeepSeekModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *DeepSeekModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &DeepSeekModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: baseUrl,
	}
}

func NewClient(config *DeepSeekModelConfig) (*DeepSeekClient, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	configImpl := impl.DefaultConfig(config.APIKey)
	configImpl.BaseURL = config.BaseUrl

	return &DeepSeekClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}, nil
}

func (c *DeepSeekClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	return openai.ExecuteChatWithFunctions(c.clientImpl, c.Config.Model, messages, functions, options)
}

func (c *DeepSeekClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	return openai.ExecuteChat(c.clientImpl, c.Config.Model, messages, options)
}
