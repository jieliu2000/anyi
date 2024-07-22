package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/message"
)

func Example_dashscope() {
	// Make sure you set OPENAI_API_KEY environment variable to your OpenAI API key.
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")
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
}
