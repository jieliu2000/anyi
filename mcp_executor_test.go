package anyi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/stretchr/testify/assert"
)

// TestMCPExecutor_Init tests the initialization of MCPExecutor
func TestMCPExecutor_Init(t *testing.T) {
	tests := []struct {
		name           string
		executor       MCPExecutor
		expectError    bool
		errorSubstring string
		checkFunc      func(t *testing.T, executor *MCPExecutor)
	}{
		{
			name:           "empty endpoint",
			executor:       MCPExecutor{},
			expectError:    true,
			errorSubstring: "MCP endpoint cannot be empty",
		},
		{
			name: "invalid transport",
			executor: MCPExecutor{
				Endpoint:  "http://localhost:8080",
				Transport: "invalid",
			},
			expectError:    true,
			errorSubstring: "invalid transport type",
		},
		{
			name: "default transport",
			executor: MCPExecutor{
				Endpoint: "http://localhost:8080",
			},
			expectError: false,
			checkFunc: func(t *testing.T, executor *MCPExecutor) {
				assert.Equal(t, "http", executor.Transport)
			},
		},
		{
			name: "default resultVarName",
			executor: MCPExecutor{
				Endpoint: "http://localhost:8080",
			},
			expectError: false,
			checkFunc: func(t *testing.T, executor *MCPExecutor) {
				assert.Equal(t, "mcpResult", executor.ResultVarName)
			},
		},
		{
			name: "custom values",
			executor: MCPExecutor{
				Endpoint:      "http://localhost:8080",
				Transport:     "sse",
				ResultVarName: "testResult",
			},
			expectError: false,
			checkFunc: func(t *testing.T, executor *MCPExecutor) {
				assert.Equal(t, "sse", executor.Transport)
				assert.Equal(t, "testResult", executor.ResultVarName)
				assert.NotNil(t, executor.httpClient)
				assert.True(t, executor.initialized)
			},
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
				if tc.checkFunc != nil {
					tc.checkFunc(t, &executor)
				}
			}
		})
	}
}

// TestMCPExecutor_FormatStringWithVariables tests the variable substitution functionality
func TestMCPExecutor_FormatStringWithVariables(t *testing.T) {
	// Table-based tests
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
			name:      "complex variables",
			format:    "/api/${obj}",
			variables: map[string]interface{}{"obj": map[string]string{"key": "value"}},
			expected:  `/api/{"key":"value"}`,
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
			executor := &MCPExecutor{Endpoint: "http://localhost:8080"}
			result := executor.formatStringWithVariables(tc.format, tc.variables)
			assert.Equal(t, tc.expected, result)
		})
	}

	// Additional test cases from builtin_executors_test.go
	executor := &MCPExecutor{
		Endpoint: "http://localhost:8080",
	}

	t.Run("complex object variables", func(t *testing.T) {
		variables := map[string]interface{}{
			"obj": map[string]interface{}{
				"key": "value",
			},
		}
		formatted := executor.formatStringWithVariables("/api/${obj}", variables)
		assert.Equal(t, `/api/{"key":"value"}`, formatted)
	})
}

// TestMCPExecutor_BasicFunctionality tests the basic functionality of MCPExecutor without mocking
func TestMCPExecutor_BasicFunctionality(t *testing.T) {
	t.Run("initialized when running", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint: "http://localhost:8080",
		}

		// Don't initialize, let Run do it
		flowContext := flow.FlowContext{}
		_, err := executor.Run(flowContext, nil)

		// Should not return init error
		assert.NoError(t, err)
		assert.True(t, executor.initialized)
	})

	t.Run("variables passed to result", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			OutputToContext: true,
			ResourceURI:     "/resources/${docId}",
			ResultVarName:   "testVar",
		}

		// Initialize
		err := executor.Init()
		assert.NoError(t, err)

		// Set up context with variables
		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"docId": "test-123",
			},
		}

		// Run - this will call the placeholder readResource method
		result, err := executor.Run(flowContext, nil)

		// Placeholder implementation should succeed
		assert.NoError(t, err)

		// Should have a variable storing the result
		resultVar, exists := result.Variables["testVar"]
		assert.True(t, exists)

		// The variable should be a map with uri and contents
		resultMap, ok := resultVar.(map[string]interface{})
		assert.True(t, ok)

		// The uri should have the variable substituted
		assert.Equal(t, "/resources/test-123", resultMap["uri"])
	})

	// Additional test case from builtin_executors_test.go
	t.Run("not initialized with step", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint: "http://localhost:8080",
		}
		flowContext := flow.FlowContext{}
		step := &flow.Step{}

		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)
		assert.True(t, executor.initialized)
		assert.Equal(t, "placeholder-session-id", result.GetVariableString("mcpSessionId", ""))
	})
}

