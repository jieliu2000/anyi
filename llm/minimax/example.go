package minimax

import (
	"fmt"
	"os"

	"github.com/jieliu2000/anyi/llm/chat"
)

// ExampleMinimaxClient demonstrates how to use the MiniMax client
func ExampleMinimaxClient() {
	// Get API key from environment variable
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set MINIMAX_API_KEY environment variable")
		os.Exit(1)
	}

	// Create MiniMax client
	client, err := NewClient(&MiniMaxModelConfig{
		APIKey:  apiKey,
		Model:   MiniMaxM2,
		BaseUrl: DefaultBaseUrl,
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Send chat request
	messages := []chat.Message{
		{Role: "user", Content: "Please introduce the main features of MiniMax-M2 model."},
	}

	chatOptions := &chat.ChatOptions{}

	// Send request
	chatResponse, _, err := client.Chat(messages, chatOptions)
	if err != nil {
		fmt.Printf("Chat error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Using model: %s\n", MiniMaxM2)
	fmt.Printf("Response: %s\n", chatResponse.Content)

	// Use stable model for second query
	client2, err := NewClient(&MiniMaxModelConfig{
		APIKey:  apiKey,
		Model:   MiniMaxM2Stable,
		BaseUrl: DefaultBaseUrl,
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Send another request
	messages2 := []chat.Message{
		{Role: "user", Content: "What are the application scenarios of large language models in code generation tasks?"},
	}

	chat2Response, _, err := client2.Chat(messages2, chatOptions)
	if err != nil {
		fmt.Printf("Chat error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("\nUsing model: %s\n", MiniMaxM2Stable)
	fmt.Printf("Response: %s\n", chat2Response.Content)
}