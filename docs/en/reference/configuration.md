# Configuration Reference

This document provides a comprehensive reference for configuring Anyi applications. It covers all configuration options, file formats, and best practices for managing settings.

## Configuration Methods

Anyi supports multiple ways to configure your application:

1. **Programmatic Configuration**: Define configuration in code
2. **Configuration Files**: Use YAML, JSON, or TOML files
3. **Environment Variables**: Override settings with environment variables
4. **Hybrid Approach**: Combine multiple methods for flexibility

## Configuration Structure

### Root Configuration

The root configuration structure contains clients and flows:

```go
type AnyiConfig struct {
    Clients []llm.ClientConfig `yaml:"clients" json:"clients" toml:"clients"`
    Flows   []FlowConfig       `yaml:"flows" json:"flows" toml:"flows"`
}
```

## Client Configuration

### ClientConfig Structure

```go
type ClientConfig struct {
    Name   string                 `yaml:"name" json:"name" toml:"name"`
    Type   string                 `yaml:"type" json:"type" toml:"type"`
    Config map[string]interface{} `yaml:"config" json:"config" toml:"config"`
}
```

**Fields:**

- `Name`: Unique identifier for the client
- `Type`: Provider type (openai, anthropic, ollama, etc.)
- `Config`: Provider-specific configuration options

### Provider-Specific Configurations

#### OpenAI Configuration

```yaml
clients:
  - name: "openai-gpt4"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"
      baseURL: "https://api.openai.com/v1" # Optional
      orgID: "$OPENAI_ORG_ID" # Optional
      temperature: 0.7 # Optional
      maxTokens: 2000 # Optional
```

**Configuration Options:**

- `apiKey` (required): OpenAI API key
- `model` (required): Model name (gpt-4o, gpt-4o-mini, etc.)
- `baseURL` (optional): Custom API endpoint
- `orgID` (optional): Organization ID
- `temperature` (optional): Default temperature (0.0-2.0)
- `maxTokens` (optional): Default max tokens

#### Anthropic Configuration

```yaml
clients:
  - name: "claude"
    type: "anthropic"
    config:
      apiKey: "$ANTHROPIC_API_KEY"
      model: "claude-3-opus-20240229"
      baseURL: "https://api.anthropic.com" # Optional
      version: "2023-06-01" # Optional
      temperature: 0.5 # Optional
      maxTokens: 1000 # Optional
```

**Configuration Options:**

- `apiKey` (required): Anthropic API key
- `model` (required): Model name
- `baseURL` (optional): Custom API endpoint
- `version` (optional): API version
- `temperature` (optional): Default temperature
- `maxTokens` (optional): Default max tokens

#### Azure OpenAI Configuration

```yaml
clients:
  - name: "azure-openai"
    type: "azure"
    config:
      apiKey: "$AZURE_OPENAI_API_KEY"
      endpoint: "$AZURE_OPENAI_ENDPOINT"
      deploymentName: "gpt-4-deployment"
      apiVersion: "2023-12-01-preview" # Optional
      temperature: 0.7 # Optional
      maxTokens: 2000 # Optional
```

**Configuration Options:**

- `apiKey` (required): Azure OpenAI API key
- `endpoint` (required): Azure OpenAI endpoint URL
- `deploymentName` (required): Deployment name
- `apiVersion` (optional): API version
- `temperature` (optional): Default temperature
- `maxTokens` (optional): Default max tokens

#### Ollama Configuration

```yaml
clients:
  - name: "local-llama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434" # Optional
      options: # Optional
        temperature: 0.8
        top_p: 0.9
        top_k: 40
```

**Configuration Options:**

- `model` (required): Ollama model name
- `baseURL` (optional): Ollama server URL
- `options` (optional): Model-specific options

#### Zhipu AI Configuration

```yaml
clients:
  - name: "zhipu"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY"
      model: "glm-4"
      baseURL: "https://open.bigmodel.cn/api/paas/v4" # Optional
      temperature: 0.7 # Optional
      maxTokens: 1000 # Optional
```

#### Dashscope Configuration

```yaml
clients:
  - name: "dashscope"
    type: "dashscope"
    config:
      apiKey: "$DASHSCOPE_API_KEY"
      model: "qwen-turbo"
      baseURL: "https://dashscope.aliyuncs.com/api/v1" # Optional
      temperature: 0.7 # Optional
      maxTokens: 1500 # Optional
```

