package anyi

import (
	"os"
	"testing"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/executors"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

type MockValidator struct {
}

func (m MockValidator) Init() error {

	return nil
}

func (m MockValidator) Validate(stepOutput string, Step *flow.Step) bool {

	return true
}

func TestNewFlowFromConfig_Success(t *testing.T) {
	// Setup
	GlobalRegistry = &registry.AnyiRegistry{
		Flows:      make(map[string]*flow.Flow),
		Clients:    make(map[string]llm.Client),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
	}
	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &executors.MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName:  "test-client",
		Name:        "test-flow",
		Description: "This is a test flow description",
		Variables: map[string]any{
			"var1": "value1",
			"var2": 123,
		},
		Steps: []StepConfig{
			{
				Name: "name1",
				Executor: &executors.ExecutorConfig{
					Type: "test-executor",
				},
				Validator: &ValidatorConfig{
					Type: "test-validator",
				},
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, flowInstance)
	assert.Equal(t, flowConfig.Name, flowInstance.Name)
	assert.Equal(t, flowConfig.Description, flowInstance.Description)
	assert.Equal(t, 1, len(flowInstance.Steps))
	step := flowInstance.Steps[0]
	assert.Equal(t, "name1", step.Name)

	var1, ok := flowInstance.GetVariable("var1")
	assert.True(t, ok)
	assert.Equal(t, "value1", var1)

	var2, ok := flowInstance.GetVariable("var2")
	assert.True(t, ok)
	assert.Equal(t, 123, var2)
}
func TestNewFlowFromConfig_WithNil(t *testing.T) {
	// Execute
	flowInstance, err := NewFlowFromConfig(nil)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
func TestNewFlowFromConfig_WithInvalidClientName(t *testing.T) {
	// Setup
	flowConfig := &FlowConfig{
		ClientName: "invalid-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Executor: &executors.ExecutorConfig{
					Type: "test-executor",
				},
				Validator: &ValidatorConfig{
					Type: "test-validator",
				},
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
func TestNewFlowFromConfig_WithInvalidStepConfig(t *testing.T) {
	// Setup
	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &executors.MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Executor: &executors.ExecutorConfig{
					Type: "invalid-executor",
				},
				Validator: &ValidatorConfig{
					Type: "test-validator",
				},
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}
func TestNewFlowFromConfig_WithEmptyStepExecutor(t *testing.T) {
	// Setup
	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &executors.MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Validator: &ValidatorConfig{
					Type: "test-validator",
				},
			},
		},
	}

	// Execute
	flowInstance, err := NewFlowFromConfig(flowConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, flowInstance)
}

func TestNewClientFromConfigWithEmptyName(t *testing.T) {
	config := llm.ClientConfig{
		Name: "",
		Type: "line",
		Config: map[string]interface{}{
			"accessToken": "test_access_token",
		},
	}
	client, err := NewClientFromConfig(&config)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestNewClientFromConfig(t *testing.T) {
	config := llm.ClientConfig{
		Name: "test",
		Type: "openai",
		Config: map[string]interface{}{
			"accessToken": "test_access_token",
		},
	}
	client, err := NewClientFromConfig(&config)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestNewClientFromConfigWithInvalidType(t *testing.T) {
	config := llm.ClientConfig{
		Name: "test",
		Type: "invalid_type",
		Config: map[string]interface{}{
			"accessToken": "test_access_token",
		},
	}
	client, err := NewClientFromConfig(&config)
	assert.Nil(t, client)
	assert.Error(t, err)
}
func TestConfig(t *testing.T) {
	RegisterExecutor("executor1", &executors.MockExecutor{})
	config := AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "client1",
				Type: "ollama",
				Config: map[string]interface{}{
					"requestTimeout": 1000,
					"model":          "qwen2",
				},
			},
		},

		Flows: []FlowConfig{
			{
				Name: "flow1",
				Variables: map[string]any{
					"testVar": "testValue",
				},
				Steps: []StepConfig{
					{
						Executor: &executors.ExecutorConfig{
							Type: "executor1",
						},
						ClientName:    "client1",
						MaxRetryTimes: 1,
					},
				},
			},
		},
	}
	err := Config(&config)
	assert.Nil(t, err)

	// Verify flow variables are set
	flow, err := GetFlow("flow1")
	assert.NoError(t, err)
	result, ok := flow.GetVariable("testVar")
	assert.True(t, ok)
	assert.Equal(t, "testValue", result)
	assert.Nil(t, err)
}

