package anyi

import (
	"testing"
	"time"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestMCPExecutor_SSETransport tests SSE transport functionality
func TestMCPExecutor_SSETransport(t *testing.T) {
	// Create mock SSE server
	mockSSEServer := test.NewMockSSEServer()
	defer mockSSEServer.Close()

	t.Run("SSE tool call", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerEndpoint: mockSSEServer.URL() + "/events",
			Transport:      TransportSSE,
			Operation:      OperationToolCall,
			ToolName:       "sse_test_tool",
			ToolArgs:       map[string]interface{}{"param1": "sse_value1"},
			Timeout:        5 * time.Second,
			RetryAttempts:  1,
		}

		// Initialize
		err := executor.Init()
		require.NoError(t, err)
		assert.True(t, executor.initialized)
		assert.NotNil(t, executor.client)

		// Test client type
		_, ok := executor.client.(*SSEMCPClient)
		assert.True(t, ok, "Expected SSEMCPClient")

		// Clean up
		executor.Close()
	})

	t.Run("SSE resource read", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerEndpoint:  mockSSEServer.URL() + "/events",
			Transport:       TransportSSE,
			Operation:       OperationResourceRead,
			ResourceURI:     "/sse-test-resource",
			OutputToContext: true,
			ResultVarName:   "sseResourceResult",
			Timeout:         5 * time.Second,
			RetryAttempts:   1,
		}

		// Initialize
		err := executor.Init()
		require.NoError(t, err)

		// Clean up
		executor.Close()
	})

	t.Run("SSE validation errors", func(t *testing.T) {
		// Missing endpoint
		executor := &MCPExecutor{
			Transport: TransportSSE,
			Operation: OperationToolCall,
			ToolName:  "test_tool",
		}

		err := executor.Init()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "serverEndpoint is required")
	})
}

// TestMCPExecutor_STDIOTransport tests STDIO transport functionality
func TestMCPExecutor_STDIOTransport(t *testing.T) {
	t.Run("STDIO client creation", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerCommand: "echo",
			ServerArgs:    []string{"test"},
			Transport:     TransportSTDIO,
			Operation:     OperationListTools,
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		}

		// Initialize
		err := executor.Init()
		require.NoError(t, err)
		assert.True(t, executor.initialized)
		assert.NotNil(t, executor.client)

		// Test client type
		_, ok := executor.client.(*STDIOMCPClient)
		assert.True(t, ok, "Expected STDIOMCPClient")

		// Clean up
		executor.Close()
	})

	t.Run("STDIO with arguments", func(t *testing.T) {
		executor := &MCPExecutor{
			ServerCommand: "python",
			ServerArgs:    []string{"-c", "print('test')"},
			Transport:     TransportSTDIO,
			Operation:     OperationToolCall,
			ToolName:      "test_tool",
			ToolArgs:      map[string]interface{}{"input": "test_data"},
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		}

		// Initialize
		err := executor.Init()
		require.NoError(t, err)

		// Clean up
		executor.Close()
	})

	t.Run("STDIO validation errors", func(t *testing.T) {
		// Missing command
		executor := &MCPExecutor{
			Transport: TransportSTDIO,
			Operation: OperationToolCall,
			ToolName:  "test_tool",
		}

		err := executor.Init()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "serverCommand is required")
	})
}

// TestMCPExecutor_TransportSpecificValidation tests transport-specific validation
func TestMCPExecutor_TransportSpecificValidation(t *testing.T) {
	tests := []struct {
		name           string
		executor       MCPExecutor
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid SSE configuration",
			executor: MCPExecutor{
				ServerEndpoint: "http://localhost:8080/events",
				Transport:      TransportSSE,
				Operation:      OperationListTools,
			},
			expectError: false,
		},
		{
			name: "invalid transport type",
			executor: MCPExecutor{
				Transport: "websocket",
				Operation: OperationToolCall,
			},
			expectError:    true,
			errorSubstring: "transport must be one of",
		},
		{
			name: "SSE without endpoint",
			executor: MCPExecutor{
				Transport: TransportSSE,
				Operation: OperationToolCall,
				ToolName:  "test_tool",
			},
			expectError:    true,
			errorSubstring: "serverEndpoint is required",
		},
		{
			name: "STDIO with empty command",
			executor: MCPExecutor{
				ServerCommand: "",
				Transport:     TransportSTDIO,
				Operation:     OperationToolCall,
				ToolName:      "test_tool",
			},
			expectError:    true,
			errorSubstring: "serverCommand is required",
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
				executor.Close()
			}
		})
	}
}

