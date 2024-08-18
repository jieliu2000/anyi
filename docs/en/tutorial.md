# Anyi Tutorials and Examples

## Installation

```bash
go get -u  github.com/jieliu2000/anyi  
```

## Accessing Large Language Models

Anyi supports access to the following large language model APIs:

- OpenAI
- Azure OpenAI
- Alibaba Cloud Model Service Lingji (Dashscope)
- Ollama

Anyi uses a unified interface for accessing different large models, allowing you to use almost the same set of code with different configurations to access various large models. Below is a simple example using OpenAI:

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

	// Before running, make sure to set the OPENAI_API_KEY environment variable to your OpenAI API key.
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

### Creating Clients

Clients can be created in Anyi using the following two methods:
- Directly using the `anyi.NewClient()` function, passing the model name and configuration.
- Using the `llm.NewClient()` function, where `llm` is a submodule of `github.com/jieliu2000/anyi/llm`.

The difference between `llm.NewClient()` and `anyi.NewClient()` is that clients created with `llm.NewClient()` do not have a client name, so you need to save the created instance in your code yourself, while clients created with `anyi.NewClient()` have a client name. After creation, you can retrieve the client instance through the `anyi.GetClient()` function.

Below is an example of creating a client using `anyi.NewClient()`:

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
_, err := anyi.NewClient("openai", config)

// Get the client instance by client name
client, err = anyi.GetClient("openai")
```

### Client Configuration

In Anyi, you can create different large model clients with different configurations. Here are some code examples. You can also find more complete code examples in the [Anyi LLM pkg.go.dev documentation](https://pkg.go.dev/github.com/jieliu2000/anyi/llm).

#### OpenAI

##### Creating a Client with Default Configuration

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := anyi.NewClient("openai", config)
```

##### Creating a Client with Specified Model Name

The default configuration uses the `gpt-3.5-turbo` model. If you want to use another model, you can create a configuration using the `openai.NewConfigWithModel()` function:

```go
config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)
client, err := anyi.NewClient("gpt4o", config)
```

You can reference the model name from any of the following documents:
* OpenAI documentation
* [completion.go in the go-openai project](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
* Constant definitions in [openai.go](../../llm/openai/openai.go) in Anyi

##### Using a Different baseURL

If you want to modify the baseURL used by OpenAI, you can create a configuration using the `openai.NewConfig()` function:

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```

Here, `gpt-4o-mini` is the model name, and `https://api.openai.com/v1` is the baseURL. This parameter is particularly useful when using other large models compatible with the OpenAI API.

#### Azure OpenAI

The configuration method for Azure OpenAI differs from OpenAI. To use Azure OpenAI, you need to set the following parameters:
* Azure OpenAI API Key
* Azure OpenAI Model Deployment ID
* Azure OpenAI Endpoint

The following code demonstrates how to use Azure OpenAI in Anyi:

```go
config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
client, err := llm.NewClient(config)
```

The `azureopenai` module is `"github.com/jieliu2000/anyi/llm/azureopenai"`. You can import it as follows:

```go
import "github.com/jieliu2000/anyi/llm/azureopenai"
```

Since all configuration items in Azure OpenAI are required and there are no default options, there is only one way to create an Azure OpenAI configuration in Anyi.

#### Dashscope

Dashscope Lingji is a large model API service provided by Alibaba Cloud. You can access it at https://dashscope.aliyun.com/. Currently, Anyi accesses Dashscope via an OpenAI-compatible API. The specific Dashscope documentation can be found at https://help.aliyun.com/zh/dashscope/developer-reference/compatibility-of-openai-with-dashscope .

##### Accessing Dashscope with Default Configuration

```go
// The first parameter of dashscope.DefaultConfig is the Dashscope API Key. If using this code, set the DASHSCOPE_API_KEY environment variable to your Dashscope API Key.
config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")
client, err := anyi.NewClient("openai", config)
```

Here, `qwen-turbo` is the model name. The `dashscope` module is `"github.com/jieliu2000/anyi/llm/dashscope"`. You can import it as follows:

```go
import "github.com/jieliu2000/anyi/llm/dashscope"
```

The default configuration here refers to using the default baseURL, which is `"https://dashscope.aliyuncs.com/compatible-mode/v1"`.

##### Accessing Dashscope with a Specified baseURL

```go
// The first parameter of NewConfig is the Dashscope API Key. If using this code, set the DASHSCOPE_API_KEY environment variable to your Dashscope API Key. The second parameter is the model name, and the third parameter is the baseURL.
config := dashscope.NewConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo", "https://your-url.com")
client, err := anyi.NewClient("openai", config)
```

#### Ollama

Ollama is a local LLM service that can run on your computer. You can visit https://ollama.ai/ for more information. Ollama supports multiple open-source models such as Llama 3.1, Mistral, qwen2, Phi-3, etc. To use Anyi to access Ollama, you need to install Ollama first and pull the model you wish to use with Ollama. If you're not familiar with Ollama usage, you can read the [Ollama documentation](https://github.com/ollama/ollama/tree/main/docs) first.

The following code demonstrates how to use Ollama in Anyi:

```go
// Ensure you have installed and started Ollama, and pulled the mistral model in Ollama.
config := ollama.DefaultConfig("mistral")
client, err := llm.NewClient(config)
```

Here, `mistral` is the model name. Refer to the [Ollama documentation](https://github.com/ollama/ollama/tree/main/docs) and the [Ollama model library list](https://ollama.com/library) for more model names.

The `ollama` module is `"github.com/jieliu2000/anyi/llm/ollama"`. You can import it as follows:

```go
import "github.com/jieliu2000/anyi/llm/ollama"
```

When creating the configuration, the `ollama.DefaultConfig()` function is used, which creates a default Anyi Ollama configuration based on the passed model name. A default configuration means using the default baseURL, which is `"http://localhost:11434"`.

If you want to access an Ollama API not hosted on localhost or if you've changed the port of the Ollama service, you can create a custom configuration using the `ollama.NewConfig()` function:

```go
config := ollama.NewConfig("mistral", "http://your-ollama-server:11434")
client, err := llm.NewClient(config)
```