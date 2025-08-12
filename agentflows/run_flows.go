package agentflows

import (
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/registry"
)

func RunPlanningFlow(planningData AgentPlanningData) (string, error) {

	planningFlow, err := registry.GetFlow("Anyi_AgentPlanningFlow")

	if err != nil {
		return "", err
	}
	context := &flow.FlowContext{}
	context.Memory = planningData

	context, err = planningFlow.Run(*context)
	if err != nil {
		return "", err
	}
	planningText := context.Text

	return planningText, nil
}
