package siliconcloud

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/config"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/tools"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://api.siliconflow.cn/v1"

	// Qwen3 series (Latest generation 2025)
	Qwen3Max       = "Qwen/Qwen3-Max"           // Latest flagship model released Sep 24, 2025
	Qwen3Plus      = "Qwen/Qwen3-Plus"
	Qwen3Turbo     = "Qwen/Qwen3-Turbo"
	Qwen3Coder     = "Qwen/Qwen3-Coder"
	Qwen3_235B     = "Qwen/Qwen3-235B-A22B"     // MoE model with 235B parameters, 22B active

	// Qwen2.5 series (Previous generation but still widely used)
	Qwen25Max72B  = "Qwen/Qwen2.5-Max"
	Qwen2572B     = "Qwen/Qwen2.5-72B-Instruct"
	Qwen2532B     = "Qwen/Qwen2.5-32B-Instruct"
	Qwen2514B     = "Qwen/Qwen2.5-14B-Instruct"
	Qwen257B      = "Qwen/Qwen2.5-7B-Instruct"
	Qwen25Coder7B = "Qwen/Qwen2.5-Coder-7B-Instruct"

	// DeepSeek series (Latest generation 2025)
	DeepSeekR1          = "deepseek-ai/DeepSeek-R1"
	DeepSeekR1Distill   = "deepseek-ai/DeepSeek-R1-Distill-Qwen-7B"
	DeepSeekV3          = "deepseek-ai/DeepSeek-V3"
	DeepSeekV31         = "deepseek-ai/DeepSeek-V3.1"
	DeepSeekV31Terminus = "deepseek-ai/DeepSeek-V3.1-Terminus" // Latest model released Sep 22, 2025
	DeepSeekChat        = "deepseek-ai/deepseek-chat"
	DeepSeekCoder       = "deepseek-ai/deepseek-coder-6.7b-instruct"
	DeepSeekMath        = "deepseek-ai/deepseek-math-7b-instruct"

	// Meta Llama4 series (Latest generation 2025)
	Llama4Maverick = "meta-llama/Llama-4-Maverick-Instruct"
	Llama4Scout    = "meta-llama/Llama-4-Scout-Instruct"
	Llama4_70B     = "meta-llama/Llama-4-70B-Instruct"
	Llama4_40B     = "meta-llama/Llama-4-40B-Instruct"

	// Meta Llama3 series (Previous generation)
	Llama32_90B  = "meta-llama/Llama-3.2-90B-Vision-Instruct"
	Llama32_11B  = "meta-llama/Llama-3.2-11B-Vision-Instruct"
	Llama32_3B   = "meta-llama/Llama-3.2-3B-Instruct"
	Llama32_1B   = "meta-llama/Llama-3.2-1B-Instruct"
	Llama31_405B = "meta-llama/Meta-Llama-3.1-405B-Instruct"
	Llama31_70B  = "meta-llama/Meta-Llama-3.1-70B-Instruct"
	Llama31_8B   = "meta-llama/Meta-Llama-3.1-8B-Instruct"

	// Google Gemma2 series
	Gemma2_27B = "google/gemma-2-27b-it"
	Gemma2_9B  = "google/gemma-2-9b-it"

	// Mistral series
	Mistral7B     = "mistralai/Mistral-7B-Instruct-v0.3"
	MistralLarge  = "mistralai/Mistral-Large-Instruct"
	MixtralMoE    = "mistralai/Mixtral-8x7B-Instruct-v0.1"
	MixtralMoE22B = "mistralai/Mixtral-8x22B-Instruct-v0.1"

	// Yi series
	Yi15_34B = "01-ai/Yi-1.5-34B-Chat"
	Yi15_9B  = "01-ai/Yi-1.5-9B-Chat"
	Yi15_6B  = "01-ai/Yi-1.5-6B-Chat"

	// GLM series
	GLM4_9B = "THUDM/glm-4-9b-chat"

	// Internlm series
	Internlm25_20B = "internlm/internlm2_5-20b-chat"
	Internlm25_7B  = "internlm/internlm2_5-7b-chat"

	// Default model - using the latest Qwen3-Max for best performance
	DefaultModel = "Qwen/Qwen3-Max"
)

type SiliconCloudConfig struct {
	config.GeneralLLMConfig
	APIKey  string `json:"apiKey" mapstructure:"apiKey"`
	BaseUrl string `json:"baseUrl" mapstructure:"baseUrl"`
	Model   string `json:"model" mapstructure:"model"`
}

type SiliconCloud struct {
	Config     *SiliconCloudConfig
	clientImpl *impl.Client
}

// Creats a default bigmodel.cn model config.
func DefaultConfig(apiKey string, model string) *SiliconCloudConfig {
	return &SiliconCloudConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *SiliconCloudConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &SiliconCloudConfig{
		GeneralLLMConfig: config.DefaultGeneralConfig(),
		APIKey:           apiKey,
		Model:            model,
		BaseUrl:          baseUrl,
	}
}

func NewClient(config *SiliconCloudConfig) (*SiliconCloud, error) {

	// Check if the config is nil to prevent panic or unexpected behavior
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Create a new default configuration implementation using the provided API key
	configImpl := impl.DefaultConfig(config.APIKey)

	// Set the BaseURL from the provided config
	configImpl.BaseURL = config.BaseUrl

	// Create a new ZhipuClient using the provided config and the configured client implementation
	client := &SiliconCloud{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	// Return the newly created ZhipuClient and nil error
	return client, nil
}

func (c *SiliconCloud) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChatWithFunctions(client, c.Config.Model, messages, functions, options, openaiConfig)
}

func (c *SiliconCloud) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	client := c.clientImpl

	// Create a temporary config object compatible with OpenAIModelConfig
	openaiConfig := &openai.OpenAIModelConfig{
		GeneralLLMConfig: c.Config.GeneralLLMConfig,
		Model:            c.Config.Model,
	}

	return openai.ExecuteChat(client, c.Config.Model, messages, options, openaiConfig)
}
