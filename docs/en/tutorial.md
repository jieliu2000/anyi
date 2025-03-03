# Anyi Tutorials and Examples

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## Table of Contents

- [Installation](#installation)
- [Accessing Large Language Models](#accessing-large-language-models)
  - [Client Creation](#creating-clients)
  - [Client Configuration](#client-configuration)
    - [OpenAI](#openai)
    - [Azure OpenAI](#azure-openai)
    - [Dashscope](#dashscope)
    - [Ollama](#ollama)
    - [Zhipu AI Open Platform](#zhipu-ai-open-platform-bigmodelcn)
    - [DeepSeek AI](#deepseek)
    - [SiliconCloud AI Platform](#siliconcloud-ai-platform-siliconflowcn)
- [Large Language Model Chat](#large-language-model-chat)
  - [Message Structure](#chatmessage-struct)
  - [Return Values](#return-values-of-large-model-invocation)
- [Multimodal Large Model Invocation](#multimodal-large-model-invocation)
  - [Basic Usage](#the-simplest-multimodal-large-model-invocation)
  - [ContentParts Property](#reading-images-to-the-large-model-via-the-chatcontentparts-property)

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
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {

	// Before running, make sure to set the OPENAI_API_KEY environment variable to your OpenAI API key.
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _, _:= client.Chat(messages, nil)

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

- OpenAI documentation
- [completion.go in the go-openai project](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
- Constant definitions in [openai.go](../../llm/openai/openai.go) in Anyi

##### Using a Different baseURL

If you want to modify the baseURL used by OpenAI, you can create a configuration using the `openai.NewConfig()` function:

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```

Here, `gpt-4o-mini` is the model name, and `https://api.openai.com/v1` is the baseURL. This parameter is particularly useful when using other large models compatible with the OpenAI API.

#### Azure OpenAI

The configuration method for Azure OpenAI differs from OpenAI. To use Azure OpenAI, you need to set the following parameters:

- Azure OpenAI API Key
- Azure OpenAI Model Deployment ID
- Azure OpenAI Endpoint

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

#### Zhipu AI Open Platform (bigmodel.cn)

##### Accessing Zhipu AI with Default Configuration

```go
// Make sure you set ZHIPU_API_KEY environment variable to your Zhipu API key.
config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")
client, err := llm.NewClient(config)
```

The zhipu package path is:

```go
import "github.com/jieliu2000/anyi/llm/zhipu"
```

#### DeepSeek

##### Default Configuration for DeepSeek

```go
// Make sure you set DEEPSEEK_API_KEY environment variable
config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")
client, err := llm.NewClient(config)
```

DeepSeek package path:

```go
import "github.com/jieliu2000/anyi/llm/deepseek"
```

#### SiliconCloud AI Platform (siliconflow.cn)

##### Accessing SiliconCloud AI Platform

```go
// Make sure you set SILICONCLOUD_API_KEY environment variable to your Siliconcloud API key.
config := siliconcloud.DefaultConfig(os.Getenv("SILICONCLOUD_API_KEY"), "glm-4-flash")
client, err := llm.NewClient(config)
```

The siliconcloud package path is:

```go
import "github.com/jieliu2000/anyi/llm/siliconcloud"
```

### Large Language Model Chat

In Anyi, the entry point for a standard LLM chat invocation is the `client.Chat()` function. This function takes a parameter of type `[]chat.Message`, which represents the chat messages received by the large model.

The aforementioned `chat` package is `"github.com/jieliu2000/anyi/llm/chat"`. You can import it with the following code:

```go
import "github.com/jieliu2000/anyi/llm/chat"
```

#### `chat.Message` Struct

The `chat.Message` struct is defined as follows:

```go
type Message struct {
	Content      string        `json:"content,omitempty"`
	Role         string        `json:"role"`
	ContentParts []ContentPart `json:"contentParts,omitempty"`
}
```

Here, `Content` is the message content, `Role` is the role of the message sender, and `ContentParts` is an attribute set up for calling images in multimodal large models. We will discuss how to use the `ContentParts` property to pass images to large models later on. If you only need to chat using text messages with a large model, the `Content` attribute will suffice, and you can completely ignore the `ContentParts` attribute.

_\* For details on how to use the `ContentParts` property, refer to [Multimodal Large Model Invocation](#multimodal-large-model-invocation)_

As demonstrated previously, you can create messages by directly assigning values to the properties of `chat.Message`, or you can create messages using the `chat.NewMessage()` function.

The following code demonstrates directly creating a user message and invoking a large model chat:

```go
messages := []chat.Message{
	{Role: "user", Content: "5+1=?"},
}
message, responseInfo, err := client.Chat(messages, nil)
```

The following code shows creating a user message using the `chat.NewMessage()` function and invoking a large model chat:

```go
messages := []chat.Message{
	chat.NewMessage("user", "Hello, world!"),
}
message, responseInfo, err := client.Chat(messages, nil)
```

#### Return Values of Large Model Invocation

The `client.Chat()` function has three return values:

- The first one is a pointer of type `*chat.Message`, representing the chat message returned by the large model. Since this is a pointer, if an error occurs during the call to the large model, this return value will be set to nil.
- The second one is a value of type `chat.ResponseInfo`, indicating response information from the large model, such as the number of tokens.
- The third one is a value of type `error`, representing any errors that occurred while calling the large model.

You may have noticed that the return value of the `client.Chat()` function is also a pointer of type `chat.Message`. This is part of our design to simplify Anyi's code. Typically, the `Role` attribute of the returned Message will be `"assistant"`, and the `Content` attribute will contain the response from the large model.

Currently, `chat.ResponseInfo` is defined as follows:

```go
type ResponseInfo struct {
	PromptTokens     int
	CompletionTokens int
}
```

As you can see, the properties within the `chat.ResponseInfo` struct are quite straightforward, each representing various token counts returned by the large model. In the future, we plan to add more attributes to this struct to support more detailed invocation information from large models.

Given that Anyi supports multiple large model interfaces and that each interface may return different information, not all attributes of the `chat.ResponseInfo` struct may be populated after each large model call. You should handle the return values according to the specific large model interface used.

### Multimodal Large Model Invocation

In Anyi, the invocation entry for the multimodal large model remains the `client.Chat()` function. Unlike single-modal large models, when invoking a multimodal large model, you need to use the `ContentParts` property in the `chat.Message` struct to pass images to the large model.

#### The Simplest Multimodal Large Model Invocation

In the `github.com/jieliu2000/anyi/llm/chat` package, we provide two simple functions for creating the `chat.Message` struct required for multimodal large model invocations. They are:

- `chat.NewImageMessageFromUrl()` is used to create a `chat.Message` struct that passes a network image URL to the large model.
- `chat.NewImageMessageFromFile()` is used to create a `chat.Message` struct that passes a local image file to the large model.

These two functions are very useful when you only need to pass a prompt string and one image to the visual large model. If you need to pass multiple images, you will need to manually create the `chat.Message` struct and then manually set the `ContentParts` property.

The following code demonstrates how to use the `chat.NewImageMessageFromUrl()` function to create the `chat.Message` struct required for multimodal large model invocation (using Dashscope):

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Make sure you set DASHSCOPE_API_KEY environment variable to your Dashscope API key.
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-vl-plus")
	client, err := anyi.NewClient("dashscope", config)

	if err!= nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		chat.NewImageMessageFromUrl("user", "What's this?", "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"),
	}

	message, responseInfo, err := client.Chat(messages, nil)

	if err!= nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
	log.Printf("Prompt tokens: %v", responseInfo.PromptTokens)
}
```

In the above code, we used the `chat.NewImageMessageFromUrl()` function to create a `chat.Message` struct that passes a network image URL to the large model. The first parameter of the `chat.NewImageMessageFromUrl()` function is the message role, the second parameter is the text message content, and the third parameter is the image URL.

It should be noted that in the above code, the `ContentParts` property in the finally created `chat.Message` struct is an array of length **2** (not just one ContentPart). The first element in the array is a `ContentPart` struct of text type, and its text content is `"What's this?"`; the second element is a `ContentPart` struct of image type, and its image URL is `https://dashscope.oss-cn-beijing.aliyuncs.com/`.

And the `Content` property of the `chat.Message` struct is an empty string. That is to say, the text message in the parameters of `chat.NewImageMessageFromUrl()` is reflected in the `ContentParts` property, not in the `Content` property of `chat.Message`. This is also a significant difference between multimodal large model invocations and single-modal large model invocations in Anyi.

`chat.NewImageMessageFromUrl()` is used to create a `chat.Message` struct that passes a network image URL to the large model.

The `chat.NewImageMessageFromFile()` function is similar to the `chat.NewImageMessageFromUrl()` function, except that it creates a `chat.Message` struct from a local image file. The following is an example code using the `chat.NewImageMessageFromFile()` function:

```go
messages := []chat.Message{
	chat.NewImageMessageFromFile("user", "What number is in the image?", "../internal/test/number_six.png"),
	}
```

As can be seen, the first parameter of the `chat.NewImageMessageFromFile()` function is the message role, the second parameter is the text message content, and the third parameter is the image file path.

The `ContentParts` property in the `chat.Message` struct created by the `chat.NewImageMessageFromFile()` function is an array of length **2** (not just one ContentPart). The first element in the array is a `ContentPart` struct of text type, and its text content is `"What's this?"`; the second element is a `ContentPart` struct of image type, and its `ImageUrl` property is the base64-encoded URL of the passed image file.

The `chat.NewImageMessageFromFile()` function will try to read the file. If the read fails, it will return a `chat.Message` value that only contains the `Role` property, and its `Content` property and `ContentParts` property are both empty.

#### Reading Images to the Large Model via the `chat.ContentParts` Property

In multimodal large model invocations, the `ContentParts` property in the `chat.Message` struct is an array, and each element in the array is a `ContentPart` struct. The `ContentPart` struct is defined as follows:

```go
type ContentPart struct {
	Text        string `json:"text"`
	ImageUrl    string `json:"imageUrl"`
	ImageDetail string `json:"imageDetail"`
}
```

The `ContentPart` struct contains three properties: `Text`, `ImageUrl`, and `ImageDetail`. The `Text` property is used to pass text messages to the large model, the `ImageUrl` property is used to pass image URLs to the large model, and the `ImageDetail` property is used to pass the level of detail of the image.

It should be noted that the **`Text` and `ImageUrl` properties are mutually exclusive**. That is, if you set both the `Text` and `ImageUrl` properties for the `ContentPart` struct at the same time, the `ImageUrl` property will be ignored. If you want to pass an image, set the `Text` property to an empty string.

The ImageUrl can be a network image URL or a base64-encoded URL of an image. Anyi provides the `chat.NewImagePartFromFile()` function to convert a local image into a `ContentPart` struct, and also provides the `chat.NewImagePartFromUrl()` function to convert a network image URL into a `ContentPart` struct.

The ImageDetail property is used to pass the level of detail of the image. For example, `"low"`, `"medium"`, `"high"`, and `"auto"` are used to represent the level of detail of the image. If you are not sure which value to use, you can directly set this parameter to an empty string.

The following code demonstrates how to use the `chat.NewImagePartFromUrl()` function to convert a network image URL into a `ContentPart` struct:

```go
imageUrl := "https://example.com/image.jpg"
contentPart, err := chat.NewImagePartFromUrl(imageUrl, "")
```

The `chat.NewImagePartFromUrl()` function does not verify the image. However, when using the `client.Chat()` function to invoke the large model, Anyi will take different actions according to the situations of different large models:

- In most cases, if the large model API supports passing image information via URL, Anyi will not check whether the image URL is valid, but directly pass the image URL to the large model API. In this case, you need to ensure that the image URL you provide is valid.
- For APIs such as ollama, which do not support passing images via URL, in this case, Anyi will read the image according to the URL and convert the image into the format required by the large model API (such as base64 encoding) and pass it out. Obviously, if the URL points to an inaccessible or invalid image, the `client.Chat()` function will return an error before actually interacting with the large model.
