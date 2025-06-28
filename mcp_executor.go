package anyi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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

// MCPOperation defines the type of MCP operation
type MCPOperation string

const (
	OperationToolCall      MCPOperation = "tool_call"
	OperationResourceRead  MCPOperation = "resource_read"
	OperationPromptGet     MCPOperation = "prompt_get"
	OperationListTools     MCPOperation = "list_tools"
	OperationListResources MCPOperation = "list_resources"
)

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
// It provides a clean abstraction for Anyi workflows to interact with various MCP servers.
type MCPExecutor struct {
	// Server configuration - only one should be specified
	ServerEndpoint string   `json:"serverEndpoint,omitempty" yaml:"serverEndpoint,omitempty" mapstructure:"serverEndpoint"`
	ServerCommand  string   `json:"serverCommand,omitempty" yaml:"serverCommand,omitempty" mapstructure:"serverCommand"`
	ServerArgs     []string `json:"serverArgs,omitempty" yaml:"serverArgs,omitempty" mapstructure:"serverArgs"`

	// Transport configuration
	Transport MCPTransport `json:"transport" yaml:"transport" mapstructure:"transport"`

	// Authentication (optional)
	APIKey string `json:"apiKey,omitempty" yaml:"apiKey,omitempty" mapstructure:"apiKey"`

	// Operation configuration
	Operation MCPOperation `json:"operation" yaml:"operation" mapstructure:"operation"`

	// Tool call parameters
	ToolName    string                 `json:"toolName,omitempty" yaml:"toolName,omitempty" mapstructure:"toolName"`
	ToolArgs    map[string]interface{} `json:"toolArgs,omitempty" yaml:"toolArgs,omitempty" mapstructure:"toolArgs"`
	ToolArgVars []string               `json:"toolArgVars,omitempty" yaml:"toolArgVars,omitempty" mapstructure:"toolArgVars"`

	// Resource parameters
	ResourceURI string `json:"resourceUri,omitempty" yaml:"resourceUri,omitempty" mapstructure:"resourceUri"`

	// Prompt parameters
	PromptName string                 `json:"promptName,omitempty" yaml:"promptName,omitempty" mapstructure:"promptName"`
	PromptArgs map[string]interface{} `json:"promptArgs,omitempty" yaml:"promptArgs,omitempty" mapstructure:"promptArgs"`

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

// Init initializes the MCPExecutor with proper validation and client setup
func (executor *MCPExecutor) Init() error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	if executor.initialized {
		return nil
	}

	// Validate configuration
	if err := executor.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	executor.setDefaults()

	// Create appropriate client based on transport
	client, err := executor.createClient()
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	executor.client = client
	executor.initialized = true

	return nil
}

// validateConfig validates the executor configuration
func (executor *MCPExecutor) validateConfig() error {
	// Validate transport
	if executor.Transport != TransportHTTP &&
		executor.Transport != TransportSSE &&
		executor.Transport != TransportSTDIO {
		return errors.New("transport must be one of: http, sse, stdio")
	}

	// Validate server configuration
	if executor.Transport == TransportSTDIO {
		if executor.ServerCommand == "" {
			return errors.New("serverCommand is required for stdio transport")
		}
	} else {
		if executor.ServerEndpoint == "" {
			return errors.New("serverEndpoint is required for http/sse transport")
		}
	}

	// Validate operation
	if executor.Operation == "" {
		return errors.New("operation must be specified")
	}

	// Validate operation-specific parameters
	switch executor.Operation {
	case OperationToolCall:
		if executor.ToolName == "" {
			return errors.New("toolName is required for tool_call operation")
		}
	case OperationResourceRead:
		if executor.ResourceURI == "" {
			return errors.New("resourceUri is required for resource_read operation")
		}
	case OperationPromptGet:
		if executor.PromptName == "" {
			return errors.New("promptName is required for prompt_get operation")
		}
	}

	return nil
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

	if executor.PromptArgs == nil {
		executor.PromptArgs = make(map[string]interface{})
	}
}

