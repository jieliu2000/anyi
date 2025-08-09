package registry

import (
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
