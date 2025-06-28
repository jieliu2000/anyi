package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockMCPServer is a mock MCP server for testing
type MockMCPServer struct {
	Server    *httptest.Server
	Responses map[string]interface{} // Predefined responses indexed by method name
	Requests  []MCPRequest           // Record of received requests
}

// MCPRequest represents an MCP request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMockMCPServer creates a new mock MCP server
func NewMockMCPServer() *MockMCPServer {
	mock := &MockMCPServer{
		Responses: make(map[string]interface{}),
		Requests:  make([]MCPRequest, 0),
	}

	// Set default responses
	mock.SetDefaultResponses()

	// Create HTTP server
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock
}

// SetDefaultResponses sets default MCP responses
func (m *MockMCPServer) SetDefaultResponses() {
	// Tool call response
	m.Responses["tools/call"] = map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Tool executed successfully",
			},
		},
	}

	// Resource read response
	m.Responses["resources/read"] = map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      "test-uri",
				"mimeType": "text/plain",
				"text":     "Mock resource content",
			},
		},
	}

	// Prompt get response
	m.Responses["prompts/get"] = map[string]interface{}{
		"description": "Mock prompt",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": map[string]interface{}{
					"type": "text",
					"text": "Mock prompt content",
				},
			},
		},
	}

	// Tools list response
	m.Responses["tools/list"] = map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "test_tool",
				"description": "A test tool",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param1": map[string]interface{}{
							"type":        "string",
							"description": "Test parameter",
						},
					},
				},
			},
		},
	}

	// Resources list response
	m.Responses["resources/list"] = map[string]interface{}{
		"resources": []map[string]interface{}{
			{
				"uri":         "/test-resource",
				"name":        "Test Resource",
				"description": "A test resource",
				"mimeType":    "text/plain",
			},
		},
	}
}

// SetResponse sets a specific response for a method
func (m *MockMCPServer) SetResponse(method string, response interface{}) {
	m.Responses[method] = response
}

// SetErrorResponse sets an error response for a specific method
func (m *MockMCPServer) SetErrorResponse(method string, code int, message string) {
	m.Responses[method] = &MCPError{
		Code:    code,
		Message: message,
	}
}

// GetRequests returns all received requests
func (m *MockMCPServer) GetRequests() []MCPRequest {
	return m.Requests
}

// GetLastRequest returns the last received request
func (m *MockMCPServer) GetLastRequest() *MCPRequest {
	if len(m.Requests) == 0 {
		return nil
	}
	return &m.Requests[len(m.Requests)-1]
}

// ClearRequests clears the request history
func (m *MockMCPServer) ClearRequests() {
	m.Requests = make([]MCPRequest, 0)
}

// Close closes the server
func (m *MockMCPServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// URL returns the server URL
func (m *MockMCPServer) URL() string {
	if m.Server != nil {
		return m.Server.URL
	}
	return ""
}

// handleRequest handles HTTP requests
func (m *MockMCPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Record the request
	m.Requests = append(m.Requests, req)

	// Find response
	response := m.findResponse(req.Method)

	// Build MCP response
	mcpResponse := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	// Check if it's an error response
	if mcpError, ok := response.(*MCPError); ok {
		mcpResponse.Error = mcpError
	} else {
		mcpResponse.Result = response
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Send response
	if err := json.NewEncoder(w).Encode(mcpResponse); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// findResponse finds a response based on the method
func (m *MockMCPServer) findResponse(method string) interface{} {
	// Direct match
	if response, exists := m.Responses[method]; exists {
		return response
	}

	// Fuzzy match (supports partial method names)
	for key, response := range m.Responses {
		if strings.Contains(method, key) || strings.Contains(key, method) {
			return response
		}
	}

	// Default response
	return map[string]interface{}{
		"message": fmt.Sprintf("Mock response for method: %s", method),
		"success": true,
	}
}
