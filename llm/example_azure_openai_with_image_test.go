package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func Example_azureOpenAIWithImage() {
	// Make sure you set the environment variables before you run this app
	config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		chat.NewImageMessageFromFile("user", "What number is in the image?", "../internal/test/number_six.png"),
	}
	message, _ := client.Chat(messages, nil)

	log.Printf("Response: %s\n", message.Content)
}
