package anyi

import (
	"errors"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm"
	"github.com/mitchellh/mapstructure"
)

// AnyiConfig represents the top-level configuration structure for the Anyi framework.
// It contains configurations for clients, flows, and formatters.
type AnyiConfig struct {
	Clients    []llm.ClientConfig
	Flows      []FlowConfig
	Formatters []FormatterConfig
	Agents     []AgentConfig
}

// AgentConfig defines the configuration structure for agents.
// Agents are autonomous entities that can plan and execute workflows.
type AgentConfig struct {
	Name             string   `mapstructure:"name" json:"name" yaml:"name"`
	Role             string   `mapstructure:"role" json:"role" yaml:"role"`
	PreferredLanguage string  `mapstructure:"preferredLanguage" json:"preferredLanguage" yaml:"preferredLanguage"`
	BackStory        string   `mapstructure:"backStory" json:"backStory" yaml:"backStory"`
	Flows            []string `mapstructure:"flows" json:"flows" yaml:"flows"`
}

// ValidatorConfig defines the configuration structure for validators.
// Validators are used to validate the output of workflow steps.
type ValidatorConfig struct {
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
}

// ExecutorConfig defines the configuration structure for executors.
// Executors are responsible for executing workflow steps.
type ExecutorConfig struct {
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
}

// FormatterConfig defines the configuration structure for formatters.
// Formatters are used to format prompts for LLM interactions.
type FormatterConfig struct {
	Name       string                 `mapstructure:"name" json:"name" yaml:"name"`
	Type       string                 `mapstructure:"type" json:"type" yaml:"type"`
	WithConfig map[string]interface{} `mapstructure:"withconfig" json:"withconfig" yaml:"withconfig"`
}

// FlowConfig defines the configuration structure for workflows.
// A workflow consists of a series of steps that are executed in sequence.
type FlowConfig struct {
	ClientName   string           `mapstructure:"clientName" json:"clientName" yaml:"clientName"`
	ClientConfig llm.ClientConfig `mapstructure:"clientConfig" json:"clientConfig" yaml:"clientConfig"`
	Name         string           `mapstructure:"name" json:"name" yaml:"name"`
	Steps        []StepConfig     `mapstructure:"steps" json:"steps" yaml:"steps"`
	Variables    map[string]any   `mapstructure:"variables" json:"variables" yaml:"variables"`
}

// StepConfig defines the configuration structure for workflow steps.
// Each step represents a unit of work within a workflow.
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

// NewClientFromConfig creates a new LLM client from a client configuration.
// It creates the appropriate model configuration, initializes the client,
// and registers it with the global registry if specified.
//
// Parameters:
//   - config: Client configuration containing type, API keys, and other settings
//
// Returns:
//   - A new LLM client instance
//   - Any error encountered during client creation
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
		RegisterNewDefaultClient("", client)
	}
	return client, nil
}

// NewStepFromConfig creates a new workflow step from a step configuration.
// It initializes the validator and executor based on the configuration,
// and associates the appropriate client with the step.
//
// Parameters:
//   - stepConfig: Step configuration containing executor, validator, and client settings
//
// Returns:
//   - A new workflow step
//   - Any error encountered during step creation
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

// NewFlowFromConfig creates a new workflow from a flow configuration.
// It initializes each step in the workflow, associates the appropriate client,
// and registers the flow with the global registry.
//
// Parameters:
//   - flowConfig: Flow configuration containing steps and client settings
//
// Returns:
//   - A new workflow
//   - Any error encountered during flow creation
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

	// Set flow variables from config
	if flowConfig.Variables != nil {
		flow.Variables = make(map[string]any)
		for k, v := range flowConfig.Variables {
			flow.Variables[k] = v
		}
	}

	err = RegisterFlow(flow.Name, flow)
	return flow, err
}

// NewAgentFromConfig creates a new agent from an agent configuration.
//
// Parameters:
//   - config: Agent configuration containing role, backstory, and flow settings
//
// Returns:
//   - A new agent instance
//   - Any error encountered during agent creation
func NewAgentFromConfig(config *AgentConfig) (*agent.Agent, error) {
	if config == nil {
		return nil, errors.New("agent config is nil")
	}

	// Convert flow names to actual flow objects
	flowObjects := make([]*flow.Flow, len(config.Flows))
	for i, flowName := range config.Flows {
		flowObj, err := GetFlow(flowName)
		if err != nil {
			return nil, fmt.Errorf("flow %q not found for agent %q", flowName, config.Name)
		}
		flowObjects[i] = flowObj
	}

	agentObj := &agent.Agent{
		Role:              config.Role,
		PreferredLanguage: config.PreferredLanguage,
		BackStory:         config.BackStory,
		Flows:             flowObjects,
	}

	// Register Agent to the global registry
	if err := RegisterAgent(config.Name, agentObj); err != nil {
		return nil, err
	}

	return agentObj, nil
}

// NewExecutorFromConfig creates a new executor from an executor configuration.
// It instantiates the appropriate executor type based on the configuration,
// decodes the configuration parameters, and initializes the executor.
//
// Parameters:
//   - executorConfig: Executor configuration containing type and parameters
//
// Returns:
//   - A new step executor
//   - Any error encountered during executor creation
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

// NewValidatorFromConfig creates a new validator from a validator configuration.
// It instantiates the appropriate validator type based on the configuration,
// decodes the configuration parameters, and initializes the validator.
//
// Parameters:
//   - validatorConfig: Validator configuration containing type and parameters
//
// Returns:
//   - A new step validator
//   - Any error encountered during validator creation
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

// Config configures the Anyi framework with the provided configuration.
// It initializes clients, flows, and formatters based on the configuration.
//
// Parameters:
//   - config: Complete configuration for the Anyi framework
//
// Returns:
//   - Any error encountered during configuration
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
	
	// Init agents (depends on flows being initialized first)
	for _, agentConfig := range config.Agents {
		_, err := NewAgentFromConfig(&agentConfig)
		if err != nil {
			return err
		}
	}

	log.Debug("Config loaded successfully")
	return nil

}

// ConfigFromFile loads configuration from a file and configures the Anyi framework.
// The file can be in any format supported by Viper (e.g., YAML, JSON, TOML).
//
// Parameters:
//   - configFile: Path to the configuration file
//
// Returns:
//   - Any error encountered during configuration loading
func ConfigFromFile(configFile string) error {

	anyiConfig, err := utils.UnmarshallConfig(configFile, &AnyiConfig{})

	if err != nil {
		return err
	}
	return Config(anyiConfig)
}

// ConfigFromString loads configuration from a string content and configures the Anyi framework.
// The string can be in any format supported by the configType parameter.
//
// Parameters:
//   - configContent: Configuration content as a string
//   - configType: Configuration format type (e.g., "yaml", "json", "toml")
//
// Returns:
//   - Any error encountered during configuration loading
func ConfigFromString(configContent string, configType string) error {
	if configContent == "" || configType == "" {
		return errors.New("configContent and configType cannot be empty")
	}
	anyiConfig, err := utils.UnmarshallConfigFromString(configContent, configType, &AnyiConfig{})

	if err != nil {
		return err
	}
	return Config(anyiConfig)
}
