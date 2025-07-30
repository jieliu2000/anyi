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
		executor       *MCPExecutor
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid preset configuration",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				Preset: PresetGitHub,
			},
			expectError: false,
		},
		{
			name: "valid custom server configuration",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				Server: &MCPServerConfig{
					Name:    "test-server",
					Type:    TransportHTTP,
					URL:     "http://localhost:8080",
					Enabled: true,
				},
			},
			expectError: false,
		},

		{
			name: "missing server configuration",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
			},
			expectError:    true,
			errorSubstring: "no server configuration provided",
		},
		{
			name: "invalid action",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action: "invalid_action",
				},
				Preset: PresetGitHub,
			},
			expectError:    true,
			errorSubstring: "invalid action",
		},
		{
			name: "missing tool name for call_tool action",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action: "call_tool",
				},
				Preset: PresetGitHub,
			},
			expectError:    true,
			errorSubstring: "toolName is required",
		},
		{
			name: "missing resource for read_resource action",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action: "read_resource",
				},
				Preset: PresetGitHub,
			},
			expectError:    true,
			errorSubstring: "resource is required",
		},
		{
			name: "missing prompt for get_prompt action",
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action: "get_prompt",
				},
				Preset: PresetGitHub,
			},
			expectError:    true,
			errorSubstring: "prompt is required",
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

func TestHTTPMCPExecutor_Init(t *testing.T) {
	tests := []struct {
		name           string
		executor       *HTTPMCPExecutor
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid HTTP configuration",
			executor: &HTTPMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name: "http-server",
					Type: TransportHTTP,
					URL:  "http://localhost:8080",
				},
			},
			expectError: false,
		},
		{
			name: "HTTP missing URL",
			executor: &HTTPMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name: "http-server",
					Type: TransportHTTP,
				},
			},
			expectError:    true,
			errorSubstring: "url is required",
		},
		{
			name: "missing tool name for call_tool action",
			executor: &HTTPMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action: "call_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name: "http-server",
					Type: TransportHTTP,
					URL:  "http://localhost:8080",
				},
			},
			expectError:    true,
			errorSubstring: "toolName is required",
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
			}
		})
	}
}

func TestSSEMCPExecutor_Init(t *testing.T) {
	tests := []struct {
		name           string
		executor       *SSEMCPExecutor
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid SSE configuration",
			executor: &SSEMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name: "sse-server",
					Type: TransportSSE,
					URL:  "http://localhost:8080/events",
				},
			},
			expectError: false,
		},
		{
			name: "SSE missing URL",
			executor: &SSEMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name: "sse-server",
					Type: TransportSSE,
				},
			},
			expectError:    true,
			errorSubstring: "url is required",
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
			}
		})
	}
}

func TestSTDIOMCPExecutor_Init(t *testing.T) {
	tests := []struct {
		name           string
		executor       *STDIOMCPExecutor
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid STDIO configuration",
			executor: &STDIOMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name:    "stdio-server",
					Type:    TransportSTDIO,
					Command: "node",
					Args:    []string{"server.js"},
				},
			},
			expectError: false,
		},
		{
			name: "STDIO missing command",
			executor: &STDIOMCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
				ServerConfig: &MCPServerConfig{
					Name: "stdio-server",
					Type: TransportSTDIO,
				},
			},
			expectError:    true,
			errorSubstring: "command is required",
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
			}
		})
	}
}

func TestMCPExecutor_PresetConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		preset MCPServerPreset
	}{
		{"GitHub preset", PresetGitHub},
		{"FileSystem preset", PresetFileSystem},
		{"Fetch preset", PresetFetch},
		{"Memory preset", PresetMemory},
		{"Slack preset", PresetSlack},
		{"Notion preset", PresetNotaion},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := getPresetConfig(tc.preset)
			assert.NoError(t, err)
			assert.NotNil(t, config)
			assert.NotEmpty(t, config.Name)
			assert.NotEmpty(t, config.Type)
		})
	}
}

