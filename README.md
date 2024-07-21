# Anyi - Open Source Autonomouse AI Agent Framework

| [English](README.md) | [中文](README-zh.md) |

## Introduction

Anyi is an open source AI Agent framework designed to help you build AI Agents that can be integrated with real-world work. We also provide APIs for accessing large models.

## Features

Anyi is a Go language programming framework that provides the following features:

- Access to big models, allowing different big models to be accessed through the same interface with different configurations.

More features are in development, so stay tuned.

## Quick start

### Installation

```bash
go get github.com/jieliu2000/anyi
```

### Accessing LLMs with anyi

Here is a simple example of using anyi to access OpenAI:

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

In the above example, an OpenAI anyi configuration is created via `openai.DefaultConfig`. Then this config instance is passed to `llm.NewClient` to create an OpenAI client, and finally a chat request is sent via `client.Chat`.

## License

Anyi is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.