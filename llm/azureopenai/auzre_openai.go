package azureopenai

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

type AzureOpenAIModelConfig struct {
	APIKey            string `json:"api_key"`
	ModelDeploymentId string `json:"model_deployment_id"`
	Endpoint          string `json:"endpoint"`
	AllowInsecureHttp bool   `json:"allowInsecureHttp" yaml:"allowInsecureHttp" mapstructure:"allowInsecureHttp"`
}

type AzureOpenAIClient struct {
	Config     *AzureOpenAIModelConfig
	clientImpl *impl.Client
}

func NewConfig(apiKey string, modelDeploymentId string, endpoint string) *AzureOpenAIModelConfig {
	return &AzureOpenAIModelConfig{APIKey: apiKey, ModelDeploymentId: modelDeploymentId, Endpoint: endpoint}
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
	return nil, chat.ResponseInfo{}, errors.New("not implemented")
}

func (c *AzureOpenAIClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {

	client := c.clientImpl
	info := chat.ResponseInfo{}

	if client == nil {
		return nil, info, errors.New("client not initialized")
	}

	messagesInput := openai.ConvertToOpenAIChatMessages(messages)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		impl.ChatCompletionRequest{
			Model:    c.Config.ModelDeploymentId,
			Messages: messagesInput,
		},
	)

	if err != nil {
		return nil, info, err
	}
	result := chat.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}
	info.PromptTokens = resp.Usage.PromptTokens
	info.CompletionTokens = resp.Usage.CompletionTokens

	return &result, info, nil
}
