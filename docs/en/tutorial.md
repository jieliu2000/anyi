# Anyi Programming Guide and Examples

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## Table of Contents

- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Usage Example](#basic-usage-example)
- [Introduction](#introduction)
- [Large Language Model Access](#large-language-model-access)
  - [Understanding Anyi's Client Architecture](#understanding-anyis-client-architecture)
  - [Client Creation Methods](#client-creation-methods)
  - [Client Configuration](#client-configuration)
  - [Supported LLM Providers](#supported-llm-providers)
    - [OpenAI](#openai)
    - [DeepSeek](#deepseek)
    - [Azure OpenAI](#azure-openai)
    - [Ollama](#ollama)
    - [Zhipu AI](#zhipu-ai)
    - [Dashscope (Alibaba Cloud)](#dashscope-alibaba-cloud)
    - [Anthropic](#anthropic)
    - [SiliconCloud](#siliconcloud)
  - [Using OpenAI-Compatible APIs](#using-openai-compatible-apis)
  - [How to Choose the Right LLM Provider](#how-to-choose-the-right-llm-provider)
- [Chat API Usage](#chat-api-usage)
  - [Understanding the Chat Lifecycle](#understanding-the-chat-lifecycle)
  - [Message Structure](#message-structure)
  - [Return Values Explained](#return-values-explained)
  - [Basic Chat Example](#basic-chat-example)
  - [Chat Options](#chat-options)
- [Multimodal Model Usage](#multimodal-model-usage)
  - [Sending Images](#sending-images)
- [Working with Functions](#working-with-functions)
  - [Function Definitions](#function-definitions)
- [Workflow System](#workflow-system)
  - [Core Workflow Concepts](#core-workflow-concepts)
  - [Flow Context](#flow-context)
- [Configuration System](#configuration-system)
  - [Dynamic Configuration](#dynamic-configuration)
  - [Configuration Files](#configuration-files)
  - [Environment Variables](#environment-variables)
- [Built-in Components](#built-in-components)
  - [Executors](#executors)
  - [Validators](#validators)
- [Advanced Usage](#advanced-usage)
  - [Multiple Client Management](#multiple-client-management)
  - [Prompt Templates](#prompt-templates)
  - [Error Handling](#error-handling)
- [Best Practices](#best-practices)
  - [Performance Optimization](#performance-optimization)
  - [Cost Management](#cost-management)
  - [Security Considerations](#security-considerations)
- [Frequently Asked Questions (FAQ)](#frequently-asked-questions-faq)

## Quick Start

If you want to get started with Anyi quickly, here are the basic steps:

### Installation

```bash
go get -u github.com/jieliu2000/anyi
```

**Requirements:** Go 1.20 or higher

### Basic Usage Example

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

This simple example demonstrates the core functionality of Anyi: creating a client and sending a request. For more detailed instructions, continue reading the full guide.

## Introduction

Anyi is an open-source Autonomous AI Agent framework written in Go, designed to help you build AI agents that integrate seamlessly with real-world workflows. This guide provides detailed programming instructions and examples for using the Anyi framework effectively.

### Key Features of Anyi

- **Unified LLM Interface**: Access multiple LLM providers through a consistent API
- **Flexible Workflow System**: Build complex, multi-step AI processes with validation and error handling
- **Configuration Management**: Support for YAML, JSON, and TOML configuration files
- **Built-in Components**: Ready-to-use executors and validators for common tasks
- **Extensible Architecture**: Create custom components to fit your specific needs

### When to Use Anyi

Anyi is particularly useful when:

- You need to orchestrate complex interactions between multiple AI models
- You want to build reliable AI workflows with validation and error handling
- You need to switch between different LLM providers without changing your code
- You're building production-grade AI applications in Go

## Large Language Model Access

Anyi provides a unified way to interact with various Large Language Models (LLMs) through a consistent interface. This approach allows you to easily switch between different providers without changing your application logic.

### Understanding Anyi's Client Architecture

Before diving into the code, it's important to understand how Anyi organizes LLM access:

1. **Providers**: Each LLM service (OpenAI, DeepSeek, etc.) has a dedicated provider module
2. **Clients**: Instances that handle communication with specific LLM services
3. **Registry**: A global store of named clients for easy retrieval throughout your application

### Client Creation Methods

Anyi provides a unified interface for accessing various large language models. There are two primary methods for creating clients:

1. Using `anyi.NewClient()` - Creates a named client registered in the global registry
2. Using `llm.NewClient()` - Creates an unregistered client instance that you manage yourself

#### When to Use Named vs. Unregistered Clients

- **Named Clients** are ideal when you need to:

  - Access the same client instance from different parts of your application
  - Configure once and reuse throughout your codebase
  - Manage multiple clients with different configurations

- **Unregistered Clients** are better when:
  - You need isolated client instances for specific tasks
  - You want to avoid potential naming conflicts
  - Your application has a simple structure with limited LLM interactions

#### Named Client Example

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
	// Create a client with a name "gpt4"
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.Model = openai.GPT4o // Use the GPT-4o model

	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Later, you can retrieve this client by name
	retrievedClient, err := anyi.GetClient("gpt4")
	if err != nil {
		log.Fatalf("Failed to retrieve client: %v", err)
	}

	// Use the client
	messages := []chat.Message{
		{Role: "user", Content: "What is the capital of France?"},
	}
	response, _, err := retrievedClient.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Response: %s", response.Content)
}
```

#### Unregistered Client Example

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create a client without registering it
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.Model = openai.GPT3Dot5Turbo

	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client directly
	messages := []chat.Message{
		{Role: "user", Content: "Explain quantum computing in simple terms"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Response: %s", response.Content)
}
```

### Client Configuration

Each LLM provider has its own configuration structure. Understanding the specific configuration options for each provider is crucial for optimizing your interactions with different models.

#### Common LLM Configuration Options

All LLM providers support a common set of configuration options provided through the `GeneralLLMConfig` structure. Understanding these options can help you optimize model outputs:

```go
// All LLM configuration structures embed GeneralLLMConfig
type SomeProviderConfig struct {
    config.GeneralLLMConfig
    // Other provider-specific configurations...
}
```

The GeneralLLMConfig includes the following options:

- **Temperature**: Controls the randomness of the output. Higher values make the output more random; lower values make it more deterministic.

  - The range is typically between 0.0 and 2.0, with a default value of 1.0
  - Example: `config.Temperature = 0.7` // More deterministic output

- **TopP**: Controls the diversity of the output. Higher values make the output more diverse; lower values make it more conservative.

  - The range is typically between 0.0 and 1.0, with a default value of 1.0
  - Example: `config.TopP = 0.9` // Maintains some diversity

- **MaxTokens**: Controls the maximum number of tokens to generate

  - 0 means no limit
  - Example: `config.MaxTokens = 2000` // Limits response length

- **PresencePenalty**: Controls how much the model avoids repeating content

  - Positive values increase the likelihood of avoiding repetition, negative values increase the likelihood of repetition
  - Example: `config.PresencePenalty = 0.5` // Moderately avoids repetition

- **FrequencyPenalty**: Controls how much the model avoids using common words

  - Positive values increase the likelihood of avoiding common words, negative values increase the likelihood of using common words
  - Example: `config.FrequencyPenalty = 0.5` // Moderately avoids common words

- **Stop**: Specifies a list of tokens that signal when to stop generating
  - Example: `config.Stop = []string{"###", "END"}` // Stops generation when encountering these tokens

##### Example Configuration

```go
import (
    "github.com/jieliu2000/anyi/llm/openai"
    "github.com/jieliu2000/anyi/llm/config"
)

// Create configuration
config := openai.DefaultConfig(apiKey)

// Adjust common parameters
config.Temperature = 0.7      // Lower temperature for more deterministic output
config.TopP = 0.9             // Slightly restrict token selection
config.MaxTokens = 500        // Limit response length
config.PresencePenalty = 0.2  // Slightly discourage repetition
config.FrequencyPenalty = 0.3 // Slightly discourage common words
config.Stop = []string{"END"} // Stop generation when encountering "END"
```

#### Configuration Best Practices

- Store API keys in environment variables rather than hardcoding them
- Use provider-specific default configurations as starting points
- Use higher Temperature (0.7-1.0) for creative tasks
- Use lower Temperature (0.1-0.3) for factual/precise tasks
- Consider setting custom timeouts for production environments
- Use custom base URLs for self-hosted models or proxy services

### Supported LLM Providers

Anyi supports a wide range of LLM providers to suit different needs and use cases. Below are detailed descriptions and examples for each supported provider, starting with the most widely used options.

#### OpenAI

OpenAI is one of the most widely used AI service providers. Access via https://platform.openai.com.

##### Configuration Example

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
	// Default configuration (gpt-3.5-turbo)
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

	// Configuration with specific model
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)

	// Create client and use example
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create OpenAI client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "What is the capital of France?"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("OpenAI response: %s", response.Content)
}
```

#### DeepSeek

DeepSeek provides specialized chat and code models, accessible via https://platform.deepseek.ai/.

##### Configuration Example

```go
package main

import (
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // Configuration with DeepSeek Chat model
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")

    // Configuration with DeepSeek Coder model
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-coder")

    // Create client and use example
    client, err := llm.NewClient(config)
    if err != nil {
        log.Fatalf("Failed to create DeepSeek client: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "Write a Go function to check if a string is a palindrome"},
    }
    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }

    log.Printf("DeepSeek response: %s", response.Content)
}
```

#### Azure OpenAI

Azure OpenAI provides Microsoft-hosted OpenAI models with enterprise-grade features and reliability.

##### Features and Advantages

- Enterprise-grade SLAs and technical support
- Compliance with various regulatory standards
- Network isolation and private network deployment options
- Integration with other Azure services

##### Configuration Example

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/azureopenai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	config := azureopenai.NewConfig(
		os.Getenv("AZ_OPENAI_API_KEY"),
		os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"),
		os.Getenv("AZ_OPENAI_ENDPOINT")
	)

	// Create client and use example
	client, err := anyi.NewClient("azure-openai", config)
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI client: %v", err)
	}

	// Use the client
	messages := []chat.Message{
		{Role: "user", Content: "What are the major differences between machine learning and deep learning?"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Azure OpenAI response: %s", response.Content)
}
```

#### Ollama

Ollama provides the ability to deploy open-source models locally, ideal for scenarios requiring offline processing or data privacy.

##### Features and Advantages

- Local deployment without requiring network connectivity
- Support for various open-source models like Llama, Mixtral, etc.
- Complete control over data flow, enhancing privacy protection
- No usage fees, suitable for large-scale experimentation

##### Configuration Example

```go
package main

import (
	"log"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Default configuration (local server)
	config := ollama.DefaultConfig("llama3")

	// Custom server configuration
	config := ollama.NewConfig("mixtral", "http://your-ollama-server:11434")

	// Create client and use example
	client, err := anyi.NewClient("local-llm", config)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Use the client for local inference
	messages := []chat.Message{
		{Role: "system", Content: "You are a math expert specializing in number theory."},
		{Role: "user", Content: "Explain the Riemann hypothesis in simple terms"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Local inference failed: %v", err)
	}

	log.Printf("Ollama model response: %s", response.Content)
}
```

#### Zhipu AI

Zhipu AI provides access to GLM series models through https://open.bigmodel.cn/. It's particularly strong for Chinese language tasks and offers competitive performance for general AI applications.

##### Features and Advantages

- Excellent Chinese language understanding and generation capabilities
- GLM-4 series models with strong reasoning abilities
- Cost-effective pricing for Chinese market
- Support for both chat and code generation tasks

##### Configuration Example

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/zhipu"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Default configuration with GLM-4-Flash model
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")

	// Configuration with GLM-4 model for more complex tasks
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")

	// Custom configuration with specific base URL
	config := zhipu.NewConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4", "https://open.bigmodel.cn/api/paas/v4/")

	// Create client and use example
	client, err := anyi.NewClient("zhipu", config)
	if err != nil {
		log.Fatalf("Failed to create Zhipu client: %v", err)
	}

	// Use the client for Chinese language tasks
	messages := []chat.Message{
		{Role: "system", Content: "你是一个专业的AI助手，擅长中文理解和生成。"},
		{Role: "user", Content: "请解释一下人工智能的发展历程"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Zhipu AI response: %s", response.Content)
}
```

#### Dashscope (Alibaba Cloud)

Dashscope provides access to Alibaba's Qwen series models via https://help.aliyun.com/en/dashscope/. It offers excellent performance for both Chinese and English tasks with strong multimodal capabilities.

##### Features and Advantages

- Qwen series models with excellent multilingual capabilities
- Strong performance in code generation and mathematical reasoning
- Integrated with Alibaba Cloud ecosystem
- Support for multimodal inputs (text and images)
- Competitive pricing and enterprise-grade reliability

##### Configuration Example

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
	// Default configuration with Qwen-Turbo model
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")

	// Configuration with Qwen-Max for complex reasoning tasks
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")

	// Custom configuration with specific base URL
	config := dashscope.NewConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-plus", "https://dashscope.aliyuncs.com/compatible-mode/v1")

	// Create client and use example
	client, err := anyi.NewClient("dashscope", config)
	if err != nil {
		log.Fatalf("Failed to create Dashscope client: %v", err)
	}

	// Use the client for code generation
	messages := []chat.Message{
		{Role: "system", Content: "You are a professional software engineer specializing in Go programming."},
		{Role: "user", Content: "Write a Go function to implement binary search algorithm"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Dashscope response: %s", response.Content)
}
```

#### Anthropic

Anthropic provides access to Claude models via https://www.anthropic.com/claude. Claude models are known for their safety, helpfulness, and excellent reasoning capabilities.

##### Features and Advantages

- Advanced reasoning and analysis capabilities
- Strong focus on AI safety and alignment
- Excellent performance in long-form content generation
- Support for large context windows
- High-quality code generation and debugging

##### Configuration Example

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/anthropic"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Default configuration with latest Claude model
	config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))

	// Configuration with specific Claude model
	config := anthropic.DefaultConfigWithModel(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-sonnet-20240229")

	// Custom configuration with all parameters
	config := anthropic.NewConfig(
		os.Getenv("ANTHROPIC_API_KEY"),
		"claude-3-opus-20240229",
		"https://api.anthropic.com/v1",
		"2023-06-01",
	)

	// Create client and use example
	client, err := anyi.NewClient("anthropic", config)
	if err != nil {
		log.Fatalf("Failed to create Anthropic client: %v", err)
	}

	// Use the client for complex analysis
	messages := []chat.Message{
		{Role: "system", Content: "You are a thoughtful analyst who provides detailed, well-reasoned responses."},
		{Role: "user", Content: "Analyze the potential implications of quantum computing on current cryptographic systems"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Anthropic response: %s", response.Content)
}
```

#### SiliconCloud

SiliconCloud provides enterprise-focused AI solutions with access to various open-source and proprietary models. It's designed for businesses requiring reliable, scalable AI services.

##### Features and Advantages

- Enterprise-grade reliability and security
- Access to multiple model families in one platform
- Competitive pricing for business applications
- Support for custom model deployment
- Strong technical support and SLA guarantees

##### Configuration Example

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/siliconcloud"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Default configuration
	config := siliconcloud.DefaultConfig(os.Getenv("SILICONCLOUD_API_KEY"), "deepseek-chat")

	// Custom configuration with specific base URL
	config := siliconcloud.NewConfig(
		os.Getenv("SILICONCLOUD_API_KEY"),
		"qwen-7b-chat",
		"https://api.siliconflow.cn/v1",
	)

	// Create client and use example
	client, err := anyi.NewClient("siliconcloud", config)
	if err != nil {
		log.Fatalf("Failed to create SiliconCloud client: %v", err)
	}

	// Use the client for business applications
	messages := []chat.Message{
		{Role: "system", Content: "You are a business analyst providing strategic insights."},
		{Role: "user", Content: "What are the key trends in enterprise AI adoption for 2024?"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("SiliconCloud response: %s", response.Content)
}
```

### Using OpenAI-Compatible APIs

Many LLM providers offer OpenAI-compatible APIs, which makes it easy to integrate them with Anyi. This approach allows you to use any service that implements the OpenAI API specification without requiring a dedicated provider module.

#### When to Use OpenAI-Compatible APIs

- **New Providers**: When you want to use a provider that doesn't have a dedicated Anyi module
- **Custom Deployments**: For self-hosted models or private deployments
- **Testing**: When evaluating new services before committing to integration
- **Flexibility**: When you need to quickly switch between different compatible providers

#### Configuration Example

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
	// Example 1: Using a custom OpenAI-compatible service
	config := openai.NewConfig(
		os.Getenv("CUSTOM_API_KEY"),
		"custom-model-name",
		"https://api.custom-provider.com/v1", // Custom base URL
	)

	// Example 2: Using Together AI (OpenAI-compatible)
	config := openai.NewConfig(
		os.Getenv("TOGETHER_API_KEY"),
		"meta-llama/Llama-2-70b-chat-hf",
		"https://api.together.xyz/v1",
	)

	// Example 3: Using Groq (OpenAI-compatible)
	config := openai.NewConfig(
		os.Getenv("GROQ_API_KEY"),
		"mixtral-8x7b-32768",
		"https://api.groq.com/openai/v1",
	)

	// Create client
	client, err := anyi.NewClient("custom-provider", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client
	messages := []chat.Message{
		{Role: "user", Content: "Hello, can you introduce yourself?"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Response: %s", response.Content)
}
```

#### Common OpenAI-Compatible Providers

Here are some popular services that offer OpenAI-compatible APIs:

- **Together AI**: Provides access to various open-source models
- **Groq**: High-speed inference for popular models
- **Perplexity AI**: Search-augmented language models
- **Fireworks AI**: Fast inference for open-source models
- **Anyscale**: Ray-based model serving platform
- **Local deployments**: vLLM, Text Generation Inference, etc.

#### Best Practices for OpenAI-Compatible APIs

1. **Check API Documentation**: Verify the exact endpoint format and authentication method
2. **Model Names**: Use the exact model names as specified by the provider
3. **Rate Limits**: Be aware of different rate limiting policies
4. **Feature Support**: Not all providers support all OpenAI features (functions, vision, etc.)
5. **Error Handling**: Different providers may return different error formats

### How to Choose the Right LLM Provider

When selecting an LLM provider, consider the following factors:

1. **Task Type**: Choose the right model based on the task, e.g., Qwen-Max (complex problems), Llama (local deployment)
2. **Language Requirements**: For Chinese language processing, prefer Zhipu AI or Dashscope
3. **Privacy Requirements**: For sensitive data, consider using Ollama for local model deployment
4. **Budget Considerations**: Balance between functionality and cost based on actual needs
5. **Latency Requirements**: Locally deployed Ollama may provide the lowest latency
6. **Scalability**: Azure OpenAI provides enterprise-grade scaling options

With Anyi framework, you can easily switch between these providers or even use multiple different LLM services within the same application.

## Chat API Usage

The core functionality of Anyi is to interact with LLMs through the Chat API. This section explains how to structure conversations, handle responses, and customize chat behavior.

### Understanding the Chat Lifecycle

A typical chat interaction with an LLM follows these steps:

1. **Prepare Messages**: Create a sequence of messages representing the conversation
2. **Configure Options**: Set parameters like temperature, max tokens, etc.
3. **Send Request**: Call the Chat method on your client
4. **Process Response**: Handle the model's reply and any metadata
5. **Continue Conversation**: Add the response to the message history for follow-ups

### Message Structure

Chat messages in Anyi use the `chat.Message` structure:

```go
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string // Text content of the message
	Name    string // Optional name (for multi-agent contexts)

	// For multimodal content
	ContentParts []ContentPart
}
```

### Return Values Explained

When calling the Chat method, you receive three values:

1. **Response Message**: The model's reply as a `chat.Message`
2. **Response Info**: Metadata about the response (tokens used, model name, etc.)
3. **Error**: Any error that occurred during the request

Understanding these return values helps you implement proper error handling and logging.

### Basic Chat Example

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
	// Create client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create message history
	messages := []chat.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "What can machine learning be used for?"},
	}

	// Send chat request
	response, info, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	// Process response
	log.Printf("Model: %s", info.Model)
	log.Printf("Response: %s", response.Content)

	// Continue the conversation
	messages = append(messages, *response) // Add assistant's response
	messages = append(messages, chat.Message{
		Role: "user",
		Content: "Can you give a specific example in healthcare?",
	})

	// Send follow-up
	response, _, err = client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Follow-up response: %s", response.Content)
}
```

### Chat Options

You can customize chat behavior using `chat.ChatOptions`:

```go
options := &chat.ChatOptions{
	Format: "json", // Specify JSON format for output (useful for structured data)
}

response, info, err := client.Chat(messages, options)
```

Currently, the Anyi framework provides a streamlined `ChatOptions` with the following functionality:

1. **Format**: When set to "json", it instructs the model to return its response in JSON format. This is particularly useful when structured data is required.

Different LLM providers may implement these options with varying behaviors, depending on the capabilities of their underlying APIs.

## Multimodal Model Usage

Many modern LLMs support multimodal inputs, allowing you to send images along with text.

### Sending Images

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
	// Create GPT-4 Vision client
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4-vision-preview")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create message with image URL
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "What's in this image?",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://example.com/image.jpg",
					},
				},
			},
		},
	}

	// Send chat request
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Description: %s", response.Content)
}
```

## Working with Functions

Many LLMs support function calling capabilities, allowing AI models to request specific actions.

### Function Definitions

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

func main() {
	// Create a client
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4")
	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Define functions
	functions := []tools.FunctionConfig{
		{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "The city and state, e.g., 'San Francisco, CA'",
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"celsius", "fahrenheit"},
						"description": "The temperature unit",
					},
				},
				"required": []string{"location"},
			},
		},
	}

	// Create message
	messages := []chat.Message{
		{Role: "user", Content: "What's the weather like in Boston?"},
	}

	// Request with function calling
	response, _, err := client.ChatWithFunctions(messages, functions, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Response type: %s", response.FunctionCall.Name)
	log.Printf("Arguments: %s", response.FunctionCall.Arguments)

	// Here you would handle the function call, execute the requested function
	// and send the result back in another message
}
```

## Workflow System

The workflow system in Anyi is one of its most powerful features, allowing you to create complex AI processing pipelines by connecting multiple steps.

### Core Workflow Concepts

- **Flow**: A sequence of steps executed in order

### Flow Context

During workflow execution, context needs to be maintained between steps. Anyi provides the `FlowContext` structure to pass and share data between various workflow steps. The `FlowContext` contains the following key properties:

- **Text**: String type, used to store the input and output text content of a step. Before a step runs, this field is the input text; after the step runs, it becomes the output text.
- **Memory**: Any type (`ShortTermMemory`), used to pass and share structured data between steps.
- **Flow**: A reference to the current flow.
- **ImageURLs**: String array, stores a list of image URLs for multimodal content processing.
- **Think**: String type, stores the content extracted from `<think>` tags in model output, used to capture the model's thinking process without affecting the final output.

#### Using ShortTermMemory

Short-term memory allows you to pass complex structured data between workflow steps, not just text. This is particularly useful in scenarios requiring multi-step processing and state maintenance.

```go
// Create workflow context with structured data
type TaskData struct {
    Objective string
    Steps     []string
    Progress  int
}

taskData := TaskData{
    Objective: "Create a website",
    Steps:     []string{"Design interface", "Develop frontend", "Develop backend", "Test and deploy"},
    Progress:  0,
}

// Initialize context with structured data in Memory
flowContext := anyi.NewFlowContextWithMemory(taskData)

// You can also set both text and memory data
flowContext := anyi.NewFlowContext("Initial input", taskData)

// Access and modify memory data in a workflow step
func (executor *MyExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Access data in Memory (requires type assertion)
    taskData := flowContext.Memory.(TaskData)

    // Update memory data
    taskData.Progress++
    flowContext.Memory = taskData

    // Update output text
    flowContext.Text = fmt.Sprintf("Current progress: %d/%d", taskData.Progress, len(taskData.Steps))

    return &flowContext, nil
}
```

#### Using the Think Field

Anyi supports special `<think>` tags where models can express their thinking process. This content doesn't affect the final output but is captured in the `Think` field. This is especially useful for models that support explicit thinking (like DeepSeek), but can also be used to prompt other models to use this format.

There are two ways to handle `<think>` tags:

1. **Automatic processing**: The `Flow.Run` method automatically detects and extracts content within `<think>` tags to the `FlowContext.Think` field, while cleaning the tag content from the `Text` field.
2. **Using DeepSeekStyleResponseFilter**: A dedicated executor for processing thinking tags:

```go
// Create a filter to process thinking tags
thinkFilter := &anyi.DeepSeekStyleResponseFilter{}
err := thinkFilter.Init()
if err != nil {
    log.Fatalf("Initialization failed: %v", err)
}

// Configure whether to output results in JSON format
thinkFilter.OutputJSON = true // When true, returns both thinking and result content as JSON format

// Use DeepSeekStyleResponseFilter as an executor
thinkStep := flow.Step{
    Executor: thinkFilter,
}

// After processing, thinking content is stored in flowContext.Think
// If OutputJSON = true, flowContext.Text will contain thinking content and results in JSON format
```

## Configuration System

Anyi's configuration system allows you to manage clients, flows, and other settings in a centralized manner. This approach brings several benefits:

- **Separation of Code and Configuration**: Keep your business logic separate from configuration details
- **Runtime Flexibility**: Change behavior without recompiling your application
- **Environment-Specific Settings**: Easily adapt to different environments (development, staging, production)
- **Centralized Management**: Define all your LLM and workflow configurations in one place

### Dynamic Configuration

Dynamic configuration allows you to programmatically define and update settings at runtime. This is useful when:

- Your configuration needs to be generated dynamically based on user input
- You're building a system that needs to adapt its behavior on the fly
- You want to test different configurations without restarting your application

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

func main() {
	// Define configuration
	config := anyi.AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "openai",
				Type: "openai",
				Config: map[string]interface{}{
					"model":  "gpt-4",
					"apiKey": os.Getenv("OPENAI_API_KEY"),
				},
			},
		},
		Flows: []anyi.FlowConfig{
			{
				Name: "content_processor",
				Steps: []anyi.StepConfig{
					{
						Name: "summarize_content",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template":      "Summarize the following content in 3 bullet points:\n\n{{.Text}}",
								"systemMessage": "You are a professional summarizer.",
							},
						},
					},
					{
						Name: "translate_summary",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "Translate the following summary to French:\n\n{{.Text}}",
							},
						},
						Validator: &anyi.ValidatorConfig{
							Type: "string",
							WithConfig: map[string]interface{}{
								"minLength": 100,
							},
						},
						MaxRetryTimes: 2,
					},
				},
			},
		},
	}

	// Apply configuration
	err := anyi.Config(&config)
	if err != nil {
		log.Fatalf("Configuration failed: %v", err)
	}

	// Get and run the flow
	flow, err := anyi.GetFlow("content_processor")
	if err != nil {
		log.Fatalf("Failed to get flow: %v", err)
	}

	input := "Artificial intelligence (AI) is intelligence demonstrated by machines, as opposed to natural intelligence displayed by animals including humans. AI research has been defined as the field of study of intelligent agents, which refers to any system that perceives its environment and takes actions that maximize its chance of achieving its goals. The term \"artificial intelligence\" had previously been used to describe machines that mimic and display \"human\" cognitive skills that are associated with the human mind, such as \"learning\" and \"problem-solving\". This definition has since been rejected by major AI researchers who now describe AI in terms of rationality and acting rationally, which does not limit how intelligence can be articulated."

	result, err := flow.RunWithInput(input)
	if err != nil {
		log.Fatalf("Flow execution failed: %v", err)
	}

	log.Printf("Result: %s", result.Text)
}
```

### Configuration Files

Using configuration files is often the most practical approach for production applications. Anyi supports multiple file formats (YAML, JSON, TOML) and provides an easy way to load them.

**Benefits of using configuration files:**

- Keep sensitive information (like API keys) out of your codebase
- Easily switch between different configurations without changing code
- Allow non-developers to modify application behavior
- Support environment-specific configurations

```go
package main

import (
	"log"
	"fmt"

	"github.com/jieliu2000/anyi"
)

func main() {
	// Load configuration from file
	err := anyi.ConfigFromFile("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Access the flow by name
	flow, err := anyi.GetFlow("content_creator")
	if err != nil {
		log.Fatalf("Failed to get flow: %v", err)
	}

	// Run the flow
	result, err := flow.RunWithInput("autonomous vehicles")
	if err != nil {
		log.Fatalf("Flow execution failed: %v", err)
	}

	fmt.Println("Generated content:", result.Text)
}
```

Example YAML configuration file (`config.yaml`):

```yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"

  - name: "anthropic"
    type: "anthropic"
    config:
      model: "claude-3-opus-20240229"
      apiKey: "$ANTHROPIC_API_KEY"

flows:
  - name: "content_creator"
    clientName: "openai"
    steps:
      - name: "research_topic"
        executor:
          type: "llm"
          withconfig:
            template: "Research the following topic and provide key facts and insights: {{.Text}}"
            systemMessage: "You are a research assistant."
        maxRetryTimes: 2

      - name: "draft_article"
        clientName: "anthropic"
        executor:
          type: "llm"
          withconfig:
            template: "Write a detailed article about this topic using the research provided:\n\n{{.Text}}"
            systemMessage: "You are a professional writer."
        validator:
          type: "string"
          withconfig:
            minLength: 500
```

### Environment Variables

Anyi supports environment variables for configuration, which is especially useful for:

- Secrets management (API keys, tokens)
- Deployment-specific settings
- CI/CD pipelines
- Container orchestration environments

Environment variables are referenced in configuration files using the `$` prefix. For example, `$OPENAI_API_KEY` in a configuration file will be replaced with the value of the `OPENAI_API_KEY` environment variable.

**Best Practices for Environment Variables:**

- Use a `.env` file for local development
- Keep sensitive information in environment variables, not in code or configuration files
- Use descriptive names for your environment variables
- Consider using a secrets manager for production environments

## Built-in Components

Anyi provides several built-in components that you can use as building blocks for your AI applications. Understanding these components will help you leverage the full power of the framework.

### Executors

Executors are the workhorses of the Anyi workflow system. They perform the actual tasks within each step.

#### Types of Built-in Executors

1. **LLMExecutor**: The most commonly used executor, it sends prompts to an LLM and captures the response.

   - Supports template-based prompts with variable substitution
   - Can use different system messages for different steps
   - Works with any registered LLM client

2. **SetContextExecutor**: Directly modifies the flow context without external calls.

   - Useful for initializing variables
   - Can overwrite or append to existing context
   - Often used at the beginning of workflows

3. **ConditionalFlowExecutor**: Enables branching logic in workflows.

   - Routes to different steps based on conditions
   - Can evaluate simple expressions
   - Allows for complex decision trees

4. **RunCommandExecutor**: Executes shell commands and captures their output.
   - Bridges the gap between AI and system operations
   - Useful for data processing or external tool integration
   - Allows workflows to interact with the operating system

### Validators

Validators are crucial components in the Anyi workflow system that ensure outputs meet specific criteria before proceeding to the next step. They serve as quality control mechanisms that can:

- Prevent low-quality or invalid outputs from propagating through your workflow
- Automatically trigger retries when outputs don't meet requirements
- Enforce data schemas and formatting requirements
- Implement business rules and logic checks

#### Types of Built-in Validators

1. **StringValidator**: Validates text output based on various criteria.
   - Length checks (minimum and maximum length)
   - Regular expression pattern matching
   - Content verification

```go
   validator := &anyi.StringValidator{
       MinLength: 100,            // Minimum length
       MaxLength: 1000,           // Maximum length
       MatchRegex: `\d{3}-\d{2}`, // Must contain pattern (e.g., SSN format)
   }
```

2. **JsonValidator**: Ensures output is valid JSON and can validate against a schema.
   - Checks for valid JSON syntax
   - Can validate against JSON Schema
   - Useful for ensuring structured data

```go
   validator := &anyi.JsonValidator{
       RequiredFields: []string{"name", "email"},
       Schema: `{"type": "object", "properties": {"name": {"type": "string"}, "email": {"type": "string", "format": "email"}}}`,
   }
```

#### Using Validators Effectively

To get the most out of validators:

- Start with simpler validations and gradually add complexity
- Use validators in combination with retry logic
- Consider creating custom validators for specific business rules
- Log validation failures to identify common issues

## Advanced Usage

### Multiple Client Management

Anyi allows you to use different LLM providers simultaneously, choosing the most appropriate model for different tasks.

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create OpenAI client for complex tasks
	openaiConfig := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	openaiClient, err := anyi.NewClient("gpt", openaiConfig)
	if err != nil {
		log.Fatalf("Failed to create OpenAI client: %v", err)
	}

	// Create Ollama local client for simple tasks
	ollamaConfig := ollama.DefaultConfig("llama3")
	ollamaClient, err := anyi.NewClient("local", ollamaConfig)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Use OpenAI client for complex problem solving
	complexMessages := []chat.Message{
		{Role: "user", Content: "Analyze the potential impact of artificial intelligence on the job market over the next decade"},
	}

	complexResponse, _, err := openaiClient.Chat(complexMessages, nil)
	if err != nil {
		log.Fatalf("OpenAI request failed: %v", err)
	}

	log.Printf("Complex question answer (GPT): %s", complexResponse.Content)

	// Use local Ollama client for simple computations
	simpleMessages := []chat.Message{
		{Role: "user", Content: "Calculate the result of 342 + 781"},
	}

	simpleResponse, _, err := ollamaClient.Chat(simpleMessages, nil)
	if err != nil {
		log.Fatalf("Ollama request failed: %v", err)
	}

	log.Printf("Simple calculation answer (Ollama): %s", simpleResponse.Content)

	// In a workflow, you could switch clients based on step requirements
	// Workflow code...
}
```

### Prompt Templates

Using templated prompts enhances the flexibility and reusability of LLM interactions. Anyi leverages Go's template system, supporting dynamic variable substitution.

#### Using FlowContext Data in Templates

In prompt templates, you can access various properties of the `FlowContext`:

1. **Using the Text field**: Directly access the current context text content with `.Text`.

```
Analyze the following text: {{.Text}}
```

2. **Using the Memory field**: Access structured data and its internal properties.

```
Task objective: {{.Memory.Objective}}
Current progress: {{.Memory.Progress}}
Task list:
{{range .Memory.Steps}}
- {{.}}
{{end}}
```

3. **Using the Think field**: Access the model's thinking process (if a previous step extracted `<think>` tag content).

```
Thinking process from the previous step: {{.Think}}

Please continue the analysis and provide a more detailed answer.
```

4. **Using Image URLs**: If image URLs are provided, you can reference them in the prompt.

A practical example integrating memory and thinking process:

```go
// Define structured data
type AnalysisData struct {
    Topic        string
    Requirements []string
    Progress     map[string]bool
}

// Create structured data
data := AnalysisData{
    Topic:        "AI Safety",
    Requirements: []string{"Current State", "Key Challenges", "Future Trends"},
    Progress:     map[string]bool{"Current State": true, "Key Challenges": false, "Future Trends": false},
}

// Create template text
templateText := `
Analyze the following topic: {{.Memory.Topic}}

Points to cover:
{{range .Memory.Requirements}}
- {{.}}
{{end}}

Current progress:
{{range $key, $value := .Memory.Progress}}
- {{$key}}: {{if $value}}Completed{{else}}Not completed{{end}}
{{end}}

{{if .Think}}
Thinking process from the previous step:
{{.Think}}
{{end}}

Please analyze the points that are not yet completed.
`

// Create context with memory
flowContext := anyi.NewFlowContextWithMemory(data)

// Previous step might have thinking content
flowContext.Think = "<think>I should focus on Key Challenges and Future Trends since Current State is already completed</think>"

// Create template
formatter, err := chat.NewPromptTemplateFormatter(templateText)
if err != nil {
    log.Fatalf("Failed to create template: %v", err)
}

// Create executor with template
executor := &anyi.LLMExecutor{
    TemplateFormatter: formatter,
    SystemMessage:     "You are a professional research analyst",
}

// Create and run flow
// ...
```

### Error Handling

Robust error handling is crucial in applications that interact with LLMs. Here are some patterns for implementing effective error handling in Anyi:

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Prepare messages
	messages := []chat.Message{
		{Role: "user", Content: "Explain the basic principles of quantum mechanics"},
	}

	// Implement retry logic
	maxRetries := 3
	backoff := 1 * time.Second

	var response *chat.Message
	var info chat.ResponseInfo

	for i := 0; i < maxRetries; i++ {
		response, info, err = client.Chat(messages, nil)

		if err == nil {
			// Successfully got response, break the loop
			break
		}

		// Check if error is retryable (like network errors, timeouts, etc.)
		if i < maxRetries-1 {
			log.Printf("Attempt %d failed: %v, retrying in %v", i+1, err, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}

	if err != nil {
		log.Fatalf("Still failed after %d attempts: %v", maxRetries, err)
	}

	// Process successful response
	log.Printf("Response: %s", response.Content)
	log.Printf("Tokens used: %d", info.PromptTokens + info.CompletionTokens)

	// Error logging and monitoring
	// In a real application, you should implement more sophisticated error logging and monitoring
	// For example, sending errors to a log management system or monitoring service
}
```

### 2. How to handle large texts that exceed token limits?

```go
// Implement chunked text processing
func processLargeText(text string, client *llm.Client) (string, error) {
    // Split text into smaller chunks
    chunks := splitIntoChunks(text, 1000) // About 1000 words per chunk

    var results []string
    // Process each chunk
    for _, chunk := range chunks {
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: "Process the following text: " + chunk},
        }, nil)
        if err != nil {
            return "", err
        }
        results = append(results, response.Content)
    }

    // Combine the results
    return combineResults(results), nil
}
```

### 3. How to integrate Anyi with existing Go web frameworks?

Anyi can seamlessly integrate with any Go web framework such as Gin, Echo, or Fiber. Here's an example with Gin:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jieliu2000/anyi"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // Initialize Anyi client
    // ...

    r.POST("/ask", func(c *gin.Context) {
        var req struct {
            Question string `json:"question"`
        }
        if err := c.BindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // Use Anyi client to process the request
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: req.Question},
        }, nil)

        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"answer": response.Content})
    })

    return r
}
```

## Best Practices

Building effective AI applications requires more than just technical knowledge. Here are comprehensive best practices to help you get the most out of the Anyi framework.

### Performance Optimization

Optimizing your Anyi workflows for performance can significantly improve user experience and reduce costs:

**1. Choose the Right Model for the Task**

- Use smaller, faster models for simple tasks
- Reserve more powerful models for complex reasoning
- Consider fine-tuned models for specialized domains

**2. Configure Generation Parameters Properly**

- Adjust `Temperature` based on the task:
  - Use 0.1-0.3 for factual responses, coding, and precision tasks
  - Use 0.7-1.0 for creative writing, brainstorming, and diverse outputs
- Set appropriate `MaxTokens` to avoid unnecessary long responses
- Use `PresencePenalty` (0.1-0.5) to reduce repetitive outputs in longer generations
- Apply `FrequencyPenalty` (0.1-0.5) to encourage more varied vocabulary
- Use `Stop` tokens to automatically end generation at appropriate points

**3. Implement Caching**

- Cache common LLM responses to avoid redundant API calls
- Use a distributed cache for multi-instance deployments
- Set appropriate cache expiration times

**4. Optimize Prompts**

- Keep prompts concise while including necessary context
- Use clear instructions to reduce back-and-forth
- Test and iterate on prompts to minimize token usage

**5. Local Deployment Options**

- For frequent, non-critical tasks, use Ollama with local models
- Balance between cloud and local models based on requirements
- Consider quantized models for resource-constrained environments

**6. Parallel Execution**

- Identify workflow steps that can run in parallel
- Use goroutines for concurrent LLM calls when appropriate
- Implement proper error handling for parallel steps

### Cost Management

Managing costs is essential when working with commercial LLM providers:

**1. Token Monitoring**

- Implement token counting to track usage
- Set up alerts for unusual spending patterns
- Regularly audit your token consumption

**2. Tiered Model Strategy**

- Use a cascading approach: try cheaper models first
- Upgrade to more expensive models only when necessary
- Implement fallbacks for service outages

**3. Response Length Control**

- Set appropriate MaxTokens limits for each use case
- Use validation to ensure outputs aren't unnecessarily verbose
- Implement truncation strategies for excessive outputs

**4. Batching Requests**

- Combine multiple small requests when possible
- Implement queue systems for non-urgent processing
- Schedule batch processing during off-peak hours

**5. Cost Attribution**

- Track costs by workflow, feature, or user
- Implement per-user quotas or rate limits
- Consider passing costs to end-users for premium features

### Security Considerations

Security is paramount when building AI systems:

**1. API Key Management**

- Never hardcode API keys in your application
- Use environment variables or a secrets manager
- Rotate keys regularly and limit their permissions

**2. Input Sanitization**

- Validate and sanitize all user inputs
- Implement rate limiting to prevent abuse
- Use context filtering to prevent prompt injection

**3. Output Validation**

- Always validate LLM outputs before using them
- Be cautious when using LLM outputs in executable contexts
- Implement content moderation for user-facing outputs

**4. Data Privacy**

- Minimize sending sensitive data to LLMs
- Implement data retention policies
- Consider using local models for processing sensitive information

**5. Audit and Logging**

- Maintain detailed logs of all LLM interactions
- Implement proper log redaction for sensitive content
- Set up monitoring for unusual patterns or security incidents

By following these best practices, you can build AI applications that are not only powerful but also efficient, cost-effective, and secure.

## Frequently Asked Questions (FAQ)

### 1. How to ensure workflows work properly in unstable network conditions?

Anyi has built-in retry mechanisms. You can set the `MaxRetryTimes` property for each step to implement automatic retries:

```go
// Set maximum retry count
step1.MaxRetryTimes = 3
```

### 2. How to handle large texts that exceed token limits?

```go
// Implement chunked text processing
func processLargeText(text string, client *llm.Client) (string, error) {
    // Split text into smaller chunks
    chunks := splitIntoChunks(text, 1000) // About 1000 words per chunk

    var results []string
    // Process each chunk
    for _, chunk := range chunks {
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: "Process the following text: " + chunk},
        }, nil)
        if err != nil {
            return "", err
        }
        results = append(results, response.Content)
    }

    // Combine the results
    return combineResults(results), nil
}
```

### 3. How to integrate Anyi with existing Go web frameworks?

Anyi can seamlessly integrate with any Go web framework such as Gin, Echo, or Fiber. Here's an example with Gin:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jieliu2000/anyi"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // Initialize Anyi client
    // ...

    r.POST("/ask", func(c *gin.Context) {
        var req struct {
            Question string `json:"question"`
        }
        if err := c.BindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // Use Anyi client to process the request
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: req.Question},
        }, nil)

        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"answer": response.Content})
    })

    return r
}
```

## Conclusion

Anyi provides a powerful framework for building AI agents and workflows. By combining different LLM providers, workflow steps, and validation techniques, you can create sophisticated AI applications that integrate with your existing systems.

For more examples and the latest documentation, visit the [GitHub repository](https://github.com/jieliu2000/anyi).

### System Requirements

- Go 1.20 or higher
- Network connectivity (for accessing LLM APIs)
- Works on all major operating systems (Linux, macOS, Windows)

### Getting Help and Contributing

If you encounter issues or have questions, consider:

- Opening an issue on GitHub
- Joining the community discussion
- Reading the API documentation
- Contributing improvements back to the project

The Anyi framework is continuously evolving, and your feedback helps make it better for everyone.
