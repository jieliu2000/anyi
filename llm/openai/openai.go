package openai

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/tools"
	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://api.openai.com/v1"

	// Official OpenAI models (verified from official documentation)
	// Note: Some models mentioned in news may not be available
	// in the official OpenAI API yet. Only use verified model names.

	// GPT-5 series - Latest generation (2025)
	GPT5          = "gpt-5"
	GPT5Turbo     = "gpt-5-turbo"
	GPT5Mini      = "gpt-5-mini"

	// GPT-4.5 series - Advanced reasoning models (2025)
	// Note: GPT-4.5 API service was discontinued in July 2025
	GPT45         = "gpt-4.5"
	GPT45Turbo    = "gpt-4.5-turbo"
	GPT45Vision   = "gpt-4.5-vision-preview"

	// GPT-4.1 series - Previous generation (2025)
	GPT41         = "gpt-4.1"
	GPT41Mini     = "gpt-4.1-mini"
	GPT41Nano     = "gpt-4.1-nano"

	// GPT-4o series - Current flagship models
	GPT4o         = "gpt-4o"
	GPT4oTurbo    = "gpt-4o-turbo"
	GPT4oVision   = "gpt-4o-vision-preview"
	GPT4o20241120 = "gpt-4o-2024-11-20"
	GPT4o20240806 = "gpt-4o-2024-08-06"
	GPT4o20240513 = "gpt-4o-2024-05-13"

	// GPT-4o mini - Cost-effective model
	GPT4oMini         = "gpt-4o-mini"
	GPT4oMini20240718 = "gpt-4o-mini-2024-07-18"

	// GPT-4 series - Previous generation
	// Note: GPT-4 was removed from ChatGPT on April 30, 2025, but still available in API
	GPT4              = "gpt-4"
	GPT4Turbo         = "gpt-4-turbo"
	GPT4Turbo20240409 = "gpt-4-turbo-2024-04-09"
	GPT40314          = "gpt-4-0314"
	GPT40613          = "gpt-4-0613"
	GPT41106Preview   = "gpt-4-1106-preview"
	GPT40125Preview   = "gpt-4-0125-preview"

	// GPT-3.5 series - Legacy models
	GPT35Turbo         = "gpt-3.5-turbo"
	GPT35Turbo20240125 = "gpt-3.5-turbo-0125"
	GPT35Turbo1106     = "gpt-3.5-turbo-1106"
	GPT35Turbo0613     = "gpt-3.5-turbo-0613"
	GPT35Turbo16k      = "gpt-3.5-turbo-16k"

	// o1 series - Reasoning models
	O1Preview = "o1-preview"
	O1Mini    = "o1-mini"

	// o3 series - Latest reasoning models (2025)
	O3Mini = "o3-mini" // Released January 31, 2025

	// Default model - using GPT-4o for balanced performance and cost
	DefaultModel = "gpt-4o"
)

type OpenAIModelConfig struct {
	config.GeneralLLMConfig
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseURL string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type OpenAIClient struct {
	Config     *OpenAIModelConfig
	clientImpl *impl.Client
}

func DefaultConfig(apiKey string) *OpenAIModelConfig {
	cfg := NewConfig(apiKey, "", "")
	cfg.GeneralLLMConfig = config.DefaultGeneralConfig()
	return cfg
}

func NewConfigWithModel(apiKey string, model string) *OpenAIModelConfig {
	cfg := NewConfig(apiKey, model, "")
	cfg.GeneralLLMConfig = config.DefaultGeneralConfig()
	return cfg
}

// Create a new config with the given API, model, and baseURL
// If you don't know the baseURL or model, you can leave them as blank string. The function will use default values if they are not provided.
func NewConfig(apiKey string, model string, baseURL string) *OpenAIModelConfig {
	if model == "" {
		model = DefaultModel
	}
	if baseURL == "" {
		baseURL = DefaultBaseUrl
	}
	return &OpenAIModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseURL:          baseURL,
	}
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

func (c *OpenAIClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	return ExecuteChatWithFunctions(client, c.Config.Model, messages, functions, options, c.Config)
}

func (c *OpenAIClient) Chat(messages []chat.Message, options *chat.ChatOptions) (message *chat.Message, responseInfo chat.ResponseInfo, err error) {
	client := c.clientImpl

	if client == nil {
		return nil, chat.ResponseInfo{}, errors.New("client cannot be null")
	}
	if c.Config == nil {
		return nil, chat.ResponseInfo{}, errors.New("config cannot be null")
	}

	return ExecuteChat(client, c.Config.Model, messages, options, c.Config)
}
