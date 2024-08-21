package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequiredParameter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "example"
		paramType := "string"
		description := "An example parameter"

		param := NewRequiredParameter(name, paramType, description)
		assert.NotNil(t, param)
		assert.Equal(t, name, param.Name)
		assert.Equal(t, paramType, param.Type)
		assert.Equal(t, description, param.Description)
		assert.True(t, param.Required)
	})

}

func TestNewParameter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "example"
		paramType := "string"
		description := "An example parameter"
		required := true

		param := NewParameter(name, paramType, description, required, nil)
		assert.NotNil(t, param)
		assert.Equal(t, name, param.Name)
		assert.Equal(t, paramType, param.Type)
		assert.Equal(t, description, param.Description)
		assert.Equal(t, required, param.Required)
	})

}