// TestMCPExecutor_DefaultSettings tests default value setting
func TestMCPExecutor_DefaultSettings(t *testing.T) {
	executor := &MCPExecutor{
		ServerEndpoint: "http://localhost:8080",
		Transport:      TransportHTTP,
		Operation:      OperationToolCall,
		ToolName:       "test_tool",
	}

	err := executor.Init()
	require.NoError(t, err)

	// Test default values
	assert.Equal(t, "mcpResult", executor.ResultVarName)
	assert.Equal(t, 30*time.Second, executor.Timeout)
	assert.Equal(t, 3, executor.RetryAttempts)
	assert.NotNil(t, executor.ToolArgs)
	assert.NotNil(t, executor.PromptArgs)

	executor.Close()
}

// TestMCPExecutor_ArgumentBuilding tests tool argument building
func TestMCPExecutor_ArgumentBuilding(t *testing.T) {
	executor := &MCPExecutor{
		ToolArgs: map[string]interface{}{
			"static_param": "static_value",
			"number_param": 42,
		},
		ToolArgVars: []string{"dynamic_param", "missing_param"},
	}

	flowContext := flow.FlowContext{
		Variables: map[string]interface{}{
			"dynamic_param": "dynamic_value",
			"other_var":     "other_value",
		},
	}

	args := executor.buildToolArguments(flowContext)

	// Check static arguments
	assert.Equal(t, "static_value", args["static_param"])
	assert.Equal(t, 42, args["number_param"])

	// Check dynamic arguments
	assert.Equal(t, "dynamic_value", args["dynamic_param"])

	// Check missing arguments are not included
	_, exists := args["missing_param"]
	assert.False(t, exists)

	// Check other variables are not included
	_, exists = args["other_var"]
	assert.False(t, exists)
}

// TestMCPExecutor_PromptArgumentBuilding tests prompt argument building
func TestMCPExecutor_PromptArgumentBuilding(t *testing.T) {
	executor := &MCPExecutor{
		PromptArgs: map[string]interface{}{
			"language":   "python",
			"complexity": "high",
			"lines":      100,
		},
	}

	flowContext := flow.FlowContext{
		Variables: map[string]interface{}{
			"unused_var": "unused_value",
		},
	}

	args := executor.buildPromptArguments(flowContext)

	// Check all prompt arguments are included
	assert.Equal(t, "python", args["language"])
	assert.Equal(t, "high", args["complexity"])
	assert.Equal(t, 100, args["lines"])

	// Check flow variables are not included in prompt args
	_, exists := args["unused_var"]
	assert.False(t, exists)
}

