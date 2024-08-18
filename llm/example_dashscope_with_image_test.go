package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/chat"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/dashscope"
)

func Example_dashscopeWithImage() {
	// Make sure you set DASHSCOPE_API_KEY environment variable to your Dashscope API key.
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-vl-plus")
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		chat.NewImageMessageFromUrl("user", "What's this?", "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"),
	}
	message, err := client.Chat(messages)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
}
