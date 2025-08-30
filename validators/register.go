package validators

import (
	"github.com/jieliu2000/anyi/registry"
)

// RegisterBuiltinValidators 注册所有内建验证器
func RegisterBuiltinValidators() {
	registry.RegisterValidator("string", &StringValidator{})
	registry.RegisterValidator("json", &JsonValidator{})
}