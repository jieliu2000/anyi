package tools

import (
	"errors"
	"fmt"
)

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

// NewFunctionConfig returns a new FunctionConfig. Note that the name and description are required.
// If the name or description are empty, then nil is returned.
func NewFunctionConfig(name string, description string, params ...ParameterConfig) (*FunctionConfig, error) {
	if name == "" || description == "" {
		return nil, errors.New("name and description are required")
	}

	paramsArray := make([]ParameterConfig, len(params))
	for index, param := range params {
		value, err := NewParameter(param.Name, param.Type, param.Description, param.Required)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error creating parameter at index %d", index), err)
		}
		paramsArray[index] = *value
	}
	return &FunctionConfig{Name: name, Description: description, Params: paramsArray}, nil
}

func NewParameter(name string, paramType string, description string, required bool) (*ParameterConfig, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	return &ParameterConfig{Name: name, Type: paramType, Description: description, Required: required}, nil
}

func NewOptionalParameter(name string, paramType string, description string) (*ParameterConfig, error) {
	return NewParameter(name, paramType, description, false)
}

func NewRequiredParameter(name string, paramType string, description string) (*ParameterConfig, error) {
	return NewParameter(name, paramType, description, true)
}
