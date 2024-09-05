package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMStepExecutor_Init(t *testing.T) {
	executor := &LLMStepExecutor{}
	err := executor.Init()
	assert.Error(t, err)
	executor = &LLMStepExecutor{
		Template: "Hello, {{.name}}!",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
	assert.Equal(t, "Hello, {{.name}}!", executor.TemplateFormatter.TemplateString)

	executor = &LLMStepExecutor{
		TemplateFile: "../internal/test/test_prompt2.tmpl",
	}

	err = executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.TemplateFormatter)
}
