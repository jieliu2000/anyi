package agent

import (
	"testing"
	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

func TestAgentCreation(t *testing.T) {
	// Create a mock flow
	mockFlow := &flow.Flow{
		Name: "test-flow",
	}

	// Create an agent with the mock flow
	agent := &Agent{
		Role:              "Test Role",
		PreferredLanguage: "English",
		BackStory:         "Test backstory",
		Flows:             []*flow.Flow{mockFlow},
	}

	// Verify the agent was created correctly
	assert.Equal(t, "Test Role", agent.Role)
	assert.Equal(t, "English", agent.PreferredLanguage)
	assert.Equal(t, "Test backstory", agent.BackStory)
	assert.Len(t, agent.Flows, 1)
	assert.Equal(t, mockFlow, agent.Flows[0])
}

func TestAgentContextCreation(t *testing.T) {
	// Create an agent context
	context := &AgentContext{
		Goal: "Test goal",
		ShortTermMemory: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		ExecuteLog: []string{
			"log entry 1",
			"log entry 2",
		},
	}

	// Verify the context was created correctly
	assert.Equal(t, "Test goal", context.Goal)
	assert.Len(t, context.ShortTermMemory, 2)
	assert.Equal(t, "value1", context.ShortTermMemory["key1"])
	assert.Equal(t, 123, context.ShortTermMemory["key2"])
	assert.Len(t, context.ExecuteLog, 2)
	assert.Equal(t, "log entry 1", context.ExecuteLog[0])
	assert.Equal(t, "log entry 2", context.ExecuteLog[1])
}

func TestAgentJobCreation(t *testing.T) {
	// Create a mock flow
	mockFlow := &flow.Flow{
		Name: "test-flow",
	}

	// Create an agent
	agent := &Agent{
		Role:              "Test Role",
		PreferredLanguage: "English",
		BackStory:         "Test backstory",
		Flows:             []*flow.Flow{mockFlow},
	}

	// Create an agent context
	context := &AgentContext{
		Goal: "Test goal",
		ShortTermMemory: map[string]interface{}{
			"key": "value",
		},
		ExecuteLog: []string{},
	}

	// Create an agent job
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
		FlowExecutionPlan: []*flow.Flow{mockFlow},
	}

	// Verify the job was created correctly
	assert.Equal(t, agent, job.Agent)
	assert.Equal(t, context, job.Context)
	assert.Equal(t, "running", job.Status)
	assert.Len(t, job.FlowExecutionPlan, 1)
	assert.Equal(t, mockFlow, job.FlowExecutionPlan[0])
}