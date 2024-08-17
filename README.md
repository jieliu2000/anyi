# Anyi - Open Source Autonomouse AI Agent Framework

| [English](README.md) | [中文](README-zh.md) |

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)


## Introduction

Anyi is an open source AI Agent framework designed to help you build AI Agents that can be integrated with real-world work. We also provide APIs for accessing large models.

## Features

Anyi is a Go language programming framework that provides the following features:

- Access to large models: with Anyi you can access various large models through a common interface with different configurations. At present Anyi supports these interfaces:
	- OpenAI
	- Azure OpenAI
	- Dashscope
	- Ollama

- For these large models, in addition to supporting regular text chat, Anyi also supports sending images to multimodal models.
- Support for generating prompts based on Go language templates.
- Workflow support: Allows chaining multiple dialogue tasks to form a workflow.
- Workflow step validation: If the output of a step does not meet expectations, it will be executed repeatedly until the output meets expectations. If the number of executions exceeds a defined limit, an error is returned.
- Different steps within a workflow can use different large model clients.
- Multiple workflows can be defined and accessed by their workflow names.

More features are in development, so stay tuned.

## Quick start

### Installation

```bash
go get github.com/jieliu2000/anyi
```

### Accessing LLMs with Anyi

Here is a simple example of using Anyi to access OpenAI:

```go
package main

import (
	"os"
	"log"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"
)

func main() {

	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
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

```

In the above example, an OpenAI Anyi configuration is created via `openai.DefaultConfig`. Then this config instance is passed to `llm.NewClient` to create an OpenAI client, and finally a chat request is sent via `client.Chat`.

## License

Anyi is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.