package llm

// GeneralLLMConfig 包含大多数LLM都支持的通用配置选项
type GeneralLLMConfig struct {
	// Temperature 控制输出的随机性。值越高，输出越随机；值越低，输出越确定。
	// 范围通常在0.0到2.0之间，默认值通常为1.0
	Temperature float32 `json:"temperature" mapstructure:"temperature"`

	// TopP 控制输出的多样性。值越高，输出越多样；值越低，输出越保守。
	// 范围通常在0.0到1.0之间，默认值通常为1.0
	TopP float32 `json:"topP" mapstructure:"topP"`

	// MaxTokens 控制生成的最大token数量
	MaxTokens int `json:"maxTokens" mapstructure:"maxTokens"`

	// PresencePenalty 控制模型避免重复内容的程度
	// 正值会增加避免重复的可能性，负值会增加重复的可能性
	PresencePenalty float32 `json:"presencePenalty" mapstructure:"presencePenalty"`

	// FrequencyPenalty 控制模型避免使用常见词的程度
	// 正值会增加避免使用常见词的可能性，负值会增加使用常见词的可能性
	FrequencyPenalty float32 `json:"frequencyPenalty" mapstructure:"frequencyPenalty"`

	// Stop 指定停止生成的标记列表
	Stop []string `json:"stop" mapstructure:"stop"`
}

// DefaultGeneralConfig 返回一个带有默认值的GeneralLLMConfig
func DefaultGeneralConfig() GeneralLLMConfig {
	return GeneralLLMConfig{
		Temperature:      1.0,
		TopP:             1.0,
		MaxTokens:        0, // 0表示不限制
		PresencePenalty:  0.0,
		FrequencyPenalty: 0.0,
		Stop:             nil,
	}
}
