package anyi

import (
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

// NewClient creates a new client based on the model config. The type of client is determined by the type of model config.
// For example, if you pass in an OpenAIModelConfig, it will return a new OpenAIClient.
func NewClient(config llm.ModelConfig) (llm.Client, error) {

	return llm.NewClient(config)
}

// NewClientFromConfigFile creates a new client based on the model config file.
// The @configFile parameter is the path to the model config file. Anyi reads config file using viper (https://github.com/spf13/viper) library.
func NewClientFromConfigFile(configFile string) (llm.Client, error) {
	return llm.NewClientFromConfigFile(configFile)
}

func NewMessage(role string, content string) message.Message {
	return message.Message{
		Role:    role,
		Content: content,
	}
}

func NewMessageTemplateFormatter(templateFile string) (*message.MessageTemplateFormatter, error) {
	return message.NewMessageTemplateFormatterFromFile(templateFile)
}

func NewFlow(client llm.Client, name string, steps ...FlowStep) *Flow {
	return &Flow{Steps: steps, Name: name, clientImpl: client}
}

func NewFlowFromConfigFile(configfile string) {
	//TODO
}
