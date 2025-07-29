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
	// 创建一个临时目录用于示例
	tempDir, err := os.MkdirTemp("", "mcp_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 在临时目录中创建一个示例文件
	exampleFile := filepath.Join(tempDir, "example.txt")
	content := "Hello, MCP! This is an example file for demonstrating Anyi's MCP functionality."
	if err := os.WriteFile(exampleFile, []byte(content), 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created example file at: %s\n", exampleFile)
	fmt.Printf("File content: %s\n\n", content)

	// 创建一个使用STDIO传输的MCP执行器
	// 这里使用文件系统MCP服务器，它可以读取和操作本地文件
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

	// 初始化执行器
	if err := executor.Init(); err != nil {
		// 如果npx命令不可用或包未找到，给出友好的提示
		if err.Error() == "exec: \"npx\": executable file not found in $PATH" {
			fmt.Println("Error: npx command not found. Please ensure Node.js is installed.")
			fmt.Println("You can download Node.js from: https://nodejs.org/")
			return
		}
		log.Fatal("Failed to initialize MCP executor:", err)
	}

	// 创建一个简单的流程上下文
	flowContext := flow.FlowContext{}

	// 运行MCP执行器
	fmt.Println("Reading file content via MCP filesystem server...")
	result, err := executor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("Error running MCP executor: %v\n", err)
		fmt.Println("This might be because the MCP server package is not available or there was a network issue.")
		fmt.Println("Please ensure you have internet connection and Node.js installed.")
		return
	}

	// 输出结果
	fmt.Println("MCP execution completed successfully!")
	fileContent := result.GetVariable("fileContent")
	if fileContent != nil {
		fmt.Printf("File content retrieved via MCP: %+v\n", fileContent)
	}

	// 展示如何使用专门的STDIO MCP执行器
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
		// 如果npx命令不可用或包未找到，给出友好的提示
		if err.Error() == "exec: \"npx\": executable file not found in $PATH" {
			fmt.Println("Error: npx command not found. Please ensure Node.js is installed.")
			fmt.Println("You can download Node.js from: https://nodejs.org/")
			return
		}
		log.Fatal("Failed to initialize STDIOMCPExecutor:", err)
	}

	stdioResult, err := stdioExecutor.Run(flowContext, nil)
	if err != nil {
		fmt.Printf("Error running STDIOMCPExecutor: %v\n", err)
		return
	}

	fmt.Println("STDIOMCPExecutor execution completed successfully!")
	stdioFileContent := stdioResult.GetVariable("fileContent")
	if stdioFileContent != nil {
		fmt.Printf("File content retrieved via STDIOMCPExecutor: %+v\n", stdioFileContent)
	}

	fmt.Println("\nExample completed!")
}