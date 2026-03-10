package executors

import (
	"github.com/jieliu2000/anyi/executors/builtin"
	"github.com/jieliu2000/anyi/registry"
)

// RegisterBuiltinExecutors registers all builtin executors.
func RegisterBuiltinExecutors() {
	builtin.RegisterBuiltinExecutors()
}