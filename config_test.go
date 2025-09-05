package anyi

import (
	"testing"
	"time"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

func TestAgentConfigStructure(t *testing.T) {
	// Test that we can create an AgentConfig struct
	config := &AgentConfig{
		Name:           "testAgent",
		Role:           "Test Role",
		BackStory:      "Test Backstory",
		AvailableFlows: []string{"flow1", "flow2"},
		MaxIterations:  5,
		MaxRetries:     3,
		Timeout:        "10m",
	}

	assert.Equal(t, "testAgent", config.Name)
	assert.Equal(t, "Test Role", config.Role)
	assert.Equal(t, "Test Backstory", config.BackStory)
	assert.Equal(t, []string{"flow1", "flow2"}, config.AvailableFlows)
	assert.Equal(t, 5, config.MaxIterations)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "10m", config.Timeout)
}

// MockExecutor is a mock executor for testing
type MockExecutor struct{}

func (m MockExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	return &flowContext, nil
}

func (m MockExecutor) Init() error {
	return nil
}

// MockValidator is a mock validator for testing
type MockValidator struct{}

func (m MockValidator) Init() error {
	return nil
}

func (m MockValidator) Validate(output string, step *flow.Step) bool {
	return true
}

func TestNewFlowFromConfig_WithDescription(t *testing.T) {
	// Setup
	registry.Global = &registry.Registry{
		Clients:    make(map[string]llm.Client),
		Flows:      make(map[string]*flow.Flow),
		Agents:     make(map[string]*agent.Agent),
		Executors:  make(map[string]flow.StepExecutor),
		Validators: make(map[string]flow.StepValidator),
		Formatters: make(map[string]chat.PromptFormatter),
	}
	registry.RegisterClient("test-client", &test.MockClient{})
	registry.RegisterExecutor("test-executor", &MockExecutor{})
	registry.RegisterValidator("test-validator", &MockValidator{})

	flowConfig := &FlowConfig{
		ClientName:  "test-client",
		Name:        "test-flow",
		Description: "This is a test flow for demonstration purposes",
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
	assert.Equal(t, flowConfig.Description, flowInstance.Description)
	assert.Equal(t, 1, len(flowInstance.Steps))
}
func TestAnyiConfigWithAgents(t *testing.T) {
	// Test that we can add agents to AnyiConfig
	config := &AnyiConfig{
		Agents: []AgentConfig{
			{
				Name:           "testAgent",
				Role:           "Test Role",
				BackStory:      "Test Backstory",
				AvailableFlows: []string{"flow1", "flow2"},
			},
		},
	}

	assert.Len(t, config.Agents, 1)
	assert.Equal(t, "testAgent", config.Agents[0].Name)
}

func TestTimeoutParsing(t *testing.T) {
	// Test timeout parsing
	duration, err := time.ParseDuration("30m")
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Minute, duration)

	duration, err = time.ParseDuration("1h")
	assert.NoError(t, err)
	assert.Equal(t, time.Hour, duration)
}
