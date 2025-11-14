package minimax

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://api.minimaxi.com/v1"

	// Official MiniMax API models (OpenAI compatible API)
	// https://platform.minimaxi.com/docs/api-reference/text-openai-api

	// Official MiniMax models from https://platform.minimaxi.com/docs/api-reference/text-intro
	MiniMaxM2         = "MiniMax-M2"
	MiniMaxM2Stable   = "MiniMax-M2-Stable"

	// Default model - using the latest model
	DefaultModel = MiniMaxM2
)

type MiniMaxModelConfig struct {
	config.GeneralLLMConfig
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseUrl string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type MiniMaxClient struct {
	Config     *MiniMaxModelConfig
	clientImpl *impl.Client
}

// DefaultConfig creates a new config with default values for the given API key and model
func DefaultConfig(apiKey string, model string) *MiniMaxModelConfig {
	if model == "" {
		model = DefaultModel
	}
	return &MiniMaxModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          DefaultBaseUrl,
	}
}

// NewConfig creates a new config with the given API key, model, and base URL
// If baseUrl is empty, it will use the default value
func NewConfig(apiKey string, model string, baseUrl string) *MiniMaxModelConfig {
	if model == "" {
		model = DefaultModel
	}
	if baseUrl == "" {
		baseUrl = DefaultBaseUrl
	}
	return &MiniMaxModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          baseUrl,
	}
}

// NewClient creates a new MiniMax client with the given configuration
func NewClient(config *MiniMaxModelConfig) (*MiniMaxClient, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	configImpl := impl.DefaultConfig(config.APIKey)
	configImpl.BaseURL = config.BaseUrl

	return &MiniMaxClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}, nil
}

// ChatWithFunctions performs a chat completion with function/tool support
func (c *MiniMaxClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChatWithFunctions(c.clientImpl, c.Config.Model, messages, functions, options, openaiConfig)
}

// Chat performs a simple chat completion
func (c *MiniMaxClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChat(c.clientImpl, c.Config.Model, messages, options, openaiConfig)
}