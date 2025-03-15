package anyi

import (
	"errors"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm"
	"github.com/mitchellh/mapstructure"
)

type AnyiConfig struct {
	Clients    []llm.ClientConfig
	Flows      []FlowConfig
	Formatters []FormatterConfig
}

type ValidatorConfig struct {
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
}

type ExecutorConfig struct {
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
}

type FormatterConfig struct {
	Name       string                 `mapstructure:"name" json:"name" yaml:"name"`
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
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

	Validator *ValidatorConfig `mapstructure:"validator" json:"validator" yaml:"validator"`
	// This is a required field. The executor name which will be used to execute the step.
	Executor *ExecutorConfig `mapstructure:"executor" json:"executor" yaml:"executor"`
	Name     string          `mapstructure:"name" json:"name" yaml:"name"`
}

func NewClientFromConfig(config *llm.ClientConfig) (llm.Client, error) {
	model, err := llm.NewModelConfigFromClientConfig(config)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(config.Name, model)
	if err != nil {
		return nil, err
	}
	if config.Default {
		defaultClient, err := GetDefaultClient()
		if err == nil || defaultClient != nil {
			log.Error("Default client is already set: ", GlobalRegistry.defaultClientName)
			log.Error("New default client: ", config.Name)
		}
		RegisterDefaultClient("", client)
	}
	return client, nil
}

func NewStepFromConfig(stepConfig *StepConfig) (*flow.Step, error) {

	if stepConfig == nil {
		return nil, errors.New("step config is nil")
	}

	var validator flow.StepValidator
	var err error
	if stepConfig.Validator != nil {
		validator, err = NewValidatorFromConfig(stepConfig.Validator)
		if err != nil {
			return nil, err
		}
	}
	var executor flow.StepExecutor
	if stepConfig.Executor != nil {
		executor, err = NewExecutorFromConfig(stepConfig.Executor)
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

	defaultClient, _ := GetDefaultClient()
	var client llm.Client = defaultClient

	if clientName != "" {
		client, err = GetClient(clientName)

		if err != nil {
			return nil, err
		}
	}
	step := flow.NewStep(executor, validator, client)
	step.Name = stepConfig.Name
	if stepConfig.MaxRetryTimes > 0 {
		step.MaxRetryTimes = stepConfig.MaxRetryTimes
	}
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

	metaExecutor, err := GetExecutor(executorConfig.Type)
	if err != nil {
		return nil, err
	}

	executor := metaExecutor

	if executor == nil {
		return nil, fmt.Errorf("executor type %s is not found", executorConfig.Type)
	}

	mapstructure.Decode(executorConfig.WithConfig, executor)
	executor.Init()
	return executor, nil
}

func NewValidatorFromConfig(validatorConfig *ValidatorConfig) (flow.StepValidator, error) {
	if validatorConfig == nil {
		return nil, errors.New("validator config is nil")
	}

	if validatorConfig.Type == "" {
		return nil, errors.New("validator type is not set")
	}

	validatorType, err := GetValidator(validatorConfig.Type)
	if err != nil {
		return nil, err
	}

	validator := reflect.New(reflect.TypeOf(validatorType).Elem()).Interface().(flow.StepValidator)

	if validatorType == nil {
		return nil, fmt.Errorf("validator type %s is not found", validatorConfig.Type)
	}

	mapstructure.Decode(validatorConfig.WithConfig, validator)
	validator.Init()
	return validator, nil

}

func Config(config *AnyiConfig) error {

	Init()

	log.Debug("Config Anyi with: ", config)
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

	log.Debug("Config loaded successfully")
	return nil

}

func ConfigFromFile(configFile string) error {

	anyiConfig, err := utils.UnmarshallConfig(configFile, &AnyiConfig{})

	if err != nil {
		return err
	}
	return Config(anyiConfig)
}

// ConfigFromString loads configuration from a string content
// Parameters:
// - configContent string: The configuration content as a string
// - configType string: The configuration type (e.g. "yaml", "json", "toml"). If empty, the format will be auto-detected.
func ConfigFromString(configContent string, configType string) error {
	anyiConfig, err := utils.UnmarshallConfigFromString(configContent, configType, &AnyiConfig{})

	if err != nil {
		return err
	}
	return Config(anyiConfig)
}
