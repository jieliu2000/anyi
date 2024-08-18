package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/chat"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/dashscope"
)

func Example_dashscope() {
	// Make sure you set DASHSCOPE_API_KEY environment variable to your Dashscope API key.
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, err := client.Chat(messages, nil)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)

}
