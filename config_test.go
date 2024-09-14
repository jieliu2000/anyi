package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/stretchr/testify/assert"
)

type MockExecutor struct {
	Param1 string
	Param2 int
}

func (m *MockExecutor) Run(flowContext flow.FlowContext, Step *flow.Step) (*flow.FlowContext, error) {

	return &flowContext, nil
}

func (m *MockExecutor) Init() error {

	return nil
}

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

	RegisterClient("test-client", &test.MockClient{})
	RegisterExecutor("test-executor", &MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Name: "name1",
				Executor: &ExecutorConfig{
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
	assert.Equal(t, 1, len(flowInstance.Steps))
	step := flowInstance.Steps[0]
	assert.Equal(t, "name1", step.Name)
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
				Executor: &ExecutorConfig{
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
	RegisterExecutor("test-executor", &MockExecutor{})
	RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName: "test-client",
		Name:       "test-flow",
		Steps: []StepConfig{
			{
				Executor: &ExecutorConfig{
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
	RegisterExecutor("test-executor", &MockExecutor{})
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

func TestNewExecutorFromConfig(t *testing.T) {

	t.Run("Invalid type", func(t *testing.T) {

		executorConfig := &ExecutorConfig{
			Type: "invalid-executor",
		}

		executor, err := NewExecutorFromConfig(executorConfig)

		assert.Error(t, err)
		assert.Nil(t, executor)
	})

	t.Run("Success path with param", func(t *testing.T) {

		executor1 := &MockExecutor{}
		RegisterExecutor("valid-executor", executor1)

		executorConfig := &ExecutorConfig{
			Type: "valid-executor",
			WithConfig: map[string]interface{}{
				"param1": "value1",
				"param2": 10,
			},
		}

		result, err := NewExecutorFromConfig(executorConfig)
		executor := result.(*MockExecutor)

		assert.NoError(t, err)
		assert.NotNil(t, executor)

		assert.Equal(t, "value1", executor.Param1)
		assert.Equal(t, 10, executor.Param2)

	})
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
	RegisterExecutor("executor1", &MockExecutor{})
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
				Steps: []StepConfig{
					{
						Executor: &ExecutorConfig{
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
						Executor: &ExecutorConfig{
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
	RegisterExecutor("executor1", &MockExecutor{})
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
						Executor: &ExecutorConfig{
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
	RegisterExecutor("executor1", &MockExecutor{})
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
						Executor: &ExecutorConfig{
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
