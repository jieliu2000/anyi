package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/mcp"
)

func Example_mcpChat() {
	// Create a client using MCP
	config := mcp.DefaultConfigWithModel(os.Getenv("MCP_API_KEY"), "mcp-model")

	// Set custom endpoint if needed
	config.Endpoint = os.Getenv("MCP_ENDPOINT")
	if config.Endpoint == "" {
		config.Endpoint = "https://example.com/api/mcp"
	}

	client, err := anyi.NewClient("mcp", config)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}

	// Create messages for chat
	messages := []chat.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Explain quantum computing in simple terms"},
	}

	// Send request to MCP API
	response, info, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	// Process response
	log.Printf("Prompt tokens: %d", info.PromptTokens)
	log.Printf("Completion tokens: %d", info.CompletionTokens)
	log.Printf("Response: %s", response.Content)

	// Output:
	// This example doesn't produce deterministic output for testing
}
