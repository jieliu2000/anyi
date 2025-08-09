package anyi

import (
	"errors"
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/registry"
	"github.com/stretchr/testify/assert"
)

func TestSimpleChat(t *testing.T) {
	// Setup - Create test environment
	// Save original registry to restore after tests
	origRegistry := GlobalRegistry
	defer func() { GlobalRegistry = origRegistry }()

	t.Run("Success case", func(t *testing.T) {
		// Setup - Register a mock client that returns a preset response
		GlobalRegistry = &registry.AnyiRegistry{
			Clients:           make(map[string]llm.Client),
			Flows:             make(map[string]*flow.Flow),
			Validators:        make(map[string]flow.StepValidator),
			Executors:         make(map[string]flow.StepExecutor),
			Formatters:        make(map[string]chat.PromptFormatter),
			DefaultClientName: "default",
		}
		mockClient := &test.MockClient{
			ChatOutput: "This is a test response",
		}
		RegisterNewDefaultClient("default", mockClient)

		// Execute
		response, err := SimpleChat("Hello")

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, "This is a test response", response)
	})

	t.Run("Empty input", func(t *testing.T) {
		// Setup
		GlobalRegistry = &registry.AnyiRegistry{
			Clients:           make(map[string]llm.Client),
			Flows:             make(map[string]*flow.Flow),
			Validators:        make(map[string]flow.StepValidator),
			Executors:         make(map[string]flow.StepExecutor),
			Formatters:        make(map[string]chat.PromptFormatter),
			DefaultClientName: "default",
		}
		mockClient := &test.MockClient{}
		RegisterNewDefaultClient("default", mockClient)

		// Execute
		response, err := SimpleChat("")

		// Verify
		assert.Error(t, err)
		assert.Equal(t, "", response)
		assert.Equal(t, "empty input", err.Error())
	})

	t.Run("No default client", func(t *testing.T) {
		// Setup - Create a registry with no default client
		GlobalRegistry = &registry.AnyiRegistry{
			Clients:    make(map[string]llm.Client),
			Flows:      make(map[string]*flow.Flow),
			Validators: make(map[string]flow.StepValidator),
			Executors:  make(map[string]flow.StepExecutor),
			Formatters: make(map[string]chat.PromptFormatter),
		}

		// Execute
		response, err := SimpleChat("Hello")

		// Verify
		assert.Error(t, err)
		assert.Equal(t, "", response)
		assert.Contains(t, err.Error(), "no default client found")
	})

	t.Run("Client error", func(t *testing.T) {
		// Setup - Register a mock client that returns an error
		GlobalRegistry = &registry.AnyiRegistry{
			Clients:           make(map[string]llm.Client),
			Flows:             make(map[string]*flow.Flow),
			Validators:        make(map[string]flow.StepValidator),
			Executors:         make(map[string]flow.StepExecutor),
			Formatters:        make(map[string]chat.PromptFormatter),
			DefaultClientName: "default",
		}
		mockClient := &test.MockClient{
			Err: errors.New("client error"),
		}
		RegisterNewDefaultClient("default", mockClient)

		// Execute
		response, err := SimpleChat("Hello")

		// Verify
		assert.Error(t, err)
		assert.Equal(t, "", response)
		assert.Equal(t, "client error", err.Error())
	})
}
