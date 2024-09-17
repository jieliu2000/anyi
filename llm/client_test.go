package llm

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/siliconcloud"
	"github.com/jieliu2000/anyi/llm/zhipu"
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

func TestNewModelConfigFromClientConfig_WithOpenAI(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "openai",
		Config: map[string]interface{}{
			"apiKey": "test_api_key",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*openai.OpenAIModelConfig)(nil), modelConfig)
	assert.Equal(t, "test_api_key", modelConfig.(*openai.OpenAIModelConfig).APIKey)
}
func TestNewModelConfigFromClientConfig_WithAzureOpenAI(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "azureopenai",
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*azureopenai.AzureOpenAIModelConfig)(nil), modelConfig)
}
func TestNewModelConfigFromClientConfig_WithDashScope(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "dashscope",
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*dashscope.DashScopeModelConfig)(nil), modelConfig)
}
func TestNewModelConfigFromClientConfig_WithZhipu(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "zhipu",
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*zhipu.ZhiPuModelConfig)(nil), modelConfig)
}
func TestNewModelConfigFromClientConfig_WithSiliconCloud(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "siliconcloud",
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*siliconcloud.SiliconCloudConfig)(nil), modelConfig)
}
func TestNewModelConfigFromClientConfig_WithOllama(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "ollama",
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*ollama.OllamaModelConfig)(nil), modelConfig)
}
func TestNewModelConfigFromClientConfig_WithUnknownType(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "unknown",
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	_, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.Error(t, err)
}
func TestNewModelConfigFromClientConfig_WithMissingType(t *testing.T) {
	clientConfig := ClientConfig{
		Config: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}
	_, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.Error(t, err)
}
func TestNewModelConfigFromClientConfig_WithNilConfig(t *testing.T) {
	_, err := NewModelConfigFromClientConfig(nil)
	assert.Error(t, err)
}
func TestNewModelConfigFromClientConfig_WithEnvironmentVariables(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test_api_key")
	clientConfig := ClientConfig{
		Type: "openai",
		Config: map[string]interface{}{
			"apiKey": "$OPENAI_API_KEY",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*openai.OpenAIModelConfig)(nil), modelConfig)
	assert.Equal(t, "test_api_key", modelConfig.(*openai.OpenAIModelConfig).APIKey)
}

// This test case will fail if the environment variable OPENAI_API_KEY is not set.
// To run this test case, set the OPENAI_API_KEY environment variable to "test_api_key".
// You can do this via the command line with: export OPENAI_API_KEY="test_api_key"
func TestNewModelConfigFromClientConfig_WithEnvironmentVariables_CrossPlatform(t *testing.T) {
	clientConfig := ClientConfig{
		Type: "openai",
		Config: map[string]interface{}{
			"api_key": "$OPENAI_API_KEY",
		},
	}
	modelConfig, err := NewModelConfigFromClientConfig(&clientConfig)
	assert.NoError(t, err)
	assert.IsType(t, (*openai.OpenAIModelConfig)(nil), modelConfig)
	assert.Equal(t, "test_api_key", modelConfig.(*openai.OpenAIModelConfig).APIKey)
}
