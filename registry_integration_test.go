package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

// TestRegistryIntegration verifies that the anyi package correctly integrates with the registry package
func TestRegistryIntegration(t *testing.T) {
	// Clear registry for clean test
	registry.Clear()
	defer registry.Clear()

	// Test client registration and retrieval
	client := &test.MockClient{}
	err := RegisterClient("test-client", client)
	assert.NoError(t, err)

	retrievedClient, err := GetClient("test-client")
	assert.NoError(t, err)
	assert.Equal(t, client, retrievedClient)

	// Test default client functionality
	err = RegisterNewDefaultClient("default", client)
	assert.NoError(t, err)

	defaultClient, err := GetDefaultClient()
	assert.NoError(t, err)
	assert.Equal(t, client, defaultClient)

	// Test that Init function registers built-in components
	Init()

	// Verify that built-in executors are registered
	_, err = registry.GetExecutor("llm")
	assert.NoError(t, err)

	_, err = registry.GetExecutor("condition")
	assert.NoError(t, err)

	_, err = registry.GetExecutor("exec")
	assert.NoError(t, err)

	// Verify that built-in validators are registered
	_, err = registry.GetValidator("json")
	assert.NoError(t, err)

	_, err = registry.GetValidator("string")
	assert.NoError(t, err)
}
