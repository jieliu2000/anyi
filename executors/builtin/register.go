package builtin

import (
	"github.com/jieliu2000/anyi/registry"
)

// RegisterBuiltinExecutors registers all builtin executors to the registry.
func RegisterBuiltinExecutors() {
	registry.RegisterExecutor("runCommand", &RunCommandExecutor{})
	registry.RegisterExecutor("conditionalFlow", &ConditionalFlowExecutor{})
	registry.RegisterExecutor("setContext", &SetContextExecutor{})
	registry.RegisterExecutor("setVariables", &SetVariablesExecutor{})
	registry.RegisterExecutor("decorated", &DecoratedExecutor{})
	registry.RegisterExecutor("llm", &LLMExecutor{})
}