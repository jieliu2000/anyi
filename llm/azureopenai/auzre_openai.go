package azureopenai

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

type AzureOpenAIModelConfig struct {
	config.GeneralLLMConfig
	APIKey            string `json:"apiKey" mapstructure:"apiKey" yaml:"apiKey"`
	ModelDeploymentId string `json:"modelDeploymentId" mapstructure:"modelDeploymentId" yaml:"modelDeploymentId"`
	Endpoint          string `json:"endpoint" mapstructure:"endpoint" yaml:"endpoint"`
	AllowInsecureHttp bool   `json:"allowInsecureHttp" yaml:"allowInsecureHttp" mapstructure:"allowInsecureHttp"`
}

type AzureOpenAIClient struct {
	Config     *AzureOpenAIModelConfig
	clientImpl *impl.Client
}

func NewConfig(apiKey string, modelDeploymentId string, endpoint string) *AzureOpenAIModelConfig {
	return &AzureOpenAIModelConfig{
		GeneralLLMConfig:  config.DefaultGeneralConfig(),
		APIKey:            apiKey,
		ModelDeploymentId: modelDeploymentId,
		Endpoint:          endpoint,
	}
}

func NewClient(config *AzureOpenAIModelConfig) (*AzureOpenAIClient, error) {

	if config == nil {
		return nil, errors.New("config is required")
	}
	if config.APIKey == "" {
		return nil, errors.New("api_key is required")
	}
	if config.ModelDeploymentId == "" {
		return nil, errors.New("model_deployment_id is required")
	}
	if config.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	configImpl := impl.DefaultAzureConfig(config.APIKey, config.Endpoint)
	configImpl.AzureModelMapperFunc = func(modelId string) string { return modelId }

	client := &AzureOpenAIClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	return client, nil
}

func (c *AzureOpenAIClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.ModelDeploymentId,
	}

	return openai.ExecuteChatWithFunctions(client, c.Config.ModelDeploymentId, messages, functions, options, openaiConfig)
}

func (c *AzureOpenAIClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.ModelDeploymentId,
	}

	return openai.ExecuteChat(client, c.Config.ModelDeploymentId, messages, options, openaiConfig)
}
