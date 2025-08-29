package agent

// AgentContext execution context - uses value type to ensure safety
type AgentContext struct {
	Variables map[string]interface{}
	Memory    interface{}
	History   []string
}
