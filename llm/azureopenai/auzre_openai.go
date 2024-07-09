package azureopenai

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"

	impl "github.com/sashabaranov/go-openai"
)

type AzureOpenAIModelConfig struct {
	APIKey            string `json:"api_key"`
	ModelDeploymentId string `json:"model_deployment_id"`
	Endpoint          string `json:"endpoint"`
}

type AzureOpenAIClient struct {
	Config     *AzureOpenAIModelConfig
	clientImpl *impl.Client
}

func (c *AzureOpenAIClient) Init() error {

	return nil
}

func NewConfig(apiKey string, modelDeploymentId string, endpoint string) *AzureOpenAIModelConfig {
	return &AzureOpenAIModelConfig{APIKey: apiKey, ModelDeploymentId: modelDeploymentId, Endpoint: endpoint}
}

func NewClient(config *AzureOpenAIModelConfig) (*AzureOpenAIClient, error) {

	configImpl := impl.DefaultAzureConfig(config.APIKey, config.Endpoint)
	configImpl.AzureModelMapperFunc = func(modelId string) string { return modelId }

	client := &AzureOpenAIClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	return client, nil
}

func (c *AzureOpenAIClient) Chat(messages []message.Message) (*message.Message, error) {

	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
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
		return nil, err
	}
	result := message.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}
	return &result, nil
}
