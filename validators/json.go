package validators

import (
	"encoding/json"

	"github.com/jieliu2000/anyi/flow"
)

// JsonValidator is a validator that checks if a text can be parsed as valid JSON.
// It is useful for validating that LLM outputs are properly formatted JSON objects.
type JsonValidator struct {
}

// Init initializes the JsonValidator.
// This validator doesn't require any initialization, so this is a no-op.
//
// Returns:
//   - Always returns nil
func (validator *JsonValidator) Init() error {
	return nil
}

// Validate checks if the text is valid JSON.
// It attempts to unmarshal the text into a generic interface{} and returns
// false if the unmarshaling fails.
//
// Parameters:
//   - stepOutput: The text to validate
//   - step: The workflow step (unused in this validator)
//
// Returns:
//   - true if the text is valid JSON
//   - false if the text is not valid JSON
func (validator *JsonValidator) Validate(stepOutput string, step *flow.Step) bool {
	var result interface{}
	err := json.Unmarshal([]byte(stepOutput), &result)
	return err == nil
}