package llm_test

import (
	"log"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/message"
)

func Example_ollama() {
	// Make sure you start ollama and pulled your target model first.
	config := ollama.DefaultConfig("mistral")
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []message.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, err := client.Chat(messages)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
	// Output: Response: 6

}
