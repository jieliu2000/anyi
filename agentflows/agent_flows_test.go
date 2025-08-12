package agentflows

import (
	"encoding/json"
	"testing"

	"github.com/jieliu2000/anyi/agent"
	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

func TestAgentPlanningData(t *testing.T) {
	// Test AgentPlanningData structure
	planningData := &AgentPlanningData{
		Role:              "Senior Data Scientist",
		BackStory:         "An experienced data scientist with expertise in machine learning and data analysis",
		PreferredLanguage: "English",
		Goal:              "Analyze customer data and provide insights",
		AvailableFlows: []FlowInfo{
			{Name: "DataAnalysisFlow", Description: "Analyzes data and generates reports"},
			{Name: "ModelTrainingFlow", Description: "Trains machine learning models"},
		},
	}

	assert.Equal(t, "Senior Data Scientist", planningData.Role)
	assert.Equal(t, "English", planningData.PreferredLanguage)
	assert.Len(t, planningData.AvailableFlows, 2)
}

func TestFlowInfo(t *testing.T) {
	// Test FlowInfo structure
	flowInfo := FlowInfo{
		Name:        "TestFlow",
		Description: "A test flow for unit testing",
	}

	assert.Equal(t, "TestFlow", flowInfo.Name)
	assert.Equal(t, "A test flow for unit testing", flowInfo.Description)
}

func TestPrepareAgentPlanningContext(t *testing.T) {
	// Create a test agent
	testAgent := &agent.Agent{
		Role:              "Software Engineer",
		BackStory:         "A skilled developer with 5 years of experience",
		PreferredLanguage: "Chinese",
		Flows: []*flow.Flow{
			{
				Name:        "CodeReviewFlow",
				Description: "Reviews code and provides feedback",
			},
			{
				Name:        "BugFixFlow",
				Description: "Identifies and fixes bugs in code",
			},
		},
	}

	goal := "Fix the authentication bug in the user login system"

	// Test the context preparation
	context := PrepareAgentPlanningContext(testAgent, goal)

	// Verify the context
	assert.NotNil(t, context)
	assert.NotNil(t, context.Memory)

	// Verify the planning data
	planningData, ok := context.Memory.(*AgentPlanningData)
	assert.True(t, ok, "Memory should contain AgentPlanningData")
	assert.Equal(t, "Software Engineer", planningData.Role)
	assert.Equal(t, "A skilled developer with 5 years of experience", planningData.BackStory)
	assert.Equal(t, "Chinese", planningData.PreferredLanguage)
	assert.Equal(t, goal, planningData.Goal)
	assert.Len(t, planningData.AvailableFlows, 2)
	assert.Equal(t, "CodeReviewFlow", planningData.AvailableFlows[0].Name)
	assert.Equal(t, "BugFixFlow", planningData.AvailableFlows[1].Name)
}

func TestCreateAgentPlanningFlow(t *testing.T) {
	// Test flow creation
	InitAgentBuiltinFlows()

	assert.NotNil(t, AgentPlanningFlow)
	assert.Equal(t, "Anyi_AgentPlanningFlow", AgentPlanningFlow.Name)
	assert.Contains(t, AgentPlanningFlow.Description, "plan the execution steps")
	assert.Len(t, AgentPlanningFlow.Steps, 1)
	assert.Equal(t, "AgentPlanningStep", AgentPlanningFlow.Steps[0].Name)
}

func TestAgentPlanningDataJSONSerialization(t *testing.T) {
	// Test JSON serialization/deserialization
	planningData := &AgentPlanningData{
		Role:              "Product Manager",
		BackStory:         "Expert in product strategy and roadmap planning",
		PreferredLanguage: "English",
		Goal:              "Create a product roadmap for Q4",
		AvailableFlows: []FlowInfo{
			{Name: "MarketResearchFlow", Description: "Conducts market research"},
			{Name: "StrategyPlanningFlow", Description: "Creates strategic plans"},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(planningData)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize from JSON
	var deserializedData AgentPlanningData
	err = json.Unmarshal(jsonData, &deserializedData)
	assert.NoError(t, err)
	assert.Equal(t, planningData.Role, deserializedData.Role)
	assert.Equal(t, planningData.Goal, deserializedData.Goal)
	assert.Len(t, deserializedData.AvailableFlows, 2)
}
