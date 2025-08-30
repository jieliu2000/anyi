package anyi

import (
	"testing"
	"time"

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
