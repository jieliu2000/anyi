package anyi

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

// anyiRegistry is the central registry for all components in the Anyi framework.
// It stores clients, flows, validators, executors, and formatters for reuse across the application.
type anyiRegistry struct {
	mu                sync.RWMutex
	Clients           map[string]llm.Client
	Flows             map[string]*flow.Flow
	Validators        map[string]flow.StepValidator
	Executors         map[string]flow.StepExecutor
	Formatters        map[string]chat.PromptFormatter
	Agents            map[string]*agent.Agent  // Add agent registry
	defaultClientName string
}

// GlobalRegistry is the singleton instance of anyiRegistry.
// All components are registered and retrieved through this global registry.
var GlobalRegistry *anyiRegistry = &anyiRegistry{
	Clients:    make(map[string]llm.Client),
	Flows:      make(map[string]*flow.Flow),
	Validators: make(map[string]flow.StepValidator),
	Executors:  make(map[string]flow.StepExecutor),
	Formatters: make(map[string]chat.PromptFormatter),
	Agents:     make(map[string]*agent.Agent), // Initialize agent registry
}

// RegisterNewDefaultClient registers a client as the default client in the global registry.
// If no name is provided, it uses "default" as the client name.
//
// Parameters:
//   - name: Name to register the client under (uses "default" if empty)
//   - client: LLM client to register as the default
//
// Returns:
//   - Any error encountered during registration
func RegisterNewDefaultClient(name string, client llm.Client) error {
	if name == "" {
		name = "default"
	}
	err := RegisterClient(name, client)
	if err != nil {
		return err
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	GlobalRegistry.defaultClientName = name
	return nil
}

// SetDefaultClient sets the default client in the global registry.
//
// Parameters:
//   - name: Name of the client to set as default
//
// Returns:
//   - Any error encountered during setting the default client
func SetDefaultClient(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	GlobalRegistry.defaultClientName = name
	return nil
}

// GetDefaultClient retrieves the default client from the global registry.
// It returns the client registered as "default", or the only client if only one exists.
//
// Returns:
//   - The default LLM client
//   - An error if no default client is found
func GetDefaultClient() (llm.Client, error) {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	if GlobalRegistry.defaultClientName != "" {
		if client, ok := GlobalRegistry.Clients[GlobalRegistry.defaultClientName]; ok {
			return client, nil
		}
	}

	if len(GlobalRegistry.Clients) == 1 {
		for _, client := range GlobalRegistry.Clients {
			return client, nil
		}
	}

	if client, ok := GlobalRegistry.Clients["default"]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("no default client found (registered clients: %v)", getClientNames(GlobalRegistry.Clients))
}

func getClientNames(clients map[string]llm.Client) []string {
	names := make([]string, 0, len(clients))
	for name := range clients {
		names = append(names, name)
	}
	return names
}

// NewClient creates a new client from a model configuration and optionally registers it.
// If a name is provided, the client is registered in the global registry under that name.
//
// Parameters:
//   - name: Name to register the client under (optional, can be empty)
//   - model: Model configuration for the client
//
// Returns:
//   - A new LLM client
//   - Any error encountered during client creation
func NewClient(name string, model llm.ModelConfig) (llm.Client, error) {
	client, err := llm.NewClient(model)
	if err != nil {
		return nil, err
	}
	// If name is not empty, Set the client to Anyi.Clients
	if name != "" {
		// Use mutex to protect access to the global registry
		GlobalRegistry.mu.Lock()
		defer GlobalRegistry.mu.Unlock()

		GlobalRegistry.Clients[name] = client
	}
	return client, nil
}

// RegisterFlow registers a flow in the global registry.
// Each flow must have a unique name.
//
// Parameters:
//   - name: Name to register the flow under
//   - flow: Workflow to register
//
// Returns:
//   - Any error encountered during registration
func RegisterFlow(name string, flow *flow.Flow) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	if _, exists := GlobalRegistry.Flows[name]; exists {
		return fmt.Errorf("flow with name %q already exists", name)
	}

	GlobalRegistry.Flows[name] = flow
	return nil
}

