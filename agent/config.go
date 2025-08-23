package agent

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AgentConfig represents the configuration structure for agents.
type AgentConfig struct {
	Agents []Agent `mapstructure:"agents"`
}

// LoadAgentsFromFile loads agent configurations from a YAML file.
func LoadAgentsFromFile(filename string) ([]Agent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	return LoadAgentsFromYAML(data)
}

// LoadAgentsFromYAML loads agent configurations from YAML data.
func LoadAgentsFromYAML(data []byte) ([]Agent, error) {
	var config AgentConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Initialize memory for each agent
	for i := range config.Agents {
		if config.Agents[i].Memory == nil {
			config.Agents[i].Memory = NewSimpleMemory()
		}
		if config.Agents[i].Config == nil {
			config.Agents[i].Config = make(map[string]any)
		}
	}

	return config.Agents, nil
}

// SaveAgentsToFile saves agent configurations to a YAML file.
func SaveAgentsToFile(agents []Agent, filename string) error {
	config := AgentConfig{Agents: agents}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal agents to YAML: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filename, err)
	}

	return nil
}

// ValidateAgent validates an agent configuration.
func ValidateAgent(agent *Agent) error {
	if agent.Name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}

	if agent.ClientName == "" {
		return fmt.Errorf("agent '%s' must have a client configured", agent.Name)
	}

	if len(agent.Flows) == 0 {
		return fmt.Errorf("agent '%s' must have at least one flow configured", agent.Name)
	}

	return nil
}

// ValidateAgents validates a list of agent configurations.
func ValidateAgents(agents []Agent) error {
	if len(agents) == 0 {
		return fmt.Errorf("at least one agent must be configured")
	}

	names := make(map[string]bool)
	for i, agent := range agents {
		// Validate individual agent
		if err := ValidateAgent(&agent); err != nil {
			return fmt.Errorf("agent %d validation failed: %w", i, err)
		}

		// Check for duplicate names
		if names[agent.Name] {
			return fmt.Errorf("duplicate agent name: %s", agent.Name)
		}
		names[agent.Name] = true
	}

	return nil
}
