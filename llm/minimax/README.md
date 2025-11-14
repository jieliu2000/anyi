# MiniMax Client

MiniMax Client is a Go language implementation of an OpenAI-compatible API client specifically designed for accessing the MiniMax API.

Based on the official API documentation: https://platform.minimaxi.com/docs/api-reference/text-openai-api

## Features

- Supports MiniMax-M2 and MiniMax-M2-Stable models
- OpenAI-style API interface compatibility
- Provides Chat and ChatWithFunctions methods
- Supports custom parameters and configuration

## Supported Models

- `MiniMax-M2`: Standard version large language model
- `MiniMax-M2-Stable`: Stable version large language model
- `DefaultModel`: Uses `MiniMax-M2` model by default

## API Base URL

- `DefaultBaseUrl`: `https://api.minimaxi.com/v1`

## Usage Example

```go
package main

import (
	"fmt"
	"os"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/minimax"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set MINIMAX_API_KEY environment variable")
		os.Exit(1)
	}

	// Create MiniMax client
	client, err := minimax.NewClient(&minimax.MiniMaxModelConfig{
		APIKey:  apiKey,
		Model:   minimax.MiniMaxM2,
		BaseUrl: minimax.DefaultBaseUrl,
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Send chat request
	messages := []chat.Message{
		{Role: "user", Content: "Please introduce the main features of MiniMax-M2 model."},
	}

	// Send request
	chatResponse, _, err := client.Chat(messages, &chat.ChatOptions{})
	if err != nil {
		fmt.Printf("Chat error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Using model: %s\n", minimax.MiniMaxM2)
	fmt.Printf("Response: %s\n", chatResponse.Content)
}
```

## Environment Variables

- `MINIMAX_API_KEY` - MiniMax API key

## Testing

Run MiniMax package tests:

```bash
go test ./llm/minimax -v
```

Run all project tests:

```bash
go test ./... -v
```

## Dependencies

This implementation depends on the following packages:

- `github.com/jieliu2000/anyi/llm/chat` - Chat message and option types
- `github.com/jieliu2000/anyi/llm/config` - General configuration
- `github.com/jieliu2000/anyi/llm/tools` - Tool functionality
- `github.com/jieliu2000/anyi/llm/openai/utils` - OpenAI-compatible utility functions