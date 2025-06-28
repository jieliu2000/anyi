package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockMCPServer(t *testing.T) {
	server := NewMockMCPServer()
	defer server.Close()

	t.Run("handles valid MCP request", func(t *testing.T) {
		// Prepare request
		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-1",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "test_tool",
				"arguments": map[string]interface{}{"param1": "value1"},
			},
		}

		requestBody, err := json.Marshal(request)
		assert.NoError(t, err)

		// Send request
		resp, err := http.Post(server.URL(), "application/json", bytes.NewBuffer(requestBody))
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response MCPResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "test-1", response.ID)
		assert.NotNil(t, response.Result)
		assert.Nil(t, response.Error)

		// Verify request was recorded
		requests := server.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "tools/call", requests[0].Method)
	})

	t.Run("handles error response", func(t *testing.T) {
		// Set error response
		server.SetErrorResponse("error/method", -1, "Test error")

		// Clear request history
		server.ClearRequests()

		// Prepare request
		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-2",
			Method:  "error/method",
		}

		requestBody, err := json.Marshal(request)
		assert.NoError(t, err)

		// Send request
		resp, err := http.Post(server.URL(), "application/json", bytes.NewBuffer(requestBody))
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response MCPResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "test-2", response.ID)
		assert.Nil(t, response.Result)
		assert.NotNil(t, response.Error)
		assert.Equal(t, -1, response.Error.Code)
		assert.Equal(t, "Test error", response.Error.Message)
	})

	t.Run("handles custom response", func(t *testing.T) {
		// Set custom response
		customResponse := map[string]interface{}{
			"custom": "data",
			"value":  42.0, // JSON decoding converts numbers to float64
		}
		server.SetResponse("custom/method", customResponse)

		// Clear request history
		server.ClearRequests()

		// Prepare request
		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-3",
			Method:  "custom/method",
		}

		requestBody, err := json.Marshal(request)
		assert.NoError(t, err)

		// Send request
		resp, err := http.Post(server.URL(), "application/json", bytes.NewBuffer(requestBody))
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		var response MCPResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Equal(t, customResponse, response.Result)
	})

	t.Run("rejects non-POST requests", func(t *testing.T) {
		resp, err := http.Get(server.URL())
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		resp, err := http.Post(server.URL(), "application/json", bytes.NewBufferString("invalid json"))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("request tracking functions", func(t *testing.T) {
		// Clear request history
		server.ClearRequests()
		assert.Len(t, server.GetRequests(), 0)
		assert.Nil(t, server.GetLastRequest())

		// Send a request
		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-tracking",
			Method:  "test/method",
		}

		requestBody, err := json.Marshal(request)
		assert.NoError(t, err)

		http.Post(server.URL(), "application/json", bytes.NewBuffer(requestBody))

		// Verify request was recorded
		requests := server.GetRequests()
		assert.Len(t, requests, 1)

		lastRequest := server.GetLastRequest()
		assert.NotNil(t, lastRequest)
		assert.Equal(t, "test/method", lastRequest.Method)
		assert.Equal(t, "test-tracking", lastRequest.ID)
	})
}
