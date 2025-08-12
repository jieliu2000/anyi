package agentflows

import (
	"encoding/json"
	"testing"

	"github.com/jieliu2000/anyi/agentflows/model"
	"github.com/stretchr/testify/assert"
)

func TestAgentPlanningData(t *testing.T) {
	// Test model.AgentPlanningData structure
	planningData := &model.AgentPlanningData{
		Role:              "Senior Data Scientist",
		BackStory:         "An experienced data scientist with expertise in machine learning and data analysis",
		PreferredLanguage: "English",
		Goal:              "Analyze customer data and provide insights",
		AvailableFlows: []model.FlowInfo{
			{Name: "DataAnalysisFlow", Description: "Analyzes data and generates reports"},
			{Name: "ModelTrainingFlow", Description: "Trains machine learning models"},
		},
	}

	assert.Equal(t, "Senior Data Scientist", planningData.Role)
	assert.Equal(t, "English", planningData.PreferredLanguage)
	assert.Len(t, planningData.AvailableFlows, 2)
}

func TestFlowInfo(t *testing.T) {
	// Test model.FlowInfo structure
	testFlowInfo := model.FlowInfo{
		Name:        "TestFlow",
		Description: "A test flow for unit testing",
	}

	assert.Equal(t, "TestFlow", testFlowInfo.Name)
	assert.Equal(t, "A test flow for unit testing", testFlowInfo.Description)
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
	planningData := &model.AgentPlanningData{
		Role:              "Product Manager",
		BackStory:         "Expert in product strategy and roadmap planning",
		PreferredLanguage: "English",
		Goal:              "Create a product roadmap for Q4",
		AvailableFlows: []model.FlowInfo{
			{Name: "MarketResearchFlow", Description: "Conducts market research"},
			{Name: "StrategyPlanningFlow", Description: "Creates strategic plans"},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(planningData)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize from JSON
	var deserializedData model.AgentPlanningData
	err = json.Unmarshal(jsonData, &deserializedData)
	assert.NoError(t, err)
	assert.Equal(t, planningData.Role, deserializedData.Role)
	assert.Equal(t, planningData.Goal, deserializedData.Goal)
	assert.Len(t, deserializedData.AvailableFlows, 2)
}
