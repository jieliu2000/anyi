# Anyi - Open Source Autonomous AI Agent Framework

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://go.dev/)

| [English](README.md) | [中文](README-zh.md) |

Anyi is a powerful autonomous AI agent framework that helps you build AI solutions that seamlessly integrate with real-world workflows through unified LLM interfaces, robust validation mechanisms, and flexible workflow systems.

## Key Features

- **Universal LLM Access** - Connect to multiple LLM providers (OpenAI, Anthropic, etc.) through a consistent API
- **Powerful Workflow System** - Chain steps together with validation and automatic retries for robust AI pipelines
- **Configuration-Driven Development** - Define workflows and clients through code or external config files (YAML, JSON, TOML)
- **Multimodal Support** - Send text and images to compatible models simultaneously
- **Go Template Integration** - Use Go's template engine to generate dynamic prompts
- **Autonomous AI Agents** - Create intelligent agents that can plan and execute complex tasks

## When to Use Anyi

Anyi is ideal for:

- **AI Application Development** - Build production-ready AI services with reliable error handling and retries
- **Multi-LLM Applications** - Create solutions that leverage different models for different tasks based on cost or capability
- **DevOps Integration** - Connect AI capabilities with your existing systems through command executors and API integrations
- **Rapid Prototyping** - Configure complex AI workflows through config files without changing code
- **Enterprise Solutions** - Maintain separation of code and configuration for production deployment across environments
- **Autonomous Agents** - Build intelligent agents that can plan, execute, and adapt to complex tasks

## Supported LLM Providers

- **OpenAI** - GPT series models
- **Azure OpenAI** - Microsoft-hosted OpenAI models
- **Anthropic** - Claude series models
- **Zhipu AI** - GLM series models
- **Alibaba Cloud Lingji** - Qwen series models
- **Ollama** - Local deployment of open-source models (Llama, Qwen, etc.)
- **DeepSeek** - DeepSeek models
- **SiliconCloud** - SiliconFlow models

## Quick Start

### Installation

```bash
go get -u github.com/jieliu2000/anyi
```

> Requires Go 1.20 or higher

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

## Creating Workflows

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

```
# config.yaml
clients:
  - name: "gpt4"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY" # References environment variable

  - name: "ollama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434" # Default Ollama server address

flows:
  - name: "story_flow"
    clientName: "gpt4" # Default client for the flow
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
        clientName: "ollama" # Override client for this step
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

## Creating Autonomous Agents

Anyi now includes a powerful Agent framework that enables you to create autonomous AI agents capable of planning and executing complex tasks:

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

	// Create an autonomous agent
	agent, err := anyi.NewAgent(
		"researcher",                                    // Agent name (for registry)
		"Research Assistant",                           // Agent role
		"Expert at researching topics and writing reports", // Agent backstory
		[]string{"research_flow", "analyze_flow"},      // Available flows
		client,                                         // LLM client for planning
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Execute a complex task
	result, _, err := agent.Execute(
		"Research the impact of AI on healthcare and write a comprehensive report",
		agent.AgentContext{
			Variables: map[string]interface{}{
				"depth":   "detailed",
				"sources": 10,
				"format":  "markdown",
			},
		},
	)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	log.Printf("Research result: %s", result)
}
```

## Built-in Components

### Executors

- **LLMExecutor** - Sends prompts to large language models
- **SetContextExecutor** - Modifies workflow context
- **ConditionalFlowExecutor** - Branches based on conditions
- **RunCommandExecutor** - Executes system commands
- **MCPExecutor** - Interfaces with Model Control Protocol to access external models, resources, and tools

### Validators

- **StringValidator** - Checks text via regex or equality
- **JsonValidator** - Ensures output is valid JSON

## Documentation

### Getting Started

- [Installation and Setup](docs/en/getting-started/installation.md) - System requirements and installation
- [Quick Start Guide](docs/en/getting-started/quickstart.md) - Your first Anyi application
- [Basic Concepts](docs/en/getting-started/concepts.md) - Understanding clients, flows, and executors

### Tutorials

- [Working with LLM Clients](docs/en/tutorials/llm-clients.md) - Comprehensive guide to all supported providers
- [Building Workflows](docs/en/tutorials/workflows.md) - Creating complex AI workflows
- [Configuration Management](docs/en/tutorials/configuration.md) - Using config files and environment variables
- [Multimodal Applications](docs/en/tutorials/multimodal.md) - Working with text and images
- [Building Autonomous Agents](docs/en/tutorials/agents.md) - Creating intelligent autonomous agents

### How-To Guides

- [Provider Setup](docs/en/how-to/provider-setup.md) - Detailed setup for each LLM provider
- [Error Handling](docs/en/how-to/error-handling.md) - Best practices for robust applications
- [Performance Optimization](docs/en/how-to/performance.md) - Speed and cost optimization
- [Web Integration](docs/en/how-to/web-integration.md) - Using Anyi with web frameworks

### Reference

- [API Reference](docs/en/reference/api.md) - Complete API documentation
- [Configuration Schema](docs/en/reference/configuration.md) - All configuration options
- [Built-in Components](docs/en/reference/components.md) - Executors and validators reference

### Advanced Topics

- [Custom Executors](docs/en/advanced/custom-executors.md) - Building your own executors
- [Security Best Practices](docs/en/advanced/security.md) - Securing your AI applications
- [Production Deployment](docs/en/advanced/deployment.md) - Production considerations

## Contributing

Contributions welcome! Anyi is under active development, and your feedback helps make it better for everyone.

## License

Anyi is licensed under the [Apache License 2.0](LICENSE).