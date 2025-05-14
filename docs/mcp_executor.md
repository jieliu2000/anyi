# MCP Executor for Anyi

The MCP (Model Context Protocol) Executor enables Anyi workflows to interact with MCP-compliant servers, providing enhanced context and tools for AI applications.

## What is MCP?

Model Context Protocol (MCP) is an open standard released by Anthropic that enables seamless integration between LLM applications and external data sources and tools. It provides a standardized way for language models to access contextual information and use external tools.

Key features of MCP include:

- Resource access: Reading contextual information from external sources
- Tool execution: Performing actions using external tools
- Standardized communication: Consistent protocol for all integrations

## MCP Executor Configuration

The MCP Executor in Anyi can be configured with the following parameters:

| Parameter         | Type     | Description                                              | Required                     |
| ----------------- | -------- | -------------------------------------------------------- | ---------------------------- |
| `endpoint`        | string   | MCP server endpoint URL                                  | Yes                          |
| `apiKey`          | string   | API key for authentication (if required)                 | No                           |
| `transport`       | string   | Transport type: "http", "sse", or "stdio"                | No (defaults to "http")      |
| `sessionId`       | string   | Session ID for tracking (if reusing an existing session) | No                           |
| `toolName`        | string   | Tool name to call                                        | No                           |
| `toolArgVars`     | []string | Variables containing arguments for tool calls            | No                           |
| `resourceUri`     | string   | Resource URI to read                                     | No                           |
| `outputToContext` | bool     | Write result to flow context text                        | No (defaults to false)       |
| `resultVarName`   | string   | Variable name to store the result                        | No (defaults to "mcpResult") |

## Usage Examples

### Initialize an MCP Session

```yaml
- name: "initialize_mcp"
  executor:
    type: "mcp"
    withconfig:
      endpoint: "http://localhost:8080"
      transport: "http"
      outputToContext: true
```

### Read a Resource

```yaml
- name: "read_resource"
  executor:
    type: "mcp"
    withconfig:
      endpoint: "http://localhost:8080"
      transport: "http"
      resourceUri: "/resources/documents/${documentId}"
      outputToContext: true
      resultVarName: "document"
```

### Call a Tool

```yaml
- name: "call_tool"
  executor:
    type: "mcp"
    withconfig:
      endpoint: "http://localhost:8080"
      transport: "http"
      toolName: "summarize"
      toolArgVars: ["document"]
      outputToContext: true
      resultVarName: "summary"
```

## Variable Substitution

The MCP Executor supports variable substitution in the `resourceUri` parameter. You can include variables in the format `${varName}` which will be replaced with the corresponding value from the flow context variables.

For example:

```yaml
resourceUri: "/resources/documents/${documentId}"
```

If the flow context has a variable `documentId` with value "example-doc-1", the actual URI will be "/resources/documents/example-doc-1".

## Response Handling

Responses from MCP operations are stored in the flow context variable specified by `resultVarName`. If `outputToContext` is true, the response will also be stored in the flow context's `Text` field, making it available for subsequent steps.

For string responses, the text is stored directly. For structured responses, the data is converted to JSON format.

## Complete Example

A complete example workflow using the MCP Executor is available in `examples/mcp_executor_example.yaml`.

## Setting Up an MCP Server

To use the MCP Executor, you'll need access to an MCP-compliant server. You can:

1. Use Anthropic's Claude Desktop app, which includes built-in MCP server support
2. Run an open-source MCP server implementation from the [MCP GitHub repository](https://github.com/anthropics/model-context-protocol)
3. Implement your own MCP server using the protocol specification

Refer to the [Model Context Protocol documentation](https://modelcontextprotocol.io/) for more information on setting up and using MCP servers.
