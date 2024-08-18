package llm

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/stretchr/testify/assert"
)

func TestReadConfigFile(t *testing.T) {

	currentDir, err := os.Getwd()

	assert.Nil(t, err)
	assert.NotNil(t, currentDir)

	configFilePath := filepath.Join(currentDir, "openai_test.toml")
	openaiClientConfig, err := utils.UnmarshallConfig(configFilePath, &ClientConfig{})

	assert.Nil(t, err)
	assert.NotNil(t, openaiClientConfig)
	assert.Equal(t, "openai", openaiClientConfig.Type)

	configMap := openaiClientConfig.Config
	assert.Equal(t, "key", configMap["apikey"])
	assert.Equal(t, "model", configMap["model"])
}

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
