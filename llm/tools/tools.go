package tools

type ParameterConfig struct {
	Name        string   `json:"name" mapstructure:"name"`
	Type        string   `json:"type" mapstructure:"type"`
	Description string   `json:"description,omitempty" mapstructure:"description"`
	Required    bool     `json:"required,omitempty" mapstructure:"required"`
	Enum        []string `json:"enum,omitempty" mapstructure:"enum"`
}

type FunctionConfig struct {
	Name        string            `json:"name" mapstructure:"name"`
	Description string            `json:"description,omitempty" mapstructure:"description"`
	Params      []ParameterConfig `json:"params,omitempty" mapstructure:"params"`
}

type ToolsChatOutput struct {
	Message string `json:"message" mapstructure:"message"`
}

func (funcConfig *FunctionConfig) AddParam(name string, paramType string, description string, required bool, paramEnum []string) *FunctionConfig {
	if funcConfig == nil {
		return nil
	}

	if funcConfig.Params == nil {
		funcConfig.Params = make([]ParameterConfig, 0)
	}
	param := NewParameter(name, paramType, description, required, paramEnum)

	funcConfig.Params = append(funcConfig.Params, param)

	return funcConfig
}

func (funcConfig *FunctionConfig) AddSimpleParam(name string, paramType string, description string) *FunctionConfig {
	return funcConfig.AddParam(name, paramType, description, true, nil)
}

// NewFunctionConfig returns a new FunctionConfig. Note that the name and description are required.
// If the name or description are empty, then nil is returned.
func NewFunctionConfig(name string, description string, params ...ParameterConfig) FunctionConfig {

	return FunctionConfig{Name: name, Description: description, Params: params}
}

func NewParameter(name string, paramType string, description string, required bool, paramEnum []string) ParameterConfig {

	return ParameterConfig{Name: name, Type: paramType, Description: description, Required: required, Enum: paramEnum}
}

func NewOptionalParameter(name string, paramType string, description string) ParameterConfig {
	return NewParameter(name, paramType, description, false, nil)
}

func NewRequiredParameter(name string, paramType string, description string) ParameterConfig {
	return NewParameter(name, paramType, description, true, nil)
}
