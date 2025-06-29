package anyi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jieliu2000/anyi/flow"
)

// MCPTransport defines the transport mechanism for MCP communication
type MCPTransport string

const (
	TransportHTTP  MCPTransport = "http"
	TransportSSE   MCPTransport = "sse"
	TransportSTDIO MCPTransport = "stdio"
)

// MCPServerPreset defines preset configurations for popular MCP servers
type MCPServerPreset string

const (
	PresetGitHub     MCPServerPreset = "github"
	PresetFileSystem MCPServerPreset = "filesystem"
	PresetFetch      MCPServerPreset = "fetch"
	PresetMemory     MCPServerPreset = "memory"
	PresetSlack      MCPServerPreset = "slack"
	PresetNotaion    MCPServerPreset = "notion"
)

// MCPServerConfig contains MCP server configuration in a simplified format
type MCPServerConfig struct {
	// Basic server identification
	Name string       `json:"name" yaml:"name" mapstructure:"name"`
	Type MCPTransport `json:"type" yaml:"type" mapstructure:"type"`

	// Server connection details
	Command string   `json:"command,omitempty" yaml:"command,omitempty" mapstructure:"command"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty" mapstructure:"args"`
	URL     string   `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url"`

	// Environment and authentication
	Env     map[string]string `json:"env,omitempty" yaml:"env,omitempty" mapstructure:"env"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" mapstructure:"headers"`

	// Optional settings
	Enabled     bool          `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`
	Timeout     time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout"`
	Tools       []string      `json:"tools,omitempty" yaml:"tools,omitempty" mapstructure:"tools"` // Filter specific tools
	AutoApprove bool          `json:"autoApprove,omitempty" yaml:"autoApprove,omitempty" mapstructure:"autoApprove"`
}

// MCPRequest represents a generic MCP request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a generic MCP response
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

// MCPClient defines the interface for MCP client operations
type MCPClient interface {
	Initialize(ctx context.Context) error
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error)
	ReadResource(ctx context.Context, uri string) (*MCPResponse, error)
	GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error)
	ListTools(ctx context.Context) (*MCPResponse, error)
	ListResources(ctx context.Context) (*MCPResponse, error)
	Close() error
}

// MCPExecutor is an optimized executor that communicates with MCP servers.
// It uses a simplified configuration format similar to VSCode/Cursor implementations.
type MCPExecutor struct {
	// Server configuration (can use preset or custom config)
	Preset MCPServerPreset  `json:"preset,omitempty" yaml:"preset,omitempty" mapstructure:"preset"`
	Server *MCPServerConfig `json:"server,omitempty" yaml:"server,omitempty" mapstructure:"server"`

	// Dynamic operation parameters (set at runtime)
	Action   string                 `json:"action" yaml:"action" mapstructure:"action"` // "call_tool", "read_resource", "get_prompt", "list_tools", "list_resources"
	ToolName string                 `json:"toolName,omitempty" yaml:"toolName,omitempty" mapstructure:"toolName"`
	ToolArgs map[string]interface{} `json:"toolArgs,omitempty" yaml:"toolArgs,omitempty" mapstructure:"toolArgs"`
	Resource string                 `json:"resource,omitempty" yaml:"resource,omitempty" mapstructure:"resource"`
	Prompt   string                 `json:"prompt,omitempty" yaml:"prompt,omitempty" mapstructure:"prompt"`

	// Output configuration
	OutputToContext bool   `json:"outputToContext" yaml:"outputToContext" mapstructure:"outputToContext"`
	ResultVarName   string `json:"resultVarName" yaml:"resultVarName" mapstructure:"resultVarName"`

	// Connection settings
	Timeout       time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout"`
	RetryAttempts int           `json:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty" mapstructure:"retryAttempts"`

	// Internal state
	client      MCPClient
	initialized bool
	mutex       sync.RWMutex
}