// createClient creates the appropriate MCP client based on transport type
func (executor *MCPExecutor) createClient() (MCPClient, error) {
	switch executor.Transport {
	case TransportHTTP:
		return NewHTTPMCPClient(executor.ServerEndpoint, executor.APIKey, executor.Timeout)
	case TransportSSE:
		return NewSSEMCPClient(executor.ServerEndpoint, executor.APIKey, executor.Timeout)
	case TransportSTDIO:
		return NewSTDIOMCPClient(executor.ServerCommand, executor.ServerArgs, executor.Timeout)
	default:
		return nil, fmt.Errorf("unsupported transport: %s", executor.Transport)
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
	switch executor.Operation {
	case OperationToolCall:
		args := executor.buildToolArguments(flowContext)
		return executor.client.CallTool(ctx, executor.ToolName, args)

	case OperationResourceRead:
		uri := executor.formatStringWithVariables(executor.ResourceURI, flowContext.Variables)
		return executor.client.ReadResource(ctx, uri)

	case OperationPromptGet:
		args := executor.buildPromptArguments(flowContext)
		return executor.client.GetPrompt(ctx, executor.PromptName, args)

	case OperationListTools:
		return executor.client.ListTools(ctx)

	case OperationListResources:
		return executor.client.ListResources(ctx)

	default:
		return nil, fmt.Errorf("unsupported operation: %s", executor.Operation)
	}
}

// buildToolArguments builds tool arguments from configuration and flow context
func (executor *MCPExecutor) buildToolArguments(flowContext flow.FlowContext) map[string]interface{} {
	args := make(map[string]interface{})

	// Copy static arguments
	for key, value := range executor.ToolArgs {
		args[key] = value
	}

	// Add arguments from flow context variables
	for _, argVar := range executor.ToolArgVars {
		if value := flowContext.GetVariable(argVar); value != nil {
			args[argVar] = value
		}
	}

	return args
}

// buildPromptArguments builds prompt arguments from configuration and flow context
func (executor *MCPExecutor) buildPromptArguments(flowContext flow.FlowContext) map[string]interface{} {
	args := make(map[string]interface{})

	// Copy static arguments
	for key, value := range executor.PromptArgs {
		args[key] = value
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
	// TODO: Implement SSE-specific fields
}

// NewSSEMCPClient creates a new SSE MCP client
func NewSSEMCPClient(endpoint, apiKey string, timeout time.Duration) (*SSEMCPClient, error) {
	return &SSEMCPClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		timeout:  timeout,
	}, nil
}

// Implement MCPClient interface for SSEMCPClient
func (c *SSEMCPClient) Initialize(ctx context.Context) error {
	// TODO: Implement SSE initialization
	return errors.New("SSE transport not yet implemented")
}

func (c *SSEMCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	return nil, errors.New("SSE transport not yet implemented")
}

func (c *SSEMCPClient) ReadResource(ctx context.Context, uri string) (*MCPResponse, error) {
	return nil, errors.New("SSE transport not yet implemented")
}

func (c *SSEMCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	return nil, errors.New("SSE transport not yet implemented")
}

func (c *SSEMCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	return nil, errors.New("SSE transport not yet implemented")
}

func (c *SSEMCPClient) ListResources(ctx context.Context) (*MCPResponse, error) {
	return nil, errors.New("SSE transport not yet implemented")
}

func (c *SSEMCPClient) Close() error {
	return nil
}

// STDIOMCPClient implements MCPClient for STDIO transport
type STDIOMCPClient struct {
	command string
	args    []string
	timeout time.Duration
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

// NewSTDIOMCPClient creates a new STDIO MCP client
func NewSTDIOMCPClient(command string, args []string, timeout time.Duration) (*STDIOMCPClient, error) {
	return &STDIOMCPClient{
		command: command,
		args:    args,
		timeout: timeout,
	}, nil
}

// Implement MCPClient interface for STDIOMCPClient
func (c *STDIOMCPClient) Initialize(ctx context.Context) error {
	// TODO: Implement STDIO initialization
	return errors.New("STDIO transport not yet implemented")
}

func (c *STDIOMCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	return nil, errors.New("STDIO transport not yet implemented")
}

func (c *STDIOMCPClient) ReadResource(ctx context.Context, uri string) (*MCPResponse, error) {
	return nil, errors.New("STDIO transport not yet implemented")
}

func (c *STDIOMCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	return nil, errors.New("STDIO transport not yet implemented")
}

func (c *STDIOMCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	return nil, errors.New("STDIO transport not yet implemented")
}

func (c *STDIOMCPClient) ListResources(ctx context.Context) (*MCPResponse, error) {
	return nil, errors.New("STDIO transport not yet implemented")
}

func (c *STDIOMCPClient) Close() error {
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// Register the optimized MCP executor
func init() {
	RegisterExecutor("mcp", &MCPExecutor{})
}
