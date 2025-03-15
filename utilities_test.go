package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/stretchr/testify/assert"
)

func TestSimpleChat(t *testing.T) {
	// Test with empty input
	t.Run("Empty input", func(t *testing.T) {
		response, err := SimpleChat("")
		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "empty input")
	})

	// Test with no default client
	t.Run("No default client", func(t *testing.T) {
		// Save the current clients
		oldClients := GlobalRegistry.Clients
		oldDefaultClient := GlobalRegistry.defaultClientName

		// Reset the registry
		GlobalRegistry.Clients = make(map[string]llm.Client)
		GlobalRegistry.defaultClientName = ""

		response, err := SimpleChat("Hello")
		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "no default client found")

		// Restore the registry
		GlobalRegistry.Clients = oldClients
		GlobalRegistry.defaultClientName = oldDefaultClient
	})

	// Test with successful chat
	t.Run("Successful chat", func(t *testing.T) {
		// Create and register a mock client that returns a successful response
		mockClient := &test.MockClient{
			ChatOutput: "This is a mock response",
		}
		RegisterDefaultClient("test-default", mockClient)

		response, err := SimpleChat("Hello, AI!")
		assert.NoError(t, err)
		assert.Equal(t, "This is a mock response", response)
	})

	// Test with client error
	t.Run("Client error", func(t *testing.T) {
		// Create and register a mock client that returns an error
		mockClient := &test.MockClient{
			Err: assert.AnError,
		}
		RegisterDefaultClient("test-error", mockClient)

		response, err := SimpleChat("Hello, AI!")
		assert.Error(t, err)
		assert.Empty(t, response)
	})
}