// getPresetConfig returns the configuration for a preset server
func getPresetConfig(preset MCPServerPreset) (*MCPServerConfig, error) {
	presets := map[MCPServerPreset]*MCPServerConfig{
		PresetGitHub: {
			Name:    "github",
			Type:    TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			Env: map[string]string{
				"GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}",
			},
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		PresetFileSystem: {
			Name:    "filesystem",
			Type:    TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		PresetFetch: {
			Name:    "fetch",
			Type:    TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-fetch"},
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		PresetMemory: {
			Name:    "memory",
			Type:    TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-memory"},
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		PresetSlack: {
			Name:    "slack",
			Type:    TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-slack"},
			Env: map[string]string{
				"SLACK_BOT_TOKEN": "${SLACK_BOT_TOKEN}",
				"SLACK_TEAM_ID":   "${SLACK_TEAM_ID}",
			},
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		PresetNotaion: {
			Name:    "notion",
			Type:    TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-notion"},
			Env: map[string]string{
				"NOTION_API_TOKEN": "${NOTION_API_TOKEN}",
			},
			Enabled: true,
			Timeout: 30 * time.Second,
		},
	}

	config, exists := presets[preset]
	if !exists {
		return nil, fmt.Errorf("unknown preset: %s", preset)
	}

	// Deep copy to avoid modifying the original
	configCopy := *config
	return &configCopy, nil
}

// resolveEnvironmentVariables resolves environment variable placeholders in the format ${VAR_NAME}
func resolveEnvironmentVariables(input string) string {
	// Simple ${VAR_NAME} replacement
	result := input
	processed := 0

	for {
		start := strings.Index(result[processed:], "${")
		if start == -1 {
			break
		}
		start += processed

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := result[start+2 : end]
		envValue, exists := os.LookupEnv(varName)

		if exists {
			// Replace the variable with its value
			result = result[:start] + envValue + result[end+1:]
			processed = start + len(envValue)
		} else {
			// Skip this variable and continue looking after it
			processed = end + 1
		}
	}
	return result
}

// Init initializes the MCPExecutor with proper validation and client setup
func (executor *MCPExecutor) Init() error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	if executor.initialized {
		return nil
	}

	// Resolve server configuration
	config, err := executor.resolveServerConfig()
	if err != nil {
		return fmt.Errorf("failed to resolve server configuration: %w", err)
	}

	// Validate configuration
	if err := executor.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	executor.setDefaults()

	// Create appropriate client based on transport
	client, err := executor.createClient(config)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	executor.client = client
	executor.initialized = true

	return nil
}

// resolveServerConfig resolves the final server configuration from preset or custom config
func (executor *MCPExecutor) resolveServerConfig() (*MCPServerConfig, error) {
	var config *MCPServerConfig
	var err error

	// Use preset if specified
	if executor.Preset != "" {
		config, err = getPresetConfig(executor.Preset)
		if err != nil {
			return nil, err
		}
	} else if executor.Server != nil {
		// Use custom server config
		config = executor.Server
	} else {
		return nil, errors.New("no server configuration provided (use 'preset' for quick setup or 'server' for custom configuration)")
	}

	// Resolve environment variables in configuration
	if config.Env != nil {
		for key, value := range config.Env {
			config.Env[key] = resolveEnvironmentVariables(value)
		}
	}

	if config.Headers != nil {
		for key, value := range config.Headers {
			config.Headers[key] = resolveEnvironmentVariables(value)
		}
	}

	if config.URL != "" {
		config.URL = resolveEnvironmentVariables(config.URL)
	}

	return config, nil
}

// validateConfig validates the resolved server configuration
func (executor *MCPExecutor) validateConfig(config *MCPServerConfig) error {
	if config == nil {
		return errors.New("server configuration is required")
	}

	// Validate transport type
	if config.Type != TransportHTTP && config.Type != TransportSSE && config.Type != TransportSTDIO {
		return fmt.Errorf("invalid transport type: %s (must be 'http', 'sse', or 'stdio')", config.Type)
	}

	// Validate transport-specific configuration
	switch config.Type {
	case TransportSTDIO:
		if config.Command == "" {
			return errors.New("command is required for stdio transport")
		}
	case TransportHTTP, TransportSSE:
		if config.URL == "" {
			return errors.New("url is required for http/sse transport")
		}
	}

	// Validate action if specified
	if executor.Action != "" {
		validActions := map[string]bool{
			"call_tool":      true,
			"read_resource":  true,
			"get_prompt":     true,
			"list_tools":     true,
			"list_resources": true,
		}
		if !validActions[executor.Action] {
			return fmt.Errorf("invalid action: %s (must be one of: call_tool, read_resource, get_prompt, list_tools, list_resources)", executor.Action)
		}

		// Validate action-specific parameters
		switch executor.Action {
		case "call_tool":
			if executor.ToolName == "" {
				return errors.New("toolName is required for call_tool action")
			}
		case "read_resource":
			if executor.Resource == "" {
				return errors.New("resource is required for read_resource action")
			}
		case "get_prompt":
			if executor.Prompt == "" {
				return errors.New("prompt is required for get_prompt action")
			}
		}
	}

	return nil
}

// createClient creates the appropriate MCP client based on transport type
func (executor *MCPExecutor) createClient(config *MCPServerConfig) (MCPClient, error) {
	switch config.Type {
	case TransportHTTP:
		apiKey := ""
		if config.Headers != nil {
			if auth, ok := config.Headers["Authorization"]; ok {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		return NewHTTPMCPClient(config.URL, apiKey, config.Timeout)
	case TransportSSE:
		apiKey := ""
		if config.Headers != nil {
			if auth, ok := config.Headers["Authorization"]; ok {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		return NewSSEMCPClient(config.URL, apiKey, config.Timeout)
	case TransportSTDIO:
		return NewSTDIOMCPClient(config.Command, config.Args, config.Timeout)
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", config.Type)
	}
}

// setDefaults sets default values for optional fields
func (executor *MCPExecutor) setDefaults() {
	if executor.ResultVarName == "" {
		executor.ResultVarName = "mcpResult"
	}

	if executor.Timeout == 0 {
		executor.Timeout = 30 * time.Second
	}

	if executor.RetryAttempts == 0 {
		executor.RetryAttempts = 3
	}

	if executor.ToolArgs == nil {
		executor.ToolArgs = make(map[string]interface{})
	}
}

// Run executes the MCP operation
func (executor *MCPExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	if !executor.initialized {
		if err := executor.Init(); err != nil {
			return &flowContext, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), executor.Timeout)
	defer cancel()

	// Initialize client if needed
	if err := executor.client.Initialize(ctx); err != nil {
		return &flowContext, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Execute operation with retry logic
	var response *MCPResponse
	var err error

	for attempt := 0; attempt < executor.RetryAttempts; attempt++ {
		response, err = executor.executeOperation(ctx, flowContext)
		if err == nil {
			break
		}

		if attempt < executor.RetryAttempts-1 {
			log.Printf("MCP operation failed (attempt %d/%d): %v, retrying...",
				attempt+1, executor.RetryAttempts, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return &flowContext, fmt.Errorf("MCP operation failed after %d attempts: %w",
			executor.RetryAttempts, err)
	}

	// Process response
	return executor.processResponse(response, flowContext)
}

// executeOperation executes the specific MCP operation
func (executor *MCPExecutor) executeOperation(ctx context.Context, flowContext flow.FlowContext) (*MCPResponse, error) {
	switch executor.Action {
	case "call_tool":
		args := executor.buildToolArguments(flowContext)
		return executor.client.CallTool(ctx, executor.ToolName, args)

	case "read_resource":
		uri := executor.formatStringWithVariables(executor.Resource, flowContext.Variables)
		return executor.client.ReadResource(ctx, uri)

	case "get_prompt":
		args := executor.buildPromptArguments(flowContext)
		return executor.client.GetPrompt(ctx, executor.Prompt, args)

	case "list_tools":
		return executor.client.ListTools(ctx)

	case "list_resources":
		return executor.client.ListResources(ctx)

	default:
		return nil, fmt.Errorf("unsupported operation: %s", executor.Action)
	}
}

// buildToolArguments builds tool arguments from configuration and flow context
func (executor *MCPExecutor) buildToolArguments(flowContext flow.FlowContext) map[string]interface{} {
	args := make(map[string]interface{})

	// Copy static arguments
	for key, value := range executor.ToolArgs {
		args[key] = value
	}

	// Process string values that might contain variable placeholders
	for key, value := range args {
		if strValue, ok := value.(string); ok {
			args[key] = executor.formatStringWithVariables(strValue, flowContext.Variables)
		}
	}

	return args
}

// buildPromptArguments builds prompt arguments from configuration and flow context
func (executor *MCPExecutor) buildPromptArguments(flowContext flow.FlowContext) map[string]interface{} {
	args := make(map[string]interface{})

	// For prompt arguments, we use ToolArgs as the general arguments container
	for key, value := range executor.ToolArgs {
		args[key] = value
	}

	// Process string values that might contain variable placeholders
	for key, value := range args {
		if strValue, ok := value.(string); ok {
			args[key] = executor.formatStringWithVariables(strValue, flowContext.Variables)
		}
	}

	return args
}

// processResponse processes the MCP response and updates flow context
func (executor *MCPExecutor) processResponse(response *MCPResponse, flowContext flow.FlowContext) (*flow.FlowContext, error) {
	if response.Error != nil {
		return &flowContext, fmt.Errorf("MCP error %d: %s", response.Error.Code, response.Error.Message)
	}

	// Store result in variables
	flowContext.SetVariable(executor.ResultVarName, response.Result)

	// Set output to context if configured
	if executor.OutputToContext {
		if err := executor.setContextOutput(response.Result, &flowContext); err != nil {
			return &flowContext, fmt.Errorf("failed to set context output: %w", err)
		}
	}

	return &flowContext, nil
}

// setContextOutput sets the flow context text based on the response
func (executor *MCPExecutor) setContextOutput(result interface{}, flowContext *flow.FlowContext) error {
	switch v := result.(type) {
	case string:
		flowContext.Text = v
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok {
			flowContext.Text = text
		} else if content, ok := v["content"].(string); ok {
			flowContext.Text = content
		} else {
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return err
			}
			flowContext.Text = string(jsonBytes)
		}
	default:
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		flowContext.Text = string(jsonBytes)
	}

	return nil
}

// formatStringWithVariables replaces variable placeholders
func (executor *MCPExecutor) formatStringWithVariables(format string, variables map[string]interface{}) string {
	result := format

	for key, value := range variables {
		placeholder := "${" + key + "}"
		if strings.Contains(result, placeholder) {
			var replacement string

			switch v := value.(type) {
			case string:
				replacement = v
			case fmt.Stringer:
				replacement = v.String()
			default:
				if jsonBytes, err := json.Marshal(value); err == nil {
					replacement = string(jsonBytes)
				} else {
					replacement = fmt.Sprintf("%v", value)
				}
			}

			result = strings.ReplaceAll(result, placeholder, replacement)
		}
	}

	return result
}

// Close closes the MCP client connection
func (executor *MCPExecutor) Close() error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	if executor.client != nil {
		return executor.client.Close()
	}

	return nil
}

// HTTPMCPClient implements MCPClient for HTTP transport
type HTTPMCPClient struct {
	endpoint string
	apiKey   string
	client   *http.Client
	timeout  time.Duration
}

// NewHTTPMCPClient creates a new HTTP MCP client
func NewHTTPMCPClient(endpoint, apiKey string, timeout time.Duration) (*HTTPMCPClient, error) {
	return &HTTPMCPClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: timeout},
		timeout:  timeout,
	}, nil
}

// Initialize initializes the HTTP client
func (c *HTTPMCPClient) Initialize(ctx context.Context) error {
	// For HTTP, we can test connectivity with a simple request
	return nil
}

// CallTool calls an MCP tool via HTTP
func (c *HTTPMCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("tool-%d", time.Now().UnixNano()),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendRequest(ctx, request)
}

// ReadResource reads an MCP resource via HTTP
func (c *HTTPMCPClient) ReadResource(ctx context.Context, uri string) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("resource-%d", time.Now().UnixNano()),
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	return c.sendRequest(ctx, request)
}

// GetPrompt gets an MCP prompt via HTTP
func (c *HTTPMCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("prompt-%d", time.Now().UnixNano()),
		Method:  "prompts/get",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendRequest(ctx, request)
}

// ListTools lists available MCP tools
func (c *HTTPMCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("list-tools-%d", time.Now().UnixNano()),
		Method:  "tools/list",
	}

	return c.sendRequest(ctx, request)
}

// ListResources lists available MCP resources
func (c *HTTPMCPClient) ListResources(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("list-resources-%d", time.Now().UnixNano()),
		Method:  "resources/list",
	}

	return c.sendRequest(ctx, request)
}

// sendRequest sends an HTTP request to the MCP server
func (c *HTTPMCPClient) sendRequest(ctx context.Context, request MCPRequest) (*MCPResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response MCPResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// Close closes the HTTP client
func (c *HTTPMCPClient) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// SSEMCPClient implements MCPClient for Server-Sent Events transport
type SSEMCPClient struct {
	endpoint string
	apiKey   string
	timeout  time.Duration
	client   *http.Client
	eventCh  chan []byte
	done     chan struct{}
	mutex    sync.RWMutex
}

// NewSSEMCPClient creates a new SSE MCP client
func NewSSEMCPClient(endpoint, apiKey string, timeout time.Duration) (*SSEMCPClient, error) {
	return &SSEMCPClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		timeout:  timeout,
		client:   &http.Client{Timeout: timeout},
		eventCh:  make(chan []byte, 100),
		done:     make(chan struct{}),
	}, nil
}

// Initialize initializes the SSE connection
func (c *SSEMCPClient) Initialize(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create SSE connection
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE endpoint: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("SSE connection failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Start reading SSE events
	go c.readSSEEvents(resp.Body)

	return nil
}

// readSSEEvents reads Server-Sent Events from the response body
func (c *SSEMCPClient) readSSEEvents(body io.ReadCloser) {
	defer body.Close()
	defer close(c.eventCh)

	scanner := bufio.NewScanner(body)
	var eventData bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		// Handle SSE format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			eventData.WriteString(data)
		} else if line == "" {
			// Empty line indicates end of event
			if eventData.Len() > 0 {
				select {
				case c.eventCh <- eventData.Bytes():
				case <-c.done:
					return
				}
				eventData.Reset()
			}
		}
	}
}

// sendSSERequest sends a request via SSE and waits for response
func (c *SSEMCPClient) sendSSERequest(ctx context.Context, request MCPRequest) (*MCPResponse, error) {
	// For SSE, we typically send requests via a separate HTTP POST endpoint
	// and receive responses via the SSE stream
	postEndpoint := strings.TrimSuffix(c.endpoint, "/events") + "/request"

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", postEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	// Wait for response via SSE
	timeout := time.NewTimer(c.timeout)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, errors.New("timeout waiting for SSE response")
		case eventData := <-c.eventCh:
			if eventData == nil {
				return nil, errors.New("SSE connection closed")
			}

			var response MCPResponse
			if err := json.Unmarshal(eventData, &response); err != nil {
				continue // Skip malformed events
			}

			// Check if this response matches our request ID
			if response.ID == request.ID {
				return &response, nil
			}
		}
	}
}

func (c *SSEMCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("tool-%d", time.Now().UnixNano()),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendSSERequest(ctx, request)
}

func (c *SSEMCPClient) ReadResource(ctx context.Context, uri string) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("resource-%d", time.Now().UnixNano()),
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	return c.sendSSERequest(ctx, request)
}

