package main

import (
	"fmt"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Initialize logging
	log.SetLevel(log.InfoLevel)

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Create OpenAI client
	openaiConfig := openai.NewConfig(apiKey, "gpt-3.5-turbo", "")
	client, err := llm.NewClient(openaiConfig)
	if err != nil {
		log.Fatalf("Failed to create OpenAI client: %v", err)
	}

	// Register client
	anyi.RegisterClient("default", client)

	// Create a new MCPExecutor for session initialization
	initExecutor := &anyi.MCPExecutor{
		Endpoint:        "http://localhost:8080", // Update with your MCP server address
		Transport:       "http",
		OutputToContext: true,
	}
	if err := initExecutor.Init(); err != nil {
		log.Fatalf("Failed to initialize MCP executor: %v", err)
	}

	// Create an MCPExecutor for resource reading
	resourceExecutor := &anyi.MCPExecutor{
		Endpoint:        "http://localhost:8080", // Update with your MCP server address
		Transport:       "http",
		ResourceURI:     "/resources/documents/example-doc",
		OutputToContext: true,
		ResultVarName:   "document",
	}
	if err := resourceExecutor.Init(); err != nil {
		log.Fatalf("Failed to initialize resource executor: %v", err)
	}

	// Create an MCPExecutor for tool calling
	toolExecutor := &anyi.MCPExecutor{
		Endpoint:        "http://localhost:8080", // Update with your MCP server address
		Transport:       "http",
		ToolName:        "summarize",
		ToolArgVars:     []string{"document"},
		OutputToContext: true,
		ResultVarName:   "summary",
	}
	if err := toolExecutor.Init(); err != nil {
		log.Fatalf("Failed to initialize tool executor: %v", err)
	}

	// Create an LLMExecutor for analysis
	analysisExecutor := &anyi.LLMExecutor{
		Template: `
You are analyzing data retrieved from an MCP server.

The document content is:
${document}

The summary is:
${summary}

Please provide your insights on this information:
`,
	}
	if err := analysisExecutor.Init(); err != nil {
		log.Fatalf("Failed to initialize LLM executor: %v", err)
	}

	// Create steps
	step1 := flow.NewStep(initExecutor, nil, client)
	step2 := flow.NewStep(resourceExecutor, nil, client)
	step3 := flow.NewStep(toolExecutor, nil, client)
	step4 := flow.NewStep(analysisExecutor, nil, client)

	// Create flow
	flow, err := flow.NewFlow(client, "mcp_example_flow", *step1, *step2, *step3, *step4)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	// Initialize variables
	variables := map[string]any{
		"documentId": "example-doc-1",
	}

	// Run flow
	result, err := flow.RunWithVariables(variables)
	if err != nil {
		log.Fatalf("Failed to run flow: %v", err)
	}

	// Print result
	fmt.Println("\n=== MCP Workflow Result ===")
	fmt.Println(result.Text)
	fmt.Println("\n=== Document from MCP ===")
	fmt.Println(result.GetVariableString("document", "Not found"))
	fmt.Println("\n=== Summary from MCP Tool ===")
	fmt.Println(result.GetVariableString("summary", "Not found"))
}