func TestConfigWithInvalidExecutor(t *testing.T) {
	config := AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name:   "client1",
				Type:   "dashscope",
				Config: map[string]interface{}{},
			},
		},

		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor: &executors.ExecutorConfig{
							Type: "invalid-executor",
						},
						ClientName:    "client1",
						MaxRetryTimes: 1,
					},
				},
			},
		},
	}
	err := Config(&config)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "no executor found with the given name: invalid-executor")
}

func TestConfigWithInvalidValidator(t *testing.T) {
	RegisterExecutor("executor1", &executors.MockExecutor{})
	config := AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "client1",
				Type: "openai",
				Config: map[string]interface{}{
					"api_key": "test_key",
				},
			},
		},

		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor: &executors.ExecutorConfig{
							Type: "executor1",
						},
						Validator: &ValidatorConfig{
							Type: "invalid",
						},
						ClientName:    "client1",
						MaxRetryTimes: 1,
					},
				},
			},
		},
	}
	err := Config(&config)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "no validator found with the given name: invalid")
}

func TestConfigWithInvalidClient(t *testing.T) {
	RegisterExecutor("executor1", &executors.MockExecutor{})
	config := AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "client1",
				Type: "openai",
				Config: map[string]interface{}{
					"api_key": "token",
				},
			},
		},

		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor: &executors.ExecutorConfig{
							Type: "executor1",
						},
						ClientName:    "no-client",
						MaxRetryTimes: 1,
					},
				},
			},
		},
	}
	err := Config(&config)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "no client found with the given name: no-client")
}

func TestNewValidatorFromConfig(t *testing.T) {

	RegisterValidator("mock", &MockValidator{})

	testCases := []struct {
		name        string
		config      *ValidatorConfig
		expectedErr string
	}{
		{
			name:        "Success",
			config:      &ValidatorConfig{Type: "mock"},
			expectedErr: "",
		},
		{
			name:        "Failure: Validator config is nil",
			config:      nil,
			expectedErr: "validator config is nil",
		},
		{
			name:        "Failure: Validator type is not set",
			config:      &ValidatorConfig{},
			expectedErr: "validator type is not set",
		},
		{
			name:        "Failure: Unrecognized validator type",
			config:      &ValidatorConfig{Type: "unknown"},
			expectedErr: "no validator found with the given name: unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator, err := NewValidatorFromConfig(tc.config)
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, validator)
				assert.IsType(t, &MockValidator{}, validator)
			}
		})
	}
}

// TestConfigFromString tests loading configuration from a string with specified format
func TestConfigFromString(t *testing.T) {
	// Setup test
	RegisterExecutor("string-executor", &executors.MockExecutor{})

	t.Run("Success: Load YAML configuration from string", func(t *testing.T) {
		yamlContent := `
clients:
  - name: string-client
    type: openai
    config:
      apiKey: test-key
flows:
  - name: string-flow
    steps:
      - name: string-step
        executor:
          type: string-executor
`
		// Execute
		err := ConfigFromString(yamlContent, "yaml")

		// Verify
		assert.NoError(t, err)

		// Verify the flow was registered
		flow, err := GetFlow("string-flow")
		assert.NoError(t, err)
		assert.NotNil(t, flow)
		assert.Equal(t, "string-flow", flow.Name)
	})

	t.Run("Success: Load JSON configuration from string", func(t *testing.T) {
		jsonContent := `{
  "clients": [
    {
      "name": "json-client",
      "type": "ollama",
      "config": {
        "model": "test-model"
      }
    }
  ],
  "flows": [
    {
      "name": "json-flow",
      "steps": [
        {
          "name": "json-step",
          "executor": {
            "type": "string-executor"
          }
        }
      ]
    }
  ]
}`
		// Execute
		err := ConfigFromString(jsonContent, "json")

		// Verify
		assert.NoError(t, err)

		// Verify the flow was registered
		flow, err := GetFlow("json-flow")
		assert.NoError(t, err)
		assert.NotNil(t, flow)
		assert.Equal(t, "json-flow", flow.Name)
	})

	t.Run("Success: Load TOML configuration from string", func(t *testing.T) {
		tomlContent := `
[[clients]]
name = "toml-client"
type = "openai"
config = { apiKey = "test-key-toml" }

[[flows]]
name = "toml-flow"
steps = [
  { name = "toml-step", executor = { type = "string-executor" } }
]
`
		// Execute
		err := ConfigFromString(tomlContent, "toml")

		// Verify
		assert.NoError(t, err)

		// Verify the flow was registered
		flow, err := GetFlow("toml-flow")
		assert.NoError(t, err)
		assert.NotNil(t, flow)
		assert.Equal(t, "toml-flow", flow.Name)
	})
	t.Run("Failure: Invalid configuration content", func(t *testing.T) {
		invalidContent := `
clients: - broken yaml
flows: []
`
		// Execute
		err := ConfigFromString(invalidContent, "yaml")

		// Verify
		assert.Error(t, err)
	})

	t.Run("Failure: Invalid configuration structure", func(t *testing.T) {
		invalidStructContent := `
clients:
  - name: invalid-client
    type: openai
flows:
  - name: invalid-flow
    steps:
      - name: invalid-step
        executor:
          type: non-existent-executor
`
		// Execute
		err := ConfigFromString(invalidStructContent, "yaml")

		// Verify
		assert.Error(t, err)
	})

	t.Run("Failure: Incorrect format type specified", func(t *testing.T) {
		yamlContent := `
clients:
  - name: yaml-client
    type: openai
flows:
  - name: yaml-flow
    steps:
      - name: yaml-step
        executor:
          type: string-executor
`
		// Execute with wrong format type
		err := ConfigFromString(yamlContent, "json")

		// Verify
		assert.Error(t, err)
	})

	t.Run("Failure: Empty configuration content", func(t *testing.T) {
		// Execute with empty configuration content
		err := ConfigFromString("", "yaml")

		// Verify
		assert.Error(t, err)
	})
}

