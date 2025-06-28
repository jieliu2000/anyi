# Working with LLM Clients

This comprehensive guide covers all LLM providers supported by Anyi, their configuration options, and practical usage examples. By the end of this tutorial, you'll know how to connect to any supported AI provider and optimize your configurations for different use cases.

## Understanding Client Architecture

Before diving into specific providers, let's understand how Anyi organizes LLM access:

1. **Providers**: Each LLM service (OpenAI, DeepSeek, etc.) has a dedicated provider module
2. **Clients**: Instances that handle communication with specific LLM services
3. **Registry**: A global store of named clients for easy retrieval throughout your application

## Client Creation Methods

Anyi provides two primary methods for creating clients:

### Named Clients (Recommended)

Use `anyi.NewClient()` to create clients that are registered globally:

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := anyi.NewClient("gpt4", config)

// Later, retrieve the client by name
retrievedClient, err := anyi.GetClient("gpt4")
```

**When to use**: When you need to access the same client from different parts of your application.

### Unregistered Clients

Use `llm.NewClient()` for isolated client instances:

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := llm.NewClient(config)
```

**When to use**: For one-off tasks or when you want to avoid naming conflicts.

## Common Configuration Options

All LLM providers support common configuration options through `GeneralLLMConfig`:

```go
config.Temperature = 0.7        // Controls randomness (0.0-2.0)
config.TopP = 0.9              // Controls diversity (0.0-1.0)
config.MaxTokens = 1000        // Maximum response length
config.PresencePenalty = 0.2   // Reduces repetition
config.FrequencyPenalty = 0.3  // Encourages varied vocabulary
config.Stop = []string{"END"}  // Stop generation tokens
```

### Configuration Guidelines

- **Temperature**: 0.1-0.3 for factual tasks, 0.7-1.0 for creative tasks
- **TopP**: Keep at 0.9 for most use cases
- **MaxTokens**: Set based on expected response length
- **Penalties**: Use sparingly (0.1-0.5) to avoid over-correction

## Supported Providers

### OpenAI

OpenAI provides access to GPT models via https://platform.openai.com.

#### Basic Configuration

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
	// Default configuration (gpt-4o-mini)
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

	// Specific model configuration
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)

	// Custom configuration
	config := openai.NewConfig(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4-turbo-preview",
		"https://api.openai.com/v1", // Custom base URL if needed
	)

	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create OpenAI client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Explain the concept of machine learning"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Response: %s", response.Content)
}
```

#### Available Models

- `openai.GPT3Dot5Turbo` - Fast and cost-effective
- `openai.GPT4` - High-quality reasoning
- `openai.GPT4o` - Optimized GPT-4 variant
- `openai.GPT4Turbo` - Latest GPT-4 with improved performance

### DeepSeek

DeepSeek offers specialized models for chat and code generation via https://platform.deepseek.ai/.

#### Configuration Examples

```go
package main

import (
	"log"
	"os"
	"github.com/jieliu2000/anyi/llm/deepseek"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Chat model for general conversation
	config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")

	// Coder model for programming tasks
	config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")

	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create DeepSeek client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Write a Go function to reverse a string"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("DeepSeek response: %s", response.Content)
}
```

### Azure OpenAI

Azure OpenAI provides Microsoft-hosted OpenAI models with enterprise features.

#### Features and Advantages

- Enterprise-grade SLAs and support
- Compliance with regulatory standards
- Network isolation options
- Integration with Azure services

#### Configuration

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
		os.Getenv("AZ_OPENAI_ENDPOINT"),
	)

	client, err := anyi.NewClient("azure-openai", config)
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Analyze the business impact of AI adoption"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Azure OpenAI response: %s", response.Content)
}
```

### Ollama

Ollama enables local deployment of open-source models, perfect for privacy-sensitive applications.

#### Features and Advantages

- Complete offline operation
- Support for Llama, Mixtral, and other open-source models
- Full data privacy control
- No usage fees

