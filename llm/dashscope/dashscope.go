// Use the openai compatible interface to access the DashScope service.
// See [Dashscope Document] for details.
// [Dashscope Document]: https://help.aliyun.com/zh/dashscope/developer-reference/compatibility-of-openai-with-dashscope/?spm=a2c4g.11186623.0.0.17504ad0abpnzJ for details
package dashscope

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/chat"
	"github.com/jieliu2000/anyi/llm/openai"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

type DashScopeModelConfig struct {
	APIKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
	Model   string `json:"model"`
}

type DashScopeClient struct {
	Config     *DashScopeModelConfig
	clientImpl *impl.Client
}

// Creats a default DashScope model config.
func DefaultConfig(apiKey string, model string) *DashScopeModelConfig {
	return &DashScopeModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *DashScopeModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &DashScopeModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: baseUrl,
	}
}

func NewClient(config *DashScopeModelConfig) (*DashScopeClient, error) {

	// Check if the config is nil to prevent panic or unexpected behavior
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Create a new default configuration implementation using the provided API key
	configImpl := impl.DefaultConfig(config.APIKey)

	// Set the BaseURL from the provided config
	configImpl.BaseURL = config.BaseUrl

	// Create a new DashScopeClient using the provided config and the configured client implementation
	client := &DashScopeClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	// Return the newly created DashScopeClient and nil error
	return client, nil
}

func (c *DashScopeClient) Chat(messages []chat.Message, options chat.ChatOptions) (*chat.Message, error) {

	// Check if the client implementation is initialized
	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
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
		return nil, err
	}

	// Check if there are no choices in the response
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices found in the response")
	}

	// Extract the first choice from the response and create a new message object
	result := chat.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}

	// Return the new message object and nil error
	return &result, nil
}
