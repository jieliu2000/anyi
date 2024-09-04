package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

func Example_azureOpenAIWithTools() {
	// Make sure you set the environment variables before you run this app
	config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		chat.NewUserMessage("What is the weather today in Paris?"),
	}

	funcConfig := tools.NewFunctionConfig("get_current_weather", "Get the current weather for a location")
	funcConfig.AddSimpleParam("location", "string", "The location to get the weather for")

	functions := []tools.FunctionConfig{
		funcConfig,
	}
	message, _, err := client.ChatWithFunctions(messages, functions, nil)

	log.Printf("Response: %s\n", message.Content)
}
