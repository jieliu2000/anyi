package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

// Registry is the central registry for all components in the Anyi framework.
// It stores clients, flows, validators, executors, formatters, and agents for reuse across the application.
type Registry struct {
	mu                sync.RWMutex
	clients           map[string]llm.Client
	flows             map[string]*flow.Flow
	validators        map[string]flow.StepValidator
	executors         map[string]flow.StepExecutor
	formatters        map[string]chat.PromptFormatter
	agents            map[string]interface{} // Use interface{} to avoid circular dependency
	defaultClientName string
}

// GlobalRegistry is the singleton instance of Registry.
// All components are registered and retrieved through this global registry.
var GlobalRegistry = NewRegistry()

// NewRegistry creates a new registry instance.
func NewRegistry() *Registry {
	return &Registry{
		clients:    make(map[string]llm.Client),
		flows:      make(map[string]*flow.Flow),
		validators: make(map[string]flow.StepValidator),
		executors:  make(map[string]flow.StepExecutor),
		formatters: make(map[string]chat.PromptFormatter),
		agents:     make(map[string]interface{}),
	}
}

// RegisterClient registers a client in the registry.
// Each client must have a unique name.
func (r *Registry) RegisterClient(name string, client llm.Client) error {
	if client == nil {
		return errors.New("client cannot be nil")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[name]; exists {
		return fmt.Errorf("client with name %q already exists", name)
	}

	r.clients[name] = client
	return nil
}

// GetClient retrieves a client from the registry by name.
func (r *Registry) GetClient(name string) (llm.Client, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.clients[name]
	if !ok {
		return nil, fmt.Errorf("no client found with name: %s", name)
	}
	return client, nil
}

// SetDefaultClient sets the default client in the registry.
func (r *Registry) SetDefaultClient(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[name]; !exists {
		return fmt.Errorf("client with name %q not found", name)
	}

	r.defaultClientName = name
	return nil
}

// GetDefaultClient retrieves the default client from the registry.
func (r *Registry) GetDefaultClient() (llm.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.defaultClientName != "" {
		if client, ok := r.clients[r.defaultClientName]; ok {
			return client, nil
		}
	}

	if len(r.clients) == 1 {
		for _, client := range r.clients {
			return client, nil
		}
	}

	if client, ok := r.clients["default"]; ok {
		return client, nil
	}

	return nil, errors.New("no default client found")
}

// RegisterFlow registers a flow in the registry.
// Each flow must have a unique name.
func (r *Registry) RegisterFlow(name string, flow *flow.Flow) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if flow == nil {
		return errors.New("flow cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.flows[name]; exists {
		return fmt.Errorf("flow with name %q already exists", name)
	}

	r.flows[name] = flow
	return nil
}

// GetFlow retrieves a flow from the registry by name.
func (r *Registry) GetFlow(name string) (*flow.Flow, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	f, ok := r.flows[name]
	if !ok {
		return nil, fmt.Errorf("no flow found with name: %s", name)
	}
	return f, nil
}

// RegisterValidator registers a validator in the registry.
func (r *Registry) RegisterValidator(name string, validator flow.StepValidator) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if validator == nil {
		return errors.New("validator cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.validators[name] != nil {
		return fmt.Errorf("validator with name %q already exists", name)
	}

	r.validators[name] = validator
	return nil
}

// GetValidator retrieves a validator from the registry by name.
func (r *Registry) GetValidator(name string) (flow.StepValidator, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	validator := r.validators[name]
	if validator == nil {
		return nil, fmt.Errorf("no validator found with name: %s", name)
	}

	return validator, nil
}

// RegisterExecutor registers an executor in the registry.
func (r *Registry) RegisterExecutor(name string, executor flow.StepExecutor) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if executor == nil {
		return errors.New("executor cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.executors[name] != nil {
		return fmt.Errorf("executor with name %q already exists", name)
	}

	r.executors[name] = executor
	return nil
}

// GetExecutor retrieves an executor from the registry by name.
func (r *Registry) GetExecutor(name string) (flow.StepExecutor, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	executor := r.executors[name]
	if executor == nil {
		return nil, fmt.Errorf("no executor found with name: %s", name)
	}

	return executor, nil
}

// RegisterFormatter registers a formatter in the registry.
func (r *Registry) RegisterFormatter(name string, formatter chat.PromptFormatter) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if formatter == nil {
		return errors.New("formatter cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.formatters[name] = formatter
	return nil
}

// GetFormatter retrieves a formatter from the registry by name.
func (r *Registry) GetFormatter(name string) chat.PromptFormatter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.formatters[name]
}

// RegisterAgent registers an agent in the registry.
// Each agent must have a unique name.
func (r *Registry) RegisterAgent(name string, agent Agent) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if agent == nil {
		return errors.New("agent cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; exists {
		return fmt.Errorf("agent with name %q already exists", name)
	}

	r.agents[name] = agent
	return nil
}

// GetAgent retrieves an agent from the registry by name.
func (r *Registry) GetAgent(name string) (interface{}, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("no agent found with name: %s", name)
	}
	return agent, nil
}

// GetFlows retrieves flows available to the given agent.
func (r *Registry) GetFlows(agent interface{}) ([]*flow.Flow, error) {
	if agent == nil {
		return nil, errors.New("agent cannot be nil")
	}

	// Type assertion to get Agent interface
	agentInterface, ok := agent.(Agent)
	if !ok {
		return nil, errors.New("invalid agent type")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var flows []*flow.Flow
	for _, flowName := range agentInterface.GetFlows() {
		if f, exists := r.flows[flowName]; exists {
			flows = append(flows, f)
		}
	}
	return flows, nil
}