func TestMCPExecutor_EnvironmentVariableResolution(t *testing.T) {
	// Set test environment variable
	testKey := "TEST_MCP_VAR"
	testValue := "test_value_123"
	t.Setenv(testKey, testValue)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single variable",
			input:    "${TEST_MCP_VAR}",
			expected: testValue,
		},
		{
			name:     "variable in string",
			input:    "prefix_${TEST_MCP_VAR}_suffix",
			expected: "prefix_" + testValue + "_suffix",
		},
		{
			name:     "missing variable",
			input:    "${MISSING_VAR}",
			expected: "${MISSING_VAR}",
		},
		{
			name:     "no variables",
			input:    "simple_string",
			expected: "simple_string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolveEnvironmentVariables(tc.input)
			assert.Equal(t, tc.expected, result)
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
				BaseMCPExecutor: BaseMCPExecutor{
					Action:   "call_tool",
					ToolName: "test_tool",
				},
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
		// Clear previous requests
		mockServer.ClearRequests()

		executor := &MCPExecutor{
			BaseMCPExecutor: BaseMCPExecutor{
				Action:   "call_tool",
				ToolName: "test_tool",
				ToolArgs: map[string]interface{}{"param1": "value1"},
			},
			Server: &MCPServerConfig{
				Name: "test-server",
				Type: TransportHTTP,
				URL:  mockServer.URL(),
			},
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
		// Clear previous requests
		mockServer.ClearRequests()

		executor := &MCPExecutor{
			BaseMCPExecutor: BaseMCPExecutor{
				Action:          "read_resource",
				Resource:        "/test-resource",
				OutputToContext: true,
				ResultVarName:   "resourceResult",
			},
			Server: &MCPServerConfig{
				Name: "test-server",
				Type: TransportHTTP,
				URL:  mockServer.URL(),
			},
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

		// Check that result was stored in variables
		assert.NotNil(t, result.GetVariable("resourceResult"))

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "resources/read", requests[0].Method)
	})

	t.Run("prompt get with mock server", func(t *testing.T) {
		// Clear previous requests
		mockServer.ClearRequests()

		executor := &MCPExecutor{
			BaseMCPExecutor: BaseMCPExecutor{
				Action:        "get_prompt",
				Prompt:        "test_prompt",
				ToolArgs:      map[string]interface{}{"param1": "value1"},
				ResultVarName: "promptResult",
			},
			Server: &MCPServerConfig{
				Name: "test-server",
				Type: TransportHTTP,
				URL:  mockServer.URL(),
			},
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

		// Check that result was stored in variables
		assert.NotNil(t, result.GetVariable("promptResult"))

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "prompts/get", requests[0].Method)
	})

	t.Run("list tools with mock server", func(t *testing.T) {
		// Clear previous requests
		mockServer.ClearRequests()

		executor := &MCPExecutor{
			BaseMCPExecutor: BaseMCPExecutor{
				Action:        "list_tools",
				ResultVarName: "toolsList",
			},
			Server: &MCPServerConfig{
				Name: "test-server",
				Type: TransportHTTP,
				URL:  mockServer.URL(),
			},
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

		// Check that result was stored in variables
		assert.NotNil(t, result.GetVariable("toolsList"))

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "tools/list", requests[0].Method)
	})

	t.Run("list resources with mock server", func(t *testing.T) {
		// Clear previous requests
		mockServer.ClearRequests()

		executor := &MCPExecutor{
			BaseMCPExecutor: BaseMCPExecutor{
				Action:        "list_resources",
				ResultVarName: "resourcesList",
			},
			Server: &MCPServerConfig{
				Name: "test-server",
				Type: TransportHTTP,
				URL:  mockServer.URL(),
			},
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

		// Check that result was stored in variables
		assert.NotNil(t, result.GetVariable("resourcesList"))

		// Verify request was sent to mock server
		requests := mockServer.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "resources/list", requests[0].Method)
	})
}

