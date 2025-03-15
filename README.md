# Anyi - Open Source Autonomous AI Agent Framework 🤖

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://go.dev/)

| [English](README.md) | [中文](README-zh.md) |

Anyi is a powerful autonomous AI agent framework that helps you build AI solutions that seamlessly integrate with real-world workflows through unified LLM interfaces, robust validation mechanisms, and flexible workflow systems.

> 📚 **Looking for detailed tutorials?** Check out our comprehensive [Anyi Programming Guide and Examples](/docs/en/tutorial.md)

## ✨ Key Features

- **Universal LLM Access** - Connect to multiple LLM providers (OpenAI, Anthropic, etc.) through a consistent API
- **Powerful Workflow System** - Chain steps together with validation and automatic retries for robust AI pipelines
- **Configuration-Driven Development** - Define workflows and clients through code or external config files (YAML, JSON, TOML)
- **Multimodal Support** - Send text and images to compatible models simultaneously
- **Go Template Integration** - Use Go's template engine to generate dynamic prompts

## 🤔 When to Use Anyi

Anyi is ideal for:

- **AI Application Development** - Build production-ready AI services with reliable error handling and retries
- **Multi-LLM Applications** - Create solutions that leverage different models for different tasks based on cost or capability
- **DevOps Integration** - Connect AI capabilities with your existing systems through command executors and API integrations
- **Rapid Prototyping** - Configure complex AI workflows through config files without changing code
- **Enterprise Solutions** - Maintain separation of code and configuration for production deployment across environments

## 📋 Supported LLM Providers

- **OpenAI** - GPT series models
- **Azure OpenAI** - Microsoft-hosted OpenAI models
- **Anthropic** - Claude series models
- **Zhipu AI** - GLM series models
- **Alibaba Cloud Lingji** - Qwen series models
- **Ollama** - Local deployment of open-source models (Llama, Qwen, etc.)
- **DeepSeek** - DeepSeek models
- **SiliconCloud** - SiliconFlow models

## 🚀 Quick Start

### Installation

```bash
go get -u github.com/jieliu2000/anyi
```

> ⚠️ Requires Go 1.20 or higher

### Basic Usage

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"  // Import your preferred provider
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create client - just change imports and config to use different providers
	config := openai.DefaultConfig("gpt-4")
	config.APIKey = os.Getenv("OPENAI_API_KEY")
	
	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Send chat request
	messages := []chat.Message{
		{Role: "user", Content: "How many countries are in Africa?"},
	}
	
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	
	log.Printf("Response: %s", response.Content)
}
```

## 🔄 Creating Workflows

### Using Code

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
	config := openai.DefaultConfig("gpt-4")
	config.APIKey = os.Getenv("OPENAI_API_KEY")
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
	
	log.Printf("Title: %s", result.Text)
}
```

### Using Configuration Files

Anyi supports configuration-driven development, allowing you to define LLM clients and workflows in external files:

```yaml
# config.yaml
clients:
  - name: "gpt4"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"  # References environment variable
  
  - name: "claude"
    type: "anthropic"
    config:
      model: "claude-3-opus-20240229"
      apiKey: "$ANTHROPIC_API_KEY"

flows:
  - name: "story_flow"
    clientName: "gpt4"  # Default client for the flow
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
        clientName: "claude"  # Override client for this step
```

Load and use this configuration:

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

## 🛠️ Built-in Components

### Executors

- **LLMExecutor** - Sends prompts to large language models
- **SetContextExecutor** - Modifies workflow context
- **ConditionalFlowExecutor** - Branches based on conditions
- **RunCommandExecutor** - Executes system commands

### Validators

- **StringValidator** - Checks text via regex or equality
- **JsonValidator** - Ensures output is valid JSON

## 📖 Documentation

For comprehensive guides and detailed examples, check out our [Programming Guide](/docs/en/tutorial.md).

Topics covered include:
- [LLM Client Configuration](/docs/en/tutorial.md#client-configuration)
- [Workflow Creation](/docs/en/tutorial.md#workflow-system)
- [Using Configuration Files](/docs/en/tutorial.md#configuration-files)
- [Best Practices](/docs/en/tutorial.md#best-practices)

## 🤝 Contributing

Contributions welcome! Anyi is under active development, and your feedback helps make it better for everyone.

## 📄 License

Anyi is licensed under the [Apache License 2.0](LICENSE).