// TestMCPExecutor_ResponseProcessing tests response processing
func TestMCPExecutor_ResponseProcessing(t *testing.T) {
	t.Run("string result", func(t *testing.T) {
		executor := &MCPExecutor{
			ResultVarName:   "testResult",
			OutputToContext: true,
		}

		flowContext := flow.FlowContext{
			Variables: make(map[string]interface{}),
		}

		response := &MCPResponse{
			JSONRPC: "2.0",
			ID:      "test",
			Result:  "test result string",
		}

		newContext, err := executor.processResponse(response, flowContext)
		require.NoError(t, err)

		// Check variable was set
		assert.Equal(t, "test result string", newContext.Variables["testResult"])

		// Check context text was set
		assert.Equal(t, "test result string", newContext.Text)
	})

	t.Run("map result with text field", func(t *testing.T) {
		executor := &MCPExecutor{
			ResultVarName:   "mapResult",
			OutputToContext: true,
		}

		flowContext := flow.FlowContext{
			Variables: make(map[string]interface{}),
		}

		response := &MCPResponse{
			JSONRPC: "2.0",
			ID:      "test",
			Result: map[string]interface{}{
				"text":   "extracted text",
				"status": "success",
			},
		}

		newContext, err := executor.processResponse(response, flowContext)
		require.NoError(t, err)

		// Check variable was set to full map
		resultMap, ok := newContext.Variables["mapResult"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "extracted text", resultMap["text"])
		assert.Equal(t, "success", resultMap["status"])

		// Check context text was set to text field
		assert.Equal(t, "extracted text", newContext.Text)
	})

	t.Run("map result with content field", func(t *testing.T) {
		executor := &MCPExecutor{
			ResultVarName:   "contentResult",
			OutputToContext: true,
		}

		flowContext := flow.FlowContext{
			Variables: make(map[string]interface{}),
		}

		response := &MCPResponse{
			JSONRPC: "2.0",
			ID:      "test",
			Result: map[string]interface{}{
				"content": "content text",
				"type":    "response",
			},
		}

		newContext, err := executor.processResponse(response, flowContext)
		require.NoError(t, err)

		// Check context text was set to content field
		assert.Equal(t, "content text", newContext.Text)
	})

	t.Run("error response", func(t *testing.T) {
		executor := &MCPExecutor{
			ResultVarName: "errorResult",
		}

		flowContext := flow.FlowContext{
			Variables: make(map[string]interface{}),
		}

		response := &MCPResponse{
			JSONRPC: "2.0",
			ID:      "test",
			Error: &MCPError{
				Code:    -1,
				Message: "test error message",
			},
		}

		_, err := executor.processResponse(response, flowContext)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error message")
		assert.Contains(t, err.Error(), "-1")
	})

	t.Run("output to context disabled", func(t *testing.T) {
		executor := &MCPExecutor{
			ResultVarName:   "noContextResult",
			OutputToContext: false,
		}

		flowContext := flow.FlowContext{
			Variables: make(map[string]interface{}),
		}

		response := &MCPResponse{
			JSONRPC: "2.0",
			ID:      "test",
			Result:  "should not set context",
		}

		newContext, err := executor.processResponse(response, flowContext)
		require.NoError(t, err)

		// Check variable was set
		assert.Equal(t, "should not set context", newContext.Variables["noContextResult"])

		// Check context text was NOT set
		assert.Empty(t, newContext.Text)
	})
}

// TestMCPExecutor_StringFormatting tests variable substitution in strings
func TestMCPExecutor_StringFormatting(t *testing.T) {
	executor := &MCPExecutor{}

	tests := []struct {
		name      string
		format    string
		variables map[string]interface{}
		expected  string
	}{
		{
			name:      "simple string substitution",
			format:    "Hello ${name}!",
			variables: map[string]interface{}{"name": "World"},
			expected:  "Hello World!",
		},
		{
			name:      "multiple substitutions",
			format:    "${greeting} ${name}, today is ${day}",
			variables: map[string]interface{}{"greeting": "Hello", "name": "Alice", "day": "Monday"},
			expected:  "Hello Alice, today is Monday",
		},
		{
			name:      "number substitution",
			format:    "Count: ${count}",
			variables: map[string]interface{}{"count": 42},
			expected:  "Count: 42",
		},
		{
			name:      "boolean substitution",
			format:    "Enabled: ${enabled}",
			variables: map[string]interface{}{"enabled": true},
			expected:  "Enabled: true",
		},
		{
			name:      "object substitution",
			format:    "Config: ${config}",
			variables: map[string]interface{}{"config": map[string]string{"key": "value"}},
			expected:  `Config: {"key":"value"}`,
		},
		{
			name:      "missing variable",
			format:    "Hello ${name}, age ${age}",
			variables: map[string]interface{}{"name": "John"},
			expected:  "Hello John, age ${age}",
		},
		{
			name:      "no variables",
			format:    "Static text with no variables",
			variables: map[string]interface{}{"unused": "value"},
			expected:  "Static text with no variables",
		},
		{
			name:      "empty variables",
			format:    "Text with ${var}",
			variables: map[string]interface{}{},
			expected:  "Text with ${var}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := executor.formatStringWithVariables(tc.format, tc.variables)
			assert.Equal(t, tc.expected, result)
		})
	}
}
