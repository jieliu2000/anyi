# API Reference

> **📚 For the most up-to-date and comprehensive API documentation, visit: [pkg.go.dev/github.com/jieliu2000/anyi](https://pkg.go.dev/github.com/jieliu2000/anyi)**

This document provides a comprehensive reference for the Anyi framework's public API. It covers the core interfaces, methods, and data structures you'll use when building AI applications.

## Core Interfaces

### Client Interface

The `Client` interface is the primary way to interact with LLM providers.

```go
type Client interface {
    Chat(messages []chat.Message, options *chat.Options) (*chat.Message, chat.ResponseInfo, error)
    GetProvider() string
    GetModel() string
}
```

#### Methods

##### Chat

```go
Chat(messages []chat.Message, options *chat.Options) (*chat.Message, chat.ResponseInfo, error)
```

Sends a chat request to the LLM provider.

**Parameters:**

- `messages`: Array of chat messages forming the conversation
- `options`: Optional configuration for the request

**Returns:**

- `*chat.Message`: The response message from the LLM
- `chat.ResponseInfo`: Metadata about the response (tokens used, model info, etc.)
- `error`: Error if the request fails

##### GetProvider

```go
GetProvider() string
```

Returns the name of the LLM provider (e.g., "openai", "anthropic").

##### GetModel

```go
GetModel() string
```

Returns the model name being used (e.g., "gpt-4", "claude-3-opus").

## Data Structures

### chat.Message

Represents a single message in a conversation.

```go
type Message struct {
    Role     string      `json:"role"`
    Content  string      `json:"content"`
    Images   []string    `json:"images,omitempty"`
    Function *Function   `json:"function,omitempty"`
}
```

**Fields:**

- `Role`: The role of the message sender ("user", "assistant", "system")
- `Content`: The text content of the message
- `Images`: Array of image URLs for multimodal messages
- `Function`: Function call information (for function calling)

### chat.Options

Configuration options for chat requests.

```go
type Options struct {
    Temperature      *float64    `json:"temperature,omitempty"`
    MaxTokens        *int        `json:"max_tokens,omitempty"`
    TopP             *float64    `json:"top_p,omitempty"`
    FrequencyPenalty *float64    `json:"frequency_penalty,omitempty"`
    PresencePenalty  *float64    `json:"presence_penalty,omitempty"`
    Stop             []string    `json:"stop,omitempty"`
    Functions        []Function  `json:"functions,omitempty"`
}
```

**Fields:**

- `Temperature`: Controls randomness (0.0-2.0, default varies by provider)
- `MaxTokens`: Maximum number of tokens to generate
- `TopP`: Nucleus sampling parameter (0.0-1.0)
- `FrequencyPenalty`: Penalty for token frequency (-2.0 to 2.0)
- `PresencePenalty`: Penalty for token presence (-2.0 to 2.0)
- `Stop`: Array of stop sequences
- `Functions`: Available functions for function calling

### chat.ResponseInfo

Contains metadata about the LLM response.

```go
type ResponseInfo struct {
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
    Model            string `json:"model"`
    Provider         string `json:"provider"`
}
```

**Fields:**

- `PromptTokens`: Number of tokens in the input
- `CompletionTokens`: Number of tokens in the response
- `TotalTokens`: Total tokens used (prompt + completion)
- `Model`: The model that generated the response
- `Provider`: The provider that handled the request

## Client Management Functions

### anyi.NewClient

```go
func NewClient(name string, config interface{}) (Client, error)
```

Creates a new named client and registers it in the global registry.

**Parameters:**

- `name`: Unique name for the client
- `config`: Provider-specific configuration

**Returns:**

- `Client`: The created client instance
- `error`: Error if client creation fails

### anyi.GetClient

```go
func GetClient(name string) (Client, error)
```

Retrieves a previously registered client by name.

**Parameters:**

- `name`: Name of the client to retrieve

**Returns:**

- `Client`: The client instance
- `error`: Error if client not found

### anyi.ListClients

```go
func ListClients() []string
```

Returns a list of all registered client names.

## Configuration Functions

### anyi.Config

```go
func Config(config *AnyiConfig) error
```

Applies a configuration to set up clients and flows.

**Parameters:**

- `config`: Configuration structure containing clients and flows

**Returns:**

- `error`: Error if configuration fails

### anyi.ConfigFromFile

```go
func ConfigFromFile(filename string) error
```

Loads configuration from a file (supports YAML, JSON, TOML).

**Parameters:**

- `filename`: Path to the configuration file

**Returns:**

- `error`: Error if loading fails

## Flow Management

### Flow Interface

```go
type Flow interface {
    Run() (*FlowContext, error)
    RunWithInput(input interface{}) (*FlowContext, error)
    GetName() string
}
```

#### Methods

##### Run

```go
Run() (*FlowContext, error)
```

Executes the flow without initial input.

##### RunWithInput

```go
RunWithInput(input interface{}) (*FlowContext, error)
```

Executes the flow with the provided input.

**Parameters:**

- `input`: Initial input for the flow (string or structured data)

**Returns:**

- `*FlowContext`: The final flow context after execution
- `error`: Error if flow execution fails

##### GetName

```go
GetName() string
```

Returns the name of the flow.

### anyi.GetFlow

```go
func GetFlow(name string) (Flow, error)
```

Retrieves a configured flow by name.

**Parameters:**

- `name`: Name of the flow to retrieve

**Returns:**

- `Flow`: The flow instance
- `error`: Error if flow not found

## Flow Context

### FlowContext Structure

```go
type FlowContext struct {
    Text   string
    Memory interface{}
    Think  string
    Images []string
}
```

**Fields:**

- `Text`: Current text content being processed
- `Memory`: Structured data for complex workflows
- `Think`: Extracted thinking process from LLM responses
- `Images`: Array of image URLs for multimodal processing

### Context Creation Functions

#### anyi.NewFlowContext

```go
func NewFlowContext(text string) *FlowContext
```

Creates a new flow context with initial text.

#### anyi.NewFlowContextWithMemory

```go
func NewFlowContextWithMemory(memory interface{}) *FlowContext
```

Creates a new flow context with structured memory data.

## Error Types

### Common Error Types

The framework defines several error types for different scenarios:

- `ErrClientNotFound`: Returned when a requested client is not registered
- `ErrFlowNotFound`: Returned when a requested flow is not found
- `ErrInvalidConfig`: Returned when configuration is invalid
- `ErrProviderNotSupported`: Returned when an unsupported provider is specified

## Provider-Specific Configurations

### OpenAI Configuration

```go
type Config struct {
    APIKey      string
    Model       string
    BaseURL     string
    OrgID       string
    Temperature *float64
    MaxTokens   *int
}
```

### Anthropic Configuration

```go
type Config struct {
    APIKey      string
    Model       string
    BaseURL     string
    Version     string
    Temperature *float64
    MaxTokens   *int
}
```

### Ollama Configuration

```go
type Config struct {
    Model   string
    BaseURL string
    Options map[string]interface{}
}
```

## Function Calling

### Function Structure

```go
type Function struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}
```

**Fields:**

- `Name`: Function name
- `Description`: Human-readable description
- `Parameters`: JSON Schema defining the function parameters

## Best Practices

### Error Handling

Always check for errors when calling API methods:

```go
client, err := anyi.GetClient("my-client")
if err != nil {
    log.Fatalf("Failed to get client: %v", err)
}

response, info, err := client.Chat(messages, options)
if err != nil {
    log.Printf("Chat failed: %v", err)
    return
}
```

### Resource Management

- Reuse clients when possible to avoid unnecessary initialization overhead
- Use appropriate timeouts for long-running operations
- Implement proper retry logic for transient failures

### Configuration

- Use environment variables for sensitive information like API keys
- Validate configuration before using it in production
- Use structured configuration files for complex setups

This API reference covers the essential interfaces and functions you'll use when working with Anyi. For implementation examples and tutorials, see the other documentation sections.
