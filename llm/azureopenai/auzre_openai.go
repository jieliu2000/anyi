package azureopenai

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/message"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type AzureOpenAIModelConfig struct {
	APIKey            string `json:"api_key"`
	ModelDeploymentId string `json:"model_deployment_id"`
	Endpoint          string `json:"endpoint"`
}

type AzureOpenAIClient struct {
	Config     *AzureOpenAIModelConfig
	clientImpl *azopenai.Client
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

	keyCredential := azcore.NewKeyCredential(config.APIKey)

	// In Azure OpenAI you must deploy a model before you can use it in your client. For more information
	// see here: https://learn.microsoft.com/azure/cognitive-services/openai/how-to/create-resource
	clientImpl, err := azopenai.NewClientWithKeyCredential(config.Endpoint, keyCredential, nil)

	if err != nil {
		return nil, err
	}
	client := &AzureOpenAIClient{
		Config:     config,
		clientImpl: clientImpl,
	}

	return client, nil
}

func (c *AzureOpenAIClient) Chat(messages []message.Message) (*message.Message, error) {

	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
	}

	messagesInput := ConvertToAzureOpenAIMessageCompletions(messages)

	resp, err := client.GetChatCompletions(
		context.Background(),
		azopenai.ChatCompletionsOptions{
			DeploymentName: &c.Config.ModelDeploymentId,
			Messages:       messagesInput,
		}, nil)

	if err != nil {
		return nil, err
	}
	result := message.Message{
		Content: *resp.Choices[0].Message.Content,
		Role:    string(*resp.Choices[0].Message.Role),
	}
	return &result, nil
}
