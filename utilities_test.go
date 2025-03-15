package anyi

import (
	"errors"
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/stretchr/testify/assert"
)

func TestSimpleChat(t *testing.T) {
	// Setup - Create test environment
	// Save original registry to restore after tests
	origRegistry := GlobalRegistry
	defer func() { GlobalRegistry = origRegistry }()

	t.Run("Success case", func(t *testing.T) {
		// Setup - Register a mock client that returns a preset response
		GlobalRegistry = &anyiRegistry{
			Clients:           make(map[string]llm.Client),
			Flows:             make(map[string]*flow.Flow),
			Validators:        make(map[string]flow.StepValidator),
			Executors:         make(map[string]flow.StepExecutor),
			Formatters:        make(map[string]chat.PromptFormatter),
			defaultClientName: "default",
		}
		mockClient := &test.MockClient{
			ChatOutput: "This is a test response",
		}
		RegisterDefaultClient("default", mockClient)

		// Execute
		response, err := SimpleChat("Hello")

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, "This is a test response", response)
	})

	t.Run("Empty input", func(t *testing.T) {
		// Setup
		GlobalRegistry = &anyiRegistry{
			Clients:           make(map[string]llm.Client),
			Flows:             make(map[string]*flow.Flow),
			Validators:        make(map[string]flow.StepValidator),
			Executors:         make(map[string]flow.StepExecutor),
			Formatters:        make(map[string]chat.PromptFormatter),
			defaultClientName: "default",
		}
		mockClient := &test.MockClient{}
		RegisterDefaultClient("default", mockClient)

		// Execute
		response, err := SimpleChat("")

		// Verify
		assert.Error(t, err)
		assert.Equal(t, "", response)
		assert.Equal(t, "empty input", err.Error())
	})

	t.Run("No default client", func(t *testing.T) {
		// Setup - Create a registry with no default client
		GlobalRegistry = &anyiRegistry{
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
		assert.Equal(t, "no default client found", err.Error())
	})

	t.Run("Client error", func(t *testing.T) {
		// Setup - Register a mock client that returns an error
		GlobalRegistry = &anyiRegistry{
			Clients:           make(map[string]llm.Client),
			Flows:             make(map[string]*flow.Flow),
			Validators:        make(map[string]flow.StepValidator),
			Executors:         make(map[string]flow.StepExecutor),
			Formatters:        make(map[string]chat.PromptFormatter),
			defaultClientName: "default",
		}
		mockClient := &test.MockClient{
			Err: errors.New("client error"),
		}
		RegisterDefaultClient("default", mockClient)

		// Execute
		response, err := SimpleChat("Hello")

		// Verify
		assert.Error(t, err)
		assert.Equal(t, "", response)
		assert.Equal(t, "client error", err.Error())
	})
}
