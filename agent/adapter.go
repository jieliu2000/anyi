package agent

import (
	"fmt"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
)

// AnyiRegistryAdapter adapts the main anyi registry to the AgentRegistry interface.
// This allows agents to access flows and clients from the main anyi framework.
type AnyiRegistryAdapter struct{}

// GetFlows retrieves flows available to the given agent from the main anyi registry.
func (a *AnyiRegistryAdapter) GetFlows(agent *Agent) ([]*flow.Flow, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	// This implementation is not used - we use FunctionalRegistryAdapter instead
	return nil, fmt.Errorf("flow access not implemented - use FunctionalRegistryAdapter")
}

// GetClient retrieves a client from the main anyi registry.
func (a *AnyiRegistryAdapter) GetClient(clientName string) (llm.Client, error) {
	// Same issue - circular dependency
	return nil, fmt.Errorf("client access not implemented - use FunctionalRegistryAdapter")
}

// RegistryFunctions holds function pointers to avoid circular dependencies.
type RegistryFunctions struct {
	GetFlow   func(name string) (*flow.Flow, error)
	GetClient func(name string) (llm.Client, error)
}

// FunctionalRegistryAdapter implements AgentRegistry using function pointers.
type FunctionalRegistryAdapter struct {
	Functions RegistryFunctions
}

// GetFlows retrieves flows available to the given agent.
func (f *FunctionalRegistryAdapter) GetFlows(agent *Agent) ([]*flow.Flow, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	var flows []*flow.Flow
	for _, flowName := range agent.Flows {
		flow, err := f.Functions.GetFlow(flowName)
		if err != nil {
			return nil, fmt.Errorf("failed to get flow '%s': %w", flowName, err)
		}
		flows = append(flows, flow)
	}
	return flows, nil
}

// GetClient retrieves a client by name.
func (f *FunctionalRegistryAdapter) GetClient(clientName string) (llm.Client, error) {
	return f.Functions.GetClient(clientName)
}