// TestMCPExecutor_ResourceRead tests the resource reading functionality
func TestMCPExecutor_ResourceRead(t *testing.T) {
	t.Run("basic resource read", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			Transport:       "http",
			ResourceURI:     "/documents/test-doc",
			OutputToContext: true,
			ResultVarName:   "testDoc",
		}

		err := executor.Init()
		assert.NoError(t, err)

		flowContext := flow.FlowContext{}
		result, err := executor.Run(flowContext, nil)

		assert.NoError(t, err)
		assert.Contains(t, result.Text, "Placeholder content for resource: /documents/test-doc")

		// Check that variable was set correctly
		docVar, exists := result.Variables["testDoc"]
		assert.True(t, exists)

		docMap, ok := docVar.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "/documents/test-doc", docMap["uri"])
		assert.Contains(t, docMap["contents"].(string), "Placeholder content for resource")
	})

	// Additional test case from builtin_executors_test.go
	t.Run("read resource with variables", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			Transport:       "http",
			ResourceURI:     "/documents/${docId}",
			OutputToContext: true,
			ResultVarName:   "testResource",
		}
		err := executor.Init()
		assert.NoError(t, err)

		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"docId": "test-123",
			},
		}
		step := &flow.Step{}

		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)

		assert.Contains(t, result.Text, "Placeholder content for resource: /documents/test-123")
		resource, ok := result.GetVariable("testResource").(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "/documents/test-123", resource["uri"])
	})
}

// TestMCPExecutor_ToolCall tests the tool calling functionality
func TestMCPExecutor_ToolCall(t *testing.T) {
	t.Run("basic tool call", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			Transport:       "http",
			ToolName:        "summarize",
			ToolArgVars:     []string{"content"},
			OutputToContext: true,
			ResultVarName:   "summary",
		}

		err := executor.Init()
		assert.NoError(t, err)

		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"content": "This is test content to summarize",
			},
		}

		result, err := executor.Run(flowContext, nil)

		assert.NoError(t, err)
		assert.Contains(t, result.Text, "Placeholder result for tool: summarize")

		// Check that variable was set correctly
		summaryVar, exists := result.Variables["summary"]
		assert.True(t, exists)

		summaryMap, ok := summaryVar.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "success", summaryMap["status"])
	})

	// Additional test case from builtin_executors_test.go
	t.Run("call tool with step", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			Transport:       "http",
			ToolName:        "summarize",
			ToolArgVars:     []string{"content"},
			OutputToContext: true,
			ResultVarName:   "toolResult",
		}
		err := executor.Init()
		assert.NoError(t, err)

		flowContext := flow.FlowContext{
			Variables: map[string]any{
				"content": "This is a test document",
			},
		}
		step := &flow.Step{}

		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)

		assert.Contains(t, result.Text, "Placeholder result for tool: summarize")
		toolResult, ok := result.GetVariable("toolResult").(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "success", toolResult["status"])
	})
}

// TestMCPExecutor_SessionInit tests the session initialization functionality
func TestMCPExecutor_SessionInit(t *testing.T) {
	t.Run("basic session initialization", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			Transport:       "http",
			OutputToContext: true,
		}

		err := executor.Init()
		assert.NoError(t, err)

		flowContext := flow.FlowContext{}
		result, err := executor.Run(flowContext, nil)

		assert.NoError(t, err)
		assert.Equal(t, "MCP session initialized with ID: placeholder-session-id", result.Text)
		assert.Equal(t, "placeholder-session-id", result.GetVariableString("mcpSessionId", ""))
	})

	// Additional test case from builtin_executors_test.go
	t.Run("session initialization with step", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        "http://localhost:8080",
			OutputToContext: true,
		}
		err := executor.Init()
		assert.NoError(t, err)

		flowContext := flow.FlowContext{}
		step := &flow.Step{}

		result, err := executor.Run(flowContext, step)
		assert.NoError(t, err)

		assert.Equal(t, "MCP session initialized with ID: placeholder-session-id", result.Text)
		assert.Equal(t, "placeholder-session-id", result.GetVariableString("mcpSessionId", ""))
	})
}

