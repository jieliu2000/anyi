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

func (m *MockExecutor) Run(memory flow.ShortTermMemory, Step *flow.Step) (*flow.ShortTermMemory, error) {

	return &memory, nil
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
				Executor:  "test-executor",
				Validator: "test-validator",
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
				Executor:  "test-executor",
				Validator: "test-validator",
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
				Executor:  "invalid-executor",
				Validator: "test-validator",
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
				Validator: "test-validator",
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
			Name: "executor1",
			Config: map[string]interface{}{
				"param1": "value1",
				"param2": 10,
			},
		}

		executor, err := NewExecutorFromConfig(executorConfig)

		assert.NoError(t, err)
		assert.NotNil(t, executor)

		assert.Equal(t, executor1.Param1, "value1")
		assert.Equal(t, executor1.Param2, 10)

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
		Executors: []ExecutorConfig{
			{
				Name: "executor1",
				Type: "llm",
				Config: map[string]interface{}{
					"requestTimeout": 1000,
				},
			},
		},
		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor:      "executor1",
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
		Executors: []ExecutorConfig{
			{
				Name: "executor1",
				Type: "llm",
				Config: map[string]interface{}{
					"requestTimeout": 1000,
				},
			},
		},
		Validators: []ValidatorConfig{
			{
				Name: "validator1",
				Type: "http",
				Config: map[string]interface{}{
					"requestTimeout": 1000,
				},
			},
		},
		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor:      "no-executor",
						Validator:     "validator1",
						ClientName:    "client1",
						MaxRetryTimes: 1,
					},
				},
			},
		},
	}
	err := Config(&config)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "step executor no-executor is not found")
}

func TestConfigWithInvalidValidator(t *testing.T) {
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
		Executors: []ExecutorConfig{
			{
				Name: "executor1",
				Type: "llm",
			},
		},
		Validators: []ValidatorConfig{
			{
				Name: "validator1",
				Type: "http",
				Config: map[string]interface{}{
					"requestTimeout": 1000,
				},
			},
		},
		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor:      "executor1",
						Validator:     "no-validator",
						ClientName:    "client1",
						MaxRetryTimes: 1,
					},
				},
			},
		},
	}
	err := Config(&config)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "validator type http is not found")
}

func TestConfigWithInvalidClient(t *testing.T) {
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
		Executors: []ExecutorConfig{
			{
				Name: "executor1",
				Type: "llm",
				Config: map[string]interface{}{
					"requestTimeout": 1000,
				},
			},
		},

		Flows: []FlowConfig{
			{
				Name: "flow1",
				Steps: []StepConfig{
					{
						Executor:      "executor1",
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
