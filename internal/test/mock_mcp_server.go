package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

// MockMCPServer is a mock MCP server for testing
type MockMCPServer struct {
	Server    *httptest.Server
	Responses map[string]interface{} // Predefined responses indexed by method name
	Requests  []MCPRequest           // Record of received requests

	// SSE support
	sseClients map[string]chan []byte // SSE clients indexed by connection ID
	sseMutex   sync.RWMutex           // Protects sseClients map
}

// MockSSEServer is a specialized mock server for SSE testing
type MockSSEServer struct {
	Server    *httptest.Server
	EventChan chan []byte
	Requests  []MCPRequest
	Responses map[string]interface{}
	mutex     sync.RWMutex
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
		Responses:  make(map[string]interface{}),
		Requests:   make([]MCPRequest, 0),
		sseClients: make(map[string]chan []byte),
	}

	// Set default responses
	mock.SetDefaultResponses()

	// Create HTTP server
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock
}

// NewMockSSEServer creates a new mock SSE server
func NewMockSSEServer() *MockSSEServer {
	mock := &MockSSEServer{
		EventChan: make(chan []byte, 100),
		Requests:  make([]MCPRequest, 0),
		Responses: make(map[string]interface{}),
	}

	// Set default responses
	mock.setDefaultResponses()

	// Create mux for handling different endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/events", mock.handleSSE)
	mux.HandleFunc("/request", mock.handleRequest)

	mock.Server = httptest.NewServer(mux)

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

// setDefaultResponses sets default responses for SSE server
func (m *MockSSEServer) setDefaultResponses() {
	// Tool call response
	m.Responses["tools/call"] = map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "SSE Tool executed successfully",
			},
		},
	}

	// Resource read response
	m.Responses["resources/read"] = map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      "test-uri",
				"mimeType": "text/plain",
				"text":     "SSE Mock resource content",
			},
		},
	}

	// Prompt get response
	m.Responses["prompts/get"] = map[string]interface{}{
		"description": "SSE Mock prompt",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": map[string]interface{}{
					"type": "text",
					"text": "SSE Mock prompt content",
				},
			},
		},
	}

	// Tools list response
	m.Responses["tools/list"] = map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "sse_test_tool",
				"description": "An SSE test tool",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param1": map[string]interface{}{
							"type":        "string",
							"description": "SSE Test parameter",
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
				"uri":         "/sse-test-resource",
				"name":        "SSE Test Resource",
				"description": "An SSE test resource",
				"mimeType":    "text/plain",
			},
		},
	}
}

// handleSSE handles SSE connections
func (m *MockSSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Flush headers
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Send initial connection message
	fmt.Fprintf(w, "data: {\"type\":\"connection\",\"status\":\"connected\"}\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Listen for events
	for {
		select {
		case event := <-m.EventChan:
			fmt.Fprintf(w, "data: %s\n\n", string(event))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			// Send keep-alive
			fmt.Fprintf(w, "data: {\"type\":\"keepalive\"}\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// handleRequest handles HTTP POST requests for SSE server
func (m *MockSSEServer) handleRequest(w http.ResponseWriter, r *http.Request) {
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
	m.mutex.Lock()
	m.Requests = append(m.Requests, req)
	m.mutex.Unlock()

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

	// Send response via SSE
	responseBytes, err := json.Marshal(mcpResponse)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	// Send to SSE channel
	select {
	case m.EventChan <- responseBytes:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"queued"}`))
	default:
		http.Error(w, "SSE channel full", http.StatusServiceUnavailable)
	}
}

// findResponse finds a response for SSE server
func (m *MockSSEServer) findResponse(method string) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Direct match
	if response, exists := m.Responses[method]; exists {
		return response
	}

	// Fuzzy match
	for key, response := range m.Responses {
		if strings.Contains(method, key) || strings.Contains(key, method) {
			return response
		}
	}

	// Default response
	return map[string]interface{}{
		"message": fmt.Sprintf("SSE Mock response for method: %s", method),
		"success": true,
	}
}

// GetRequests returns all received requests for SSE server
func (m *MockSSEServer) GetRequests() []MCPRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return append([]MCPRequest{}, m.Requests...)
}

// ClearRequests clears the request history for SSE server
func (m *MockSSEServer) ClearRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Requests = make([]MCPRequest, 0)
}

// SetResponse sets a specific response for SSE server
func (m *MockSSEServer) SetResponse(method string, response interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Responses[method] = response
}

// SetErrorResponse sets an error response for SSE server
func (m *MockSSEServer) SetErrorResponse(method string, code int, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Responses[method] = &MCPError{
		Code:    code,
		Message: message,
	}
}

// Close closes the SSE server
func (m *MockSSEServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
	close(m.EventChan)
}

// URL returns the SSE server URL
func (m *MockSSEServer) URL() string {
	if m.Server != nil {
		return m.Server.URL
	}
	return ""
}
