package agent

import (
	"testing"

	"github.com/jieliu2000/anyi/registry"
)

func TestAgent(t *testing.T) {
	// Create an agent
	agent := &Agent{
		Name:        "test_agent",
		Description: "A test agent",
		Flows:       []string{"test_flow"},
		ClientName:  "test_client",
		Memory:      NewSimpleMemory(),
		Config:      make(map[string]any),
	}

	// Test basic functionality
	if agent.GetName() != "test_agent" {
		t.Errorf("Expected name 'test_agent', got '%s'", agent.GetName())
	}

	if agent.GetClientName() != "test_client" {
		t.Errorf("Expected client name 'test_client', got '%s'", agent.GetClientName())
	}

	flows := agent.GetFlows()
	if len(flows) != 1 || flows[0] != "test_flow" {
		t.Errorf("Expected flows ['test_flow'], got %v", flows)
	}
}

func TestSimpleMemory(t *testing.T) {
	memory := NewSimpleMemory()

	// Test storing and retrieving tasks
	task := &TaskResult{
		Objective:   "test objective",
		Status:      "completed",
		FinalOutput: "test output",
	}

	memory.StoreTask(task)

	// Get task history
	history := memory.GetTaskHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 task in history, got %d", len(history))
	}

	if history[0].Objective != "test objective" {
		t.Errorf("Expected objective 'test objective', got '%s'", history[0].Objective)
	}

	// Get specific task
	retrieved, err := memory.GetTask("test objective")
	if err != nil {
		t.Errorf("Failed to retrieve task: %v", err)
	}

	if retrieved.Objective != "test objective" {
		t.Errorf("Expected objective 'test objective', got '%s'", retrieved.Objective)
	}

	// Test clearing memory
	memory.Clear()
	history = memory.GetTaskHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d items", len(history))
	}
}

func TestAgentRegistry(t *testing.T) {
	reg := registry.NewRegistry()

	// Create and register an agent
	agent := Agent{
		Name:        "test_agent",
		Description: "A test agent",
		Flows:       []string{"test_flow"},
		ClientName:  "test_client",
		Memory:      NewSimpleMemory(),
		Config:      make(map[string]any),
	}

	err := reg.RegisterAgent(agent.Name, agent)
	if err != nil {
		t.Errorf("Failed to register agent: %v", err)
	}

	// Retrieve the agent
	retrievedInterface, err := reg.GetAgent("test_agent")
	if err != nil {
		t.Errorf("Failed to get agent: %v", err)
	}

	retrieved, ok := retrievedInterface.(Agent)
	if !ok {
		t.Error("Retrieved agent is not of correct type")
	}

	if retrieved.Name != "test_agent" {
		t.Errorf("Expected name 'test_agent', got '%s'", retrieved.Name)
	}
}

func TestAgentConfig(t *testing.T) {
	yamlData := `
agents:
  - name: test_agent
    description: A test agent
    flows:
      - flow1
      - flow2
    clientName: test_client
    config:
      param1: value1
      param2: 42
`

	agents, err := LoadAgentsFromYAML([]byte(yamlData))
	if err != nil {
		t.Errorf("Failed to load agents from YAML: %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	agent := agents[0]
	if agent.Name != "test_agent" {
		t.Errorf("Expected name 'test_agent', got '%s'", agent.Name)
	}

	if len(agent.Flows) != 2 {
		t.Errorf("Expected 2 flows, got %d", len(agent.Flows))
	}

	if agent.ClientName != "test_client" {
		t.Errorf("Expected client 'test_client', got '%s'", agent.ClientName)
	}
}

func TestValidateAgent(t *testing.T) {
	// Valid agent
	validAgent := &Agent{
		Name:       "test_agent",
		ClientName: "test_client",
		Flows:      []string{"flow1"},
	}

	err := ValidateAgent(validAgent)
	if err != nil {
		t.Errorf("Valid agent failed validation: %v", err)
	}

	// Invalid agent - no name
	invalidAgent1 := &Agent{
		Name:       "",
		ClientName: "test_client",
		Flows:      []string{"flow1"},
	}

	err = ValidateAgent(invalidAgent1)
	if err == nil {
		t.Error("Agent with empty name should fail validation")
	}

	// Invalid agent - no client
	invalidAgent2 := &Agent{
		Name:       "test_agent",
		ClientName: "",
		Flows:      []string{"flow1"},
	}

	err = ValidateAgent(invalidAgent2)
	if err == nil {
		t.Error("Agent with empty client should fail validation")
	}

	// Invalid agent - no flows
	invalidAgent3 := &Agent{
		Name:       "test_agent",
		ClientName: "test_client",
		Flows:      []string{},
	}

	err = ValidateAgent(invalidAgent3)
	if err == nil {
		t.Error("Agent with no flows should fail validation")
	}
}
