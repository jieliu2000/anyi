# Basic Concepts

This guide introduces the fundamental concepts in Anyi that you'll encounter when building AI applications. Understanding these concepts will help you design better AI workflows and make the most of Anyi's features.

## Core Architecture

Anyi is built around several key concepts that work together to create powerful AI applications:

### Clients

**Clients** are the foundation of Anyi. They provide a unified interface to interact with different Large Language Model (LLM) providers.

```go
// Create a client for OpenAI
config := openai.DefaultConfig(apiKey)
client, err := anyi.NewClient("openai", config)

// Create a client for Anthropic
anthropicConfig := anthropic.DefaultConfig(anthropicKey)
anthropicClient, err := anyi.NewClient("anthropic", anthropicConfig)
```

Key features of clients:

- **Unified Interface**: All clients implement the same interface, making it easy to switch providers
- **Named Registry**: Clients can be registered with names for easy retrieval throughout your application
- **Provider Abstraction**: Hide the complexity of different API formats behind a consistent interface

### Messages

**Messages** represent the conversation between users and AI models. They follow the standard chat format used by most LLM providers.

```go
messages := []chat.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "What is machine learning?"},
    {Role: "assistant", Content: "Machine learning is..."},
    {Role: "user", Content: "Can you give me an example?"},
}
```

Message roles:

- **system**: Sets the behavior and context for the AI
- **user**: Represents input from the human user
- **assistant**: Represents responses from the AI model

### Flows and Workflows

**Flows** are sequences of steps that process data through multiple stages. They're the core of Anyi's workflow system.

```go
// Create a flow with multiple steps
step1 := anyi.NewLLMStepWithTemplate("Analyze: {{.Text}}", "You are an analyst", client)
step2 := anyi.NewLLMStepWithTemplate("Summarize: {{.Text}}", "You are a summarizer", client)

flow, err := anyi.NewFlow("analysis_flow", client, *step1, *step2)
```

Flows enable:

- **Sequential Processing**: Data flows from one step to the next
- **Error Handling**: Built-in retry mechanisms and error recovery
- **Validation**: Ensure outputs meet quality standards before proceeding

### Steps

**Steps** are individual units of work within a flow. Each step has an executor that performs the actual work.

```go
step := flow.Step{
    Name: "content_analysis",
    Executor: &anyi.LLMExecutor{
        Template: "Analyze the sentiment of: {{.Text}}",
        SystemMessage: "You are a sentiment analysis expert.",
    },
    Validator: &anyi.StringValidator{
        MinLength: 10,
    },
    MaxRetryTimes: 3,
}
```

Step components:

- **Executor**: Performs the actual work (LLM call, command execution, etc.)
- **Validator**: Checks if the output meets requirements
- **Retry Logic**: Automatically retries failed steps

### Executors

**Executors** are components that perform specific tasks within workflow steps. Anyi provides several built-in executors:

#### LLMExecutor

Sends prompts to language models:

```go
executor := &anyi.LLMExecutor{
    Template: "Translate to French: {{.Text}}",
    SystemMessage: "You are a professional translator.",
}
```

#### SetContextExecutor

Modifies the workflow context:

```go
executor := &anyi.SetContextExecutor{
    Text: "Starting analysis of user input",
}
```

#### ConditionalFlowExecutor

Implements branching logic:

```go
executor := &anyi.ConditionalFlowExecutor{
    Condition: "{{.Text}} contains 'urgent'",
    TrueFlow: urgentFlow,
    FalseFlow: normalFlow,
}
```

#### RunCommandExecutor

Executes system commands:

```go
executor := &anyi.RunCommandExecutor{
    Command: "python analyze.py {{.Text}}",
}
```

### Validators

**Validators** ensure that step outputs meet specific criteria before proceeding to the next step.

#### StringValidator

Validates text output:

```go
validator := &anyi.StringValidator{
    MinLength: 50,
    MaxLength: 500,
    MatchRegex: `\d{3}-\d{3}-\d{4}`, // Phone number format
}
```

#### JsonValidator

Ensures valid JSON output:

```go
validator := &anyi.JsonValidator{
    RequiredFields: []string{"name", "email", "phone"},
}
```

### Flow Context

**FlowContext** carries data between workflow steps. It contains several important fields:

