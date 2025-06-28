// Use the openai compatible interface to access the bigmodel.cn service.
// See [bigmodel.cn Document] for details.
// [bigmodel.cn Document]: https://open.bigmodel.cn/dev/api#openai_sdk
package zhipu

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://open.bigmodel.cn/api/paas/v4/"

	// GLM-4 series - Latest and most popular
	GLM4Plus  = "glm-4-plus"
	GLM4      = "glm-4"
	GLM4Air   = "glm-4-air"
	GLM4AirX  = "glm-4-airx"
	GLM4Flash = "glm-4-flash"
	GLM4Long  = "glm-4-long"

	// GLM-4V series - Vision models
	GLM4V     = "glm-4v"
	GLM4VPlus = "glm-4v-plus"

	// GLM-3 series - Legacy
	GLM3Turbo = "glm-3-turbo"

	// Code generation models
	CodeGeeX4 = "codegeex-4"

	// Embedding models
	Embedding2 = "embedding-2"
	Embedding3 = "embedding-3"

	// Default model - using the most popular GLM-4-Flash
	DefaultModel = "glm-4-flash"
)

type ZhiPuModelConfig struct {
	config.GeneralLLMConfig
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
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *ZhiPuModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &ZhiPuModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          baseUrl,
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

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChatWithFunctions(client, c.Config.Model, messages, functions, options, openaiConfig)
}

func (c *ZhipuClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChat(client, c.Config.Model, messages, options, openaiConfig)
}
