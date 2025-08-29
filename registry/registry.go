package registry

import (
	"fmt"
	"sync"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

// Registry unified registration table
type Registry struct {
	mu                sync.RWMutex
	Clients           map[string]llm.Client
	Flows             map[string]*flow.Flow
	Agents            map[string]*agent.Agent // Use concrete type pointer
	Executors         map[string]flow.StepExecutor
	Validators        map[string]flow.StepValidator
	Formatters        map[string]chat.PromptFormatter
	defaultClientName string
}

// Global global registry instance
var Global = &Registry{
	Clients:    make(map[string]llm.Client),
	Flows:      make(map[string]*flow.Flow),
	Agents:     make(map[string]*agent.Agent),
	Executors:  make(map[string]flow.StepExecutor),
	Validators: make(map[string]flow.StepValidator),
	Formatters: make(map[string]chat.PromptFormatter),
}

// GetFlow implements agent.FlowGetter interface
func (r *Registry) GetFlow(name string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	flow, exists := r.Flows[name]
	if !exists {
		return nil, fmt.Errorf("flow %s not found", name)
	}
	return flow, nil
}

// RegisterAgent registers an agent
func RegisterAgent(name string, agent *agent.Agent) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	if _, exists := Global.Agents[name]; exists {
		return fmt.Errorf("agent %s already exists", name)
	}

	Global.Agents[name] = agent
	return nil
}

// GetAgent retrieves an agent by name
func GetAgent(name string) (*agent.Agent, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	agent, exists := Global.Agents[name]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", name)
	}
	return agent, nil
}

// RegisterFlow registers a flow
func RegisterFlow(name string, flow *flow.Flow) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	if _, exists := Global.Flows[name]; exists {
		return fmt.Errorf("flow %s already exists", name)
	}

	Global.Flows[name] = flow
	return nil
}

// GetFlow retrieves a flow by name
func GetFlow(name string) (*flow.Flow, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	flow, exists := Global.Flows[name]
	if !exists {
		return nil, fmt.Errorf("flow %s not found", name)
	}
	return flow, nil
}

// RegisterClient registers an LLM client
func RegisterClient(name string, client llm.Client) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	if _, exists := Global.Clients[name]; exists {
		return fmt.Errorf("client %s already exists", name)
	}

	Global.Clients[name] = client
	return nil
}

// GetClient retrieves a client by name
func GetClient(name string) (llm.Client, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	client, exists := Global.Clients[name]
	if !exists {
		return nil, fmt.Errorf("client %s not found", name)
	}
	return client, nil
}

// RegisterExecutor registers a step executor
func RegisterExecutor(name string, executor flow.StepExecutor) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	if _, exists := Global.Executors[name]; exists {
		return fmt.Errorf("executor %s already exists", name)
	}

	Global.Executors[name] = executor
	return nil
}

// GetExecutor retrieves an executor by name
func GetExecutor(name string) (flow.StepExecutor, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	executor, exists := Global.Executors[name]
	if !exists {
		return nil, fmt.Errorf("executor %s not found", name)
	}
	return executor, nil
}

// RegisterValidator registers a step validator
func RegisterValidator(name string, validator flow.StepValidator) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	if _, exists := Global.Validators[name]; exists {
		return fmt.Errorf("validator %s already exists", name)
	}

	Global.Validators[name] = validator
	return nil
}

// GetValidator retrieves a validator by name
func GetValidator(name string) (flow.StepValidator, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	validator, exists := Global.Validators[name]
	if !exists {
		return nil, fmt.Errorf("validator %s not found", name)
	}
	return validator, nil
}

// ListAgents returns a list of all registered agent names
func ListAgents() []string {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	names := make([]string, 0, len(Global.Agents))
	for name := range Global.Agents {
		names = append(names, name)
	}
	return names
}

// ListFlows returns a list of all registered flow names
func ListFlows() []string {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	names := make([]string, 0, len(Global.Flows))
	for name := range Global.Flows {
		names = append(names, name)
	}
	return names
}

// ListClients returns a list of all registered client names
func ListClients() []string {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	names := make([]string, 0, len(Global.Clients))
	for name := range Global.Clients {
		names = append(names, name)
	}
	return names
}

// Clear clears all registrations (useful for testing)
func Clear() {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	Global.Clients = make(map[string]llm.Client)
	Global.Flows = make(map[string]*flow.Flow)
	Global.Agents = make(map[string]*agent.Agent)
	Global.Executors = make(map[string]flow.StepExecutor)
	Global.Validators = make(map[string]flow.StepValidator)
	Global.Formatters = make(map[string]chat.PromptFormatter)
	Global.defaultClientName = ""
}

// RegisterNewDefaultClient registers a client as the default client
func RegisterNewDefaultClient(name string, client llm.Client) error {
	if name == "" {
		name = "default"
	}
	err := RegisterClient(name, client)
	if err != nil {
		return err
	}

	Global.mu.Lock()
	defer Global.mu.Unlock()

	Global.defaultClientName = name
	return nil
}

// SetDefaultClient sets the default client
func SetDefaultClient(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	Global.mu.Lock()
	defer Global.mu.Unlock()

	Global.defaultClientName = name
	return nil
}

// GetDefaultClient retrieves the default client
func GetDefaultClient() (llm.Client, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	if Global.defaultClientName != "" {
		if client, ok := Global.Clients[Global.defaultClientName]; ok {
			return client, nil
		}
	}

	if len(Global.Clients) == 1 {
		for _, client := range Global.Clients {
			return client, nil
		}
	}

	if client, ok := Global.Clients["default"]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("no default client found")
}

// RegisterFormatter registers a prompt formatter
func RegisterFormatter(name string, formatter chat.PromptFormatter) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()

	if _, exists := Global.Formatters[name]; exists {
		return fmt.Errorf("formatter %s already exists", name)
	}

	Global.Formatters[name] = formatter
	return nil
}

// GetFormatter retrieves a formatter by name
func GetFormatter(name string) (chat.PromptFormatter, error) {
	Global.mu.RLock()
	defer Global.mu.RUnlock()

	formatter, exists := Global.Formatters[name]
	if !exists {
		return nil, fmt.Errorf("formatter %s not found", name)
	}
	return formatter, nil
}
