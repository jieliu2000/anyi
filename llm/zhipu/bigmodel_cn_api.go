// Use the openai compatible interface to access the bigmodel.cn service.
// See [bigmodel.cn Document] for details.
// [bigmodel.cn Document]: https://open.bigmodel.cn/dev/api#openai_sdk
package zhipu

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://open.bigmodel.cn/api/paas/v4/"
)

type ZhiPuModelConfig struct {
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseUrl string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type ZhipuClient struct {
	Config     *ZhiPuModelConfig
	clientImpl *impl.Client
}

// Creats a default bigmodel.cn model config.
func DefaultConfig(apiKey string, model string) *ZhiPuModelConfig {
	return &ZhiPuModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *ZhiPuModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &ZhiPuModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: baseUrl,
	}
}

func NewClient(config *ZhiPuModelConfig) (*ZhipuClient, error) {

	// Check if the config is nil to prevent panic or unexpected behavior
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Create a new default configuration implementation using the provided API key
	configImpl := impl.DefaultConfig(config.APIKey)

	// Set the BaseURL from the provided config
	configImpl.BaseURL = config.BaseUrl

	// Create a new ZhipuClient using the provided config and the configured client implementation
	client := &ZhipuClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	// Return the newly created ZhipuClient and nil error
	return client, nil
}

func (c *ZhipuClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	return openai.ExecuteChatWithFunctions(client, c.Config.Model, messages, functions, options)
}

func (c *ZhipuClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {

	info := chat.ResponseInfo{}
	// Check if the client implementation is initialized
	client := c.clientImpl
	if client == nil {
		return nil, info, errors.New("client not initialized")
	}

	// Convert the messages to OpenAI ChatMessages format
	messagesInput := openai.ConvertToOpenAIChatMessages(messages)

	// Create a ChatCompletion request using the client and the converted messages
	resp, err := client.CreateChatCompletion(
		context.Background(),
		impl.ChatCompletionRequest{
			Model:    c.Config.Model,
			Messages: messagesInput,
		},
	)

	// Check if there was an error in creating the ChatCompletion
	if err != nil {
		return nil, info, err
	}

	// Check if there are no choices in the response
	if len(resp.Choices) == 0 {
		return nil, info, errors.New("no choices found in the response")
	}

	// Extract the first choice from the response and create a new message object
	result := chat.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}

	info.CompletionTokens = resp.Usage.CompletionTokens
	info.PromptTokens = resp.Usage.PromptTokens

	// Return the new message object and nil error
	return &result, info, nil
}
