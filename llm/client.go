package llm

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"
	"github.com/mitchellh/mapstructure"
	config "github.com/spf13/viper"
)

// ClientConfig is the configuration for a client. In Anyi, this struct is mainly used for reading the client config file. The config file can be in any formats that viper (https://github.com/spf13/viper) supports.
// If you create clients based on programmed ModelConfig then you don't need to use this struct.
// The function NewModelConfigFromFile is provided to help you read the config file and convert it to corresponding ModelConfig.
type ClientConfig struct {
	// The model to use. Currently, it supports these values:
	// "openai" - OpenAI model
	//"azureopenai" - Azure OpenAI model
	//"dashscope" - DashScope model
	Model string

	// The model config. The type of this field depends on the model. We define this property as map[string]interface{} for extensibility.
	// You can refer to the ModelConfig type of your model to see what properties you need to define hee.
	// For example, for openai, you need to define properties based on openai.OpenAIModelConfig struct.
	Config map[string]interface{}
}

type ModelConfig interface {
}

type Client interface {
	Init() error
	Chat(messages []message.Message) (*message.Message, error)
}

func readConfigFile(configFile string) (*ClientConfig, error) {
	config.SetConfigFile(configFile)

	err := config.ReadInConfig() // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		return nil, err
	}

	clientConfig := ClientConfig{}
	err = config.Unmarshal(&clientConfig)
	if err != nil {

		return nil, err
	}
	return &clientConfig, nil
}

func NewModelConfigFromFile(configFile string) (ModelConfig, error) {
	clientConfig, err := readConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	var modelConfig ModelConfig
	switch clientConfig.Model {
	case "openai":
		modelConfig = &openai.OpenAIModelConfig{}
	case "azureopenai":
		modelConfig = &azureopenai.AzureOpenAIModelConfig{}
	case "dashscope":
		modelConfig = &dashscope.DashScopeModelConfig{}
	default:
		return nil, errors.New("unknown model")
	}

	err = mapstructure.Decode(clientConfig.Config, modelConfig)
	return modelConfig, err
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

	}
	return nil, errors.New("unknown model config")
}

// NewClientFromConfigFile creates a new client based on the model config file.
// The @configFile parameter is the path to the model config file. Anyi reads config file using viper (https://github.com/spf13/viper) library.
// Refer to the ClientConfig struct on what contents can be speified in the config file.
func NewClientFromConfigFile(configFile string) (Client, error) {

	config, err := NewModelConfigFromFile(configFile)
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}