// TestConfigFromFile tests loading configuration from a file
func TestConfigFromFile(t *testing.T) {
	// Setup test
	RegisterExecutor("file-executor", &executors.MockExecutor{})

	// Create a temporary test config file
	yamlContent := `
clients:
  - name: file-client
    type: openai
    config:
      apiKey: test-key-file
flows:
  - name: file-flow
    steps:
      - name: file-step
        executor:
          type: file-executor
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	t.Run("Success: Load configuration from file", func(t *testing.T) {
		// Execute
		err := ConfigFromFile(tmpFile.Name())

		// Verify
		assert.NoError(t, err)

		// Verify the flow was registered
		flow, err := GetFlow("file-flow")
		assert.NoError(t, err)
		assert.NotNil(t, flow)
		assert.Equal(t, "file-flow", flow.Name)
	})

	t.Run("Failure: File does not exist", func(t *testing.T) {
		// Execute
		err := ConfigFromFile("non-existent-file.yaml")

		// Verify
		assert.Error(t, err)
	})

	// Create an invalid config file
	invalidContent := `
clients: - broken yaml
flows: []
`
	invalidFile, err := os.CreateTemp("", "invalid-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(invalidFile.Name())

	_, err = invalidFile.WriteString(invalidContent)
	assert.NoError(t, err)
	err = invalidFile.Close()
	assert.NoError(t, err)

	t.Run("Failure: Invalid configuration file format", func(t *testing.T) {
		// Execute
		err := ConfigFromFile(invalidFile.Name())

		// Verify
		assert.Error(t, err)
	})
}

func TestNewAgentFromConfig_Success(t *testing.T) {
	// Setup
	GlobalRegistry = &registry.AnyiRegistry{
		Flows:      make(map[string]*flow.Flow),
		Clients:    make(map[string]llm.Client),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
		Agents:     make(map[string]*agent.Agent),
	}

	// Register a mock flow
	mockFlow := &flow.Flow{Name: "test-flow"}
	err := RegisterFlow("test-flow", mockFlow)
	assert.NoError(t, err)

	// Create agent config
	agentConfig := &agent.AgentConfig{
		Name:              "test-agent",
		Role:              "Test Role",
		PreferredLanguage: "English",
		BackStory:         "Test backstory",
		Flows:             []string{"test-flow"},
	}

	// Execute
	agentInstance, err := NewAgentFromConfig(agentConfig)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, agentInstance)
	assert.Equal(t, "Test Role", agentInstance.Role)
	assert.Equal(t, "English", agentInstance.PreferredLanguage)
	assert.Equal(t, "Test backstory", agentInstance.BackStory)
	assert.Len(t, agentInstance.Flows, 1)
	assert.Equal(t, mockFlow, agentInstance.Flows[0])

	// Verify agent was registered
	registeredAgent, err := GetAgent("test-agent")
	assert.NoError(t, err)
	assert.Equal(t, agentInstance, registeredAgent)
}

func TestNewAgentFromConfig_WithNil(t *testing.T) {
	// Execute
	agentInstance, err := NewAgentFromConfig(nil)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, agentInstance)
	assert.Equal(t, "agent config is nil", err.Error())
}

func TestNewAgentFromConfig_WithNonExistentFlow(t *testing.T) {
	// Setup
	GlobalRegistry = &registry.AnyiRegistry{
		Flows:      make(map[string]*flow.Flow),
		Clients:    make(map[string]llm.Client),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
		Agents:     make(map[string]*agent.Agent),
	}

	// Create agent config with non-existent flow
	agentConfig := &agent.AgentConfig{
		Name:  "test-agent",
		Role:  "Test Role",
		Flows: []string{"non-existent-flow"},
	}

	// Execute
	agentInstance, err := NewAgentFromConfig(agentConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, agentInstance)
	assert.Contains(t, err.Error(), "flow \"non-existent-flow\" not found for agent \"test-agent\"")
}

func TestNewAgentFromConfig_WithClientName(t *testing.T) {
	// Setup
	GlobalRegistry = &registry.AnyiRegistry{
		Flows:      make(map[string]*flow.Flow),
		Clients:    make(map[string]llm.Client),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
		Agents:     make(map[string]*agent.Agent),
	}

	// Register a mock flow
	mockFlow := &flow.Flow{Name: "test-flow"}
	err := RegisterFlow("test-flow", mockFlow)
	assert.NoError(t, err)

	// Register a mock client
	mockClient := &test.MockClient{}
	err = RegisterClient("test-client", mockClient)
	assert.NoError(t, err)

	// Create agent config with client name
	agentConfig := &agent.AgentConfig{
		Name:              "test-agent",
		Role:              "Test Role",
		PreferredLanguage: "English",
		BackStory:         "Test backstory",
		ClientName:        "test-client",
		Flows:             []string{"test-flow"},
	}

	// Execute
	agentInstance, err := NewAgentFromConfig(agentConfig)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, agentInstance)
	assert.Equal(t, "Test Role", agentInstance.Role)
	assert.Equal(t, "English", agentInstance.PreferredLanguage)
	assert.Equal(t, "Test backstory", agentInstance.BackStory)
	assert.Equal(t, mockClient, agentInstance.Client)
	assert.Len(t, agentInstance.Flows, 1)
	assert.Equal(t, mockFlow, agentInstance.Flows[0])

	// Verify agent was registered
	registeredAgent, err := GetAgent("test-agent")
	assert.NoError(t, err)
	assert.Equal(t, agentInstance, registeredAgent)
}

func TestNewAgentFromConfig_WithNonExistentClient(t *testing.T) {
	// Setup
	GlobalRegistry = &registry.AnyiRegistry{
		Flows:      make(map[string]*flow.Flow),
		Clients:    make(map[string]llm.Client),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
		Agents:     make(map[string]*agent.Agent),
	}

	// Register a mock flow
	mockFlow := &flow.Flow{Name: "test-flow"}
	err := RegisterFlow("test-flow", mockFlow)
	assert.NoError(t, err)

	// Create agent config with non-existent client
	agentConfig := &agent.AgentConfig{
		Name:       "test-agent",
		Role:       "Test Role",
		ClientName: "non-existent-client",
		Flows:      []string{"test-flow"},
	}

	// Execute
	agentInstance, err := NewAgentFromConfig(agentConfig)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, agentInstance)
	assert.Contains(t, err.Error(), "client \"non-existent-client\" not found for agent \"test-agent\"")
}

func TestConfig_WithAgents(t *testing.T) {
	// Setup
	GlobalRegistry = &registry.AnyiRegistry{
		Flows:      make(map[string]*flow.Flow),
		Clients:    make(map[string]llm.Client),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
		Agents:     make(map[string]*agent.Agent),
	}

	// Register a mock flow
	mockFlow := &flow.Flow{Name: "test-flow"}
	err := RegisterFlow("test-flow", mockFlow)
	assert.NoError(t, err)

	// Create config with agents
	config := &AnyiConfig{
		Agents: []agent.AgentConfig{
			{
				Name:  "test-agent",
				Role:  "Test Role",
				Flows: []string{"test-flow"},
			},
		},
	}

	// Execute
	err = Config(config)

	// Verify
	assert.NoError(t, err)

	// Verify agent was created and registered
	agentInstance, err := GetAgent("test-agent")
	assert.NoError(t, err)
	assert.NotNil(t, agentInstance)
	assert.Equal(t, "Test Role", agentInstance.Role)
	assert.Len(t, agentInstance.Flows, 1)
	assert.Equal(t, mockFlow, agentInstance.Flows[0])
}
