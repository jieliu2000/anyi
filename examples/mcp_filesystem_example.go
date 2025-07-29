package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

// ExampleMCPExecutor_ExtendedFileSystem demonstrates how to use Anyi's MCP executors
// with a filesystem server in a more extended example.
//
// This example shows:
// 1. Creating and using a filesystem MCP server
// 2. Reading file content via MCP protocol
// 3. Using both generic and specialized MCP executors
func ExampleMCPExecutor_ExtendedFileSystem() {
	// Create a temporary directory for our example
	tempDir, err := os.MkdirTemp("", "mcp_example")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Create an example file in our temporary directory
	exampleFile := filepath.Join(tempDir, "example.txt")
	content := "Hello, MCP! This is an example file for demonstrating Anyi's MCP functionality."
	if err := os.WriteFile(exampleFile, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Created example file at: %s\n", exampleFile)
	fmt.Printf("File content: %s\n\n", content)

	// Create an MCP executor that uses the filesystem server via STDIO
	executor := &anyi.MCPExecutor{
		Server: &anyi.MCPServerConfig{
			Name:    "filesystem",
			Type:    anyi.TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", tempDir},
		},
		Action:        "read_resource",
		Resource:      "file://" + exampleFile,
		ResultVarName: "fileContent",
	}

	// Initialize the executor
	if err := executor.Init(); err != nil {
		// If npx command is not available, provide a friendly message
		if err.Error() == "exec: \"npx\": executable file not found in $PATH" {
			fmt.Println("Example requires Node.js to be installed")
			return
		}
		fmt.Printf("Error initializing executor: %v\n", err)
		return
	}

	// Create flow context
	flowContext := flow.FlowContext{}

	// Run the MCP executor
	fmt.Println("Reading file content via MCP filesystem server...")
	result, err := executor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("Error running MCP executor: %v\n", err)
		return
	}

	// Print the result
	fmt.Println("MCP execution completed successfully!")
	fileContent := result.GetVariable("fileContent")
	if fileContent != nil {
		fmt.Printf("File content retrieved via MCP: %s\n", fileContent)
	}

	// Demonstrate how to use the specialized STDIO MCP executor
	fmt.Println("\n--- Using specialized STDIOMCPExecutor ---")
	stdioExecutor := &anyi.STDIOMCPExecutor{
		BaseMCPExecutor: anyi.BaseMCPExecutor{
			Action:        "read_resource",
			Resource:      "file://" + exampleFile,
			ResultVarName: "fileContent",
		},
		ServerConfig: &anyi.MCPServerConfig{
			Name:    "filesystem",
			Type:    anyi.TransportSTDIO,
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", tempDir},
		},
	}

	if err := stdioExecutor.Init(); err != nil {
		// If npx command is not available, provide a friendly message
		if err.Error() == "exec: \"npx\": executable file not found in $PATH" {
			fmt.Println("Example requires Node.js to be installed")
			return
		}
		fmt.Printf("Error initializing executor: %v\n", err)
		return
	}

	stdioResult, err := stdioExecutor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("Error running STDIOMCPExecutor: %v\n", err)
		return
	}

	fmt.Println("STDIOMCPExecutor execution completed successfully!")
	stdioFileContent := stdioResult.GetVariable("fileContent")
	if stdioFileContent != nil {
		fmt.Printf("File content retrieved via STDIOMCPExecutor: %s\n", stdioFileContent)
	}

	fmt.Println("\nExample completed!")

	// Output:
	// Created example file at: /tmp/mcp_exampleXXXXXX/example.txt
	// File content: Hello, MCP! This is an example file for demonstrating Anyi's MCP functionality.
	//
	// Reading file content via MCP filesystem server...
	// MCP execution completed successfully!
	// File content retrieved via MCP: Hello, MCP! This is an example file for demonstrating Anyi's MCP functionality.
	//
	// --- Using specialized STDIOMCPExecutor ---
	// STDIOMCPExecutor execution completed successfully!
	// File content retrieved via STDIOMCPExecutor: Hello, MCP! This is an example file for demonstrating Anyi's MCP functionality.
	//
	// Example completed!
}