func (c *SSEMCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("prompt-%d", time.Now().UnixNano()),
		Method:  "prompts/get",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendSSERequest(ctx, request)
}

func (c *SSEMCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("list-tools-%d", time.Now().UnixNano()),
		Method:  "tools/list",
	}

	return c.sendSSERequest(ctx, request)
}

func (c *SSEMCPClient) ListResources(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("list-resources-%d", time.Now().UnixNano()),
		Method:  "resources/list",
	}

	return c.sendSSERequest(ctx, request)
}

func (c *SSEMCPClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	close(c.done)
	return nil
}

// STDIOMCPClient implements MCPClient for STDIO transport
type STDIOMCPClient struct {
	command   string
	args      []string
	timeout   time.Duration
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	scanner   *bufio.Scanner
	responses map[string]chan *MCPResponse
	mutex     sync.RWMutex
	done      chan struct{}
}

// NewSTDIOMCPClient creates a new STDIO MCP client
func NewSTDIOMCPClient(command string, args []string, timeout time.Duration) (*STDIOMCPClient, error) {
	return &STDIOMCPClient{
		command:   command,
		args:      args,
		timeout:   timeout,
		responses: make(map[string]chan *MCPResponse),
		done:      make(chan struct{}),
	}, nil
}

// Initialize initializes the STDIO connection by starting the MCP server process
func (c *STDIOMCPClient) Initialize(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cmd != nil {
		return nil // Already initialized
	}

	// Create the command
	c.cmd = exec.CommandContext(ctx, c.command, c.args...)

	// Set up pipes
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		c.stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = stdout

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		c.stdin.Close()
		c.stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = stderr

	// Start the process
	if err := c.cmd.Start(); err != nil {
		c.stdin.Close()
		c.stdout.Close()
		c.stderr.Close()
		return fmt.Errorf("failed to start MCP server process: %w", err)
	}

	// Set up scanner for reading responses
	c.scanner = bufio.NewScanner(c.stdout)

	// Start goroutines to handle I/O
	go c.readResponses()
	go c.readErrors()

	// Send initialization request
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      "init",
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
				"sampling": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "anyi-mcp-client",
				"version": "1.0.0",
			},
		},
	}

	response, err := c.sendSTDIORequest(ctx, initRequest)
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to initialize MCP server: %w", err)
	}

	if response.Error != nil {
		c.Close()
		return fmt.Errorf("MCP server initialization error: %s", response.Error.Message)
	}

	return nil
}

