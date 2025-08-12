package registry

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/jieliu2000/anyi/agent/agentmodel"
	"github.com/jieliu2000/anyi/executors"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/mitchellh/mapstructure"
)

// AnyiRegistry is the central registry for all components in the Anyi framework.
// It stores clients, flows, validators, executors, and formatters for reuse across the application.
type AnyiRegistry struct {
	Mu                sync.RWMutex
	Clients           map[string]llm.Client
	Flows             map[string]*flow.Flow
	Validators        map[string]flow.StepValidator
	Executors         map[string]flow.StepExecutor
	Formatters        map[string]chat.PromptFormatter
	Agents            map[string]*agentmodel.Agent // Add agent registry
	DefaultClientName string
}

// GlobalRegistry is the singleton instance of AnyiRegistry.
// All components are registered and retrieved through this global registry.
var GlobalRegistry *AnyiRegistry = &AnyiRegistry{
	Clients:    make(map[string]llm.Client),
	Flows:      make(map[string]*flow.Flow),
	Validators: make(map[string]flow.StepValidator),
	Executors:  make(map[string]flow.StepExecutor),
	Formatters: make(map[string]chat.PromptFormatter),
	Agents:     make(map[string]*agentmodel.Agent), // Initialize agent registry
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

	GlobalRegistry.Mu.RLock()
	defer GlobalRegistry.Mu.RUnlock()

	f, ok := GlobalRegistry.Flows[name]
	if !ok {
		return nil, errors.New("no flow found with the given name: " + name)
	}
	return f, nil
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

	GlobalRegistry.Mu.RLock()
	defer GlobalRegistry.Mu.RUnlock()

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

	GlobalRegistry.Mu.Lock()
	defer GlobalRegistry.Mu.Unlock()

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

	GlobalRegistry.Mu.Lock()
	defer GlobalRegistry.Mu.Unlock()

	if GlobalRegistry.Validators[name] != nil {
		return fmt.Errorf("validator type with the name %s already exists", name)
	}
	GlobalRegistry.Validators[name] = validator
	return nil
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

	GlobalRegistry.Mu.Lock()
	defer GlobalRegistry.Mu.Unlock()

	if _, exists := GlobalRegistry.Flows[name]; exists {
		return fmt.Errorf("flow with name %q already exists", name)
	}

	GlobalRegistry.Flows[name] = flow
	return nil
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

	GlobalRegistry.Mu.Lock()
	defer GlobalRegistry.Mu.Unlock()

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

	GlobalRegistry.Mu.RLock()
	defer GlobalRegistry.Mu.RUnlock()

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

	GlobalRegistry.Mu.RLock()
	defer GlobalRegistry.Mu.RUnlock()

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
//   - agent: Agent to register
//
// Returns:
//   - Any error encountered during registration
func RegisterAgent(a *agentmodel.Agent) error {
	if a == nil {
		return errors.New("agent cannot be nil")
	}
	if a.Role == "" {
		return errors.New("agent role cannot be empty")
	}

	GlobalRegistry.Mu.Lock()
	defer GlobalRegistry.Mu.Unlock()

	if _, exists := GlobalRegistry.Agents[a.Role]; exists {
		return fmt.Errorf("agent with role %q already exists", a.Role)
	}

	GlobalRegistry.Agents[a.Role] = a
	return nil
}

// GetAgent retrieves an agent from the global registry by name.
//
// Parameters:
//   - role: Role of the agent to retrieve
//
// Returns:
//   - The requested agent
//   - An error if the agent is not found
func GetAgent(role string) (*agentmodel.Agent, error) {
	if role == "" {
		return nil, errors.New("name cannot be empty")
	}

	GlobalRegistry.Mu.RLock()
	defer GlobalRegistry.Mu.RUnlock()

	a, ok := GlobalRegistry.Agents[role]
	if !ok {
		return nil, errors.New("no agent found with the given name: " + role)
	}
	return a, nil
}

// GetFormatter retrieves a formatter from the global registry by name.
//
// Parameters:
//   - name: Name of the formatter to retrieve
//
// Returns:
//   - The requested prompt formatter, or nil if not found
func GetFormatter(name string) chat.PromptFormatter {
	GlobalRegistry.Mu.RLock()
	defer GlobalRegistry.Mu.RUnlock()

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

	GlobalRegistry.Mu.Lock()
	defer GlobalRegistry.Mu.Unlock()

	GlobalRegistry.Formatters[name] = formatter
	return nil
}

// NewExecutorFromConfig creates a new executor from an executor configuration.
// It instantiates the appropriate executor type based on the configuration,
// decodes the configuration parameters, and initializes the executor.
//
// Parameters:
//   - executorConfig: Executor configuration containing type and parameters
//
// Returns:
//   - A new step executor
//   - Any error encountered during executor creation
func NewExecutorFromConfig(executorConfig *executors.ExecutorConfig) (flow.StepExecutor, error) {
	if executorConfig == nil {
		return nil, errors.New("executor config is nil")
	}

	if executorConfig.Type == "" {
		return nil, errors.New("executor type is not set")
	}

	metaExecutor, err := GetExecutor(executorConfig.Type)
	if err != nil {
		return nil, err
	}

	executor := metaExecutor

	if executor == nil {
		return nil, fmt.Errorf("executor type %s is not found", executorConfig.Type)
	}

	mapstructure.Decode(executorConfig.WithConfig, executor)
	executor.Init()
	return executor, nil
}
