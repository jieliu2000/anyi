package anyi

import "github.com/jieliu2000/anyi/llm"

func NewClient(config llm.ModelConfig) (llm.Client, error) {

	return llm.NewClient(config)
}
