package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/chat"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
)

func Example_openAI() {
	// Make sure you set OPENAI_API_KEY environment variable to your OpenAI API key.
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _ := client.Chat(messages)

	log.Printf("Response: %s\n", message.Content)
}
