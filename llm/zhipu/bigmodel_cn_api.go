// Use the openai compatible interface to access the bigmodel.cn service.
// See [bigmodel.cn Document] for details.
// [bigmodel.cn Document]: https://open.bigmodel.cn/dev/api#openai_sdk
package zhipu

import (
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
	client := c.clientImpl

	return openai.ExecuteChat(client, c.Config.Model, messages, options)
}
