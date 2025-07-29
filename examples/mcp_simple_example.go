// mcp_simple_example.go demonstrates how to use Anyi's MCP executors
// with a simple file system server that can read local files.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

func main() {
	// Create a temporary directory for our example
	tempDir, err := os.MkdirTemp("", "mcp_simple_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Create an example file in our temporary directory
	exampleFile := filepath.Join(tempDir, "hello.txt")
	content := "Hello, Model Context Protocol!"
	err = os.WriteFile(exampleFile, []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created example file at: %s\n", exampleFile)
	fmt.Printf("File content: %s\n\n", content)

	// Example 1: Using the generic MCPExecutor with STDIO transport
	fmt.Println("=== Example 1: Generic MCPExecutor ===")
	
	// Create an MCP executor that uses the filesystem server via STDIO
	executor := &anyi.MCPExecutor{
		Server: &anyi.MCPServerConfig{
			Name:    "filesystem-server",
			Type:    anyi.TransportSTDIO,      // Use STDIO transport
			Command: "npx",                    // Command to start the server
			Args:    []string{                 // Arguments for the command
				"-y", 
				"@modelcontextprotocol/server-filesystem", 
				tempDir,                       // Root directory for the filesystem server
			},
		},
		Action:        "read_resource",      // Action to perform
		Resource:      "file://" + exampleFile, // Resource to read
		ResultVarName: "fileContent",        // Variable name to store the result
	}

	// Initialize the executor
	fmt.Println("Initializing MCP executor...")
	if err := executor.Init(); err != nil {
		handleInitError(err)
		return
	}

	// Create flow context
	flowContext := flow.FlowContext{}

	// Run the executor
	fmt.Println("Reading file via MCP filesystem server...")
	result, err := executor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the result
	fileContent := result.GetVariable("fileContent")
	fmt.Printf("Success! File content: %s\n", fileContent)

	// Example 2: Using the specialized STDIOMCPExecutor
	fmt.Println("\n=== Example 2: Specialized STDIOMCPExecutor ===")
	
	// For STDIO-specific use cases, we can use the specialized executor
	stdioExecutor := &anyi.STDIOMCPExecutor{
		BaseMCPExecutor: anyi.BaseMCPExecutor{
			Action:        "read_resource",
			Resource:      "file://" + exampleFile,
			ResultVarName: "fileContent",
		},
		ServerConfig: &anyi.MCPServerConfig{
			Name:    "filesystem-server",
			Type:    anyi.TransportSTDIO,
			Command: "npx",
			Args: []string{
				"-y", 
				"@modelcontextprotocol/server-filesystem", 
				tempDir,
			},
		},
	}

	fmt.Println("Initializing STDIOMCPExecutor...")
	if err := stdioExecutor.Init(); err != nil {
		handleInitError(err)
		return
	}

	fmt.Println("Reading file via specialized STDIOMCPExecutor...")
	stdioResult, err := stdioExecutor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	stdioFileContent := stdioResult.GetVariable("fileContent")
	fmt.Printf("Success! File content: %s\n", stdioFileContent)

	fmt.Println("\nExample completed successfully!")
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