#### Configuration

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

	client, err := anyi.NewClient("local-llm", config)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	messages := []chat.Message{
		{Role: "system", Content: "You are a helpful assistant running locally."},
		{Role: "user", Content: "What are the advantages of local AI deployment?"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Local inference failed: %v", err)
	}

	log.Printf("Ollama response: %s", response.Content)
}
```

### Zhipu AI

Zhipu AI provides GLM series models via https://open.bigmodel.cn/, with excellent Chinese language capabilities.

#### Features and Advantages

- Superior Chinese language understanding
- GLM-4 series with strong reasoning
- Cost-effective pricing
- Support for chat and code generation

#### Configuration

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
	// Default configuration with GLM-4-Flash
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")

	// GLM-4 for complex tasks
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")

	// Custom configuration
	config := zhipu.NewConfig(
		os.Getenv("ZHIPU_API_KEY"),
		"glm-4",
		"https://open.bigmodel.cn/api/paas/v4/",
	)

	client, err := anyi.NewClient("zhipu", config)
	if err != nil {
		log.Fatalf("Failed to create Zhipu client: %v", err)
	}

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

### Dashscope (Alibaba Cloud)

Dashscope provides Qwen series models via https://help.aliyun.com/en/dashscope/.

#### Features and Advantages

- Excellent multilingual capabilities
- Strong code generation and math reasoning
- Alibaba Cloud ecosystem integration
- Multimodal support
- Enterprise reliability

#### Configuration

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
	// Default with Qwen-Turbo
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")

	// Qwen-Max for complex reasoning
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")

	// Custom configuration
	config := dashscope.NewConfig(
		os.Getenv("DASHSCOPE_API_KEY"),
		"qwen-plus",
		"https://dashscope.aliyuncs.com/compatible-mode/v1",
	)

	client, err := anyi.NewClient("dashscope", config)
	if err != nil {
		log.Fatalf("Failed to create Dashscope client: %v", err)
	}

	messages := []chat.Message{
		{Role: "system", Content: "You are a professional software engineer."},
		{Role: "user", Content: "Write a Go function to implement binary search"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Dashscope response: %s", response.Content)
}
```

### Anthropic

Anthropic provides Claude models via https://www.anthropic.com/claude.

#### Features and Advantages

- Advanced reasoning and analysis
- Strong AI safety focus
- Excellent long-form content generation
- Large context windows
- High-quality code generation

