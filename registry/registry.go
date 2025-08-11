package registry

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
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
	Agents            map[string]*agent.Agent // Add agent registry
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
	Agents:     make(map[string]*agent.Agent), // Initialize agent registry
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
