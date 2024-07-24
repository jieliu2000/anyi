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

		param, err := NewRequiredParameter(name, paramType, description)
		assert.Nil(t, err)
		assert.NotNil(t, param)
		assert.Equal(t, name, param.Name)
		assert.Equal(t, paramType, param.Type)
		assert.Equal(t, description, param.Description)
		assert.True(t, param.Required)
	})

	t.Run("EmptyName", func(t *testing.T) {
		_, err := NewRequiredParameter("", "string", "An example parameter")
		assert.Error(t, err)
	})

	t.Run("InvalidDescription", func(t *testing.T) {
		name := "example"
		paramType := "string"

		_, err := NewRequiredParameter(name, paramType, "")
		assert.Nil(t, err)
	})
}

func TestNewParameter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "example"
		paramType := "string"
		description := "An example parameter"
		required := true

		param, err := NewParameter(name, paramType, description, required)
		assert.Nil(t, err)
		assert.NotNil(t, param)
		assert.Equal(t, name, param.Name)
		assert.Equal(t, paramType, param.Type)
		assert.Equal(t, description, param.Description)
		assert.Equal(t, required, param.Required)
	})

	t.Run("EmptyName", func(t *testing.T) {
		_, err := NewParameter("", "string", "An example parameter", true)
		assert.Error(t, err)
	})

	t.Run("InvalidDescription", func(t *testing.T) {
		name := "example"
		paramType := "string"
		required := true

		_, err := NewParameter(name, paramType, "", required)
		assert.Nil(t, err)
	})
}