// GetFlow retrieves a flow from the global registry by name.
//
// Parameters:
//   - name: Name of the flow to retrieve
//
// Returns:
//   - The requested workflow
//   - An error if the flow is not found
func GetFlow(name string) (*flow.Flow, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	f, ok := GlobalRegistry.Flows[name]
	if !ok {
		return nil, errors.New("no flow found with the given name: " + name)
	}
	return f, nil
}

// RegisterClient registers a client in the global registry.
// Each client must have a unique name.
//
// Parameters:
//   - name: Name to register the client under
//   - client: LLM client to register
//
// Returns:
//   - Any error encountered during registration
func RegisterClient(name string, client llm.Client) error {
	if client == nil {
		return errors.New("client cannot be empty")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	if _, exists := GlobalRegistry.Clients[name]; exists {
		return fmt.Errorf("client with name %q already exists", name)
	}

	GlobalRegistry.Clients[name] = client
	return nil
}

// GetValidator retrieves a validator from the global registry by name.
// It returns a new instance of the validator with the same configuration.
//
// Parameters:
//   - name: Name of the validator to retrieve
//
// Returns:
//   - A new instance of the requested validator
//   - An error if the validator is not found
func GetValidator(name string) (flow.StepValidator, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	validatorType := GlobalRegistry.Validators[name]
	if validatorType == nil {
		return nil, errors.New("no validator found with the given name: " + name)
	}

	val := reflect.ValueOf(validatorType)
	if val.Kind() == reflect.Ptr {
		elem := val.Elem()
		newVal := reflect.New(elem.Type())
		newVal.Elem().Set(elem)
		return newVal.Interface().(flow.StepValidator), nil
	}

	return validatorType, nil
}

// GetExecutor retrieves an executor from the global registry by name.
// It returns a new instance of the executor with the same configuration.
//
// Parameters:
//   - name: Name of the executor to retrieve
//
// Returns:
//   - A new instance of the requested executor
//   - An error if the executor is not found
func GetExecutor(name string) (flow.StepExecutor, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	executor := GlobalRegistry.Executors[name]
	if executor == nil {
		return nil, errors.New("no executor found with the given name: " + name)
	}

	val := reflect.ValueOf(executor)
	if val.Kind() == reflect.Ptr {
		elem := val.Elem()
		newVal := reflect.New(elem.Type())
		newVal.Elem().Set(elem)
		return newVal.Interface().(flow.StepExecutor), nil
	}
	return executor, nil
}

// GetClient retrieves a client from the global registry by name.
//
// Parameters:
//   - name: Name of the client to retrieve
//
// Returns:
//   - The requested LLM client
//   - An error if the client is not found
func GetClient(name string) (llm.Client, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	client, ok := GlobalRegistry.Clients[name]
	if !ok {
		return nil, errors.New("no client found with the given name: " + name)
	}
	return client, nil
}

// RegisterAgent registers an agent in the global registry.
// Each agent must have a unique name.
//
// Parameters:
//   - name: Name to register the agent under
//   - agent: Agent to register
//
// Returns:
//   - Any error encountered during registration
func RegisterAgent(name string, agent *agent.Agent) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	if _, exists := GlobalRegistry.Agents[name]; exists {
		return fmt.Errorf("agent with name %q already exists", name)
	}

	GlobalRegistry.Agents[name] = agent
	return nil
}

// GetAgent retrieves an agent from the global registry by name.
//
// Parameters:
//   - name: Name of the agent to retrieve
//
// Returns:
//   - The requested agent
//   - An error if the agent is not found
func GetAgent(name string) (*agent.Agent, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	a, ok := GlobalRegistry.Agents[name]
	if !ok {
		return nil, errors.New("no agent found with the given name: " + name)
	}
	return a, nil
}

