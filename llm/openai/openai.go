package openai

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/message"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-3.5-turbo"
)

type OpenAIModelConfig struct {
	APIKey            string `json:"apiKey" yaml:"apiKey" mapstructure:"apiKey"`
	BaseURL           string `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"`
	Model             string `json:"model" yaml:"model" mapstructure:"model"`
	AllowInsecureHttp bool   `json:"allowInsecureHttp" yaml:"allowInsecureHttp" mapstructure:"allowInsecureHttp"`
}

type OpenAIClient struct {
	Config     *OpenAIModelConfig
	clientImpl *azopenai.Client
}

func DefaultConfig(apiKey string) *OpenAIModelConfig {
	return NewConfig(apiKey, "", "")
}

func NewConfigWithModel(apiKey string, model string) *OpenAIModelConfig {
	return NewConfig(apiKey, model, "")
}

// Create a new config with the given API, model, and baseURL
// If you don't know the baseURL or model, you can leave them as blank string. The function will use default values if they are not provided.
func NewConfig(apiKey string, model string, baseURL string) *OpenAIModelConfig {
	if model == "" {
		model = DefaultModel
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &OpenAIModelConfig{APIKey: apiKey, Model: model, BaseURL: baseURL}
}

func NewClient(config *OpenAIModelConfig) (*OpenAIClient, error) {

	if config == nil {
		return nil, errors.New("config cannot be null")
	}

	// Create a new default configuration implementation using the provided API key
	keyCredential := azcore.NewKeyCredential(config.APIKey)

	options := &azopenai.ClientOptions{}
	options.InsecureAllowCredentialWithHTTP = config.AllowInsecureHttp
	// Set the BaseURL from the provided config
	clientImpl, err := azopenai.NewClientForOpenAI(config.BaseURL, keyCredential, options)

	if err != nil {
		return nil, err
	}

	client := &OpenAIClient{
		Config:     config,
		clientImpl: clientImpl,
	}

	return client, nil
}

func (c *OpenAIClient) Chat(messages []message.Message) (*message.Message, error) {
	// Check if the client implementation is initialized
	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
	}

	// Convert the messages to OpenAI ChatMessages format
	messagesInput := azureopenai.ConvertToAzureOpenAIMessageCompletions(messages)

	// Create a ChatCompletion request using the client and the converted messages

	resp, err := client.GetChatCompletions(
		context.Background(),
		azopenai.ChatCompletionsOptions{
			DeploymentName: &c.Config.Model,
			Messages:       messagesInput,
		}, nil)

	if err != nil {
		return nil, err
	}
	result := message.Message{
		Content: *resp.Choices[0].Message.Content,
		Role:    string(*resp.Choices[0].Message.Role),
	}
	// Return the new message object and nil error
	return &result, nil
}
