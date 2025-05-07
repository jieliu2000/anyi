package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/anthropic"
	"github.com/jieliu2000/anyi/llm/chat"
)

func Example_anthropic() {
	// Make sure to set ANTHROPIC_API_KEY environment variable
	config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))

	// You can specify a specific model, such as Claude 3 Sonnet
	config.Model = "claude-3-sonnet-20240229"

	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Can you briefly introduce yourself?"},
	}
	message, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Response: %s\n", message.Content)
}
