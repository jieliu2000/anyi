package anyi

import (
	"log"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestNewFlow(t *testing.T) {

	client := test.MockClient{}
	flow := NewFlow(&client, "flow1")

	assert.NotNil(t, flow)
	assert.Equal(t, "flow1", flow.Name)
	assert.Equal(t, &client, flow.clientImpl)

}

func TestNewLLMStepWithTemplateFile(t *testing.T) {

	step, err := NewLLMStepWithTemplateFile("internal/test/test_prompt1.tmpl", "system_message", nil)

	assert.NoError(t, err)
	assert.NotNil(t, step)

	stepConfig := step.StepConfig.(LLMFlowStepConfig)
	assert.Equal(t, "system_message", stepConfig.SystemMessage)

	formatter := stepConfig.TemplateFormatter
	assert.NotNil(t, formatter)
	assert.Equal(t, "internal/test/test_prompt1.tmpl", formatter.File)

	type AgentTasks struct {
		Tasks     []string
		Objective string
	}

	tasks := AgentTasks{
		Tasks:     []string{"task1", "task2"},
		Objective: "objective",
	}

	output, err := formatter.Format(tasks)
	assert.Nil(t, err)
	log.Printf("output: %s", output)

	assert.Greater(t, len(output), 10)

}

func TestNewLLMStepWithTemplateString(t *testing.T) {
	step, err := NewLLMStepWithTemplate("Analyze this target and break it into action plans: {{.}}", "system_message", nil)

	assert.NoError(t, err)
	assert.NotNil(t, step)

	stepConfig := step.StepConfig.(LLMFlowStepConfig)

	assert.Equal(t, "system_message", stepConfig.SystemMessage)

	formatter := stepConfig.TemplateFormatter

	assert.NotNil(t, formatter)

	output, err := formatter.Format("Build an AI operating system")
	assert.Nil(t, err)

	assert.Equal(t, "Analyze this target and break it into action plans: Build an AI operating system", output)

}
