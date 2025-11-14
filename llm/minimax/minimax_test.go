package minimax

import (
	"fmt"
	"os"
	"testing"

	"github.com/jieliu2000/anyi/llm/chat"
)

func TestChat(t *testing.T) {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		t.Skip("MINIMAX_API_KEY is not set")
	}

	// Test MiniMaxM2 model
	client, err := NewClient(&MiniMaxModelConfig{
		APIKey:  apiKey,
		Model:   MiniMaxM2,
		BaseUrl: DefaultBaseUrl,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Hello, MiniMax-M2!"},
	}

	chatOptions := &chat.ChatOptions{}
	response, _, err := client.Chat(messages, chatOptions)
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}

	fmt.Printf("Response from model %s: %s\n", MiniMaxM2, response.Content)
}

func TestChatWithMiniMaxM2Stable(t *testing.T) {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		t.Skip("MINIMAX_API_KEY is not set")
	}

	// Test MiniMaxM2Stable model
	client, err := NewClient(&MiniMaxModelConfig{
		APIKey:  apiKey,
		Model:   MiniMaxM2Stable,
		BaseUrl: DefaultBaseUrl,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Hello, MiniMax-M2-Stable!"},
	}

	chatOptions := &chat.ChatOptions{}

	// Send request
	response, _, err := client.Chat(messages, chatOptions)
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}

	// Print results
	fmt.Printf("Model: %s\n", MiniMaxM2Stable)
	fmt.Printf("Response: %s\n", response.Content)
}

func TestMiniMaxConfig(t *testing.T) {
	// Test creating default config
	apiKey := "test-api-key"
	config := DefaultConfig(apiKey, MiniMaxM2)

	// Verify default config
	if config.APIKey != apiKey {
		t.Errorf("Expected APIKey to be %s, got %s", apiKey, config.APIKey)
	}
	if config.Model != MiniMaxM2 {
		t.Errorf("Expected Model to be %s, got %s", MiniMaxM2, config.Model)
	}
	if config.BaseUrl != DefaultBaseUrl {
		t.Errorf("Expected BaseUrl to be %s, got %s", DefaultBaseUrl, config.BaseUrl)
	}

	// Test empty model parameter, using default value
	config2 := DefaultConfig(apiKey, "")
	if config2.Model != DefaultModel {
		t.Errorf("Expected Model to be %s, got %s", DefaultModel, config2.Model)
	}
}