#### Configuration

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
	// Default configuration
	config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))

	// Specific model
	config := anthropic.DefaultConfigWithModel(
		os.Getenv("ANTHROPIC_API_KEY"),
		"claude-3-sonnet-20240229",
	)

	// Full configuration
	config := anthropic.NewConfig(
		os.Getenv("ANTHROPIC_API_KEY"),
		"claude-3-opus-20240229",
		"https://api.anthropic.com/v1",
		"2023-06-01",
	)

	client, err := anyi.NewClient("anthropic", config)
	if err != nil {
		log.Fatalf("Failed to create Anthropic client: %v", err)
	}

	messages := []chat.Message{
		{Role: "system", Content: "You are a thoughtful analyst."},
		{Role: "user", Content: "Analyze the implications of quantum computing on cryptography"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Anthropic response: %s", response.Content)
}
```

### SiliconCloud

SiliconCloud provides enterprise AI solutions with multiple model access.

#### Features and Advantages

- Enterprise-grade reliability
- Multiple model families
- Competitive pricing
- Custom deployment support
- Strong technical support

#### Configuration

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
	config := siliconcloud.DefaultConfig(
		os.Getenv("SILICONCLOUD_API_KEY"),
		"deepseek-chat",
	)

	// Custom configuration
	config := siliconcloud.NewConfig(
		os.Getenv("SILICONCLOUD_API_KEY"),
		"qwen-7b-chat",
		"https://api.siliconflow.cn/v1",
	)

	client, err := anyi.NewClient("siliconcloud", config)
	if err != nil {
		log.Fatalf("Failed to create SiliconCloud client: %v", err)
	}

	messages := []chat.Message{
		{Role: "system", Content: "You are a business analyst."},
		{Role: "user", Content: "What are key trends in enterprise AI adoption?"},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("SiliconCloud response: %s", response.Content)
}
```

## Using OpenAI-Compatible APIs

Many providers offer OpenAI-compatible APIs, allowing you to use them with Anyi's OpenAI client.

### When to Use OpenAI-Compatible APIs

- **New Providers**: Services without dedicated Anyi modules
- **Custom Deployments**: Self-hosted or private models
- **Testing**: Evaluating new services
- **Flexibility**: Quick switching between compatible providers

### Configuration Examples

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
	// Together AI
	config := openai.NewConfig(
		os.Getenv("TOGETHER_API_KEY"),
		"meta-llama/Llama-2-70b-chat-hf",
		"https://api.together.xyz/v1",
	)

	// Groq
	config := openai.NewConfig(
		os.Getenv("GROQ_API_KEY"),
		"mixtral-8x7b-32768",
		"https://api.groq.com/openai/v1",
	)

	// Custom service
	config := openai.NewConfig(
		os.Getenv("CUSTOM_API_KEY"),
		"custom-model-name",
		"https://api.custom-provider.com/v1",
	)

	client, err := anyi.NewClient("compatible-provider", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

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

### Common Compatible Providers

- **Together AI**: Various open-source models
- **Groq**: High-speed inference
- **Perplexity AI**: Search-augmented models
- **Fireworks AI**: Fast open-source model serving
- **Anyscale**: Ray-based model platform
- **Local deployments**: vLLM, Text Generation Inference

## Multiple Client Management

Use different providers for different tasks to optimize cost and performance:

```go
package main

import (
	"log"
	"os"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/ollama"
)

func main() {
	// Fast, cheap model for simple tasks
	quickConfig := openai.NewConfigWithModel(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4o-mini",
	)
	quickClient, _ := anyi.NewClient("quick", quickConfig)

	// Powerful model for complex tasks
	powerConfig := openai.NewConfigWithModel(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4",
	)
	powerClient, _ := anyi.NewClient("power", powerConfig)

	// Local model for privacy-sensitive tasks
	localConfig := ollama.DefaultConfig("llama3")
	localClient, _ := anyi.NewClient("local", localConfig)

	// Use appropriate client based on task complexity
	simpleTask := "What is 2+2?"
	quickResponse, _, _ := quickClient.Chat([]chat.Message{
		{Role: "user", Content: simpleTask},
	}, nil)

	complexTask := "Analyze the economic implications of AI automation"
	powerResponse, _, _ := powerClient.Chat([]chat.Message{
		{Role: "user", Content: complexTask},
	}, nil)

	privateTask := "Analyze this confidential business data: [data]"
	localResponse, _, _ := localClient.Chat([]chat.Message{
		{Role: "user", Content: privateTask},
	}, nil)
}
```

## Provider Selection Guidelines

### By Use Case

- **General Chat**: OpenAI GPT-4o-mini, Zhipu GLM-4-Flash
- **Complex Reasoning**: OpenAI GPT-4o, DeepSeek-Reasoner, Anthropic Claude-3.5-Sonnet
- **Code Generation**: Qwen2.5-Coder, GPT-4o, Claude-3.5-Sonnet
- **Chinese Language**: Zhipu GLM-4-Plus, Qwen2.5-Max, DeepSeek-Chat
- **Privacy/Offline**: Ollama with Llama-3.2, Qwen2.5 local models
- **Enterprise**: Azure OpenAI GPT-4o, SiliconCloud

### By Performance Requirements

- **Speed**: GPT-4o-mini, GLM-4-Flash, Qwen2.5-Turbo
- **Quality**: GPT-4o, Claude-3.5-Sonnet, DeepSeek-Reasoner
- **Cost**: GPT-4o-mini, GLM-4-Flash, local models
- **Reliability**: Azure OpenAI, enterprise providers

### By Deployment Constraints

- **Cloud**: Any cloud provider
- **On-premises**: Ollama, Azure OpenAI (private)
- **Air-gapped**: Ollama only
- **Compliance**: Azure OpenAI, enterprise providers

## Best Practices

### Configuration Management

```go
// Use environment variables for sensitive data
config.APIKey = os.Getenv("API_KEY")

// Set reasonable defaults
config.Temperature = 0.7
config.MaxTokens = 1000

// Configure timeouts for production
config.Timeout = 30 * time.Second
```

### Error Handling

```go
response, info, err := client.Chat(messages, nil)
if err != nil {
    log.Printf("Chat failed: %v", err)
    // Implement fallback logic
    return
}

log.Printf("Used %d tokens", info.PromptTokens + info.CompletionTokens)
```

### Token Management

```go
// Monitor token usage
response, info, err := client.Chat(messages, nil)
if err == nil {
    totalTokens := info.PromptTokens + info.CompletionTokens
    log.Printf("Request used %d tokens", totalTokens)
}
```

## Next Steps

Now that you understand LLM clients:

1. **Build Workflows**: Learn to create complex workflows in [Building Workflows](workflows.md)
2. **Configuration**: Master configuration management in [Configuration Management](configuration.md)
3. **Multimodal**: Explore image and text processing in [Multimodal Applications](multimodal.md)
4. **Production**: Check out [Performance Optimization](../how-to/performance.md) for production tips
