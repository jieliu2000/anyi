package llm

type ParameterConfig struct {
	Name        string `json:"name" mapstructure:"name"`
	Type        string `json:"type" mapstructure:"type"`
	Description string `json:"description,omitempty" mapstructure:"description"`
	Required    bool   `json:"required,omitempty" mapstructure:"required"`
}

type FunctionConfig struct {
	Name        string            `json:"name" mapstructure:"name"`
	Description string            `json:"description,omitempty" mapstructure:"description"`
	Params      []ParameterConfig `json:"params,omitempty" mapstructure:"params"`
}

func NewFunctionConfig(name string, description string, params ...ParameterConfig) *FunctionConfig {
	return &FunctionConfig{Name: name, Description: description, Params: params}
}
