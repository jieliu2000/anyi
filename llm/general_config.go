package llm

// GeneralLLMConfig contains common configuration options supported by most LLMs
type GeneralLLMConfig struct {
	// Temperature controls the randomness of output. Higher values produce more random output; lower values produce more deterministic output.
	// Range is typically between 0.0 and 2.0, with default usually being 1.0
	Temperature float32 `json:"temperature" mapstructure:"temperature"`

	// TopP controls the diversity of output. Higher values produce more diverse output; lower values produce more conservative output.
	// Range is typically between 0.0 and 1.0, with default usually being 1.0
	TopP float32 `json:"topP" mapstructure:"topP"`

	// MaxTokens controls the maximum number of tokens to generate
	MaxTokens int `json:"maxTokens" mapstructure:"maxTokens"`

	// PresencePenalty controls the degree to which the model avoids repeating content
	// Positive values increase the likelihood of avoiding repetition, negative values increase the likelihood of repetition
	PresencePenalty float32 `json:"presencePenalty" mapstructure:"presencePenalty"`

	// FrequencyPenalty controls the degree to which the model avoids using common words
	// Positive values increase the likelihood of avoiding common words, negative values increase the likelihood of using common words
	FrequencyPenalty float32 `json:"frequencyPenalty" mapstructure:"frequencyPenalty"`

	// Stop specifies a list of tokens that will stop generation
	Stop []string `json:"stop" mapstructure:"stop"`
}

// DefaultGeneralConfig returns a GeneralLLMConfig with default values
func DefaultGeneralConfig() GeneralLLMConfig {
	return GeneralLLMConfig{
		Temperature:      1.0,
		TopP:             1.0,
		MaxTokens:        0, // 0 means no limit
		PresencePenalty:  0.0,
		FrequencyPenalty: 0.0,
		Stop:             nil,
	}
}
