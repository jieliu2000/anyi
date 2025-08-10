package agent

import (
	"github.com/jieliu2000/anyi/flow"
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
		Name: "Anyi_AgentPlanningFlow",
	}

	AgentReflectionFlow = &flow.Flow{
		Name: "Anyi_AgentReflectionFlow",
	}
}