// NewClientFromConfigFile creates a new client from a configuration file and optionally registers it.
// The file can be in any format supported by Viper (e.g., YAML, JSON, TOML).
//
// Parameters:
//   - name: Name to register the client under (optional, can be empty)
//   - configFile: Path to the client configuration file
//
// Returns:
//   - A new LLM client
//   - Any error encountered during client creation
func NewClientFromConfigFile(name string, configFile string) (llm.Client, error) {
	client, err := llm.NewClientFromConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	// If name is not empty, Set the client to Anyi.Clients
	if name != "" {
		// Use mutex to protect access to the global registry
		GlobalRegistry.mu.Lock()
		defer GlobalRegistry.mu.Unlock()

		GlobalRegistry.Clients[name] = client
	}
	return client, nil
}

// NewMessage creates a new chat message with the specified role and content.
//
// Parameters:
//   - role: Role of the message sender (e.g., "user", "assistant", "system")
//   - content: Content of the message
//
// Returns:
//   - A new chat message
func NewMessage(role string, content string) chat.Message {
	return chat.Message{
		Role:    role,
		Content: content,
	}
}

// NewFlowContextWithText creates a new flow context with the specified text.
// It's a convenience function for creating a context with only text and no memory.
//
// Parameters:
//   - text: Text content for the flow context
//
// Returns:
//   - A new flow context with the specified text
func NewFlowContextWithText(text string) *flow.FlowContext {
	return NewFlowContext(text, nil)
}

// NewFlowContext creates a new flow context with the specified text and memory.
// This is the core function for creating flow contexts.
//
// Parameters:
//   - text: Text content for the flow context
//   - memory: Short-term memory for the flow context (can be any type)
//
// Returns:
//   - A new flow context with the specified text and memory
func NewFlowContext(text string, memory flow.ShortTermMemory) *flow.FlowContext {
	flowContext := flow.FlowContext{
		Text:      text,
		Memory:    memory,
		Variables: make(map[string]any),
	}

	return &flowContext
}

// NewFlowContextWithMemory creates a new flow context with the specified memory and empty text.
// It's a convenience function for creating a context with only memory.
//
// Parameters:
//   - memory: Short-term memory for the flow context
//
// Returns:
//   - A new flow context with the specified memory and empty text
func NewFlowContextWithMemory(memory flow.ShortTermMemory) *flow.FlowContext {
	return NewFlowContext("", memory)
}

// NewFlowContextWithVariables creates a FlowContext with initial variables
//
// Parameters:
//   - text: Text content
//   - memory: Short-term memory
//   - variables: Initial variable collection
//
// Returns:
//   - A new FlowContext with initial variables
func NewFlowContextWithVariables(text string, memory flow.ShortTermMemory, variables map[string]any) *flow.FlowContext {
	if variables == nil {
		variables = make(map[string]any)
	}
	flowContext := flow.FlowContext{
		Text:      text,
		Memory:    memory,
		Variables: variables,
	}

	return &flowContext
}

// GetFormatter retrieves a formatter from the global registry by name.
//
// Parameters:
//   - name: Name of the formatter to retrieve
//
// Returns:
//   - The requested prompt formatter, or nil if not found
func GetFormatter(name string) chat.PromptFormatter {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	return GlobalRegistry.Formatters[name]
}

// RegisterFormatter registers a formatter in the global registry.
// Each formatter must have a unique name.
//
// Parameters:
//   - name: Name to register the formatter under
//   - formatter: Prompt formatter to register
//
// Returns:
//   - Any error encountered during registration
func RegisterFormatter(name string, formatter chat.PromptFormatter) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	GlobalRegistry.Formatters[name] = formatter
	return nil
}

// NewPromptTemplateFormatterFromFile creates a new template formatter from a file and registers it.
// The file should contain a Go template for formatting prompts.
//
// Parameters:
//   - name: Name to register the formatter under
//   - templateFile: Path to the template file
//
// Returns:
//   - A new template formatter
//   - Any error encountered during formatter creation
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

