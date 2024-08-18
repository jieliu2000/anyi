package anyi

import (
	"errors"
	"fmt"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm"
	"github.com/mitchellh/mapstructure"
)

type AnyiConfig struct {
	Clients    []llm.ClientConfig
	Flows      []FlowConfig
	Formatters []FormatterConfig
	Executors  []ExecutorConfig
	Validators []ValidatorConfig
}

type ValidatorConfig struct {
	Name   string                 `mapstructure:"name" json:"name" yaml:"name"`
	Type   string                 `mapstructure:"type" json:"type" yaml:"type"`
	Config map[string]interface{} `mapstructure:"config" json:"config" yaml:"config"`
}

type ExecutorConfig struct {
	Name   string                 `mapstructure:"name" json:"name" yaml:"name"`
	Type   string                 `mapstructure:"type" json:"type" yaml:"type"`
	Config map[string]interface{} `mapstructure:"config" json:"config" yaml:"config"`
}

type FormatterConfig struct {
	Name   string                 `mapstructure:"name" json:"name" yaml:"name"`
	Type   string                 `mapstructure:"type" json:"type" yaml:"type"`
	Config map[string]interface{} `mapstructure:"config" json:"config" yaml:"config"`
}

type FlowConfig struct {
	ClientName   string           `mapstructure:"clientName" json:"clientName" yaml:"clientName"`
	ClientConfig llm.ClientConfig `mapstructure:"clientConfig" json:"clientConfig" yaml:"clientConfig"`
	Name         string           `mapstructure:"name" json:"name" yaml:"name"`
	Steps        []StepConfig     `mapstructure:"steps" json:"steps" yaml:"steps"`
}

type StepConfig struct {
	ClientName string `mapstructure:"clientName" json:"clientName" yaml:"clientName"`
	// The client name which will be used to validate the step output. If not set, validator will use the default client of the step (which is identified by the ClientName field). If the step doesn't have a default client, the validator will use the default client of the flow.
	ValidatorClientName string `mapstructure:"validatorClientName" json:"validatorClientName" yaml:"validatorClientName"`
	MaxRetryTimes       int    `mapstructure:"maxRetryTimes" json:"maxRetryTimes" yaml:"maxRetryTimes"`

	Validator string `mapstructure:"validator" json:"validator" yaml:"validator"`
	// This is a required field. The executor name which will be used to execute the step.
	Executor string `mapstructure:"executor" json:"executor" yaml:"executor"`
}

func NewClientFromConfig(config *llm.ClientConfig) (llm.Client, error) {
	model, err := llm.NewModelConfigFromClientConfig(config)
	if err != nil {
		return nil, err
	}

	return NewClient(config.Name, model)
}

func NewStepFromConfig(stepConfig *StepConfig) (*flow.Step, error) {

	if stepConfig == nil {
		return nil, errors.New("step config is nil")
	}

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
		if executor == nil {
			return nil, fmt.Errorf("step executor %s is not found", stepConfig.Executor)
		}
	} else {
		return nil, errors.New("step executor is not set")
	}
	clientName := stepConfig.ClientName

	var client llm.Client = nil
	if clientName != "" {
		client, err = GetClient(clientName)

		if err != nil {
			return nil, err
		}
	}

	step := flow.NewStep(executor, validator, client)

	return step, nil
}

func NewFlowFromConfig(flowConfig *FlowConfig) (*flow.Flow, error) {

	if flowConfig == nil {
		return nil, errors.New("flow config is nil")
	}

	var client llm.Client = nil
	var err error
	if flowConfig.ClientName != "" {
		client, err = GetClient(flowConfig.ClientName)
		if err != nil {
			return nil, err
		}
	} else if defaultClient, err := GetDefaultClient(); err == nil && defaultClient != nil {
		client = defaultClient
	}

	steps := make([]flow.Step, len(flowConfig.Steps))
	for i, stepConfig := range flowConfig.Steps {
		step, err := NewStepFromConfig(&stepConfig)
		if err != nil {
			return nil, err
		}
		steps[i] = *step
	}

	flow, err := flow.NewFlow(client, flowConfig.Name, steps...)

	if err != nil {
		return nil, err
	}
	err = RegisterFlow(flow.Name, flow)
	return flow, err
}

func NewExecutorFromConfig(executorConfig *ExecutorConfig) (flow.StepExecutor, error) {
	if executorConfig == nil {
		return nil, errors.New("executor config is nil")
	}

	if executorConfig.Type == "" {
		return nil, errors.New("executor type is not set")
	}

	if executorConfig.Name == "" {
		return nil, errors.New("executor name is not set")
	}

	executor := GetExecutorType(executorConfig.Type)

	if executor == nil {
		return nil, fmt.Errorf("executor type %s is not found", executorConfig.Type)
	}

	mapstructure.Decode(executorConfig.Config, executor)
	executor.Init()
	RegisterExecutor(executorConfig.Name, executor)
	return executor, nil
}

func Config(config *AnyiConfig) error {

	Init()
	// Init clients
	for _, clientConfig := range config.Clients {
		if clientConfig.Name != "" {
			_, err := NewClientFromConfig(&clientConfig)
			if err != nil {
				return err
			}

		}
	}
	// Init executors
	for _, executorConfig := range config.Executors {
		_, err := NewExecutorFromConfig(&executorConfig)
		if err != nil {
			return err
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

func ConfigFromFile(configFile string) error {

	anyiConfig, err := utils.UnmarshallConfig(configFile, &AnyiConfig{})

	if err != nil {
		return err
	}
	return Config(anyiConfig)
}
