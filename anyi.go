package anyi

import (
	"errors"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/message"
)

type anyiRegistry struct {
	Clients    map[string]llm.Client
	Flows      map[string]*flow.Flow
	Validators map[string]flow.StepValidator
	Executors  map[string]flow.StepExecutor
	Formatters map[string]message.PromptFormatter

	executorTypes map[string]flow.StepExecutor
}

var GlobalRegistry *anyiRegistry = &anyiRegistry{
	Clients:       make(map[string]llm.Client),
	Flows:         make(map[string]*flow.Flow),
	Validators:    make(map[string]flow.StepValidator),
	Executors:     make(map[string]flow.StepExecutor),
	Formatters:    make(map[string]message.PromptFormatter),
	executorTypes: make(map[string]flow.StepExecutor),
}

// The function creates a new client based on the given configuration and, if a non-empty name is provided, Set that client to the global Anyi instance.
// The name is used to identify the client in Anyi. After a client is Seted to Anyi with a name, you can access it by calling [GetClient].
// Please note that if the name is empty but the config is valid, the client will still be created but it won't be Seted to Anyi. No error will be returned in this case.
// If the config is invalid, an error will be returned.
func NewClient(name string, model llm.ModelConfig) (llm.Client, error) {
	client, err := llm.NewClient(model)
	if err != nil {
		return nil, err
	}
	// If name is not empty, Set the client to Anyi.Clients
	if name != "" {
		GlobalRegistry.Clients[name] = client
	}
	return client, nil
}

func RegisterFlow(name string, flow *flow.Flow) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.Flows[name] = flow
	return nil
}

func GetFlow(name string) (*flow.Flow, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	f, ok := GlobalRegistry.Flows[name]
	if !ok {
		return nil, errors.New("no flow found with the given name: " + name)
	}
	return f, nil
}

// The function Sets a client to the global Anyi instance.
// If the client or name is nil, an error will be returned.
func RegisterClient(name string, client llm.Client) error {
	if client == nil {
		return errors.New("client cannot be empty")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.Clients[name] = client
	return nil
}

func GetValidator(name string) (flow.StepValidator, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	return GlobalRegistry.Validators[name], nil
}

func GetExecutor(name string) (flow.StepExecutor, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	return GlobalRegistry.Executors[name], nil
}

func GetClient(name string) (llm.Client, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	client, ok := GlobalRegistry.Clients[name]
	if !ok {
		return nil, errors.New("no client found with the given name: " + name)
	}
	return client, nil
}

// NewClientFromConfigFile creates a new client based on the model config file.
// The configFile parameter is the path to the model config file. Anyi reads config file using [viper] library.
// The name parameter is used to identify the client in Anyi. After a client is Seted to Anyi with a name, you can access it by calling Anyi.GetClient(name).
// Please note that if the name is empty but the config is valid, the client will still be created but it won't be Seted to Anyi. No error will be returned in this case.
// If the config is invalid, an error will be returned.
//
// [viper]: https://github.com/spf13/viper
func NewClientFromConfigFile(name string, configFile string) (llm.Client, error) {
	client, err := llm.NewClientFromConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	// If name is not empty, Set the client to Anyi.Clients
	if name != "" {
		GlobalRegistry.Clients[name] = client
	}
	return client, nil
}

func NewMessage(role string, content string) message.Message {
	return message.Message{
		Role:    role,
		Content: content,
	}
}

func NewContext(input string) *flow.FlowContext {
	context := flow.FlowContext{
		Context: input,
	}

	return &context
}

func GetFormatter(name string) message.PromptFormatter {
	return GlobalRegistry.Formatters[name]
}

func RegisterFormatter(name string, formatter message.PromptFormatter) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.Formatters[name] = formatter
	return nil
}

func NewPromptTemplateFormatterFromFile(name string, templateFile string) (*message.PromptyTemplateFormatter, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	formatter, err := message.NewPromptTemplateFormatterFromFile(templateFile)

	if err != nil {
		return nil, err
	}
	err = RegisterFormatter(name, formatter)
	return formatter, err
}

func NewPromptTemplateFormatter(name string, template string) (*message.PromptyTemplateFormatter, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	formatter, err := message.NewPromptTemplateFormatter(template)
	if err != nil {
		return nil, err
	}
	err = RegisterFormatter(name, formatter)
	return formatter, err
}

func NewFlow(name string, client llm.Client, steps ...flow.Step) (*flow.Flow, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	f, err := flow.NewFlow(client, name, steps...)

	if err != nil {
		return nil, err
	}

	GlobalRegistry.Flows[name] = f
	return f, nil
}

func DefineExecutorType(typeName string, executor flow.StepExecutor) error {
	if typeName == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.executorTypes[typeName] = executor
	return nil
}

func GetExecutorType(typeName string) flow.StepExecutor {
	return GlobalRegistry.executorTypes[typeName]
}

func RegisterExecutor(name string, executor flow.StepExecutor) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.Executors[name] = executor
	return nil
}

func RegisterValidator(name string, validator flow.StepValidator) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.Validators[name] = validator
	return nil
}

func NewLLMStepExecutorWithFormatter(name string, formatter *message.PromptyTemplateFormatter, systemMessage string, client llm.Client) *flow.LLMStepExecutor {

	stepExecutor := flow.LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}

	RegisterExecutor(name, &stepExecutor)
	return &stepExecutor
}

func NewLLMStep(tmplate string, systemMessage string, client llm.Client) (*flow.Step, error) {
	return flow.NewLLMStepWithTemplate(tmplate, systemMessage, client)
}

func Init() {
	DefineExecutorType("llm", &flow.LLMStepExecutor{})
}
