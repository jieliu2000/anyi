package llm

import (
	"reflect"
	"testing"
)

func TestNewFunctionConfig(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		params      []ParameterConfig
		expected    *FunctionConfig
	}{
		{
			name:        "EmptyConfig",
			description: "",
			params:      nil,
			expected:    &FunctionConfig{},
		},
		{
			name:        "ConfigWithName",
			description: "",
			params:      nil,
			expected:    &FunctionConfig{Name: "ConfigWithName"},
		},
		{
			name:        "ConfigWithDescription",
			description: "Some description",
			params:      nil,
			expected:    &FunctionConfig{Description: "Some description"},
		},
		{
			name:        "ConfigWithParams",
			description: "",
			params: []ParameterConfig{
				{Name: "param1", Type: "type1", Description: "", Required: false},
				{Name: "param2", Type: "type2", Description: "", Required: true},
			},
			expected: &FunctionConfig{
				Params: []ParameterConfig{
					{Name: "param1", Type: "type1", Description: "", Required: false},
					{Name: "param2", Type: "type2", Description: "", Required: true},
				},
			},
		},
		{
			name:        "ConfigWithAllFields",
			description: "Full configuration",
			params: []ParameterConfig{
				{Name: "param1", Type: "type1", Description: "desc1", Required: false},
				{Name: "param2", Type: "type2", Description: "desc2", Required: true},
			},
			expected: &FunctionConfig{
				Name:        "ConfigWithAllFields",
				Description: "Full configuration",
				Params: []ParameterConfig{
					{Name: "param1", Type: "type1", Description: "desc1", Required: false},
					{Name: "param2", Type: "type2", Description: "desc2", Required: true},
				},
			},
		},
	}

	for _, tc := range testCases {
		actual := NewFunctionConfig(tc.name, tc.description, tc.params...)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Test '%s' failed: got %#v, expected %#v", tc.name, actual, tc.expected)
		}
	}
}
