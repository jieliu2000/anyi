// Package mcp provides implementation of the Model Context Protocol (MCP) for Anyi.
package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

// MCPModelConfig defines the configuration for MCP client
type MCPModelConfig struct {
	// Endpoint is the URL of the MCP server
	Endpoint string `json:"endpoint" mapstructure:"endpoint" yaml:"endpoint"`

	// APIKey is the API key for authentication
	APIKey string `json:"apiKey" mapstructure:"apiKey" yaml:"apiKey"`

	// ModelID is the ID of the model to use
	ModelID string `json:"modelId" mapstructure:"modelId" yaml:"modelId"`

	// Timeout specifies the timeout for requests
	Timeout time.Duration `json:"timeout" mapstructure:"timeout" yaml:"timeout"`

	// AllowInsecureHTTP allows using HTTP instead of HTTPS
	AllowInsecureHTTP bool `json:"allowInsecureHttp" mapstructure:"allowInsecureHttp" yaml:"allowInsecureHttp"`
}

// MCPClient implements the llm.Client interface for MCP
type MCPClient struct {
	Config     *MCPModelConfig
	HTTPClient *http.Client
}

// DefaultEndpoint is the default MCP endpoint
const DefaultEndpoint = "https://api.example.com/mcp/v1"

// DefaultModel is the default model ID
const DefaultModel = "default-model"

// DefaultConfig creates a default configuration with the given API key
func DefaultConfig(apiKey string) *MCPModelConfig {
	return &MCPModelConfig{
		Endpoint: DefaultEndpoint,
		APIKey:   apiKey,
		ModelID:  DefaultModel,
		Timeout:  30 * time.Second,
	}
}

// DefaultConfigWithModel creates a default configuration with the given API key and model ID
func DefaultConfigWithModel(apiKey string, modelID string) *MCPModelConfig {
	config := DefaultConfig(apiKey)
	config.ModelID = modelID
	return config
}

// NewConfig creates a new configuration with all parameters
func NewConfig(apiKey string, modelID string, endpoint string, timeout time.Duration) *MCPModelConfig {
	return &MCPModelConfig{
		Endpoint: endpoint,
		APIKey:   apiKey,
		ModelID:  modelID,
		Timeout:  timeout,
	}
}

// NewClient creates a new MCP client with the given configuration
func NewClient(config *MCPModelConfig) (*MCPClient, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if config.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	if config.Endpoint == "" {
		config.Endpoint = DefaultEndpoint
	}

	if config.ModelID == "" {
		config.ModelID = DefaultModel
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &MCPClient{
		Config:     config,
		HTTPClient: httpClient,
	}, nil
}

// mcpRequest represents a request to the MCP API
type mcpRequest struct {
	ModelID   string                 `json:"model_id" mapstructure:"model_id" yaml:"model_id"`
	Messages  []chat.Message         `json:"messages" mapstructure:"messages" yaml:"messages"`
	Options   *mcpRequestOptions     `json:"options,omitempty" mapstructure:"options,omitempty" yaml:"options,omitempty"`
	Functions []tools.FunctionConfig `json:"functions,omitempty" mapstructure:"functions,omitempty" yaml:"functions,omitempty"`
}

// mcpRequestOptions represents the options for a request to the MCP API
type mcpRequestOptions struct {
	Temperature      float64 `json:"temperature,omitempty" mapstructure:"temperature,omitempty" yaml:"temperature,omitempty"`
	TopP             float64 `json:"top_p,omitempty" mapstructure:"topP,omitempty" yaml:"topP,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty" mapstructure:"maxTokens,omitempty" yaml:"maxTokens,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty" mapstructure:"presencePenalty,omitempty" yaml:"presencePenalty,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty" mapstructure:"frequencyPenalty,omitempty" yaml:"frequencyPenalty,omitempty"`
	Format           string  `json:"format,omitempty" mapstructure:"format,omitempty" yaml:"format,omitempty"`
}

// mcpResponse represents a response from the MCP API
type mcpResponse struct {
	ID      string       `json:"id" mapstructure:"id" yaml:"id"`
	Message chat.Message `json:"message" mapstructure:"message" yaml:"message"`
	Model   string       `json:"model" mapstructure:"model" yaml:"model"`
	Usage   mcpUsage     `json:"usage" mapstructure:"usage" yaml:"usage"`
	Error   *mcpError    `json:"error,omitempty" mapstructure:"error,omitempty" yaml:"error,omitempty"`
}

// mcpUsage represents the token usage in a response
type mcpUsage struct {
	PromptTokens     int `json:"prompt_tokens" mapstructure:"prompt_tokens" yaml:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens" mapstructure:"completion_tokens" yaml:"completion_tokens"`
	TotalTokens      int `json:"total_tokens" mapstructure:"total_tokens" yaml:"total_tokens"`
}

// mcpError represents an error returned by the MCP API
type mcpError struct {
	Code    string `json:"code" mapstructure:"code" yaml:"code"`
	Message string `json:"message" mapstructure:"message" yaml:"message"`
}

// convertChatOptions converts chat.ChatOptions to mcpRequestOptions
func convertChatOptions(options *chat.ChatOptions) *mcpRequestOptions {
	if options == nil {
		return nil
	}

	return &mcpRequestOptions{
		Format: options.Format,
		// Add default values for other fields
		Temperature: 0.7,
		MaxTokens:   2048,
	}
}

// Chat sends a chat request to the MCP API
func (c *MCPClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	request := mcpRequest{
		ModelID:  c.Config.ModelID,
		Messages: messages,
		Options:  convertChatOptions(options),
	}

	return c.sendRequest(request)
}

// ChatWithFunctions sends a chat request with functions to the MCP API
func (c *MCPClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	request := mcpRequest{
		ModelID:   c.Config.ModelID,
		Messages:  messages,
		Options:   convertChatOptions(options),
		Functions: functions,
	}

	return c.sendRequest(request)
}

// sendRequest sends a request to the MCP API and processes the response
func (c *MCPClient) sendRequest(request mcpRequest) (*chat.Message, chat.ResponseInfo, error) {
	// Serialize the request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, chat.ResponseInfo{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.Config.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, chat.ResponseInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, chat.ResponseInfo{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, chat.ResponseInfo{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-2xx status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, chat.ResponseInfo{}, fmt.Errorf("MCP API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var mcpResp mcpResponse
	if err := json.Unmarshal(bodyBytes, &mcpResp); err != nil {
		return nil, chat.ResponseInfo{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API error
	if mcpResp.Error != nil {
		return nil, chat.ResponseInfo{}, fmt.Errorf("MCP API error: %s: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	// Create response info
	responseInfo := chat.ResponseInfo{
		PromptTokens:     mcpResp.Usage.PromptTokens,
		CompletionTokens: mcpResp.Usage.CompletionTokens,
	}

	return &mcpResp.Message, responseInfo, nil
}
