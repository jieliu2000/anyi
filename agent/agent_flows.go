package agent

import (
	"github.com/jieliu2000/anyi/executors"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
)

// AgentPlanningFlow is a built-in flow that plans the execution steps for an agent's goal
var AgentPlanningFlow *flow.Flow

// AgentReflectionFlow is a built-in flow that checks if the agent's goal has been achieved
var AgentReflectionFlow *flow.Flow

func InitAgentBuiltinFlows() {
	// Initialize the Agent Planning Flow
	AgentPlanningFlow = createAgentPlanningFlow()

	// Initialize the Agent Reflection Flow
	AgentReflectionFlow = &flow.Flow{
		Name:        "Anyi_AgentReflectionFlow",
		Description: "This flow checks if the agent's goal has been achieved and decides next steps",
	}

}

// createAgentPlanningFlow creates and returns the agent planning flow
func createAgentPlanningFlow() *flow.Flow {
	// Create the planning template in English
	planningTemplate := `You are an AI assistant helping to plan execution steps for an agent.

Agent Information:
- Role: {{.Memory.Role}}
- Background: {{.Memory.BackStory}}
{{if .Memory.PreferredLanguage}}- Preferred Language: {{.Memory.PreferredLanguage}}{{end}}

Goal to Achieve: {{.Memory.Goal}}

Available Flows:
{{range .Memory.AvailableFlows}}- {{.Name}}: {{.Description}}
{{end}}

Please create a step-by-step execution plan to achieve the goal. Each step should include:
1. The name of the flow to use (must be from the available flows list above)
2. A clear description of what this step will accomplish

Requirements:
- Output must be in valid JSON format
- Each step should be a JSON object with "flowName" and "description" fields
- The response should be a JSON array of step objects
{{if .Memory.PreferredLanguage}}- All descriptions should be written in {{.Memory.PreferredLanguage}}{{else}}- All descriptions should be written in English{{end}}

Example format:
[
  {
    "flowName": "ExampleFlow1",
    "description": "This step will..."
  },
  {
    "flowName": "ExampleFlow2", 
    "description": "This step will..."
  }
]

Please provide only the JSON array, no additional text.`

	// Create LLM executor with the planning template
	llmExecutor := &executors.LLMExecutor{
		Template:   planningTemplate,
		OutputJSON: true,
	}

	// Create the step with the LLM executor
	planningStep := flow.NewStep(llmExecutor, nil, nil)
	planningStep.Name = "AgentPlanningStep"

	// Create and return the flow
	planningFlow := &flow.Flow{
		Name:        "Anyi_AgentPlanningFlow",
		Description: "This flow uses LLM to plan the execution steps for an agent's goal based on available flows",
		Steps:       []flow.Step{*planningStep},
	}

	return planningFlow
}

// CreateAgentPlanningFlowWithClient creates an agent planning flow with a specific LLM client
func CreateAgentPlanningFlowWithClient(client llm.Client) *flow.Flow {
	planningFlow := createAgentPlanningFlow()
	planningFlow.ClientImpl = client
	return planningFlow
}
