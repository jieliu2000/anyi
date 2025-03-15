# Anyi Programming Guide and Examples

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Large Language Model Access](#large-language-model-access)
  - [Client Creation Methods](#client-creation-methods)
  - [Client Configuration](#client-configuration)
    - [OpenAI](#openai)
    - [Anthropic](#anthropic)
    - [Azure OpenAI](#azure-openai)
    - [Ollama](#ollama)
    - [Other Providers](#other-providers)
- [Chat API Usage](#chat-api-usage)
  - [Message Structure](#message-structure)
  - [Return Values](#return-values)
  - [Chat Options](#chat-options)
- [Multimodal Model Usage](#multimodal-model-usage)
  - [Sending Images](#sending-images)
  - [Using ContentParts](#using-contentparts)
- [Working with Functions](#working-with-functions)
  - [Function Definitions](#function-definitions)
  - [Function Calling](#function-calling)
- [Workflow System](#workflow-system)
  - [Flow Creation](#flow-creation)
  - [Steps and Executors](#steps-and-executors)
  - [Data Passing Between Steps](#data-passing-between-steps)
  - [Validation and Retries](#validation-and-retries)
  - [Conditional Flows](#conditional-flows)
- [Configuration System](#configuration-system)
  - [Dynamic Configuration](#dynamic-configuration)
  - [Configuration Files](#configuration-files)
  - [Environment Variables](#environment-variables)
- [Built-in Components](#built-in-components)
  - [Executors](#executors)
  - [Validators](#validators)
  - [Formatting](#formatting)
- [Advanced Usage](#advanced-usage)
  - [Multiple Client Management](#multiple-client-management)
  - [Prompt Templates](#prompt-templates)
  - [Error Handling](#error-handling)
- [Best Practices](#best-practices)
  - [Performance Optimization](#performance-optimization)
  - [Cost Management](#cost-management)
  - [Security Considerations](#security-considerations)

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

## Installation

To start using Anyi, install it via Go modules:

```bash
go get -u github.com/jieliu2000/anyi
```

Anyi requires Go version 1.20 or higher.

### Verifying Your Installation

You can verify your installation by creating a simple program that imports the Anyi package:

```go
package main

import (
	"fmt"
	"github.com/jieliu2000/anyi"
)

func main() {
	fmt.Println("Anyi version:", anyi.Version)
}
```

## Large Language Model Access

Anyi provides a unified way to interact with various Large Language Models (LLMs) through a consistent interface. This approach allows you to easily switch between different providers without changing your application logic.

### Understanding Anyi's Client Architecture

Before diving into the code, it's important to understand how Anyi organizes LLM access:

1. **Providers**: Each LLM service (OpenAI, Anthropic, etc.) has a dedicated provider module
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
	config := openai.DefaultConfig("gpt-4")
	config.APIKey = os.Getenv("OPENAI_API_KEY")
	
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
	"github.com/jieliu2000/anyi/llm/anthropic"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create a client without registering it
	config := anthropic.DefaultConfig("claude-3-opus-20240229")
	config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	
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

#### Configuration Best Practices

- Store API keys in environment variables rather than hardcoding them
- Use provider-specific default configurations as starting points
- Consider setting custom timeouts for production environments
- Use custom base URLs for self-hosted models or proxy services

#### OpenAI

OpenAI's API is widely used and provides access to models like GPT-4 and GPT-3.5.

```go
// Default configuration (gpt-3.5-turbo)
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

// Configuration with specific model
config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")

// Configuration with custom base URL (for self-hosted or proxy services)
config := openai.NewConfig(
    os.Getenv("OPENAI_API_KEY"), 
    "gpt-4", 
    "https://your-openai-proxy.com/v1"
)
```

#### Anthropic

Anthropic's Claude models are known for their safety, helpfulness, and honesty.

```go
// Default configuration
config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-opus-20240229")

// Custom configuration
config := anthropic.NewConfig(
    os.Getenv("ANTHROPIC_API_KEY"),
    "claude-3-haiku-20240307",
    "https://api.anthropic.com"
)
```

#### Azure OpenAI

Azure OpenAI provides Microsoft-hosted OpenAI models with enterprise features.

```go
config := azureopenai.NewConfig(
    os.Getenv("AZ_OPENAI_API_KEY"),
    os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"),
    os.Getenv("AZ_OPENAI_ENDPOINT")
)
```

#### Ollama

Ollama provides access to locally deployed open-source models.

```go
// Default configuration (local server)
config := ollama.DefaultConfig("llama3")

// Custom server configuration
config := ollama.NewConfig("mistral", "http://your-ollama-server:11434")
```

#### Other Providers

Anyi supports many other LLM providers, including:

- **DeepSeek**: `deepseek.DefaultConfig()`
- **Zhipu AI**: `zhipu.DefaultConfig()`
- **Dashscope (Alibaba)**: `dashscope.DefaultConfig()`
- **SiliconCloud**: `siliconcloud.DefaultConfig()`

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
	Temperature: 0.7,               // Controls randomness (0.0-2.0)
	MaxTokens:   1000,              // Maximum response length
	TopP:        0.9,               // Nucleus sampling parameter
	Stream:      true,              // Enable streaming responses
	Stop:        []string{"STOP"},  // Custom stop sequences
}

response, info, err := client.Chat(messages, options)
```

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

Anyi's workflow system is one of its most powerful features, allowing you to create sophisticated AI pipelines by connecting multiple steps.

### Core Workflow Concepts

- **Flow**: A sequence of steps executed in order
- **Step**: A single unit of work with an executor and optional validator
- **Executor**: Performs the actual work (e.g., calling an LLM, setting context)
- **Validator**: Ensures output meets requirements before proceeding
- **Context**: Shared data that passes between steps

### When to Use Workflows

Workflows are particularly useful for:
- Multi-step reasoning processes
- Content generation pipelines
- Data transformation and enrichment
- Decision trees with conditional logic
- Tasks requiring validation and retries

### Flow Creation

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/flow"
)

func main() {
	// Create a client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	
	// Create individual steps
	step1, err := anyi.NewLLMStepWithTemplate(
		"Generate a short story about {{.Text}}",
		"You are a creative fiction writer.",
		client,
	)
	if err != nil {
		log.Fatalf("Failed to create step: %v", err)
	}
	step1.Name = "story_generation"
	
	step2, err := anyi.NewLLMStepWithTemplate(
		"Create a title for this story:\n\n{{.Text}}",
		"You are an expert at creating compelling titles.",
		client,
	)
	if err != nil {
		log.Fatalf("Failed to create step: %v", err)
	}
	step2.Name = "title_creation"
	
	// Create the flow
	myFlow, err := anyi.NewFlow("story_flow", client, *step1, *step2)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}
	
	// Register the flow
	err = anyi.RegisterFlow("story_flow", myFlow)
	if err != nil {
		log.Fatalf("Failed to register flow: %v", err)
	}
	
	// Run the flow
	result, err := myFlow.RunWithInput("a detective in future Tokyo")
	if err != nil {
		log.Fatalf("Flow execution failed: %v", err)
	}
	
	log.Printf("Title: %s", result.Text)
}
```

### Steps and Executors Explained

Each step in a workflow uses an executor to perform its task. Anyi provides several built-in executors:

1. **LLMExecutor**: The most common executor that sends prompts to an LLM
2. **SetContextExecutor**: Modifies the workflow context directly
3. **ConditionalFlowExecutor**: Directs flow based on conditions
4. **RunCommandExecutor**: Executes shell commands

Steps can be chained together, with the output of one step becoming the input to the next.

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
	err := anyi.ConfigFromFile("./config/workflow.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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

Example YAML configuration file (`config/workflow.yaml`):

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

### Formatting

Anyi includes formatters to help process and transform text in your workflows. Formatters can:
- Standardize output formats
- Extract specific information
- Transform data between different representations
- Apply consistent styling and formatting

## Best Practices

Building effective AI applications requires more than just technical knowledge. Here are comprehensive best practices to help you get the most out of the Anyi framework.

### Performance Optimization

Optimizing your Anyi workflows for performance can significantly improve user experience and reduce costs:

**1. Choose the Right Model for the Task**
- Use smaller, faster models for simple tasks
- Reserve more powerful models for complex reasoning
- Consider fine-tuned models for specialized domains

**2. Implement Caching**
- Cache common LLM responses to avoid redundant API calls
- Use a distributed cache for multi-instance deployments
- Set appropriate cache expiration times

**3. Optimize Prompts**
- Keep prompts concise while including necessary context
- Use clear instructions to reduce back-and-forth
- Test and iterate on prompts to minimize token usage

**4. Local Deployment Options**
- For frequent, non-critical tasks, use Ollama with local models
- Balance between cloud and local models based on requirements
- Consider quantized models for resource-constrained environments

**5. Parallel Execution**
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

## Conclusion

Anyi provides a powerful framework for building AI agents and workflows. By combining different LLM providers, workflow steps, and validation techniques, you can create sophisticated AI applications that integrate with your existing systems.

For more examples and the latest documentation, visit the [GitHub repository](https://github.com/jieliu2000/anyi).

### Getting Help and Contributing

If you encounter issues or have questions, consider:
- Opening an issue on GitHub
- Joining the community discussion
- Reading the API documentation
- Contributing improvements back to the project

The Anyi framework is continuously evolving, and your feedback helps make it better for everyone.
