package examples

import (
	"fmt"
	"testing"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

// ExampleMCPExecutor_Tools demonstrates how to use Anyi's MCP executors to call tools
// using the memory MCP server which provides tools for working with in-memory data.
//
// This example shows:
// 1. Calling MCP tools with arguments using 'call_tool' action
// 2. Reading data created by MCP tools
// 3. Listing available tools from an MCP server with 'list_tools' action
func ExampleMCPExecutor_Tools() {
	fmt.Println("=== MCP Tools Example ===")
	fmt.Println("This example demonstrates calling tools provided by an MCP server.")
	fmt.Println()

	// Example 1: Using the memory MCP server to call tools
	fmt.Println("1. Using MCPExecutor with memory server to call tools:")

	// Create an MCP executor that uses the memory server via STDIO
	// The memory server provides tools for working with in-memory data
	executor := &anyi.MCPExecutor{
		Server: &anyi.MCPServerConfig{
			Name:    "memory-server",
			Type:    anyi.TransportSTDIO, // Use STDIO transport
			Command: "npx",               // Command to start the server
			Args: []string{ // Arguments for the command
				"-y",
				"@modelcontextprotocol/server-memory",
			},
		},
		Action:        "call_tool",          // Action to perform - calling a tool
		ToolName:      "create_memory_item", // Tool to call
		ToolArgs: map[string]interface{}{    // Arguments for the tool
			"key":   "example_key",
			"value": "Hello, MCP Tools!",
		},
		ResultVarName: "toolResult", // Variable name to store the result
	}

	// Initialize the executor
	fmt.Println("   Initializing MCP executor...")
	if err := executor.Init(); err != nil {
		handleInitError(err)
		return
	}

	// Create flow context
	flowContext := flow.FlowContext{}

	// Run the executor to call the tool
	fmt.Println("   Calling 'create_memory_item' tool...")
	result, err := executor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("   Error calling tool: %v\n", err)
		return
	}

	// Print the result
	toolResult := result.GetVariable("toolResult")
	fmt.Printf("   Tool call result: %+v\n", toolResult)

	// Example 2: Reading the memory item we just created
	fmt.Println("\n2. Reading the memory item we created:")

	readExecutor := &anyi.MCPExecutor{
		Server: &anyi.MCPServerConfig{
			Name:    "memory-server",
			Type:    anyi.TransportSTDIO,
			Command: "npx",
			Args: []string{
				"-y",
				"@modelcontextprotocol/server-memory",
			},
		},
		Action:   "call_tool",
		ToolName: "read_memory_item",
		ToolArgs: map[string]interface{}{
			"key": "example_key",
		},
		ResultVarName: "readResult",
	}

	if err := readExecutor.Init(); err != nil {
		handleInitError(err)
		return
	}

	fmt.Println("   Calling 'read_memory_item' tool...")
	readResult, err := readExecutor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("   Error calling tool: %v\n", err)
		return
	}

	readResultValue := readResult.GetVariable("readResult")
	fmt.Printf("   Read result: %+v\n", readResultValue)

	// Example 3: Listing available tools
	fmt.Println("\n3. Listing available tools from the memory server:")

	listToolsExecutor := &anyi.MCPExecutor{
		Server: &anyi.MCPServerConfig{
			Name:    "memory-server",
			Type:    anyi.TransportSTDIO,
			Command: "npx",
			Args: []string{
				"-y",
				"@modelcontextprotocol/server-memory",
			},
		},
		Action:        "list_tools", // Action to list tools
		ResultVarName: "toolsList",
	}

	if err := listToolsExecutor.Init(); err != nil {
		handleInitError(err)
		return
	}

	fmt.Println("   Listing available tools...")
	listResult, err := listToolsExecutor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("   Error listing tools: %v\n", err)
		return
	}

	toolsList := listResult.GetVariable("toolsList")
	fmt.Printf("   Available tools: %+v\n", toolsList)

	fmt.Println("\nExample completed successfully!")
	fmt.Println("\nThis example demonstrated:")
	fmt.Println(" - Calling MCP tools with arguments")
	fmt.Println(" - Reading data created by MCP tools")
	fmt.Println(" - Listing available tools from an MCP server")

	// Output:
	// === MCP Tools Example ===
	// This example demonstrates calling tools provided by an MCP server.
	//
	// 1. Using MCPExecutor with memory server to call tools:
	//    Initializing MCP executor...
	//    Calling 'create_memory_item' tool...
	//    Tool call result: map[created:true]
	//
	// 2. Reading the memory item we created:
	//    Calling 'read_memory_item' tool...
	//    Read result: map[content:Hello, MCP Tools! key:example_key]
	//
	// 3. Listing available tools from the memory server:
	//    Listing available tools...
	//    Available tools: [{create_memory_item Create a new item in memory [] map[key:The key for the memory item value:The value to store in memory]} {read_memory_item Read an item from memory [] map[key:The key of the item to read]} {list_memory_items List all items in memory [] map[]}]
	//
	// Example completed successfully!
	//
	// This example demonstrated:
	//  - Calling MCP tools with arguments
	//  - Reading data created by MCP tools
	//  - Listing available tools from an MCP server
}

// handleInitError provides user-friendly error messages
func handleInitError(err error) {
	switch err.Error() {
	case "exec: \"npx\": executable file not found in $PATH":
		fmt.Println("Error: npx command not found.")
		fmt.Println("Please ensure Node.js is installed (https://nodejs.org/)")
	default:
		fmt.Printf("Error initializing executor: %v\n", err)
	}
}