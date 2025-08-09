# Frequently Asked Questions (FAQ)

This document answers common questions about using Anyi. If you don't find your question here, please check our [GitHub Discussions](https://github.com/jieliu2000/anyi/discussions) or open an issue.

## Getting Started

### Q: What are the system requirements for Anyi?

**A:** Anyi requires:

- Go 1.20 or higher
- Internet connectivity for accessing LLM APIs (unless using local models)
- Sufficient memory for your chosen models (especially for local deployment)

### Q: Can I use Anyi without an internet connection?

**A:** Yes, but only with local models via Ollama. You'll need to:

1. Install Ollama
2. Download models locally
3. Use only the Ollama provider in your Anyi applications

### Q: Which LLM provider should I choose?

**A:** It depends on your needs:

- **General use**: OpenAI GPT-4o-mini for cost-effectiveness, GPT-4o for quality
- **Chinese language**: Zhipu AI or Dashscope
- **Code generation**: DeepSeek Chat or Dashscope Qwen-Max
- **Privacy/offline**: Ollama with local models
- **Enterprise**: Azure OpenAI or SiliconCloud

## Configuration and Setup

### Q: How do I manage API keys securely?

**A:** Best practices for API key management:

```go
// Use environment variables
config.APIKey = os.Getenv("OPENAI_API_KEY")

// For development, use .env files
// For production, use cloud secret managers
```

Never hardcode API keys in your source code.

### Q: Can I use multiple LLM providers in the same application?

**A:** Yes! This is one of Anyi's key features:

```go
// Create multiple clients
openaiClient, _ := anyi.NewClient("openai", openaiConfig)
ollamaClient, _ := anyi.NewClient("ollama", ollamaConfig)

// Use different clients for different tasks
quickResponse, _, _ := openaiClient.Chat(simpleMessages, nil)
privateResponse, _, _ := ollamaClient.Chat(sensitiveMessages, nil)
```

### Q: How do I switch between different models from the same provider?

**A:** Create separate clients with different configurations:

```go
// Fast model for simple tasks
fastConfig := openai.NewConfigWithModel(apiKey, "gpt-4o-mini")
fastClient, _ := anyi.NewClient("fast", fastConfig)

// Powerful model for complex tasks
powerConfig := openai.NewConfigWithModel(apiKey, "gpt-4")
powerClient, _ := anyi.NewClient("power", powerConfig)
```

## Workflows and Development

### Q: What are the VarsImmutable, TextImmutable, and MemoryImmutable properties in Step?

**A:** These properties control how a step modifies the context during execution:

- **TextImmutable**: When set to true, the step won't modify the context text. This is useful when you want to preserve the original text for multiple analyses.

```yaml
steps:
  - name: "analysis_step"
    textImmutable: true  # Text content will remain unchanged
    executor:
      type: "llm"
      withconfig:
        template: "Analyze this text: {{.Text}}"
```

- **MemoryImmutable**: When set to true, the step won't modify the context memory. Use this when you need to preserve structured data.

```yaml
steps:
  - name: "query_step"
    memoryImmutable: true  # Memory data will remain unchanged
    executor:
      type: "llm"
      withconfig:
        template: "Answer based on this data: {{.Memory}}"
```

- **VarsImmutable**: When set to true, the step won't modify context variables. This is useful when you need to preserve variable state.

These properties are particularly useful in multi-step workflows where you want certain steps to read data without modifying it.

### Q: How do I handle large texts that exceed token limits?

**A:** Several approaches:

1. **Chunking**: Split large texts into smaller pieces

```go
func processLargeText(text string, client *llm.Client) (string, error) {
    chunks := splitIntoChunks(text, 1000) // ~1000 words per chunk
    var results []string

    for _, chunk := range chunks {
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: "Process: " + chunk},
        }, nil)
        if err != nil {
            return "", err
        }
        results = append(results, response.Content)
    }

    return combineResults(results), nil
}
```

2. **Summarization**: Use a multi-step workflow to summarize first, then process
3. **Streaming**: For providers that support it, use streaming responses

### Q: How do I ensure workflows work properly in unstable network conditions?

**A:** Use Anyi's built-in retry mechanisms:

```go
step := flow.Step{
    Name: "robust_step",
    Executor: myExecutor,
    MaxRetryTimes: 3,  // Retry up to 3 times
    Validator: myValidator,
}
```

You can also implement exponential backoff:

```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    backoff := 1 * time.Second

    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if i < maxRetries-1 {
            time.Sleep(backoff)
            backoff *= 2 // Exponential backoff
        }
    }

    return fmt.Errorf("failed after %d retries", maxRetries)
}
```

### Q: How do I integrate Anyi with existing Go web frameworks?