// NewPromptTemplateFormatter creates a new template formatter from a string and registers it.
// The string should contain a Go template for formatting prompts.
//
// Parameters:
//   - name: Name to register the formatter under
//   - template: Template string
//
// Returns:
//   - A new template formatter
//   - Any error encountered during formatter creation
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

// NewFlow creates a new workflow with the specified name, client, and steps.
// The workflow is registered in the global registry.
//
// Parameters:
//   - name: Name for the workflow
//   - client: LLM client to use for the workflow
//   - steps: Workflow steps to include
//
// Returns:
//   - A new workflow
//   - Any error encountered during workflow creation
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

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	GlobalRegistry.Flows[name] = f
	return f, nil
}

// RegisterExecutor registers an executor in the global registry.
// Executors are used to execute steps in workflows.
//
// Parameters:
//   - name: Name to register the executor under
//   - executor: Step executor to register
//
// Returns:
//   - Any error encountered during registration
func RegisterExecutor(name string, executor flow.StepExecutor) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	if GlobalRegistry.Executors[name] != nil {
		return fmt.Errorf("executor type with the name %s already exists", name)
	}

	GlobalRegistry.Executors[name] = executor
	return nil
}

// RegisterValidator registers a validator in the global registry.
// Validators are used to validate the output of workflow steps.
//
// Parameters:
//   - name: Name to register the validator under
//   - validator: Step validator to register
//
// Returns:
//   - Any error encountered during registration
func RegisterValidator(name string, validator flow.StepValidator) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	if GlobalRegistry.Validators[name] != nil {
		return fmt.Errorf("validator type with the name %s already exists", name)
	}
	GlobalRegistry.Validators[name] = validator
	return nil
}

// NewLLMStepExecutorWithFormatter creates a new LLM step executor with a template formatter.
// The executor is registered in the global registry.
//
// Parameters:
//   - name: Name to register the executor under
//   - formatter: Template formatter for generating prompts
//   - systemMessage: System message to include in the conversation
//   - client: LLM client to use for execution
//
// Returns:
//   - A new LLM executor
func NewLLMStepExecutorWithFormatter(name string, formatter *chat.PromptyTemplateFormatter, systemMessage string, client llm.Client) *LLMExecutor {

	stepExecutor := LLMExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}

	RegisterExecutor(name, &stepExecutor)
	return &stepExecutor
}

// NewLLMStep creates a new workflow step with an LLM executor.
// This is a convenience function that calls NewLLMStepWithTemplate.
//
// Parameters:
//   - tmplate: Template string for generating prompts
//   - systemMessage: System message to include in the conversation
//   - client: LLM client to use for the step
//
// Returns:
//   - A new workflow step
//   - Any error encountered during step creation
func NewLLMStep(tmplate string, systemMessage string, client llm.Client) (*flow.Step, error) {
	return NewLLMStepWithTemplate(tmplate, systemMessage, client)
}

// Init initializes the Anyi framework by registering built-in executors and validators.
// This should be called before using the framework, but is automatically called by Config.
func Init() {

	log.Debug("Initializing Anyi...")
	RegisterExecutor("llm", &LLMExecutor{})
	RegisterExecutor("condition", &ConditionalFlowExecutor{})
	RegisterExecutor("exec", &RunCommandExecutor{})
	RegisterExecutor("setContext", &SetContextExecutor{})
	RegisterExecutor("setVariables", &SetVariablesExecutor{})
	// Register with old name for backward compatibility
	RegisterExecutor("setVariable", &SetVariablesExecutor{})
	// Register MCP executor
	RegisterExecutor("mcp", &MCPExecutor{})

	RegisterValidator("string", &StringValidator{})
	RegisterValidator("json", &JsonValidator{})

	log.Debug("Anyi initialized successfully.")
}
