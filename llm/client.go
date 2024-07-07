package llm

import (
	"errors"

	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/openai"
)

type ModelConfig interface {
}

type Client interface {
	Init() error
	Chat(messages []chat.Message) (*chat.Message, error)
}

func NewClient(config ModelConfig) (Client, error) {

	//lint:ignore S1034 config variable will be used in future so we ignore this linter for now
	switch config.(type) {

	case *openai.OpenAIModelConfig:
		return openai.NewClient(config.(*openai.OpenAIModelConfig))

	case *azureopenai.AzureOpenAIModelConfig:
		return azureopenai.NewClient(config.(*azureopenai.AzureOpenAIModelConfig))

	}
	return nil, errors.New("unknown model config")
}
