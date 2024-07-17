package llm

import (
	"reflect"
	"testing"

	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {

	openaiConfig := openai.DefaultConfig("test")
	client, err := NewClient(openaiConfig)

	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, reflect.TypeOf(&openai.OpenAIClient{}), reflect.TypeOf(client))

	azureOpenAIConfig := azureopenai.NewConfig("test", "test", "test")
	client, err = NewClient(azureOpenAIConfig)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, reflect.TypeOf(&azureopenai.AzureOpenAIClient{}), reflect.TypeOf(client))

	dashscopeConfig := dashscope.NewConfig("key", "model", "baseUrl")
	client, err = NewClient(dashscopeConfig)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, reflect.TypeOf(&dashscope.DashScopeClient{}), reflect.TypeOf(client))
}
