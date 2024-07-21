package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

type AnyiData struct {
	Clients map[string]llm.Client
	Flows   map[string]*Flow
}

var Anyi *AnyiData = &AnyiData{
	Clients: make(map[string]llm.Client),
	Flows:   make(map[string]*Flow),
}

// The function creates a new client based on the given configuration and, if a non-empty name is provided, add that client to the global Anyi instance.
// The name is used to identify the client in Anyi. After a client is added to Anyi with a name, you can access it by calling [GetClient].
// Please note that if the name is empty but the config is valid, the client will still be created but it won't be added to Anyi. No error will be returned in this case.
// If the config is invalid, an error will be returned.
func NewClient(config llm.ModelConfig, name string) (llm.Client, error) {
	client, err := llm.NewClient(config)
	if err != nil {
		return nil, err
	}
	// If name is not empty, add the client to Anyi.Clients
	if name != "" {
		Anyi.Clients[name] = client
	}
	return client, nil
}

// The function adds a client to the global Anyi instance.
// If the client or name is nil, an error will be returned.
func AddClient(client llm.Client, name string) error {
	if client == nil {
		return errors.New("client cannot be empty")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}
	Anyi.Clients[name] = client
	return nil
}

func GetClient(name string) (llm.Client, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	client, ok := Anyi.Clients[name]
	if !ok {
		return nil, errors.New("no client found with the given name: " + name)
	}
	return client, nil
}

// NewClientFromConfigFile creates a new client based on the model config file.
// The configFile parameter is the path to the model config file. Anyi reads config file using [viper] library.
// The name parameter is used to identify the client in Anyi. After a client is added to Anyi with a name, you can access it by calling Anyi.GetClient(name).
// Please note that if the name is empty but the config is valid, the client will still be created but it won't be added to Anyi. No error will be returned in this case.
// If the config is invalid, an error will be returned.
//
// [viper]: https://github.com/spf13/viper
func NewClientFromConfigFile(configFile string, name string) (llm.Client, error) {
	client, err := llm.NewClientFromConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	// If name is not empty, add the client to Anyi.Clients
	if name != "" {
		Anyi.Clients[name] = client
	}
	return client, nil
}

func NewMessage(role string, content string) message.Message {
	return message.Message{
		Role:    role,
		Content: content,
	}
}

func NewPromptTemplateFormatterFromFile(templateFile string) (*message.PromptyTemplateFormatter, error) {
	return message.NewPromptTemplateFormatterFromFile(templateFile)
}

func NewPromptTemplateFormatter(template string) (*message.PromptyTemplateFormatter, error) {
	return message.NewPromptTemplateFormatter(template)
}
