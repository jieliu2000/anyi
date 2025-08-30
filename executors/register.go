package executors

import (
	"github.com/jieliu2000/anyi/registry"
)

// RegisterBuiltinExecutors 注册所有内建执行器
func RegisterBuiltinExecutors() {
	registry.RegisterExecutor("set_context", &SetContextExecutor{})
	registry.RegisterExecutor("set_variables", &SetVariablesExecutor{})
	registry.RegisterExecutor("decorated", &DecoratedExecutor{})
	registry.RegisterExecutor("condition", &ConditionalFlowExecutor{})
	registry.RegisterExecutor("exec", &RunCommandExecutor{})
	registry.RegisterExecutor("llm", &LLMExecutor{})
	registry.RegisterExecutor("deepseek_style_response_filter", &DeepSeekStyleResponseFilter{})
}