#### DeepSeek Configuration

```yaml
clients:
  - name: "deepseek"
    type: "deepseek"
    config:
      apiKey: "$DEEPSEEK_API_KEY"
      model: "deepseek-chat"
      baseURL: "https://api.deepseek.com/v1" # Optional
      temperature: 0.7 # Optional
      maxTokens: 2000 # Optional
```

#### SiliconCloud Configuration

```yaml
clients:
  - name: "siliconcloud"
    type: "siliconcloud"
    config:
      apiKey: "$SILICONCLOUD_API_KEY"
      model: "meta-llama/Llama-2-7b-chat-hf"
      baseURL: "https://api.siliconflow.cn/v1" # Optional
      temperature: 0.7 # Optional
      maxTokens: 1000 # Optional
```

## Flow Configuration

### FlowConfig Structure

```go
type FlowConfig struct {
    Name       string       `yaml:"name" json:"name" toml:"name"`
    ClientName string       `yaml:"clientName,omitempty" json:"clientName,omitempty" toml:"clientName,omitempty"`
    Steps      []StepConfig `yaml:"steps" json:"steps" toml:"steps"`
}
```

**Fields:**

- `Name`: Unique identifier for the flow
- `ClientName`: Default client to use for all steps
- `Steps`: Array of step configurations

### StepConfig Structure

```go
type StepConfig struct {
    Name          string           `yaml:"name" json:"name" toml:"name"`
    ClientName    string           `yaml:"clientName,omitempty" json:"clientName,omitempty" toml:"clientName,omitempty"`
    Executor      *ExecutorConfig  `yaml:"executor,omitempty" json:"executor,omitempty" toml:"executor,omitempty"`
    Validator     *ValidatorConfig `yaml:"validator,omitempty" json:"validator,omitempty" toml:"validator,omitempty"`
    MaxRetryTimes int              `yaml:"maxRetryTimes,omitempty" json:"maxRetryTimes,omitempty" toml:"maxRetryTimes,omitempty"`
}
```

**Fields:**

- `Name`: Step identifier
- `ClientName`: Client to use for this step (overrides flow default)
- `Executor`: Executor configuration
- `Validator`: Validator configuration
- `MaxRetryTimes`: Maximum retry attempts

### ExecutorConfig Structure

```go
type ExecutorConfig struct {
    Type       string                 `yaml:"type" json:"type" toml:"type"`
    WithConfig map[string]interface{} `yaml:"withconfig,omitempty" json:"withconfig,omitempty" toml:"withconfig,omitempty"`
}
```

**Executor Types:**

- `llm`: LLM executor for AI processing
- `setcontext`: Context manipulation
- `conditional`: Conditional branching
- `command`: Shell command execution

#### LLM Executor Configuration

```yaml
executor:
  type: "llm"
  withconfig:
    template: "Analyze the following text: {{.Text}}"
    systemMessage: "You are a professional analyst."
    temperature: 0.7
    maxTokens: 1000
```

**Options:**

- `template`: Prompt template with variable substitution
- `systemMessage`: System message for the LLM
- `temperature`: Temperature override
- `maxTokens`: Max tokens override

#### SetContext Executor Configuration

```yaml
executor:
  type: "setcontext"
  withconfig:
    text: "Initial text content"
    memory:
      key1: "value1"
      key2: "value2"
```

**Options:**

- `text`: Set context text
- `memory`: Set structured memory data

### ValidatorConfig Structure

```go
type ValidatorConfig struct {
    Type       string                 `yaml:"type" json:"type" toml:"type"`
    WithConfig map[string]interface{} `yaml:"withconfig,omitempty" json:"withconfig,omitempty" toml:"withconfig,omitempty"`
}
```

**Validator Types:**

- `string`: String validation
- `json`: JSON validation
- `regex`: Regular expression validation

#### String Validator Configuration

```yaml
validator:
  type: "string"
  withconfig:
    minLength: 100
    maxLength: 2000
    contains: "required phrase"
```

**Options:**

- `minLength`: Minimum string length
- `maxLength`: Maximum string length
- `contains`: Required substring

#### JSON Validator Configuration

```yaml
validator:
  type: "json"
  withconfig:
    schema: |
      {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "age": {"type": "number"}
        },
        "required": ["name"]
      }
```

**Options:**

- `schema`: JSON Schema for validation

## Configuration File Formats

### YAML Format

