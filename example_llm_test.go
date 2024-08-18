package anyi_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/chat"
	"github.com/jieliu2000/anyi/llm/openai"
)

func Example_lLMClient() {
	// For more documentation and examples, see github.com/jieliu2000/anyi/llm package documentation.
	// Make sure you set OPENAI_API_KEY environment variable to your OpenAI API key.
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _ := client.Chat(messages, nil)

	log.Printf("Response: %s\n", message.Content)
}
