package llm

import (
	"errors"

	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/mitchellh/mapstructure"
)

// ClientConfig is the configuration for a client. In Anyi, this struct is mainly used for reading the client config file. The config file can be in any formats that [viper] supports.
// If you create clients based on programmed ModelConfig then you don't need to use this struct.
// The function NewModelConfigFromFile is provided to help you read the config file and convert it to corresponding ModelConfig.
//
// [viper]: https://github.com/spf13/viper
type ClientConfig struct {

	// The name of the client. This property is only used when you want anyi to have multiple client configurations and allows workflows/steps to configure clients via the name.
	// If you don't need to use multiple clients, you can ignore this property.
	Name string `mapstructure:"name" json:"name,omitempty"`

	// The type to use. Currently, it supports these values:
	//	* "openai" - OpenAI model
	//	* "azureopenai" - Azure OpenAI model
	//	* "dashscope" - DashScope model
	//	* "ollama" - Ollama model
	Type string `mapstructure:"type" json:"type"`

	// The model config. The type of this field depends on the model. We define this property as map[string]interface{} for extensibility.
	// You can refer to the ModelConfig type of your model to see what properties you need to define hee.
	// For example, for openai, you need to define properties based on openai.OpenAIModelConfig struct.
	Config map[string]interface{} `mapstructure:"config" json:"config"`
}

type ModelConfig interface {
}

type Client interface {
	Chat(messages []chat.Message, options chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error)
}

// NewModelConfigFromClientConfig creates a new ModelConfig instance based on the provided ClientConfig.
// Parameters:
// - clientConfig *ClientConfig: The ClientConfig object containing configuration information.
// Return values:
// - ModelConfig: The newly created ModelConfig instance.
// - error: An error object if any occurs during the process, nil otherwise.
func NewModelConfigFromClientConfig(clientConfig *ClientConfig) (ModelConfig, error) {
	if clientConfig == nil {
		return nil, errors.New("client config is null")
	}
	var modelConfig ModelConfig
	switch clientConfig.Type {
	case "openai":
		modelConfig = &openai.OpenAIModelConfig{}
	case "azureopenai":
		modelConfig = &azureopenai.AzureOpenAIModelConfig{}
	case "dashscope":
		modelConfig = &dashscope.DashScopeModelConfig{}
	case "ollama":
		modelConfig = &ollama.OllamaModelConfig{}
	default:
		return nil, errors.New("unknown model")
	}
	err := mapstructure.Decode(clientConfig.Config, modelConfig)
	return modelConfig, err
}

// NewModelConfigFromFile function creates a new ModelConfig object from a configuration file.
// Parameters:
// - configFile string: The path to the configuration file.
// Return values:
// - ModelConfig: The created ModelConfig object.
// - error: If an error occurs during the process, the corresponding error message is returned.
func NewModelConfigFromFile(configFile string) (ModelConfig, error) {
	clientConfig, err := utils.UnmarshallConfig(configFile, &ClientConfig{})
	if err != nil {
		return nil, err
	}
	return NewModelConfigFromClientConfig(clientConfig)
}

// NewClient creates a new client based on the model config. The type of client is determined by the type of model config.
// For example, if you pass in an OpenAIModelConfig, it will return a new OpenAIClient.
func NewClient(config ModelConfig) (Client, error) {

	//lint:ignore S1034 config variable will be used in future so we ignore this linter for now
	switch config.(type) {

	case *openai.OpenAIModelConfig:
		return openai.NewClient(config.(*openai.OpenAIModelConfig))

	case *azureopenai.AzureOpenAIModelConfig:
		return azureopenai.NewClient(config.(*azureopenai.AzureOpenAIModelConfig))

	case *dashscope.DashScopeModelConfig:
		return dashscope.NewClient(config.(*dashscope.DashScopeModelConfig))

	case *ollama.OllamaModelConfig:
		return ollama.NewClient(config.(*ollama.OllamaModelConfig))
	}
	return nil, errors.New("unknown model config")
}

// NewClientFromConfigFile creates a new client based on the model config file.
// The @configFile parameter is the path to the model config file. Anyi reads config file using [viper] library.
// Refer to the ClientConfig struct on what contents can be speified in the config file.
//
// [viper]: https://github.com/spf13/viper
func NewClientFromConfigFile(configFile string) (Client, error) {

	config, err := NewModelConfigFromFile(configFile)
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}