// TestMCPExecutor_WithHTTPServer tests the executor with a mock HTTP server
func TestMCPExecutor_WithHTTPServer(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Respond based on request path
		switch {
		case strings.HasPrefix(r.URL.Path, "/resources/"):
			resourceID := strings.TrimPrefix(r.URL.Path, "/resources/")
			// Return a simple response based on the resource ID
			response := map[string]interface{}{
				"resourceID": resourceID,
				"content":    fmt.Sprintf("Content for resource %s", resourceID),
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/tools/"):
			// Parse the request body to get the tool params
			var requestBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			if err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			toolName := strings.TrimPrefix(r.URL.Path, "/tools/")
			response := map[string]interface{}{
				"toolName": toolName,
				"status":   "success",
				"result":   fmt.Sprintf("Result from tool %s", toolName),
				"params":   requestBody,
			}
			json.NewEncoder(w).Encode(response)

		default:
			// Session initialization or other requests
			response := map[string]interface{}{
				"sessionID": "test-session-456",
				"status":    "initialized",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// Test that the server works as expected
	t.Run("server test", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/resources/testfile")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "testfile", result["resourceID"])
	})

	// Test with our executor (using the placeholder implementation)
	t.Run("executor with server endpoint", func(t *testing.T) {
		executor := &MCPExecutor{
			Endpoint:        server.URL,
			Transport:       "http",
			OutputToContext: true,
			ResultVarName:   "serverTest",
		}

		err := executor.Init()
		assert.NoError(t, err)
		assert.Equal(t, server.URL, executor.Endpoint)

		// We won't actually make HTTP requests from the test since our
		// MCPExecutor uses placeholder implementations, but this confirms
		// the structure works correctly
		flowContext := flow.FlowContext{}
		result, err := executor.Run(flowContext, nil)

		assert.NoError(t, err)
		assert.Contains(t, result.Text, "placeholder-session-id")
	})
}

// TestMCPExecutor_ComplexStructures tests handling of complex data structures
func TestMCPExecutor_ComplexStructures(t *testing.T) {
	// Create a MCPExecutor for resource reading
	executor := &MCPExecutor{
		Endpoint:        "http://localhost:8080",
		Transport:       "http",
		ResourceURI:     "/resources/test",
		OutputToContext: true,
		ResultVarName:   "resourceResult",
	}

	err := executor.Init()
	assert.NoError(t, err)

	// The placeholder implementation of readResource returns a string,
	// which should be handled correctly by the executor
	flowContext := flow.FlowContext{}
	result, err := executor.Run(flowContext, nil)

	assert.NoError(t, err)
	resourceVar, exists := result.Variables["resourceResult"]
	assert.True(t, exists)

	resourceMap, ok := resourceVar.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "/resources/test", resourceMap["uri"])

	// Make sure the content is what we expect from the placeholder
	assert.Contains(t, resourceMap["contents"].(string), "Placeholder content for resource")
}

// TestMCPExecutor_ErrorHandling tests basic error handling
func TestMCPExecutor_ErrorHandling(t *testing.T) {
	// Try to initialize with invalid parameters
	badExecutor := &MCPExecutor{}
	err := badExecutor.Init()
	assert.Error(t, err)

	// Test that a step reference is required
	emptyExecutor := &MCPExecutor{
		Endpoint: "http://localhost:8080",
	}
	err = emptyExecutor.Init()
	assert.NoError(t, err)

	// Run should succeed with the placeholder implementation
	flowContext := flow.FlowContext{}
	_, err = emptyExecutor.Run(flowContext, nil)
	assert.NoError(t, err)
}
