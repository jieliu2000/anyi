package llm_test

import (
	"log"

	"github.com/jieliu2000/anyi/chat"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/ollama"
)

func Example_ollamaWithImage() {
	// Make sure you start ollama and pulled your target model first.
	config := ollama.DefaultConfig("moondream")
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		chat.NewImageMessageFromFile("user", "What number is in the image?", "../internal/test/number_six.png"),
	}
	message, err := client.Chat(messages)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
}
