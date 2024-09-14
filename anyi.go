package anyi

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

type anyiRegistry struct {
	Clients    map[string]llm.Client
	Flows      map[string]*flow.Flow
	Validators map[string]flow.StepValidator
	Executors  map[string]flow.StepExecutor
	Formatters map[string]chat.PromptFormatter
}

var GlobalRegistry *anyiRegistry = &anyiRegistry{
	Clients:    make(map[string]llm.Client),
	Flows:      make(map[string]*flow.Flow),
	Validators: make(map[string]flow.StepValidator),
	Executors:  make(map[string]flow.StepExecutor),
	Formatters: make(map[string]chat.PromptFormatter),
}

// RegisterDefaultClient registers the default client to the global registry.
// Parameters:
// - client llm.Client: The client to be registered as the default client.
func RegisterDefaultClient(client llm.Client) {
	GlobalRegistry.Clients["default"] = client
}

// GetDefaultClient function retrieves the default client from the Anyi global registry. A default client is a client that meets any of the following conditions:
// - It has been Set to the global Anyi instance with name "default"
// - There is only one client in the registry. Then this client will be the default client and returned.
//
// If no client is found, it returns an error indicating that no default client was found.
func GetDefaultClient() (llm.Client, error) {
	client, ok := GlobalRegistry.Clients["default"]
	if !ok {
		if len(GlobalRegistry.Clients) == 1 {
			for _, client = range GlobalRegistry.Clients {
				return client, nil
			}
		}
		return nil, errors.New("no default client found")
	}
	return client, nil
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
	validatorType := GlobalRegistry.Validators[name]
	if validatorType == nil {
		return nil, errors.New("no validator found with the given name: " + name)
	}
	return validatorType, nil
}

func GetExecutor(name string) (flow.StepExecutor, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	executorType := GlobalRegistry.Executors[name]
	if executorType == nil {
		return nil, errors.New("no executor found with the given name: " + name)
	}
	return executorType, nil
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

func NewMessage(role string, content string) chat.Message {
	return chat.Message{
		Role:    role,
		Content: content,
	}
}

func NewFlowContext(input string) *flow.FlowContext {
	flowContext := flow.FlowContext{
		Text: input,
	}

	return &flowContext
}

func GetFormatter(name string) chat.PromptFormatter {
	return GlobalRegistry.Formatters[name]
}

func RegisterFormatter(name string, formatter chat.PromptFormatter) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	GlobalRegistry.Formatters[name] = formatter
	return nil
}

func NewPromptTemplateFormatterFromFile(name string, templateFile string) (*chat.PromptyTemplateFormatter, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	formatter, err := chat.NewPromptTemplateFormatterFromFile(templateFile)

	if err != nil {
		return nil, err
	}
	err = RegisterFormatter(name, formatter)
	return formatter, err
}

func NewPromptTemplateFormatter(name string, template string) (*chat.PromptyTemplateFormatter, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	formatter, err := chat.NewPromptTemplateFormatter(template)
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
	if len(steps) == 0 {
		return nil, errors.New("no steps provided")
	}

	f, err := flow.NewFlow(client, name, steps...)

	if err != nil {
		return nil, err
	}

	GlobalRegistry.Flows[name] = f
	return f, nil
}

// RegisterExecutor function registers a StepExecutor to the global registry with a specified name.
// Note here you can simply pass an empty StepExecutor instance to the GlobalRegistry. The executors are used by steps. You can config the properties of the executors in step config or actual execution.
// Parameters:
// - name string: The name of the executor to be registered.
// - executor flow.StepExecutor: The executor to be registered.
// Return value:
// - error: If an error occurs during registration, the corresponding error message is returned.
func RegisterExecutor(name string, executor flow.StepExecutor) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if GlobalRegistry.Executors[name] != nil {
		return fmt.Errorf("executor type with the name %s already exists", name)
	}

	GlobalRegistry.Executors[name] = executor
	return nil
}

// RegisterValidator registers a validator with a given name to the global registry.
// Note here you can simply pass an empty StepValidator instance to the GlobalRegistry. The validators are used by steps. You can config the properties of the validators in step config or actual validation.
// Parameters:
// - name string: The name of the validator to be registered.
// - validator flow.StepValidator: The validator to be registered.
// Return value:
// - error: If an error occurs during registration, the corresponding error message is returned.
func RegisterValidator(name string, validator flow.StepValidator) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if GlobalRegistry.Validators[name] != nil {
		return fmt.Errorf("validator type with the name %s already exists", name)
	}
	GlobalRegistry.Validators[name] = validator
	return nil
}

func NewLLMStepExecutorWithFormatter(name string, formatter *chat.PromptyTemplateFormatter, systemMessage string, client llm.Client) *LLMStepExecutor {

	stepExecutor := LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}

	RegisterExecutor(name, &stepExecutor)
	return &stepExecutor
}

func NewLLMStep(tmplate string, systemMessage string, client llm.Client) (*flow.Step, error) {
	return NewLLMStepWithTemplate(tmplate, systemMessage, client)
}

func Init() {

	log.Debug("Initializing Anyi...")
	RegisterExecutor("llm", &LLMStepExecutor{})
	RegisterValidator("string", &StringValidator{})
	log.Debug("Anyi initialized successfully.")
}
