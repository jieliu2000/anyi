package anyi

import (
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/utils"
)

type AnyiConfig struct {
	Clients []llm.ClientConfig
	Flows   []FlowConfig
}

type FlowConfig struct {
	ClientName   string
	ClientConfig llm.ClientConfig
	Name         string
	Steps        []StepConfig
}

type StepConfig struct {
	ClientName string
	// The client name which will be used to validate the step output. If not set, validator will use the default client of the step (which is identified by the ClientName field). If the step doesn't have a default client, the validator will use the default client of the flow.
	ValidatorClientName string
	MaxRetryTimes       int
	Validator           string
	Executor            string
}

func NewStepFromConfig(stepConfig *StepConfig) (*flow.Step, error) {

	var validator flow.StepValidator
	var err error
	if stepConfig.Validator != "" {
		validator, err = GetValidator(stepConfig.Validator)
		if err != nil {
			return nil, err
		}
	}
	var executor flow.StepExecutor
	if stepConfig.Executor != "" {
		executor, err = GetExecutor(stepConfig.Executor)
		if err != nil {
			return nil, err
		}
	}
	clientName := stepConfig.ClientName

	client, err := GetClient(clientName)

	if err != nil {
		return nil, err
	}

	step := flow.NewStep(executor, validator, client)

	return step, nil
}

func NewFlowFromConfig(flowConfig *FlowConfig) (*flow.Flow, error) {
	client, err := GetClient(flowConfig.ClientName)

	if err != nil {
		return nil, err
	}

	steps := make([]flow.Step, len(flowConfig.Steps))
	for i, stepConfig := range flowConfig.Steps {
		step, err := NewStepFromConfig(&stepConfig)
		if err != nil {
			return nil, err
		}
		steps[i] = *step
	}

	return flow.NewFlow(client, flowConfig.Name, steps...)

}
func InitFromConfig(config *AnyiConfig) error {

	// Init clients
	for _, clientConfig := range config.Clients {
		if clientConfig.Name != "" {
			_, err := NewClientFromConfig(&clientConfig)
			if err != nil {
				return err
			}

		}
	}

	// Init flows
	for _, flowConfig := range config.Flows {
		_, err := NewFlowFromConfig(&flowConfig)
		if err != nil {
			return err
		}
	}

	return nil

}

func InitFromConfigFile(configFile string) error {

	anyiConfig, err := utils.UnmarshallConfig(configFile, &AnyiConfig{})

	if err != nil {
		return err
	}
	return InitFromConfig(anyiConfig)
}
