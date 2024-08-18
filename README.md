# Anyi - Open Source Autonomouse AI Agent Framework

| [English](README.md) | [中文](README-zh.md) |

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)

## Introduction

Anyi is an open source AI Agent framework designed to help you build AI Agents that can be integrated with real-world work. We also provide APIs for accessing large language models.

Anyi requires [Go](https://go.dev/) version [1.18](https://go.dev/doc/devel/release#go1.18) or higher.

## Features

Anyi, as a Go programming framework, offers the following features:

- **Access to Large Language Models**: Allows access to large language models through a uniform interface, using different configurations to access various large models. Currently supported large model interfaces include:

    - OpenAI
    - Azure OpenAI
    - Dashscope
    - Ollama

- **Multimodal Model Support**: In addition to supporting regular text-based conversations, Anyi also supports sending images to multimodal large models for processing.
- **Prompt Generation Based on Go Templates**: Supports generating prompts based on Go language templates.
- **Workflow Support**: Allows chaining multiple conversation tasks into workflows.
- **Step Validation in Workflows**: Repeats steps whose outputs do not meet expectations until the output is as expected. If the number of retries exceeds a predefined limit, an error is returned.
- **Using Different Large Model Clients in Workflow Steps**: Different workflow steps can use different large model clients.
- **Defining Multiple Workflows**: Allows defining multiple workflows and accessing them by workflow name.
- **Configuration-Based Workflow Definition**: Allows dynamic configuration of workflows through program code or static configuration files.

More features are under development. Stay tuned for updates.

## Quick start

### Installation

```bash
go get -u github.com/jieliu2000/anyi
```

### Accessing LLMs with Anyi

Here is a simple example of using Anyi to access OpenAI:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/chat"
)

func main() {
	// For more documentation and examples, see github.com/jieliu2000/anyi/llm package documentation.
	// Make sure you set OPENAI_API_KEY environment variable to your OpenAI API key.
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _ := client.Chat(messages, nil)

	log.Printf("Response: %s\n", message.Content)
}

```

In the above example, an OpenAI Anyi configuration is created via `openai.DefaultConfig`. Then this config instance is passed to `anyi.NewClient` to create an OpenAI client, and finally a chat request is sent via `client.Chat`.

## Tutorials and Examples

Please refer to the [Anyi Tutorials and Examples](/docs/en/tutorial.md)

## License

Anyi is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.