package validators

import (
	"errors"
	"regexp"

	"github.com/jieliu2000/anyi/flow"
)

// StringValidator is a validator that checks if a text matches a regular expression pattern.
// It is useful for validating that LLM outputs follow specific text patterns.
type StringValidator struct {
	Pattern string `json:"pattern" yaml:"pattern" mapstructure:"pattern"`
	re      *regexp.Regexp
}

// Init initializes the StringValidator by compiling the regular expression pattern.
//
// Returns:
//   - An error if the pattern is empty or if compilation fails
func (validator *StringValidator) Init() error {
	if validator.Pattern == "" {
		return errors.New("pattern is required")
	}

	re, err := regexp.Compile(validator.Pattern)
	if err != nil {
		return err
	}
	validator.re = re
	return nil
}

// Validate checks if the text matches the regular expression pattern.
//
// Parameters:
//   - stepOutput: the text to validate
//   - step: the workflow step (unused in this validator)
//
// Returns:
//   - true if the text matches the pattern
//   - false if the text does not match the pattern
func (validator *StringValidator) Validate(stepOutput string, step *flow.Step) bool {
	if validator.re == nil {
		return false
	}
	return validator.re.MatchString(stepOutput)
}
