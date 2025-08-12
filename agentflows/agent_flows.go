package agentflows

import (
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/registry"
)

// AgentPlanningFlow is a built-in flow that plans the execution steps for an agent's goal
var AgentPlanningFlow *flow.Flow

// AgentReflectionFlow is a built-in flow that checks if the agent's goal has been achieved
var AgentReflectionFlow *flow.Flow

func InitAgentBuiltinFlows() {
	// Initialize the built-in flows
	// These would typically be complex flows with multiple steps
	// For now, we just create empty flows as placeholders
	AgentPlanningFlow = &flow.Flow{
		Name:        "Anyi_AgentPlanningFlow",
		Description: "This flow uses LLM to plan the execution steps for an agent's goal based on available flows",
	}

	AgentReflectionFlow = &flow.Flow{
		Name:        "Anyi_AgentReflectionFlow",
		Description: "This flow checks if the agent's goal has been achieved and decides next steps",
	}

	registry.RegisterFlow("Anyi_AgentPlanningFlow", AgentPlanningFlow)
	registry.RegisterFlow("Anyi_AgentReflectionFlow", AgentReflectionFlow)
}
