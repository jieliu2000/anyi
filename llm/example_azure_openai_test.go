package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/message"
)

func Example_azureOpenAI() {
	// Make sure you set the environment variables before you run this app
	config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []message.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _ := client.Chat(messages)

	log.Printf("Response: %s\n", message.Content)
}
