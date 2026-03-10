// Use the openai compatible interface to access the DashScope service.
// See [Dashscope Document] for details.
// [Dashscope Document]: https://help.aliyun.com/en/dashscope/developer-reference/compatibility-of-openai-with-dashscope/
package dashscope

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"

	// Qwen3 series - Latest generation (2025)
	Qwen3Max       = "qwen3-max"           // Latest flagship model released Sep 24, 2025
	Qwen3Plus      = "qwen3-plus"
	Qwen3Turbo     = "qwen3-turbo"
	Qwen3Coder     = "qwen3-coder"
	Qwen3CoderPlus = "qwen3-coder-plus"    // Enhanced code model
	Qwen3Coder480B = "qwen3-coder-480b-a35b-instruct" // MoE model with 480B parameters, 35B active
	Qwen3Omni      = "qwen3-omni"          // Open-source multimodal model

	// Qwen2.5 series - Previous generation but still widely used
	Qwen25Max   = "qwen-max"
	Qwen25Plus  = "qwen-plus"
	Qwen25Turbo = "qwen-turbo"
	Qwen2572B   = "qwen2.5-72b-instruct"
	Qwen2532B   = "qwen2.5-32b-instruct"
	Qwen2514B   = "qwen2.5-14b-instruct"
	Qwen257B    = "qwen2.5-7b-instruct"
	Qwen253B    = "qwen2.5-3b-instruct"
	Qwen251B    = "qwen2.5-1.5b-instruct"
	Qwen250B    = "qwen2.5-0.5b-instruct"

	// Qwen2.5-Coder series
	Qwen25Coder32B = "qwen2.5-coder-32b-instruct"
	Qwen25Coder14B = "qwen2.5-coder-14b-instruct"
	Qwen25Coder7B  = "qwen2.5-coder-7b-instruct"
	Qwen25Coder1B  = "qwen2.5-coder-1.5b-instruct"

	// Qwen2.5-Math series
	Qwen25Math72B = "qwen2.5-math-72b-instruct"
	Qwen25Math7B  = "qwen2.5-math-7b-instruct"
	Qwen25Math1B  = "qwen2.5-math-1.5b-instruct"

	// Qwen-VL series - Vision models
	QwenVLMax  = "qwen-vl-max"
	QwenVLPlus = "qwen-vl-plus"
	QwenVL     = "qwen-vl-v1"

	// Qwen-Audio series
	QwenAudio = "qwen-audio-turbo"

	// Legacy Qwen2 series
	Qwen272B       = "qwen2-72b-instruct"
	Qwen257BLegacy = "qwen2-7b-instruct"
	Qwen21B        = "qwen2-1.5b-instruct"
	Qwen20B        = "qwen2-0.5b-instruct"

	// Default model - using the latest Qwen3-Max for best performance
	DefaultModel = "qwen3-max"
)

type DashScopeModelConfig struct {
	config.GeneralLLMConfig
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseUrl string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type DashScopeClient struct {
	Config     *DashScopeModelConfig
	clientImpl *impl.Client
}

// Creats a default DashScope model config.
func DefaultConfig(apiKey string, model string) *DashScopeModelConfig {
	return &DashScopeModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *DashScopeModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &DashScopeModelConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          baseUrl,
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

func (c *DashScopeClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChatWithFunctions(client, c.Config.Model, messages, functions, options, openaiConfig)
}

func (c *DashScopeClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChat(client, c.Config.Model, messages, options, openaiConfig)
}
