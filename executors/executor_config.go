package executors

// ExecutorConfig defines the configuration structure for executors.
// Executors are responsible for executing workflow steps.
type ExecutorConfig struct {
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
}
