package registry

import (
	"testing"

	"github.com/jieliu2000/anyi/agent/agentmodel"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/stretchr/testify/assert"
)

type mockExecutor struct{}

func (e *mockExecutor) Init() error {
	return nil
}

func (e *mockExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	return &flowContext, nil
}

type mockValidator struct{}

func (v *mockValidator) Init() error {
	return nil
}

func (v *mockValidator) Validate(stepOutput string, step *flow.Step) bool {
	return true
}

func TestRegisterAndGetFlow(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.Flows = make(map[string]*flow.Flow)
	GlobalRegistry.Mu.Unlock()

	client := &test.MockClient{}
	flowInstance, _ := flow.NewFlow(client, "test-flow")

	// Test registering a flow
	err := RegisterFlow("test-flow", flowInstance)
	assert.NoError(t, err)

	// Test getting a flow
	retrievedFlow, err := GetFlow("test-flow")
	assert.NoError(t, err)
	assert.Equal(t, flowInstance, retrievedFlow)

	// Test getting a non-existent flow
	_, err = GetFlow("non-existent")
	assert.Error(t, err)
	assert.Equal(t, "no flow found with the given name: non-existent", err.Error())

	// Test getting flow with empty name
	_, err = GetFlow("")
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())

	// Test registering a flow with existing name
	err = RegisterFlow("test-flow", flowInstance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flow with name \"test-flow\" already exists")
}

func TestRegisterAndGetClient(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.Clients = make(map[string]llm.Client)
	GlobalRegistry.Mu.Unlock()

	client := &test.MockClient{}

	// Test registering a client
	err := RegisterClient("test-client", client)
	assert.NoError(t, err)

	// Test getting a client
	retrievedClient, err := GetClient("test-client")
	assert.NoError(t, err)
	assert.Equal(t, client, retrievedClient)

	// Test getting a non-existent client
	_, err = GetClient("non-existent")
	assert.Error(t, err)
	assert.Equal(t, "no client found with the given name: non-existent", err.Error())

	// Test getting client with empty name
	_, err = GetClient("")
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())

	// Test registering a client with existing name
	err = RegisterClient("test-client", client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client with name \"test-client\" already exists")

	// Test registering a nil client
	err = RegisterClient("nil-client", nil)
	assert.Error(t, err)
	assert.Equal(t, "client cannot be empty", err.Error())
}

func TestRegisterAndGetExecutor(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.Executors = make(map[string]flow.StepExecutor)
	GlobalRegistry.Mu.Unlock()

	executor := &mockExecutor{}

	// Test registering an executor
	err := RegisterExecutor("test-executor", executor)
	assert.NoError(t, err)

	// Test getting an executor
	retrievedExecutor, err := GetExecutor("test-executor")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedExecutor)

	// Test getting a non-existent executor
	_, err = GetExecutor("non-existent")
	assert.Error(t, err)
	assert.Equal(t, "no executor found with the given name: non-existent", err.Error())

	// Test getting executor with empty name
	_, err = GetExecutor("")
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())

	// Test registering an executor with existing name
	err = RegisterExecutor("test-executor", executor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor type with the name test-executor already exists")
}

func TestRegisterAndGetValidator(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.Validators = make(map[string]flow.StepValidator)
	GlobalRegistry.Mu.Unlock()

	validator := &mockValidator{}

	// Test registering a validator
	err := RegisterValidator("test-validator", validator)
	assert.NoError(t, err)

	// Test getting a validator
	retrievedValidator, err := GetValidator("test-validator")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedValidator)

	// Test getting a non-existent validator
	_, err = GetValidator("non-existent")
	assert.Error(t, err)
	assert.Equal(t, "no validator found with the given name: non-existent", err.Error())

	// Test getting validator with empty name
	_, err = GetValidator("")
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())

	// Test registering a validator with existing name
	err = RegisterValidator("test-validator", validator)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validator type with the name test-validator already exists")
}

func TestRegisterAndGetFormatter(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.Formatters = make(map[string]chat.PromptFormatter)
	GlobalRegistry.Mu.Unlock()

	// Test registering a formatter
	err := RegisterFormatter("test-formatter", nil)
	assert.NoError(t, err)

	// Test getting a formatter
	retrievedFormatter := GetFormatter("test-formatter")
	assert.Nil(t, retrievedFormatter)

	// Test getting a non-existent formatter (should return nil)
	nonExistentFormatter := GetFormatter("non-existent")
	assert.Nil(t, nonExistentFormatter)

	// Test getting formatter with empty name
	emptyNameFormatter := GetFormatter("")
	assert.Nil(t, emptyNameFormatter)
}

func TestRegisterAndGetAgent(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.Agents = make(map[string]*agentmodel.Agent)
	GlobalRegistry.Mu.Unlock()

	agentInstance := &agentmodel.Agent{Role: "test-agent"}

	// Test registering an agent
	err := RegisterAgent(agentInstance)
	assert.NoError(t, err)

	// Test getting an agent
	retrievedAgent, err := GetAgent("test-agent")
	assert.NoError(t, err)
	assert.Equal(t, agentInstance, retrievedAgent)

	// Test getting a non-existent agent
	_, err = GetAgent("non-existent")
	assert.Error(t, err)
	assert.Equal(t, "no agent found with the given name: non-existent", err.Error())

	// Test getting agent with empty name
	_, err = GetAgent("")
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())

	// Test registering an agent with existing name
	err = RegisterAgent(agentInstance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent with role \"test-agent\" already exists")
}

func TestDefaultClientName(t *testing.T) {
	// Reset registry for clean test
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.DefaultClientName = ""
	GlobalRegistry.Clients = make(map[string]llm.Client)
	GlobalRegistry.Mu.Unlock()

	// Test setting default client name
	GlobalRegistry.Mu.Lock()
	GlobalRegistry.DefaultClientName = "default-client"
	GlobalRegistry.Mu.Unlock()

	// Verify it was Set
	GlobalRegistry.Mu.RLock()
	defaultName := GlobalRegistry.DefaultClientName
	GlobalRegistry.Mu.RUnlock()
	assert.Equal(t, "default-client", defaultName)
}
