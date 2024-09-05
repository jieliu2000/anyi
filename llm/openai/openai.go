package openai

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
	impl "github.com/sashabaranov/go-openai"
)

const (
	//Copied from https://github.com/sashabaranov/go-openai/blob/master/completion.go.
	GPT432K0613           = "gpt-4-32k-0613"
	GPT432K0314           = "gpt-4-32k-0314"
	GPT432K               = "gpt-4-32k"
	GPT40613              = "gpt-4-0613"
	GPT40314              = "gpt-4-0314"
	GPT4o                 = "gpt-4o"
	GPT4o20240513         = "gpt-4o-2024-05-13"
	GPT4o20240806         = "gpt-4o-2024-08-06"
	GPT4oMini             = "gpt-4o-mini"
	GPT4oMini20240718     = "gpt-4o-mini-2024-07-18"
	GPT4Turbo             = "gpt-4-turbo"
	GPT4Turbo20240409     = "gpt-4-turbo-2024-04-09"
	GPT4Turbo0125         = "gpt-4-0125-preview"
	GPT4Turbo1106         = "gpt-4-1106-preview"
	GPT4TurboPreview      = "gpt-4-turbo-preview"
	GPT4VisionPreview     = "gpt-4-vision-preview"
	GPT4                  = "gpt-4"
	GPT3Dot5Turbo0125     = "gpt-3.5-turbo-0125"
	GPT3Dot5Turbo1106     = "gpt-3.5-turbo-1106"
	GPT3Dot5Turbo0613     = "gpt-3.5-turbo-0613"
	GPT3Dot5Turbo0301     = "gpt-3.5-turbo-0301"
	GPT3Dot5Turbo16K      = "gpt-3.5-turbo-16k"
	GPT3Dot5Turbo16K0613  = "gpt-3.5-turbo-16k-0613"
	GPT3Dot5Turbo         = "gpt-3.5-turbo"
	GPT3Dot5TurboInstruct = "gpt-3.5-turbo-instruct"
	GPT3Davinci           = "davinci"
	GPT3Davinci002        = "davinci-002"
	GPT3Curie             = "curie"
	GPT3Curie002          = "curie-002"
	GPT3Ada002            = "ada-002"
	GPT3Babbage002        = "babbage-002"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-3.5-turbo"
)

type OpenAIModelConfig struct {
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseURL string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type OpenAIClient struct {
	Config     *OpenAIModelConfig
	clientImpl *impl.Client
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

	return ExecuteChatWithFunctions(client, c.Config.Model, messages, functions, options)
}

func (c *OpenAIClient) Chat(messages []chat.Message, options *chat.ChatOptions) (message *chat.Message, responseInfo chat.ResponseInfo, err error) {
	if c.clientImpl == nil {
		return nil, responseInfo, errors.New("client not initialized")
	}

	messagesInput := ConvertToOpenAIChatMessages(messages)
	request := impl.ChatCompletionRequest{
		Model:    c.Config.Model,
		Messages: messagesInput,
	}

	resp, err := c.clientImpl.CreateChatCompletion(
		context.Background(),
		request,
	)
	if err != nil {
		return nil, responseInfo, err
	}

	if len(resp.Choices) == 0 {
		return nil, responseInfo, errors.New("no chat completion choices returned")
	}

	responseInfo.CompletionTokens = resp.Usage.CompletionTokens
	responseInfo.PromptTokens = resp.Usage.PromptTokens

	return &chat.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}, responseInfo, nil
}
