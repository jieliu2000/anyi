package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/validators"
	"github.com/stretchr/testify/assert"
)

func TestStringValidator_Init(t *testing.T) {
	validator := &validators.StringValidator{
		Pattern: "foo",
	}
	err := validator.Init()
	assert.NoError(t, err, "expected no error, got %s", err)
}
func TestStringValidator_Init_MatchRegex(t *testing.T) {
	validator := &validators.StringValidator{
		Pattern: "[a-z",
	}
	err := validator.Init()
	assert.Error(t, err, "expected error, got nil")
}
func TestStringValidator_Init_BothSet(t *testing.T) {
	// this test case is not applicable since stringvalidator only has pattern field
	// and we can't set both equalto and matchregex
	t.Skip("stringvalidator only has pattern field")
}
func TestStringValidator_Init_NeitherSet(t *testing.T) {
	validator := &validators.StringValidator{}
	err := validator.Init()

	assert.Error(t, err, "expected error, got nil")
	assert.EqualError(t, err, "pattern is required")
}

func TestStringValidator_Validate(t *testing.T) {
	t.Run("exact match - valid", func(t *testing.T) {
		validator := validators.StringValidator{
			Pattern: "^expected output$",
		}
		err := validator.Init()
		assert.NoError(t, err)
		step := &flow.Step{}
		assert.True(t, validator.Validate("expected output", step))
	})
	t.Run("exact match - invalid", func(t *testing.T) {
		validator := validators.StringValidator{
			Pattern: "^expected output$",
		}
		err := validator.Init()
		assert.NoError(t, err)
		step := &flow.Step{}
		assert.False(t, validator.Validate("wrong output", step))
	})
	t.Run("pattern match - valid", func(t *testing.T) {
		validator := validators.StringValidator{
			Pattern: "[a-z]*",
		}
		err := validator.Init()
		assert.NoError(t, err)
		step := &flow.Step{}
		assert.True(t, validator.Validate("expected output", step))
	})
	t.Run("pattern match - invalid", func(t *testing.T) {
		validator := validators.StringValidator{
			Pattern: "^[a-z]*$", // only allow lowercase
		}
		err := validator.Init()
		assert.NoError(t, err)
		step := &flow.Step{}
		assert.False(t, validator.Validate("1234567890", step))
	})

}

func TestStringValidator_Validate_InvalidRegex(t *testing.T) {
	validator := validators.StringValidator{
		Pattern: "[", // invalid regular expression
	}

	// init should fail with invalid regex
	err := validator.Init()
	assert.Error(t, err)

	step := &flow.Step{}

	// validate should return false when init fails
	assert.False(t, validator.Validate("any output", step))
}

func TestStringValidator_Validate_NilStep(t *testing.T) {
	validator := validators.StringValidator{
		Pattern: "^expected output$",
	}
	err := validator.Init()
	assert.NoError(t, err)

	assert.False(t, validator.Validate("wrong output", nil))
}

func TestStringValidator_Validate_NilValues(t *testing.T) {
	validator := validators.StringValidator{}

	// init should fail with empty pattern
	err := validator.Init()
	assert.Error(t, err)

	step := &flow.Step{}

	// validate should return false when init fails
	assert.False(t, validator.Validate("any output", step))
}

func TestJsonValidator_Validate(t *testing.T) {
	validator := &validators.JsonValidator{}
	step := &flow.Step{}
	assert.False(t, validator.Validate("", step))
	assert.False(t, validator.Validate("not json string", step))
	assert.True(t, validator.Validate(`{"key": "value"}`, step))
}
