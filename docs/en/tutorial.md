# Anyi Tutorials and Examples

## Installation

```bash
go get github.com/jieliu2000/anyi
```

## Access Large Language Models

Anyi supports access to the following large model APIs:

- OpenAI
- Azure OpenAI
- Alibaba Cloud Model Service Lingji (Dashscope)
- Ollama

Anyi uses a unified interface for large model access, allowing you to use almost the same code with different configurations to access various large models. Below is a simple example using OpenAI:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"
)

func main() {

	// Before running, remember to set the OPENAI_API_KEY environment variable to your OpenAI API Key
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)

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
In the above code, we create the OpenAI configuration using the `openai.DefaultConfig()` function and then create an Anyi client using the `anyi.NewClient()` function.

### Client Creation Methods

You can create Anyi clients in the following two ways:
- Directly using the `anyi.NewClient()` function, passing in the model name and configuration.
- Using the `llm.NewClient()` function, where llm is a submodule of `github.com/jieliu2000/anyi/llm`.

The difference between `llm.NewClient()` and `anyi.NewClient()` is that the client created by `llm.NewClient()` does not have a client name, so you need to store the created instance in your code yourself. The client created by `anyi.NewClient()` has a client name, and after creation, you can retrieve the client instance using the `anyi.GetClient()` function.

Here's an example of creating a client using `anyi.NewClient()`:

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
_, err := anyi.NewClient("openai", config)

// Retrieve the client instance by client name
client, err = anyi.GetClient("openai")
```

### Client Configuration

In Anyi, you can create different large model clients with different configurations. Here are some code examples.

#### OpenAI

##### Creating a Client with Default Configuration

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := anyi.NewClient("openai", config)
```

##### Creating a Client with a Specified Model Name

The default configuration uses the `gpt-3.5-turbo` model. If you want to use another model, you can create the configuration using the `openai.NewConfigWithModel()` function:

```go
config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)
client, err := anyi.NewClient("gpt4o", config)
```

The model name can be referenced from any of the following documents:
* OpenAI documentation
* [go-openai project's completion.go](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
* Or refer to the constant definitions in Anyi’s [openai.go](../../llm/openai/openai.go)

##### Using a Different baseURL

If you want to modify the baseURL used for calling OpenAI, you can create the configuration using the `openai.NewConfig()` function:

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```
Here, `gpt-4o-mini` is the model name, and `https://api.openai.com/v1` is the baseURL. This parameter is particularly useful when using other large models compatible with the OpenAI API.

#### Accessing Other Large Models

For accessing other large models, please refer to the instructions in [Other Large Model Configurations](./llm_config.md).