# Quick Start Guide

This guide will help you create your first Anyi application in just a few minutes. By the end of this tutorial, you'll have a working AI application that can chat with users.

## Prerequisites

Before starting, make sure you have:

- [Anyi installed](installation.md)
- An API key for at least one LLM provider (we'll use OpenAI in this example)

## Your First Anyi Application

### Step 1: Create a New Project

Create a new directory for your project:

```bash
mkdir my-anyi-app
cd my-anyi-app
go mod init my-anyi-app
```

### Step 2: Install Anyi

Add Anyi to your project:

```bash
go get -u github.com/jieliu2000/anyi
```

### Step 3: Set Up Environment Variables

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
```

### Step 4: Create Your First Chat Application

Create a file called `main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 1. Create a client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 2. Send a simple request
	messages := []chat.Message{
		{Role: "user", Content: "Please briefly explain quantum computing"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	fmt.Println("Response:", response.Content)
}
```

### Step 5: Run Your Application

```bash
go run main.go
```

You should see a response from the AI explaining quantum computing!

## Understanding What Happened

Let's break down the code:

1. **Client Creation**: We created an OpenAI client with default configuration
2. **Message Structure**: We created a message array with a user message
3. **Chat Request**: We sent the messages to the AI and received a response
4. **Response Handling**: We printed the AI's response

## Your First Workflow

Now let's create a more complex example using Anyi's workflow system:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
)

func main() {
	// Create client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create a two-step workflow
	step1, _ := anyi.NewLLMStepWithTemplate(
		"Generate a short story about {{.Text}}",
		"You are a creative fiction writer.",
		client,
	)
	step1.Name = "story_generation"

	step2, _ := anyi.NewLLMStepWithTemplate(
		"Create an engaging title for the following story:\n\n{{.Text}}",
		"You are an editor skilled at creating titles.",
		client,
	)
	step2.Name = "title_creation"

	// Create and register the flow
	myFlow, _ := anyi.NewFlow("story_flow", client, *step1, *step2)
	anyi.RegisterFlow("story_flow", myFlow)

	// Run the workflow
	result, _ := myFlow.RunWithInput("a detective in future London")

	log.Printf("Generated title: %s", result.Text)
}
```

This workflow:

1. Generates a short story based on your input
2. Creates a title for that story
3. Returns the final title

## Using Configuration Files

For more complex applications, you can use configuration files. Create a `config.yaml`:

```yaml
clients:
  - name: "gpt4"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"

flows:
  - name: "story_flow"
    clientName: "gpt4"
    steps:
      - name: "story_generation"
        executor:
          type: "llm"
          withconfig:
            template: "Generate a short story about {{.Text}}"
            systemMessage: "You are a creative fiction writer."
        maxRetryTimes: 2

      - name: "title_creation"
        executor:
          type: "llm"
          withconfig:
            template: "Create an engaging title for the following story:\n\n{{.Text}}"
            systemMessage: "You are an editor skilled at creating titles."
```

Then load and use it:

```go
package main

import (
	"log"
	"github.com/jieliu2000/anyi"
)

func main() {
	// Load configuration from file
	err := anyi.ConfigFromFile("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get and run the configured flow
	flow, err := anyi.GetFlow("story_flow")
	if err != nil {
		log.Fatalf("Failed to get flow: %v", err)
	}

	result, err := flow.RunWithInput("a detective in future London")
	if err != nil {
		log.Fatalf("Flow execution failed: %v", err)
	}

	log.Printf("Result: %s", result.Text)
}
```

## Common Patterns

### Error Handling

Always handle errors appropriately:

```go
response, info, err := client.Chat(messages, nil)
if err != nil {
    log.Printf("Chat failed: %v", err)
    return
}

log.Printf("Used %d tokens", info.PromptTokens + info.CompletionTokens)
```

### Multiple Providers

You can use different providers for different tasks:

```go
// Fast, cheap model for simple tasks
quickConfig := openai.NewConfigWithModel(apiKey, "gpt-3.5-turbo")
quickClient, _ := anyi.NewClient("quick", quickConfig)

// Powerful model for complex tasks
powerConfig := openai.NewConfigWithModel(apiKey, "gpt-4")
powerClient, _ := anyi.NewClient("power", powerConfig)
```

### Message History

Maintain conversation context:

```go
messages := []chat.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "What is the capital of France?"},
}

response, _, err := client.Chat(messages, nil)
if err != nil {
    log.Fatal(err)
}

// Add response to history
messages = append(messages, *response)

// Continue conversation
messages = append(messages, chat.Message{
    Role: "user",
    Content: "What's the population of that city?",
})

response, _, err = client.Chat(messages, nil)
```

## Next Steps

Now that you have a basic understanding of Anyi:

1. **Learn the Concepts**: Read about [Basic Concepts](concepts.md) to understand Anyi's architecture
2. **Explore Providers**: Check out the [LLM Clients Tutorial](../tutorials/llm-clients.md) to learn about different AI providers
3. **Build Workflows**: Learn to create complex workflows in the [Workflows Tutorial](../tutorials/workflows.md)
4. **Configuration**: Master configuration management in the [Configuration Tutorial](../tutorials/configuration.md)

## Troubleshooting

### Common Issues

**"Failed to create client" error**

- Check that your API key is correctly set
- Verify the API key has the necessary permissions

**"Request failed" error**

- Check your internet connection
- Verify the API service is available
- Ensure you have sufficient API credits

**Import errors**

- Run `go mod tidy` to resolve dependencies
- Check that you're using Go 1.20 or higher

### Getting Help

If you need help:

- Check the [FAQ](../reference/faq.md)
- Browse the [examples](../../../examples/) directory
- Ask questions on [GitHub Discussions](https://github.com/jieliu2000/anyi/discussions)
