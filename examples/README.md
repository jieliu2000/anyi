# Anyi MCP Examples

This directory contains examples demonstrating how to use Anyi's MCP (Model Context Protocol) executors.

## Prerequisites

To run these examples, you need:

1. Go 1.20 or higher
2. Node.js and npm (for the MCP server examples)

## Examples

### mcp_simple_example.go

A simple example showing how to use Anyi's MCP executors with a filesystem server.

The example demonstrates:
- Using the generic `MCPExecutor` with STDIO transport
- Using the specialized `STDIOMCPExecutor`
- Reading a file through the MCP filesystem server

To run this example:

```bash
go run mcp_simple_example.go
```

This will:
1. Create a temporary directory with an example file
2. Start an MCP filesystem server using STDIO transport
3. Read the file content through the MCP protocol
4. Display the results

The example uses the [@modelcontextprotocol/server-filesystem](https://github.com/modelcontextprotocol/servers) package, which is automatically downloaded and installed when you run the example.

### mcp_tools_example.go

An example showing how to call tools provided by MCP servers using the memory server.

The example demonstrates:
- Calling MCP tools with arguments using `call_tool` action
- Reading data created by MCP tools
- Listing available tools from an MCP server with `list_tools` action

To run this example:

```bash
go run mcp_tools_example.go
```

This uses the [@modelcontextprotocol/server-memory](https://github.com/modelcontextprotocol/servers) package which provides tools for working with in-memory data.

### mcp_filesystem_example.go

An extended example showing how to use Anyi's MCP executors with a filesystem server, demonstrating both generic and specialized executors.

To run this example:

```bash
go run mcp_filesystem_example.go
```

### Example test files (Go documentation format)

In addition to the runnable examples, we also provide example tests in the project root directory which demonstrate the same functionality but in the form of Go example tests. These follow Go's convention for example functions and can be run with:

```bash
go test -run Example
```

These examples show:
- How to use `MCPExecutor` with filesystem servers
- How to use specialized `STDIOMCPExecutor` for STDIO transport
- How to call tools provided by MCP servers

The test files are:
- `mcp_simple_example_test.go` - Simple filesystem example
- `mcp_tools_example_test.go` - Tools calling example
- `mcp_filesystem_example_test.go` - Extended filesystem example

## Understanding the MCP Executors

Anyi provides several MCP executors for different use cases:

1. **MCPExecutor** - Generic executor that supports all transport types (HTTP, SSE, STDIO)
2. **HTTPMCPExecutor** - Specialized for HTTP transport
3. **SSEMCPExecutor** - Specialized for Server-Sent Events transport
4. **STDIOMCPExecutor** - Specialized for STDIO transport

The specialized executors provide a cleaner API when you know which transport type you'll be using, while the generic executor offers flexibility.

## Troubleshooting

If you encounter errors when running the examples:

1. **"npx: command not found"** - Make sure Node.js is installed from [nodejs.org](https://nodejs.org/)
2. **Network errors** - The examples need to download the MCP server packages from npm, so make sure you have internet connectivity
3. **Permission errors** - On some systems, you might need to adjust permissions for the temporary directories