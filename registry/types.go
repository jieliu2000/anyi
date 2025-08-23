package registry

// Agent defines the interface that agents must implement for registry integration.
// This interface is defined here to avoid circular dependencies.
type Agent interface {
	GetName() string
	GetFlows() []string
	GetClientName() string
}
