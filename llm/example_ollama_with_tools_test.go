package llm_test

import (
	"log"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/tools"
)

func Example_ollamaWithTools() {
	// Make sure you start ollama and pulled your target model first.
	config := ollama.DefaultConfig("mistral")
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

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
	// Output: Response: It's cloudy today in Paris.
}
