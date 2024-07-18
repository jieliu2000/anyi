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

type ClientConfig struct {
	Model  string
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

func NewConfigFromConfigFile(configFile string) (ModelConfig, error) {
	clientConfig, err := readConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	switch clientConfig.Model {

	case "openai":
		openaiConfig := openai.OpenAIModelConfig{}
		err := mapstructure.Decode(clientConfig.Config, &openaiConfig)
		return &openaiConfig, err

	case "azureopenai":
		azureOpenAIConfig := azureopenai.AzureOpenAIModelConfig{}
		err := mapstructure.Decode(clientConfig.Config, &azureOpenAIConfig)
		return &azureOpenAIConfig, err

	case "dashscope":
		dashScopeConfig := dashscope.DashScopeModelConfig{}
		err := mapstructure.Decode(clientConfig.Config, &dashScopeConfig)
		return &dashScopeConfig, err
	}

	return nil, errors.New("unknown model")
}

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

func NewClientFromConfigFile(configFile string) (Client, error) {

	config, err := NewConfigFromConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}