```yaml
# config.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"

  - name: "local"
    type: "ollama"
    config:
      model: "llama3"

flows:
  - name: "content_processor"
    clientName: "openai"
    steps:
      - name: "analyze"
        executor:
          type: "llm"
          withconfig:
            template: "Analyze: {{.Text}}"
        validator:
          type: "string"
          withconfig:
            minLength: 50
        maxRetryTimes: 2
```

### JSON Format

```json
{
  "clients": [
    {
      "name": "openai",
      "type": "openai",
      "config": {
        "apiKey": "$OPENAI_API_KEY",
        "model": "gpt-4"
      }
    }
  ],
  "flows": [
    {
      "name": "content_processor",
      "clientName": "openai",
      "steps": [
        {
          "name": "analyze",
          "executor": {
            "type": "llm",
            "withconfig": {
              "template": "Analyze: {{.Text}}"
            }
          }
        }
      ]
    }
  ]
}
```

### TOML Format

```toml
# config.toml
[[clients]]
name = "openai"
type = "openai"

[clients.config]
apiKey = "$OPENAI_API_KEY"
model = "gpt-4"

[[flows]]
name = "content_processor"
clientName = "openai"

[[flows.steps]]
name = "analyze"

[flows.steps.executor]
type = "llm"

[flows.steps.executor.withconfig]
template = "Analyze: {{.Text}}"
```

## Environment Variables

### Variable Substitution

Environment variables are automatically substituted in configuration files using the `$VARIABLE_NAME` syntax:

```yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY" # Substituted at runtime
      orgID: "$OPENAI_ORG_ID" # Optional, empty if not set
      model: "${MODEL_NAME:-gpt-4}" # Default value syntax
```

### Common Environment Variables

**OpenAI:**

- `OPENAI_API_KEY`: API key
- `OPENAI_ORG_ID`: Organization ID
- `OPENAI_BASE_URL`: Custom base URL

**Anthropic:**

- `ANTHROPIC_API_KEY`: API key
- `ANTHROPIC_BASE_URL`: Custom base URL

**Azure OpenAI:**

- `AZURE_OPENAI_API_KEY`: API key
- `AZURE_OPENAI_ENDPOINT`: Endpoint URL

**Zhipu AI:**

- `ZHIPU_API_KEY`: API key

**Dashscope:**

- `DASHSCOPE_API_KEY`: API key

**DeepSeek:**

- `DEEPSEEK_API_KEY`: API key

**SiliconCloud:**

- `SILICONCLOUD_API_KEY`: API key

## Loading Configuration

### From File

```go
// Load from YAML
err := anyi.ConfigFromFile("config.yaml")

// Load from JSON
err := anyi.ConfigFromFile("config.json")

// Load from TOML
err := anyi.ConfigFromFile("config.toml")
```

### Programmatic Configuration

```go
config := anyi.AnyiConfig{
    Clients: []llm.ClientConfig{
        {
            Name: "openai",
            Type: "openai",
            Config: map[string]interface{}{
                "apiKey": os.Getenv("OPENAI_API_KEY"),
                "model":  "gpt-4",
            },
        },
    },
    Flows: []anyi.FlowConfig{
        {
            Name:       "processor",
            ClientName: "openai",
            Steps: []anyi.StepConfig{
                {
                    Name: "analyze",
                    Executor: &anyi.ExecutorConfig{
                        Type: "llm",
                        WithConfig: map[string]interface{}{
                            "template": "Analyze: {{.Text}}",
                        },
                    },
                },
            },
        },
    },
}

err := anyi.Config(&config)
```

## Best Practices

### Security

1. **Never hardcode API keys** in configuration files
2. **Use environment variables** for sensitive information
3. **Restrict file permissions** on configuration files
4. **Use secrets management** in production environments

### Organization

1. **Use descriptive names** for clients and flows
2. **Group related configurations** logically
3. **Document custom configurations** with comments
4. **Version control** configuration templates

### Environment Management

1. **Use separate configurations** for different environments
2. **Override settings** with environment variables
3. **Validate configurations** before deployment
4. **Test configurations** in staging environments

### Performance

1. **Reuse clients** across flows when possible
2. **Set appropriate timeouts** and retry limits
3. **Configure connection pooling** for high-throughput scenarios
4. **Monitor resource usage** and adjust accordingly

This configuration reference provides all the information needed to set up and manage Anyi applications effectively. For implementation examples, see the tutorials and how-to guides.