```go
type FlowContext struct {
    Text      string                 // Main text content
    Memory    interface{}           // Structured data storage
    ImageURLs []string             // Image URLs for multimodal content
    Think     string               // Captured thinking process
    Flow      *Flow                // Reference to current flow
}
```

#### Using Memory for Structured Data

The Memory field allows you to pass complex data structures between steps:

```go
type TaskData struct {
    Objective string
    Steps     []string
    Progress  int
}

// Initialize context with structured data
taskData := TaskData{
    Objective: "Write a blog post",
    Steps:     []string{"Research", "Outline", "Write", "Edit"},
    Progress:  0,
}

flowContext := anyi.NewFlowContextWithMemory(taskData)
```

#### Template Variables

In templates, you can access FlowContext properties:

```go
template := `
Current task: {{.Memory.Objective}}
Progress: {{.Memory.Progress}}/{{len .Memory.Steps}}
Current text: {{.Text}}
Previous thinking: {{.Think}}
`
```

## Configuration System

Anyi supports multiple configuration approaches:

### Code-Based Configuration

Define everything in Go code:

```go
config := openai.DefaultConfig(apiKey)
client, _ := anyi.NewClient("openai", config)

step1, _ := anyi.NewLLMStepWithTemplate("Process: {{.Text}}", "", client)
flow, _ := anyi.NewFlow("my_flow", client, *step1)
```

### File-Based Configuration

Use YAML, JSON, or TOML files:

```yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"

flows:
  - name: "my_flow"
    clientName: "openai"
    steps:
      - name: "process"
        executor:
          type: "llm"
          withconfig:
            template: "Process: {{.Text}}"
```

### Environment Variables

Reference environment variables in configuration:

```yaml
config:
  apiKey: "$OPENAI_API_KEY" # References environment variable
  baseURL: "$CUSTOM_BASE_URL"
```

## Best Practices

### Client Management

- Use named clients for components that will be reused
- Create unregistered clients for one-off tasks
- Choose appropriate models for different task complexities

### Flow Design

- Keep steps focused on single responsibilities
- Use validators to ensure quality outputs
- Implement proper error handling with retries

### Template Design

- Keep templates clear and specific
- Use system messages to set proper context
- Test templates with various inputs

### Memory Usage

- Use Memory for structured data that needs to persist across steps
- Keep Memory objects reasonably sized
- Document the structure of your Memory objects

## Common Patterns

### Multi-Step Analysis

```go
// Step 1: Extract key information
extractStep := anyi.NewLLMStepWithTemplate(
    "Extract key facts from: {{.Text}}",
    "You are a fact extraction expert.",
    client,
)

// Step 2: Analyze the extracted facts
analyzeStep := anyi.NewLLMStepWithTemplate(
    "Analyze these facts: {{.Text}}",
    "You are a data analyst.",
    client,
)

// Step 3: Generate recommendations
recommendStep := anyi.NewLLMStepWithTemplate(
    "Based on this analysis, provide recommendations: {{.Text}}",
    "You are a strategic advisor.",
    client,
)
```

### Content Pipeline

```go
// Research -> Write -> Edit -> Publish
researchStep := anyi.NewLLMStepWithTemplate("Research: {{.Text}}", "", client)
writeStep := anyi.NewLLMStepWithTemplate("Write article: {{.Text}}", "", client)
editStep := anyi.NewLLMStepWithTemplate("Edit and improve: {{.Text}}", "", client)
```

### Decision Making

```go
conditionalExecutor := &anyi.ConditionalFlowExecutor{
    Condition: "{{.Text}} contains 'urgent'",
    TrueFlow:  urgentProcessingFlow,
    FalseFlow: normalProcessingFlow,
}
```

## Next Steps

Now that you understand Anyi's core concepts:

1. **Try the Tutorials**: Start with [LLM Clients](../tutorials/llm-clients.md) to learn about different providers
2. **Build Workflows**: Learn to create complex workflows in [Building Workflows](../tutorials/workflows.md)
3. **Explore Configuration**: Master configuration in [Configuration Management](../tutorials/configuration.md)
4. **Advanced Topics**: When ready, explore [Custom Executors](../advanced/custom-executors.md)

Understanding these concepts will help you build more effective and maintainable AI applications with Anyi.
