package deepseek

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://api.deepseek.com/v1"

	// Official DeepSeek API models (verified from official documentation)
	// https://api-docs.deepseek.com/api/list-models

	// DeepSeek Chat model - points to DeepSeek-V3-0324
	DeepSeekChat = "deepseek-chat"

	// DeepSeek Reasoning model - points to DeepSeek-R1-0528
	DeepSeekReasoner = "deepseek-reasoner"

	// Default model - using the general chat model
	DefaultModel = "deepseek-chat"
)

type DeepSeekModelConfig struct {
	config.GeneralLLMConfig
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
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *DeepSeekModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &DeepSeekModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          baseUrl,
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
	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChatWithFunctions(c.clientImpl, c.Config.Model, messages, functions, options, openaiConfig)
}

func (c *DeepSeekClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChat(c.clientImpl, c.Config.Model, messages, options, openaiConfig)
}
