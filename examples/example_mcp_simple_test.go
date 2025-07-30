package examples

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

// ExampleMCPExecutor_FileSystem demonstrates how to use Anyi's MCP executors
// with a simple file system server that can read local files.
//
// This example shows:
// 1. Using the generic MCPExecutor with STDIO transport
// 2. Using the specialized STDIOMCPExecutor
// 3. Reading a file through the MCP filesystem server
//
// Note: This example requires Node.js and may take some time on first run
// as it downloads the MCP filesystem server via npm.
// The test has been modified to handle timeouts gracefully.
func Example_mcpExecutor_FileSystem() {
	fmt.Println("=== MCP FileSystem Example ===")
	fmt.Println("Note: This example downloads MCP server on first run, which may take time.")
	fmt.Println()

	// Quick environment check
	fmt.Println("Checking environment...")

	// Check if this is a CI environment or should skip external dependencies
	if os.Getenv("CI") != "" || os.Getenv("SKIP_MCP_TESTS") != "" {
		fmt.Println("Skipping MCP external dependency test in CI/test environment")
		fmt.Println("To run this test locally, ensure Node.js is installed and run:")
		fmt.Println("go test -v -run ExampleMCPExecutor_FileSystem")
		return
	}
	// Create a temporary directory for our example
	tempDir, err := os.MkdirTemp("", "mcp_simple_example")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Create an example file in our temporary directory
	exampleFile := filepath.Join(tempDir, "hello.txt")
	content := "Hello, Model Context Protocol!"
	err = os.WriteFile(exampleFile, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Created example file at: %s\n", exampleFile)
	fmt.Printf("File content: %s\n\n", content)

	// Example 1: Using the generic MCPExecutor with STDIO transport
	fmt.Println("=== Example 1: Generic MCPExecutor ===")

	// Create an MCP executor that uses the filesystem server via STDIO
	executor := &anyi.MCPExecutor{
		BaseMCPExecutor: anyi.BaseMCPExecutor{
			Action:        "read_resource",         // Action to perform
			Resource:      "file://" + exampleFile, // Resource to read
			ResultVarName: "fileContent",           // Variable name to store the result
			Timeout:       15 * time.Second,        // Add timeout
		},
		Server: &anyi.MCPServerConfig{
			Name:    "filesystem-server",
			Type:    anyi.TransportSTDIO, // Use STDIO transport
			Command: "npx",               // Command to start the server
			Args: []string{ // Arguments for the command
				"-y",
				"@modelcontextprotocol/server-filesystem",
				tempDir, // Root directory for the filesystem server
			},
			Timeout: 10 * time.Second, // Server startup timeout
		},
	}

	// Initialize the executor with timeout
	fmt.Println("Initializing MCP executor...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Set up a channel to handle initialization in a goroutine
	initDone := make(chan error, 1)
	go func() {
		initDone <- executor.Init()
	}()

	select {
	case err := <-initDone:
		if err != nil {
			handleSimpleInitError(err)
			return
		}
	case <-ctx.Done():
		fmt.Println("Error: MCP executor initialization timed out after 20 seconds")
		fmt.Println("This might be due to:")
		fmt.Println("- Slow internet connection while downloading MCP server")
		fmt.Println("- npm configuration issues")
		fmt.Println("- Node.js environment problems")
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
			Timeout:       15 * time.Second,
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
			Timeout: 10 * time.Second,
		},
	}

	fmt.Println("Initializing STDIOMCPExecutor...")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel2()

	initDone2 := make(chan error, 1)
	go func() {
		initDone2 <- stdioExecutor.Init()
	}()

	select {
	case err := <-initDone2:
		if err != nil {
			handleSimpleInitError(err)
			return
		}
	case <-ctx2.Done():
		fmt.Println("Error: STDIOMCPExecutor initialization timed out after 20 seconds")
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

	// Output:
	// Created example file at: /tmp/mcp_simple_exampleXXXXXX/hello.txt
	// File content: Hello, Model Context Protocol!
	//
	// === Example 1: Generic MCPExecutor ===
	// Initializing MCP executor...
	// Reading file via MCP filesystem server...
	// Success! File content: Hello, Model Context Protocol!
	//
	// === Example 2: Specialized STDIOMCPExecutor ===
	// Initializing STDIOMCPExecutor...
	// Reading file via specialized STDIOMCPExecutor...
	// Success! File content: Hello, Model Context Protocol!
	//
	// Example completed successfully!
}

// handleSimpleInitError provides user-friendly error messages for simple example
func handleSimpleInitError(err error) {
	fmt.Printf("Initialization error: %v\n", err)

	switch {
	case err.Error() == "exec: \"npx\": executable file not found in $PATH":
		fmt.Println("\nSolution:")
		fmt.Println("- Install Node.js from https://nodejs.org/")
		fmt.Println("- Ensure 'npx' is in your PATH")
	case err.Error() == "context deadline exceeded":
		fmt.Println("\nThis error indicates a timeout during initialization.")
		fmt.Println("Possible causes:")
		fmt.Println("- Slow internet connection")
		fmt.Println("- npm registry issues")
		fmt.Println("- First-time package download taking too long")
		fmt.Println("\nSolutions:")
		fmt.Println("- Check your internet connection")
		fmt.Println("- Try running: npm config set registry https://registry.npmjs.org/")
		fmt.Println("- Pre-install the package: npm install -g @modelcontextprotocol/server-filesystem")
	default:
		fmt.Println("\nFor troubleshooting:")
		fmt.Println("- Ensure Node.js and npm are properly installed")
		fmt.Println("- Check that you can run: npx --version")
		fmt.Println("- Try installing the MCP server manually:")
		fmt.Println("  npm install -g @modelcontextprotocol/server-filesystem")
	}
}
