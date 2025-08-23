package anyi

import (
	"testing"
)

func TestAgentConfigurationBasic(t *testing.T) {
	// Test basic agent configuration loading without full mock setup
	
	// Initialize the framework
	Init()

	// Test agent configuration storage
	agentConfig := &AgentConfig{
		Name:        "test-agent",
		Description: "A test agent",
		Flows:       []string{"test-flow"},
		ClientName:  "test-client",
		Config:      map[string]interface{}{"test": "value"},
	}

	// Store the agent configuration directly (bypassing validation for this test)
	GlobalRegistry.mu.Lock()
	GlobalRegistry.AgentConfigs[agentConfig.Name] = agentConfig
	GlobalRegistry.mu.Unlock()

	// Test that agent configuration was stored
	agents := ListAgents()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	if agents[0] != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agents[0])
	}

	// Test getting agent configuration
	agentWrapper, err := GetAgent("test-agent")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if agentWrapper.Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agentWrapper.Name)
	}

	if agentWrapper.Description != "A test agent" {
		t.Errorf("Expected description 'A test agent', got '%s'", agentWrapper.Description)
	}

	if len(agentWrapper.Flows) != 1 || agentWrapper.Flows[0] != "test-flow" {
		t.Errorf("Expected flows ['test-flow'], got %v", agentWrapper.Flows)
	}

	if agentWrapper.ClientName != "test-client" {
		t.Errorf("Expected client name 'test-client', got '%s'", agentWrapper.ClientName)
	}
}

func TestAgentConfigTypes(t *testing.T) {
	// Test that AgentConfig has the expected structure
	config := &AgentConfig{
		Name:        "test",
		Description: "test description",
		Flows:       []string{"flow1", "flow2"},
		ClientName:  "client",
		Config:      map[string]interface{}{"key": "value"},
	}

	if config.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", config.Name)
	}

	if len(config.Flows) != 2 {
		t.Errorf("Expected 2 flows, got %d", len(config.Flows))
	}

	if config.Config["key"] != "value" {
		t.Errorf("Expected config value 'value', got '%v'", config.Config["key"])
	}
}
