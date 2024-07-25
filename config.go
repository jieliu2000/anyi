package anyi

import "github.com/jieliu2000/anyi/llm"

type AnyiConfig struct {
	Clients []llm.ClientConfig
	Flows   []FlowConfig
}

type FlowConfig struct {
	ClientName   string
	ClientConfig llm.ClientConfig
	Name         string
}

type FlowStepConfig struct {
	ClientName string
	// The client name which will be used to validate the step output. If not set, validator will use the default client of the step (which is identified by the ClientName field). If the step doesn't have a default client, the validator will use the default client of the flow.
	ValidatorClientName string
	MaxRetryTimes       int
}
