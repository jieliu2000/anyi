package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

func TestStringValidator_Init(t *testing.T) {
	validator := &StringValidator{
		EqualTo: "foo",
	}
	err := validator.Init()
	assert.NoError(t, err, "expected no error, got %s", err)
}
func TestStringValidator_Init_MatchRegex(t *testing.T) {
	validator := &StringValidator{
		MatchRegex: "[a-z",
	}
	err := validator.Init()
	assert.Error(t, err, "expected error, got nil")
}
func TestStringValidator_Init_BothSet(t *testing.T) {
	validator := &StringValidator{
		EqualTo:    "foo",
		MatchRegex: "[a-z]*",
	}
	err := validator.Init()

	assert.Error(t, err, "expected error, got nil")
	assert.EqualError(t, err, "StringValidator should have either EqualTo or MatchRegex set, not both")
}
func TestStringValidator_Init_NeitherSet(t *testing.T) {
	validator := &StringValidator{}
	err := validator.Init()

	assert.Error(t, err, "expected error, got nil")
	assert.EqualError(t, err, "StringValidator should have either EqualTo or MatchRegex set")
}

func TestStringValidator_Validate(t *testing.T) {
	t.Run("EqualTo - Valid", func(t *testing.T) {
		validator := StringValidator{
			EqualTo: "expected output",
		}
		step := &flow.Step{}
		assert.True(t, validator.Validate("expected output", step))
	})
	t.Run("EqualTo - Invalid", func(t *testing.T) {
		validator := StringValidator{
			EqualTo: "expected output",
		}
		step := &flow.Step{}
		assert.False(t, validator.Validate("wrong output", step))
	})
	t.Run("MatchRegex - Valid", func(t *testing.T) {
		validator := StringValidator{
			MatchRegex: "[a-z]*",
		}
		step := &flow.Step{}
		assert.True(t, validator.Validate("expected output", step))
	})
	t.Run("MatchRegex - Invalid", func(t *testing.T) {
		validator := StringValidator{
			MatchRegex: "^[a-z]*$", // only allow lowercase
		}
		step := &flow.Step{}
		assert.False(t, validator.Validate("1234567890", step))
	})

}

func TestStringValidator_Validate_InvalidRegex(t *testing.T) {
	validator := StringValidator{
		MatchRegex: "[", // Invalid regular expression
	}

	step := &flow.Step{}

	assert.False(t, validator.Validate("any output", step))
}

func TestStringValidator_Validate_NilStep(t *testing.T) {
	validator := StringValidator{
		EqualTo: "expected output",
	}

	assert.False(t, validator.Validate("wrong output", nil))
}

func TestStringValidator_Validate_NilValues(t *testing.T) {
	validator := StringValidator{}

	step := &flow.Step{}

	assert.False(t, validator.Validate("any output", step))
}

func TestJsonValidator_Validate(t *testing.T) {
	validator := &JsonValidator{}
	step := &flow.Step{}
	assert.False(t, validator.Validate("", step))
	assert.False(t, validator.Validate("not json string", step))
	assert.True(t, validator.Validate(`{"key": "value"}`, step))
}
