# Anyi Tutorial - Getting Started

Welcome to Anyi! This tutorial will guide you through the essential concepts and help you build your first AI application with the Anyi framework.

## 📚 Documentation Structure

We've organized the Anyi documentation to make it easier to find what you need:

### 🚀 New to Anyi? Start Here

**[📖 Complete Documentation Hub →](README.md)**

For the full documentation experience, visit our main documentation hub which provides organized learning paths for different user needs.

### ⚡ Quick Start (5 minutes)

If you want to jump right in:

1. **[Installation →](getting-started/installation.md)** - Set up Anyi on your system
2. **[Quick Start Guide →](getting-started/quickstart.md)** - Build your first AI app in 5 minutes
3. **[Core Concepts →](getting-started/concepts.md)** - Understand how Anyi works

### 📋 Learning Path

Follow this structured learning path:

#### Step 1: Foundation

- [Installation and Setup](getting-started/installation.md)
- [Basic Concepts](getting-started/concepts.md)
- [Your First Application](getting-started/quickstart.md)

#### Step 2: Core Skills

- [Working with LLM Clients](tutorials/llm-clients.md) - Connect to OpenAI, Anthropic, Ollama, and more
- [Building Workflows](tutorials/workflows.md) - Create multi-step AI processes
- [Configuration Management](tutorials/configuration.md) - Organize your settings

#### Step 3: Advanced Features

- [Multimodal Applications](tutorials/multimodal.md) - Work with text and images
- [Error Handling](how-to/error-handling.md) - Build robust applications
- [Performance Optimization](how-to/performance.md) - Speed and cost optimization

#### Step 4: Production Ready

- [Web Integration](how-to/web-integration.md) - Use with web frameworks
- [Security Best Practices](advanced/security.md) - Secure your applications
- [Production Deployment](advanced/deployment.md) - Deploy to production

## 🎯 Common Use Cases

### I want to...

**Connect to an LLM provider**
→ [Provider Setup Guide](how-to/provider-setup.md)

**Build a multi-step AI workflow**
→ [Building Workflows Tutorial](tutorials/workflows.md)

**Handle errors and retries**
→ [Error Handling Guide](how-to/error-handling.md)

**Work with images and text**
→ [Multimodal Applications](tutorials/multimodal.md)

**Integrate with my web app**
→ [Web Integration Guide](how-to/web-integration.md)

**Optimize performance and costs**
→ [Performance Optimization](how-to/performance.md)

**Deploy to production**
→ [Deployment Guide](advanced/deployment.md)

**Create custom components**
→ [Custom Executors Guide](advanced/custom-executors.md)

## 📖 Reference Materials

When you need to look up specific information:

- **[API Reference](https://pkg.go.dev/github.com/jieliu2000/anyi)** - Complete API documentation
- **[Configuration Reference](reference/configuration.md)** - All configuration options
- **[Components Reference](reference/components.md)** - Built-in executors and validators
- **[FAQ](reference/faq.md)** - Frequently asked questions

## 🔍 Looking for the Original Tutorial?

The comprehensive tutorial content has been reorganized into focused, modular guides for better navigation and learning. If you need the original single-page tutorial, it's available as:

**[Legacy Tutorial →](tutorial-legacy.md)**

However, we recommend using the new modular structure above for a better learning experience.

## 💡 Quick Examples

### Simple Chat Example

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
    // Create client
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, err := anyi.NewClient("openai", config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Send message
    messages := []chat.Message{
        {Role: "user", Content: "Hello, how are you?"},
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("Chat failed: %v", err)
    }

    fmt.Println("Response:", response.Content)
}
```

### Configuration-Based Workflow

```yaml
# config.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"

flows:
  - name: "content_processor"
    clientName: "openai"
    steps:
      - name: "analyze"
        executor:
          type: "llm"
          withconfig:
            template: "Analyze the following text: {{.Text}}"
        validator:
          type: "string"
          withconfig:
            minLength: 50
```

```go
// Load configuration and run workflow
func main() {
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    flow, err := anyi.GetFlow("content_processor")
    if err != nil {
        log.Fatal(err)
    }

    result, err := flow.RunWithInput("Your text here...")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Result:", result.Text)
}
```

## 🤝 Getting Help

- **Questions?** Check the [FAQ](reference/faq.md)
- **Issues?** Open an issue on [GitHub](https://github.com/jieliu2000/anyi/issues)
- **Examples?** Browse the [examples directory](../../examples/)

## 🚀 Ready to Start?

Choose your path:

- **Beginner**: Start with [Installation](getting-started/installation.md)
- **Experienced**: Jump to [LLM Clients](tutorials/llm-clients.md)
- **Reference**: Browse [API Documentation](https://pkg.go.dev/github.com/jieliu2000/anyi)

Happy coding with Anyi! 🎉
