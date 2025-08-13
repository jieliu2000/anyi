package agent

import (
	"github.com/jieliu2000/anyi/agentflows/model"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
)

func RunPlanningFlow(planningData model.AgentPlanningData, client llm.Client) (string, error) {

	context := &flow.FlowContext{}
	context.Memory = planningData

	AgentPlanningFlow.ClientImpl = client
	context, err := AgentPlanningFlow.Run(*context)
	if err != nil {
		return "", err
	}
	planningText := context.Text

	return planningText, nil
}
