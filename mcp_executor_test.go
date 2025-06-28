package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMCPExecutor_Init(t *testing.T) {
	tests := []struct {
		name           string
		executor       MCPExecutor
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid http configuration",
			executor: MCPExecutor{
				ServerEndpoint: "http://localhost:8080",
				Transport:      TransportHTTP,
				Operation:      OperationToolCall,
				ToolName:       "test_tool",
			},
			expectError: false,
		},
		{
			name: "valid stdio configuration",
			executor: MCPExecutor{
				ServerCommand: "node",
				ServerArgs:    []string{"server.js"},
				Transport:     TransportSTDIO,
				Operation:     OperationToolCall,
				ToolName:      "test_tool",
			},
			expectError: false,
		},
		{
			name: "missing server endpoint for http",
			executor: MCPExecutor{
				Transport: TransportHTTP,
				Operation: OperationToolCall,
				ToolName:  "test_tool",
			},
			expectError:    true,
			errorSubstring: "serverEndpoint is required",
		},
		{
			name: "missing server command for stdio",
			executor: MCPExecutor{
				Transport: TransportSTDIO,
				Operation: OperationToolCall,
				ToolName:  "test_tool",
			},
			expectError:    true,
			errorSubstring: "serverCommand is required",
		},
		{
			name: "missing operation",
			executor: MCPExecutor{
				ServerEndpoint: "http://localhost:8080",
				Transport:      TransportHTTP,
			},
			expectError:    true,
			errorSubstring: "operation must be specified",
		},
		{
			name: "missing tool name for tool call",
			executor: MCPExecutor{
				ServerEndpoint: "http://localhost:8080",
				Transport:      TransportHTTP,
				Operation:      OperationToolCall,
			},
			expectError:    true,
			errorSubstring: "toolName is required",
		},
		{
			name: "missing resource URI for resource read",
			executor: MCPExecutor{
				ServerEndpoint: "http://localhost:8080",
				Transport:      TransportHTTP,
				Operation:      OperationResourceRead,
			},
			expectError:    true,
			errorSubstring: "resourceUri is required",
		},
		{
			name: "missing prompt name for prompt get",
			executor: MCPExecutor{
				ServerEndpoint: "http://localhost:8080",
				Transport:      TransportHTTP,
				Operation:      OperationPromptGet,
			},
			expectError:    true,
			errorSubstring: "promptName is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := tc.executor
			err := executor.Init()

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorSubstring != "" {
					assert.Contains(t, err.Error(), tc.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, executor.initialized)
				assert.NotNil(t, executor.client)
			}
		})
	}
}

func TestMCPExecutor_FormatStringWithVariables(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		variables map[string]interface{}
		expected  string
	}{
		{
			name:      "string variables",
			format:    "/users/${name}/documents/${id}",
			variables: map[string]interface{}{"name": "John", "id": "12345"},
			expected:  "/users/John/documents/12345",
		},
		{
			name:      "non-string variables",
			format:    "/items/${id}/status/${flag}",
			variables: map[string]interface{}{"id": 42, "flag": true},
			expected:  "/items/42/status/true",
		},
		{
			name:      "missing variables",
			format:    "/users/${name}/items/${itemId}",
			variables: map[string]interface{}{"name": "John"},
			expected:  "/users/John/items/${itemId}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MCPExecutor{
				ServerEndpoint: "http://localhost:8080",
				Transport:      TransportHTTP,
				Operation:      OperationToolCall,
				ToolName:       "test_tool",
			}
			result := executor.formatStringWithVariables(tc.format, tc.variables)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMCPExecutor_WithMockServer(t *testing.T) {
	// Create mock server
	mockServer := test.NewMockMCPServer()
	defer mockServer.Close()

	t.Run("tool call with mock server", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerEndpoint: mockServer.URL(),
			Transport:      TransportHTTP,
			Operation:      OperationToolCall,
			ToolName:       "test_tool",
			ToolArgs:       map[string]interface{}{"param1": "value1"},
		}

		// Initialize
		err := executor.Init()
		assert.NoError(t, err)

		// Run
		flowContext := flow.FlowContext{}
		result, err := executor.Run(flowContext, nil)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "tools/call", requests[0].Method)
	})

	t.Run("resource read with mock server", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerEndpoint:  mockServer.URL(),
			Transport:       TransportHTTP,
			Operation:       OperationResourceRead,
			ResourceURI:     "/test-resource",
			OutputToContext: true,
			ResultVarName:   "resourceResult",
		}

		// Clear previous requests
		mockServer.ClearRequests()

		// Initialize
		err := executor.Init()
		assert.NoError(t, err)

		// Run
		flowContext := flow.FlowContext{
			Variables: make(map[string]any),
		}
		result, err := executor.Run(flowContext, nil)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "resources/read", requests[0].Method)
	})

	t.Run("prompt get with mock server", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerEndpoint: mockServer.URL(),
			Transport:      TransportHTTP,
			Operation:      OperationPromptGet,
			PromptName:     "test_prompt",
			PromptArgs:     map[string]interface{}{"arg1": "value1"},
		}

		// Clear previous requests
		mockServer.ClearRequests()

		// Initialize
		err := executor.Init()
		assert.NoError(t, err)

		// Run
		flowContext := flow.FlowContext{}
		result, err := executor.Run(flowContext, nil)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "prompts/get", requests[0].Method)
	})

	t.Run("variables substitution with mock server", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerEndpoint:  mockServer.URL(),
			Transport:       TransportHTTP,
			Operation:       OperationResourceRead,
			ResourceURI:     "/resources/${docId}",
			OutputToContext: true,
			ResultVarName:   "testVar",
		}

		// Clear previous requests
		mockServer.ClearRequests()

		// Initialize
		err := executor.Init()
		assert.NoError(t, err)

		// Set context variables
		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"docId": "test-123",
			},
		}

		// Run
		result, err := executor.Run(flowContext, nil)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify variable substitution worked
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		// Could further verify that the URI in request parameters had variables correctly substituted
	})

	t.Run("error handling with mock server", func(t *testing.T) {
		// Set error response
		mockServer.SetErrorResponse("tools/call", -1, "Mock error")

		executor := &MCPExecutor{
			ServerEndpoint: mockServer.URL(),
			Transport:      TransportHTTP,
			Operation:      OperationToolCall,
			ToolName:       "error_tool",
			RetryAttempts:  1, // Reduce retry attempts to speed up test
		}

		// Clear previous requests
		mockServer.ClearRequests()

		// Initialize
		err := executor.Init()
		assert.NoError(t, err)

		// Run (should fail)
		flowContext := flow.FlowContext{}
		_, err = executor.Run(flowContext, nil)

		// Verify error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mock error")

		// Restore normal response
		mockServer.SetResponse("tools/call", map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Tool executed successfully",
				},
			},
		})
	})
}

func TestMCPExecutor_BasicFunctionality(t *testing.T) {
	t.Run("initialized when running", func(t *testing.T) {
		// Create mock server
		mockServer := test.NewMockMCPServer()
		defer mockServer.Close()

		executor := &MCPExecutor{
			ServerEndpoint: mockServer.URL(),
			Transport:      TransportHTTP,
			Operation:      OperationToolCall,
			ToolName:       "test_tool",
		}

		// Don't initialize, let Run do it
		flowContext := flow.FlowContext{}
		executor.Run(flowContext, nil)

		// Should be initialized after Run
		assert.True(t, executor.initialized)
	})
}
