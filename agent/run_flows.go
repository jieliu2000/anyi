package agent

import (
	"github.com/jieliu2000/anyi/agentflows/model"
	"github.com/jieliu2000/anyi/flow"
)

func RunPlanningFlow(planningData model.AgentPlanningData) (string, error) {

	context := &flow.FlowContext{}
	context.Memory = planningData

	context, err := AgentPlanningFlow.Run(*context)
	if err != nil {
		return "", err
	}
	planningText := context.Text

	return planningText, nil
}