**A:** Anyi integrates seamlessly with any Go web framework. Example with Gin:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jieliu2000/anyi"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // Initialize Anyi client
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, _ := anyi.NewClient("web", config)

    r.POST("/ask", func(c *gin.Context) {
        var req struct {
            Question string `json:"question"`
        }

        if err := c.BindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

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

## Performance and Optimization

### Q: How can I optimize performance and reduce costs?

**A:** Several strategies:

1. **Choose appropriate models**: Use cheaper models for simple tasks
2. **Implement caching**: Cache common responses
3. **Set token limits**: Use `MaxTokens` to control response length
4. **Use local models**: For frequent, non-critical tasks
5. **Batch processing**: Group similar requests together

```go
// Example: Tiered approach
func processRequest(complexity string, content string) (*chat.Message, error) {
    var client *llm.Client

    switch complexity {
    case "simple":
        client, _ = anyi.GetClient("fast")     // GPT-4o-mini
    case "complex":
        client, _ = anyi.GetClient("power")    // GPT-4
    case "private":
        client, _ = anyi.GetClient("local")    // Ollama
    }

    return client.Chat([]chat.Message{
        {Role: "user", Content: content},
    }, nil)
}
```

### Q: How do I monitor token usage and costs?

**A:** Track usage through response info:

```go
response, info, err := client.Chat(messages, nil)
if err == nil {
    totalTokens := info.PromptTokens + info.CompletionTokens
    log.Printf("Request used %d tokens (prompt: %d, completion: %d)",
        totalTokens, info.PromptTokens, info.CompletionTokens)

    // Estimate cost (example for OpenAI)
    cost := estimateOpenAICost(info.Model, info.PromptTokens, info.CompletionTokens)
    log.Printf("Estimated cost: $%.4f", cost)
}
```

## Troubleshooting

### Q: I'm getting "Failed to create client" errors. What should I check?

**A:** Common causes and solutions:

1. **API Key Issues**:

   - Verify the API key is correct
   - Check if the key has necessary permissions
   - Ensure billing is set up (for paid services)

2. **Network Issues**:

   - Check internet connectivity
   - Verify firewall settings
   - Try using a different base URL if available

3. **Configuration Issues**:
   - Check that all required environment variables are set
   - Verify the model name is correct
   - For Azure: ensure the model is deployed

### Q: My requests are failing with rate limit errors. How do I handle this?

**A:** Implement rate limiting and backoff:

```go
func handleRateLimit(client *llm.Client, messages []chat.Message) (*chat.Message, error) {
    maxRetries := 3
    baseDelay := 1 * time.Second

    for i := 0; i < maxRetries; i++ {
        response, _, err := client.Chat(messages, nil)

        if err == nil {
            return response, nil
        }

        // Check if it's a rate limit error
        if strings.Contains(err.Error(), "rate limit") {
            delay := baseDelay * time.Duration(1<<i) // Exponential backoff
            log.Printf("Rate limited, waiting %v before retry %d", delay, i+1)
            time.Sleep(delay)
            continue
        }

        return nil, err // Non-rate-limit error
    }

    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

### Q: How do I debug workflow execution issues?

**A:** Enable detailed logging and use step validation:

```go
// Add logging to your executors
type LoggingExecutor struct {
    *anyi.LLMExecutor
    Logger *log.Logger
}

func (e *LoggingExecutor) Run(ctx flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    e.Logger.Printf("Executing step: %s with input: %s", step.Name, ctx.Text)

    result, err := e.LLMExecutor.Run(ctx, step)

    if err != nil {
        e.Logger.Printf("Step %s failed: %v", step.Name, err)
        return nil, err
    }

    e.Logger.Printf("Step %s completed: %s", step.Name, result.Text)
    return result, nil
}
```

## Advanced Usage

### Q: Can I create custom executors?

**A:** Yes! Implement the `Executor` interface:

```go
type CustomExecutor struct {
    // Your fields
}

func (e *CustomExecutor) Run(ctx flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Your custom logic
    result := processCustomLogic(ctx.Text)

    return &flow.FlowContext{
        Text: result,
        Memory: ctx.Memory,
        // ... other fields
    }, nil
}

func (e *CustomExecutor) Init() error {
    // Initialization logic
    return nil
}
```

### Q: How do I implement conditional logic in workflows?

**A:** Use the `ConditionalFlowExecutor`:

```go
conditionalExecutor := &anyi.ConditionalFlowExecutor{
    Condition: "{{.Text}} contains 'urgent'",
    TrueFlow:  urgentProcessingFlow,
    FalseFlow: normalProcessingFlow,
}

step := flow.Step{
    Name: "routing",
    Executor: conditionalExecutor,
}
```

### Q: Can I use Anyi in microservices architecture?

**A:** Absolutely! Each microservice can have its own Anyi configuration:

```go
// Service A: Text processing service
func main() {
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, _ := anyi.NewClient("text-processor", config)

    // Set up HTTP handlers that use the client
    setupTextProcessingHandlers(client)
}

// Service B: Image analysis service
func main() {
    config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4-vision")
    client, _ := anyi.NewClient("image-analyzer", config)

    setupImageAnalysisHandlers(client)
}
```

## Community and Support

### Q: Where can I get help if I'm stuck?

**A:** Several options:

1. **Documentation**: Check this documentation and the [API reference](api.md)
2. **Examples**: Browse the [examples directory](../../../examples/)
3. **GitHub Issues**: Search existing [issues](https://github.com/jieliu2000/anyi/issues) or create a new one
4. **Discussions**: Join [GitHub Discussions](https://github.com/jieliu2000/anyi/discussions)
5. **Community**: Connect with other users in the community

### Q: How can I contribute to Anyi?

**A:** We welcome contributions! You can:

1. **Report bugs**: Open issues for any bugs you find
2. **Suggest features**: Propose new features or improvements
3. **Submit code**: Create pull requests for bug fixes or new features
4. **Improve documentation**: Help make the documentation better
5. **Share examples**: Contribute example applications

### Q: Is Anyi suitable for production use?

**A:** Yes! Anyi is designed for production use with features like:

- Robust error handling and retry mechanisms
- Support for multiple LLM providers
- Configuration management for different environments
- Comprehensive testing and validation
- Performance optimization capabilities

Many users successfully run Anyi in production environments.

## Still Have Questions?

If your question isn't answered here:

1. Search the [GitHub repository](https://github.com/jieliu2000/anyi)
2. Check [GitHub Discussions](https://github.com/jieliu2000/anyi/discussions)
3. Open a new [issue](https://github.com/jieliu2000/anyi/issues/new) with detailed information
4. Browse the [examples](../../../examples/) for code samples
