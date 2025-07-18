package anthropic

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://api.anthropic.com"

	// Official Anthropic API models (verified from official documentation)
	// https://docs.anthropic.com/en/docs/about-claude/models/overview

	// Claude 4 series - Latest generation (2025)
	ClaudeOpus4   = "claude-opus-4-20250514"
	ClaudeSonnet4 = "claude-sonnet-4-20250514"

	// Claude 3.7 series - Extended thinking model
	Claude37Sonnet = "claude-3-7-sonnet-20250219"

	// Claude 3.5 series - Current generation
	Claude35Sonnet   = "claude-3-5-sonnet-20241022" // v2
	Claude35SonnetV1 = "claude-3-5-sonnet-20240620" // v1
	Claude35Haiku    = "claude-3-5-haiku-20241022"

	// Claude 3 series - Previous generation
	Claude3Opus   = "claude-3-opus-20240229"
	Claude3Sonnet = "claude-3-sonnet-20240229"
	Claude3Haiku  = "claude-3-haiku-20240307"

	// Default model - using proven reliable model
	DefaultModel = "claude-3-5-sonnet-20241022"
)

// AnthropicModelConfig defines configuration parameters for Anthropic models
type AnthropicModelConfig struct {
	config.GeneralLLMConfig
	APIKey     string `json:"apiKey" mapstructure:"apiKey"`
	BaseUrl    string `json:"baseUrl" mapstructure:"baseUrl"`
	Model      string `json:"model" mapstructure:"model"`
	APIVersion string `json:"apiVersion" mapstructure:"apiVersion"`
}

// AnthropicClient defines the structure for Anthropic client
type AnthropicClient struct {
	Config     *AnthropicModelConfig
	clientImpl *impl.Client
}

// DefaultConfig creates default Anthropic configuration
func DefaultConfig(apiKey string) *AnthropicModelConfig {
	return &AnthropicModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            DefaultModel,
		BaseUrl:          DefaultBaseUrl,
		APIVersion:       "2023-06-01", // Use the latest API version
	}
}

// DefaultConfigWithModel creates default configuration with specified model
func DefaultConfigWithModel(apiKey string, model string) *AnthropicModelConfig {
	config := DefaultConfig(apiKey)
	config.Model = model
	return config
}

// NewConfig creates a new configuration allowing all parameters to be specified
func NewConfig(apiKey string, model string, baseUrl string, apiVersion string) *AnthropicModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	if len(apiVersion) == 0 {
		apiVersion = "2023-06-01"
	}
	return &AnthropicModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          baseUrl,
		APIVersion:       apiVersion,
	}
}

// NewClient creates a new Anthropic client
func NewClient(config *AnthropicModelConfig) (*AnthropicClient, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("api key cannot be empty")
	}

	if config.Model == "" {
		config.Model = DefaultModel
	}

	// Create OpenAI compatible client
	configImpl := impl.DefaultConfig(config.APIKey)
	configImpl.BaseURL = config.BaseUrl

	// Note: This is a simplified implementation, actual integration may require adding Anthropic-specific request headers
	// such as 'anthropic-version'
	// The current go-openai library doesn't directly expose methods to set custom headers
	// A custom HTTP implementation might be needed in a production environment

	return &AnthropicClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}, nil
}

// Chat implements chat functionality for Anthropic
func (c *AnthropicClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChat(c.clientImpl, c.Config.Model, messages, options, openaiConfig)
}

// ChatWithFunctions implements function calling functionality for Anthropic
func (c *AnthropicClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChatWithFunctions(c.clientImpl, c.Config.Model, messages, functions, options, openaiConfig)
}