func TestMCPExecutor_TransportSpecificValidation(t *testing.T) {
	tests := []struct {
		name           string
		serverConfig   *MCPServerConfig
		expectError    bool
		errorSubstring string
	}{
		{
			name: "valid HTTP config",
			serverConfig: &MCPServerConfig{
				Name: "http-server",
				Type: TransportHTTP,
				URL:  "http://localhost:8080",
			},
			expectError: false,
		},
		{
			name: "valid STDIO config",
			serverConfig: &MCPServerConfig{
				Name:    "stdio-server",
				Type:    TransportSTDIO,
				Command: "node",
				Args:    []string{"server.js"},
			},
			expectError: false,
		},
		{
			name: "HTTP missing URL",
			serverConfig: &MCPServerConfig{
				Name: "http-server",
				Type: TransportHTTP,
			},
			expectError:    true,
			errorSubstring: "url is required",
		},
		{
			name: "STDIO missing command",
			serverConfig: &MCPServerConfig{
				Name: "stdio-server",
				Type: TransportSTDIO,
			},
			expectError:    true,
			errorSubstring: "command is required",
		},
		{
			name: "invalid transport type",
			serverConfig: &MCPServerConfig{
				Name: "invalid-server",
				Type: "invalid",
			},
			expectError:    true,
			errorSubstring: "invalid transport type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					Action: "list_tools",
				},
				Server: tc.serverConfig,
			}

			err := executor.Init()

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorSubstring != "" {
					assert.Contains(t, err.Error(), tc.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMCPExecutor_DefaultSettings(t *testing.T) {
	executor := &MCPExecutor{
		BaseMCPExecutor: BaseMCPExecutor{
			Action:   "call_tool",
			ToolName: "test_tool",
		},
		Preset: PresetMemory, // Add a preset to provide server configuration
	}

	err := executor.Init()
	require.NoError(t, err)

	// Check default values
	assert.Equal(t, "mcpResult", executor.ResultVarName)
	assert.Equal(t, 30*time.Second, executor.Timeout)
	assert.Equal(t, 3, executor.RetryAttempts)
	assert.NotNil(t, executor.ToolArgs)
}

func TestMCPExecutor_ArgumentBuilding(t *testing.T) {
	executor := &MCPExecutor{
		BaseMCPExecutor: BaseMCPExecutor{
			ToolArgs: map[string]interface{}{
				"static_param":   "static_value",
				"template_param": "${user_id}",
				"number_param":   42,
			},
		},
	}

	flowContext := flow.FlowContext{
		Variables: map[string]interface{}{
			"user_id": "12345",
		},
	}

	args := executor.buildToolArguments(flowContext)

	assert.Equal(t, "static_value", args["static_param"])
	assert.Equal(t, "12345", args["template_param"])
	assert.Equal(t, 42, args["number_param"])
}

func TestMCPExecutor_PromptArgumentBuilding(t *testing.T) {
	executor := &MCPExecutor{
		BaseMCPExecutor: BaseMCPExecutor{
			ToolArgs: map[string]interface{}{
				"context":  "${context_data}",
				"language": "en",
			},
		},
	}

	flowContext := flow.FlowContext{
		Variables: map[string]interface{}{
			"context_data": "test context",
		},
	}

	args := executor.buildPromptArguments(flowContext)

	assert.Equal(t, "test context", args["context"])
	assert.Equal(t, "en", args["language"])
}

func TestMCPExecutor_ResponseProcessing(t *testing.T) {
	tests := []struct {
		name     string
		response *MCPResponse
		executor *MCPExecutor
		expected string
	}{
		{
			name: "string result",
			response: &MCPResponse{
				JSONRPC: "2.0",
				ID:      "test",
				Result:  "Simple text result",
			},
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					OutputToContext: true,
					ResultVarName:   "testResult",
				},
			},
			expected: "Simple text result",
		},
		{
			name: "map result with text field",
			response: &MCPResponse{
				JSONRPC: "2.0",
				ID:      "test",
				Result: map[string]interface{}{
					"text": "Text from map",
					"meta": "additional data",
				},
			},
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					OutputToContext: true,
					ResultVarName:   "testResult",
				},
			},
			expected: "Text from map",
		},
		{
			name: "map result with content field",
			response: &MCPResponse{
				JSONRPC: "2.0",
				ID:      "test",
				Result: map[string]interface{}{
					"content": "Content from map",
					"meta":    "additional data",
				},
			},
			executor: &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{
					OutputToContext: true,
					ResultVarName:   "testResult",
				},
			},
			expected: "Content from map",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flowContext := flow.FlowContext{
				Variables: make(map[string]interface{}),
			}

			result, err := tc.executor.processResponse(tc.response, flowContext)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.expected, result.Text)
			assert.Equal(t, tc.response.Result, result.GetVariable(tc.executor.ResultVarName))
		})
	}
}

