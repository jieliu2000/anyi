package anyi

import "github.com/jieliu2000/anyi/llm"

type AnyiConfig struct {
	Clients []llm.ClientConfig
	Flows   []*Flow
}