// readResponses reads JSON-RPC responses from stdout
func (c *STDIOMCPClient) readResponses() {
	defer func() {
		c.mutex.Lock()
		for _, ch := range c.responses {
			close(ch)
		}
		c.mutex.Unlock()
	}()

	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var response MCPResponse
		if err := json.Unmarshal(line, &response); err != nil {
			log.Printf("Failed to unmarshal MCP response: %v", err)
			continue
		}

		c.mutex.RLock()
		if ch, exists := c.responses[response.ID]; exists {
			select {
			case ch <- &response:
			case <-c.done:
				c.mutex.RUnlock()
				return
			}
		}
		c.mutex.RUnlock()
	}
}

// readErrors reads error messages from stderr
func (c *STDIOMCPClient) readErrors() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		log.Printf("MCP server stderr: %s", scanner.Text())
	}
}

// sendSTDIORequest sends a JSON-RPC request via STDIO and waits for response
func (c *STDIOMCPClient) sendSTDIORequest(ctx context.Context, request MCPRequest) (*MCPResponse, error) {
	// Create response channel
	responseCh := make(chan *MCPResponse, 1)

	c.mutex.Lock()
	c.responses[request.ID] = responseCh
	c.mutex.Unlock()

	// Clean up response channel when done
	defer func() {
		c.mutex.Lock()
		delete(c.responses, request.ID)
		c.mutex.Unlock()
		close(responseCh)
	}()

	// Marshal and send request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Add newline for JSON-RPC over STDIO
	jsonData = append(jsonData, '\n')

	if _, err := c.stdin.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to write request to stdin: %w", err)
	}

	// Wait for response
	timeout := time.NewTimer(c.timeout)
	defer timeout.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeout.C:
		return nil, errors.New("timeout waiting for STDIO response")
	case response := <-responseCh:
		if response == nil {
			return nil, errors.New("STDIO connection closed")
		}
		return response, nil
	case <-c.done:
		return nil, errors.New("STDIO client closed")
	}
}