func TestMCPExecutor_ErrorHandling(t *testing.T) {
	response := &MCPResponse{
		JSONRPC: "2.0",
		ID:      "test",
		Error: &MCPError{
			Code:    -1,
			Message: "Test error message",
		},
	}

	executor := &MCPExecutor{
		BaseMCPExecutor: BaseMCPExecutor{
			ResultVarName: "testResult",
		},
	}

	flowContext := flow.FlowContext{
		Variables: make(map[string]interface{}),
	}

	result, err := executor.processResponse(response, flowContext)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MCP error -1: Test error message")
	assert.NotNil(t, result)
}

func TestMCPExecutor_CustomServerConfiguration(t *testing.T) {
	// Test that custom server configuration works
	executor := &MCPExecutor{
		BaseMCPExecutor: BaseMCPExecutor{
			Action: "list_tools",
		},
		Server: &MCPServerConfig{
			Name:    "custom-server",
			Type:    TransportSTDIO,
			Command: "node",
			Args:    []string{"server.js"},
		},
	}

	err := executor.Init()
	assert.NoError(t, err)
	assert.True(t, executor.initialized)
	assert.NotNil(t, executor.client)
}

func TestMCPExecutor_StringFormatting(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		variables map[string]interface{}
		expected  string
	}{
		{
			name:      "basic string replacement",
			format:    "Hello ${name}!",
			variables: map[string]interface{}{"name": "World"},
			expected:  "Hello World!",
		},
		{
			name:   "multiple variables",
			format: "${greeting} ${name}, welcome to ${place}!",
			variables: map[string]interface{}{
				"greeting": "Hello",
				"name":     "Alice",
				"place":    "Wonderland",
			},
			expected: "Hello Alice, welcome to Wonderland!",
		},
		{
			name:   "non-string values",
			format: "Count: ${count}, Active: ${active}",
			variables: map[string]interface{}{
				"count":  42,
				"active": true,
			},
			expected: "Count: 42, Active: true",
		},
		{
			name:   "json objects",
			format: "Config: ${config}",
			variables: map[string]interface{}{
				"config": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
			},
			expected: `Config: {"host":"localhost","port":8080}`,
		},
		{
			name:      "no variables to replace",
			format:    "Static string without variables",
			variables: map[string]interface{}{"unused": "value"},
			expected:  "Static string without variables",
		},
		{
			name:      "missing variables",
			format:    "Hello ${name}, age ${age}",
			variables: map[string]interface{}{"name": "Bob"},
			expected:  "Hello Bob, age ${age}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MCPExecutor{
				BaseMCPExecutor: BaseMCPExecutor{},
			}
			result := executor.formatStringWithVariables(tc.format, tc.variables)
			assert.Equal(t, tc.expected, result)
		})
	}
}
