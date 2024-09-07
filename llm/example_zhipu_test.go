package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/zhipu"
)

func Example_zhipu() {
	// Make sure you set ZHIPU_API_KEY environment variable to your Zhipu API key.
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Perform one task based on the following objective: Use python to create an AI digital employee project which can generate code for Quasar hybrid mobile app based on user input requirements.\nTake into account these previously completed tasks: \n\n\nYour task: {Create a new project in python}\nResponse:"},
	}
	message, _, err := client.Chat(messages, nil)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
}
