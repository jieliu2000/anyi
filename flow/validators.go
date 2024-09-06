package flow

import (
	"errors"
	"regexp"
)

// StepValidator is the interface for validators of step output.
// In a flow if a step validator is set, the output of the step will be checked against the validator's Validate method.
type StepValidator interface {
	Init() error
	Validate(stepOutput string, Step *Step) bool
}

// StringValidator is a validator for string output. It can be used to check if the step's output matches a given regular expression or equals a specific string.
// Note that the EqualTo and MatchRegex fields are mutually exclusive. If both are set, an error is returned during initialization.
type StringValidator struct {
	EqualTo    string `json:"eqaulTo" mapstructure:"eqaulTo" yaml:"eqaulTo"`
	MatchRegex string `json:"matchRegex" mapstructure:"matchRegex" yaml:"matchRegex"`
}

// Init initializes the StringValidator.
// It checks the validity of the regular expression and ensures that either EqualTo or MatchRegex is set, but not both.
// If any error occurs during initialization, the corresponding error message is returned.
func (validator *StringValidator) Init() error {
	if validator.MatchRegex != "" {
		_, err := regexp.Compile(validator.MatchRegex)
		if err != nil {
			return err
		}
	}
	if validator.EqualTo == "" && validator.MatchRegex == "" {
		return errors.New("StringValidator should have either EqualTo or MatchRegex set")
	}

	if validator.EqualTo != "" && validator.MatchRegex != "" {
		return errors.New("StringValidator should have either EqualTo or MatchRegex set, not both")
	}

	return nil
}

// Validate function checks if the stepOutput matches the validation criteria set in the StringValidator struct.
// For regular expressions, see [Golang regexp documentation]
//
// Parameters:
// - stepOutput string: The output string to be validated.
// - Step *Step: The step object containing the validation information.
// Return value:
// - bool: True if the validation passes, false otherwise.
//
// [Golang regexp documentation]: https://pkg.go.dev/regexp
func (validator *StringValidator) Validate(stepOutput string, Step *Step) bool {

	if validator.EqualTo != "" && stepOutput == validator.EqualTo {
		return true
	}

	if validator.MatchRegex != "" {
		matched, err := regexp.MatchString(validator.MatchRegex, stepOutput)
		if err != nil {
			return false
		}
		return matched
	}

	return false
}
