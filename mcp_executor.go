package anyi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jieliu2000/anyi/flow"
)

// MCPExecutor is an executor that communicates with MCP servers.
// It allows the flow to interact with Model Context Protocol servers for enhanced context and tool usage.
type MCPExecutor struct {
	// Server endpoint URL for MCP server
	Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`

	// API key for authentication (if required)
	APIKey string `json:"apiKey" yaml:"apiKey" mapstructure:"apiKey"`

	// Transport type (http, sse, or stdio)
	Transport string `json:"transport" yaml:"transport" mapstructure:"transport"`

	// SessionID for tracking the MCP session
	SessionID string `json:"sessionId" yaml:"sessionId" mapstructure:"sessionId"`

	// Tool name to call (if empty, will use the text as input)
	ToolName string `json:"toolName" yaml:"toolName" mapstructure:"toolName"`

	// Arguments for tool call, these will be extracted from the flow context variables
	ToolArgVars []string `json:"toolArgVars" yaml:"toolArgVars" mapstructure:"toolArgVars"`

	// Resource URI to read (if empty, will use tool call)
	ResourceURI string `json:"resourceUri" yaml:"resourceUri" mapstructure:"resourceUri"`

	// Output to context flag, if true will write the result to the flow context text
	OutputToContext bool `json:"outputToContext" yaml:"outputToContext" mapstructure:"outputToContext"`

	// Variable name to store the result in, if not set will use "mcpResult"
	ResultVarName string `json:"resultVarName" yaml:"resultVarName" mapstructure:"resultVarName"`

	// HTTP client for making requests
	httpClient *http.Client

	// Whether the executor has been initialized
	initialized bool
}

// Init initializes the MCPExecutor.
// It checks if the required parameters are set and initializes the HTTP client if needed.
func (executor *MCPExecutor) Init() error {
	if executor.Endpoint == "" {
		return errors.New("MCP endpoint cannot be empty")
	}

	// Default to HTTP transport if not specified
	if executor.Transport == "" {
		executor.Transport = "http"
	}

	// Validate transport type
	if !isValidTransport(executor.Transport) {
		return fmt.Errorf("invalid transport type: %s, must be one of: http, sse, stdio", executor.Transport)
	}

	// Initialize HTTP client if using HTTP or SSE transport
	if executor.Transport == "http" || executor.Transport == "sse" {
		executor.httpClient = &http.Client{}
	}

	// Set default result variable name if not specified
	if executor.ResultVarName == "" {
		executor.ResultVarName = "mcpResult"
	}

	executor.initialized = true
	return nil
}

// isValidTransport checks if the transport type is valid
func isValidTransport(transport string) bool {
	return transport == "http" || transport == "sse" || transport == "stdio"
}

// Run executes the MCP request.
// It communicates with the MCP server based on the configured parameters and handles the response.
//
// Parameters:
//   - flowContext: The current flow context
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context with the MCP response
//   - Any error encountered during execution
func (executor *MCPExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	if !executor.initialized {
		if err := executor.Init(); err != nil {
			return &flowContext, err
		}
	}

	// Create context for request
	ctx := context.Background()

	// Handle based on the requested operation
	if executor.ResourceURI != "" {
		// Handle resource read operation
		return executor.handleResourceRead(ctx, flowContext)
	} else if executor.ToolName != "" {
		// Handle tool call operation
		return executor.handleToolCall(ctx, flowContext)
	} else {
		// Use text as input for a default operation (session context)
		return executor.handleDefault(ctx, flowContext)
	}
}

// handleResourceRead processes a resource read operation
func (executor *MCPExecutor) handleResourceRead(ctx context.Context, flowContext flow.FlowContext) (*flow.FlowContext, error) {
	// Format the URI with any template parameters from variables
	uri := executor.formatStringWithVariables(executor.ResourceURI, flowContext.Variables)

	log.Printf("Reading MCP resource: %s", uri)

	// Construct the resource read request
	resource, err := executor.readResource(ctx, uri)
	if err != nil {
		return &flowContext, fmt.Errorf("failed to read MCP resource: %w", err)
	}

	// Process the response
	result := map[string]interface{}{
		"uri":      uri,
		"contents": resource,
	}

	// Store the result in variables
	flowContext.SetVariable(executor.ResultVarName, result)

	// If configured, set the text output as well
	if executor.OutputToContext {
		// If resource is a string, set it directly
		if content, ok := resource.(string); ok {
			flowContext.Text = content
		} else {
			// Otherwise, marshal to JSON
			jsonContent, err := json.MarshalIndent(resource, "", "  ")
			if err != nil {
				return &flowContext, fmt.Errorf("error marshaling resource content to JSON: %w", err)
			}
			flowContext.Text = string(jsonContent)
		}
	}

	return &flowContext, nil
}

// handleToolCall processes a tool call operation
func (executor *MCPExecutor) handleToolCall(ctx context.Context, flowContext flow.FlowContext) (*flow.FlowContext, error) {
	log.Printf("Calling MCP tool: %s", executor.ToolName)

	// Extract arguments from variables
	args := make(map[string]interface{})
	for _, argVar := range executor.ToolArgVars {
		if value := flowContext.GetVariable(argVar); value != nil {
			args[argVar] = value
		}
	}

	// Call the tool
	result, err := executor.callTool(ctx, executor.ToolName, args)
	if err != nil {
		return &flowContext, fmt.Errorf("failed to call MCP tool: %w", err)
	}

	// Store the result in variables
	flowContext.SetVariable(executor.ResultVarName, result)

	// If configured, set the text output as well
	if executor.OutputToContext {
		// If result contains a "text" field, use that
		if textContent, ok := result["text"].(string); ok {
			flowContext.Text = textContent
		} else {
			// Otherwise, marshal to JSON
			jsonContent, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return &flowContext, fmt.Errorf("error marshaling tool result to JSON: %w", err)
			}
			flowContext.Text = string(jsonContent)
		}
	}

	return &flowContext, nil
}

// handleDefault processes a default session operation
func (executor *MCPExecutor) handleDefault(ctx context.Context, flowContext flow.FlowContext) (*flow.FlowContext, error) {
	// For now, we'll just initialize a session and store the session ID
	sessionID, err := executor.initializeSession(ctx)
	if err != nil {
		return &flowContext, fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	// Store the session ID in variables
	flowContext.SetVariable("mcpSessionId", sessionID)

	// If configured, set the text output as well
	if executor.OutputToContext {
		flowContext.Text = fmt.Sprintf("MCP session initialized with ID: %s", sessionID)
	}

	return &flowContext, nil
}

// formatStringWithVariables replaces variable placeholders in the format ${varName}
// with their corresponding values from the variables map
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
				// Try to convert to JSON
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

// The following are placeholder implementations that would need to be replaced
// with actual MCP client implementations in a production environment

func (executor *MCPExecutor) readResource(ctx context.Context, uri string) (interface{}, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use the MCP client to read a resource
	return fmt.Sprintf("Placeholder content for resource: %s", uri), nil
}

func (executor *MCPExecutor) callTool(ctx context.Context, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use the MCP client to call a tool
	return map[string]interface{}{
		"status": "success",
		"text":   fmt.Sprintf("Placeholder result for tool: %s with args: %v", toolName, args),
	}, nil
}

func (executor *MCPExecutor) initializeSession(ctx context.Context) (string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use the MCP client to initialize a session
	return "placeholder-session-id", nil
}

// Register the MCP executor
func init() {
	// Register the executor with the name "mcp"
	RegisterExecutor("mcp", &MCPExecutor{})
}
