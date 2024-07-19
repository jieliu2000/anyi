package azureopenai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("key", "deploy", "endpoint")

	assert.Equal(t, "key", config.APIKey, "Expected token %s, got %s", "key", config.APIKey)
	assert.Equal(t, "deploy", config.ModelDeploymentId, "Expected model %s, got %s", "deploy", config.ModelDeploymentId)
	assert.Equal(t, "endpoint", config.Endpoint, "Expected URL %s, got %s", "endpoint", config.Endpoint)
}

func TestNewClient(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		APIKey:            "test_api_key",
		ModelDeploymentId: "test_model_deployment_id",
		Endpoint:          "test_endpoint",
	}
	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config)
	assert.NotNil(t, client.clientImpl)
}

func TestNewClient_NilConfig(t *testing.T) {
	client, err := NewClient(nil)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyConfig(t *testing.T) {
	config := &AzureOpenAIModelConfig{}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyAPIKey(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		ModelDeploymentId: "test_model_deployment_id",
		Endpoint:          "test_endpoint",
	}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyModelDeploymentId(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		APIKey:   "test_api_key",
		Endpoint: "test_endpoint",
	}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
func TestNewClient_EmptyEndpoint(t *testing.T) {
	config := &AzureOpenAIModelConfig{
		APIKey:            "test_api_key",
		ModelDeploymentId: "test_model_deployment_id",
	}
	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}