func (c *STDIOMCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("tool-%d", time.Now().UnixNano()),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendSTDIORequest(ctx, request)
}

func (c *STDIOMCPClient) ReadResource(ctx context.Context, uri string) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("resource-%d", time.Now().UnixNano()),
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	return c.sendSTDIORequest(ctx, request)
}

func (c *STDIOMCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("prompt-%d", time.Now().UnixNano()),
		Method:  "prompts/get",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendSTDIORequest(ctx, request)
}

func (c *STDIOMCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("list-tools-%d", time.Now().UnixNano()),
		Method:  "tools/list",
	}

	return c.sendSTDIORequest(ctx, request)
}

func (c *STDIOMCPClient) ListResources(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("list-resources-%d", time.Now().UnixNano()),
		Method:  "resources/list",
	}

	return c.sendSTDIORequest(ctx, request)
}

func (c *STDIOMCPClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Signal shutdown
	close(c.done)

	// Close pipes
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	// Terminate process
	if c.cmd != nil && c.cmd.Process != nil {
		// Try graceful shutdown first
		c.cmd.Process.Signal(os.Interrupt)

		// Wait for process to exit or kill it
		done := make(chan error, 1)
		go func() {
			done <- c.cmd.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			// Force kill if not exited within 5 seconds
			c.cmd.Process.Kill()
			c.cmd.Wait()
		case <-done:
			// Process exited gracefully
		}
	}

	return nil
}

// Register the optimized MCP executor
func init() {
	RegisterExecutor("mcp", &MCPExecutor{})
}
