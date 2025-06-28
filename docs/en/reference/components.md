# Components Reference

This document provides a comprehensive reference for all built-in components in the Anyi framework, including executors, validators, and other reusable components.

## Overview

Anyi components are the building blocks of workflows. They provide specific functionality that can be composed together to create complex AI applications. Components are designed to be:

- **Reusable**: Can be used across different workflows
- **Configurable**: Accept parameters to customize behavior
- **Composable**: Can be combined to create complex logic
- **Extensible**: Can be extended or replaced with custom implementations

## Executors

Executors are components that perform specific tasks within workflow steps. They process input, perform operations, and produce output.

### LLMExecutor

The most commonly used executor that sends prompts to LLM providers and processes responses.

#### Configuration

```yaml
executor:
  type: "llm"
  withconfig:
    template: "Analyze the following text: {{.Text}}"
    systemMessage: "You are a professional analyst."
    temperature: 0.7
    maxTokens: 1000
    extractThink: true
```

#### Configuration Options

- `template` (string): Prompt template with Go template syntax
- `systemMessage` (string, optional): System message for the LLM
- `temperature` (float64, optional): Temperature override (0.0-2.0)
- `maxTokens` (int, optional): Maximum tokens to generate
- `topP` (float64, optional): Top-p sampling parameter
- `frequencyPenalty` (float64, optional): Frequency penalty (-2.0 to 2.0)
- `presencePenalty` (float64, optional): Presence penalty (-2.0 to 2.0)
- `stop` ([]string, optional): Stop sequences
- `extractThink` (bool, optional): Extract `<think>` tags from response

#### Template Variables

Templates can access the following variables from FlowContext:

- `{{.Text}}`: Current text content
- `{{.Memory}}`: Structured memory data
- `{{.Memory.FieldName}}`: Specific memory fields
- `{{.Think}}`: Previous thinking process
- `{{.Images}}`: Array of image URLs

#### Template Functions

Built-in template functions:

- `{{len .Text}}`: Get text length
- `{{contains .Text "substring"}}`: Check if text contains substring
- `{{upper .Text}}`: Convert to uppercase
- `{{lower .Text}}`: Convert to lowercase
- `{{trim .Text}}`: Trim whitespace

#### Example Usage

```go
// Programmatic usage
executor := &anyi.LLMExecutor{
    Template: "Summarize this text in 3 bullet points:\n\n{{.Text}}",
    SystemMessage: "You are a professional summarizer.",
    Temperature: 0.3,
    MaxTokens: 500,
}
```

### SetContextExecutor

Directly modifies the flow context without external API calls.

#### Configuration

```yaml
executor:
  type: "setcontext"
  withconfig:
    text: "New text content"
    memory:
      key1: "value1"
      key2: "value2"
    think: "Initial thinking process"
    images:
      - "https://example.com/image1.jpg"
      - "https://example.com/image2.jpg"
```

#### Configuration Options

- `text` (string, optional): Set context text
- `memory` (map, optional): Set structured memory data
- `think` (string, optional): Set thinking content
- `images` ([]string, optional): Set image URLs
- `append` (bool, optional): Append to existing content instead of replacing

#### Use Cases

- Initialize workflow context
- Set up structured data for complex workflows
- Reset context between workflow phases
- Prepare data for subsequent steps

#### Example Usage

```go
// Initialize workflow with structured data
executor := &anyi.SetContextExecutor{
    Memory: map[string]interface{}{
        "task": "content_analysis",
        "requirements": []string{"accuracy", "brevity", "clarity"},
        "status": "initialized",
    },
}
```

### ConditionalFlowExecutor

Enables branching logic in workflows based on conditions.

#### Configuration

```yaml
executor:
  type: "conditional"
  withconfig:
    condition: '{{contains .Text "error"}}'
    trueFlow: "error_handler"
    falseFlow: "normal_processor"
    trueSteps:
      - name: "log_error"
        executor:
          type: "llm"
          withconfig:
            template: "Log this error: {{.Text}}"
    falseSteps:
      - name: "continue_processing"
        executor:
          type: "llm"
          withconfig:
            template: "Process normally: {{.Text}}"
```

#### Configuration Options

- `condition` (string): Go template expression that evaluates to boolean
- `trueFlow` (string, optional): Flow to execute if condition is true
- `falseFlow` (string, optional): Flow to execute if condition is false
- `trueSteps` ([]StepConfig, optional): Steps to execute if condition is true
- `falseSteps` ([]StepConfig, optional): Steps to execute if condition is false

#### Condition Expressions

Supported condition patterns:

- `{{eq .Memory.Status "complete"}}`: Equality check
- `{{ne .Text ""}}`: Not equal check
- `{{gt (len .Text) 100}}`: Greater than comparison
- `{{lt .Memory.Count 5}}`: Less than comparison
- `{{contains .Text "keyword"}}`: String contains check
- `{{and (gt (len .Text) 50) (lt (len .Text) 500)}}`: Logical AND
- `{{or (eq .Memory.Type "urgent") (eq .Memory.Type "critical")}}`: Logical OR

### RunCommandExecutor

Executes shell commands and captures their output.

#### Configuration

```yaml
executor:
  type: "command"
  withconfig:
    command: "python"
    args:
      - "script.py"
      - "{{.Text}}"
    workingDir: "/path/to/scripts"
    timeout: 30
    captureOutput: true
    captureError: true
```

#### Configuration Options

- `command` (string): Command to execute
- `args` ([]string, optional): Command arguments (supports templates)
- `workingDir` (string, optional): Working directory
- `timeout` (int, optional): Timeout in seconds (default: 30)
- `captureOutput` (bool, optional): Capture stdout (default: true)
- `captureError` (bool, optional): Capture stderr (default: false)
- `env` (map[string]string, optional): Environment variables

#### Security Considerations

- Always validate input before using in commands
- Use allowlists for permitted commands
- Run with minimal privileges
- Consider using containerization for isolation

#### Example Usage

```go
// Execute a Python data processing script
executor := &anyi.RunCommandExecutor{
    Command: "python",
    Args: []string{"process_data.py", "--input", "{{.Text}}"},
    WorkingDir: "/opt/scripts",
    Timeout: 60,
}
```

## Validators

Validators ensure that executor outputs meet specific criteria before proceeding to the next step.

### StringValidator

Validates text output based on various string criteria.

#### Configuration

```yaml
validator:
  type: "string"
  withconfig:
    minLength: 100
    maxLength: 2000
    contains: "required phrase"
    notContains: "forbidden word"
    matchRegex: "^[A-Z].*\\.$"
    notMatchRegex: "\\b(spam|scam)\\b"
    startsWith: "Summary:"
    endsWith: "."
```

#### Configuration Options

- `minLength` (int, optional): Minimum string length
- `maxLength` (int, optional): Maximum string length
- `contains` (string, optional): Required substring
- `notContains` (string, optional): Forbidden substring
- `matchRegex` (string, optional): Required regex pattern
- `notMatchRegex` (string, optional): Forbidden regex pattern
- `startsWith` (string, optional): Required prefix
- `endsWith` (string, optional): Required suffix

#### Example Usage

```go
// Validate that output is a proper summary
validator := &anyi.StringValidator{
    MinLength: 50,
    MaxLength: 500,
    StartsWith: "Summary:",
    EndsWith: ".",
    NotContains: "I cannot",
}
```

### JsonValidator

Validates that output is valid JSON and optionally validates against a JSON Schema.

#### Configuration

```yaml
validator:
  type: "json"
  withconfig:
    schema: |
      {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "summary": {"type": "string"},
          "tags": {
            "type": "array",
            "items": {"type": "string"}
          },
          "confidence": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          }
        },
        "required": ["title", "summary"]
      }
    requiredFields:
      - "title"
      - "summary"
```

#### Configuration Options

- `schema` (string, optional): JSON Schema for validation
- `requiredFields` ([]string, optional): List of required field names
- `allowEmpty` (bool, optional): Allow empty JSON objects (default: false)

#### Example Usage

```go
// Validate structured analysis output
validator := &anyi.JsonValidator{
    RequiredFields: []string{"analysis", "recommendations", "confidence"},
    Schema: `{
        "type": "object",
        "properties": {
            "analysis": {"type": "string", "minLength": 100},
            "recommendations": {
                "type": "array",
                "items": {"type": "string"},
                "minItems": 1
            },
            "confidence": {"type": "number", "minimum": 0, "maximum": 1}
        }
    }`,
}
```

### RegexValidator

Validates output against regular expression patterns.

#### Configuration

```yaml
validator:
  type: "regex"
  withconfig:
    pattern: "^\\d{3}-\\d{2}-\\d{4}$"
    flags: "i"
    multiline: true
    dotAll: true
```

#### Configuration Options

- `pattern` (string): Regular expression pattern
- `flags` (string, optional): Regex flags (i=case insensitive, m=multiline, s=dotall)
- `multiline` (bool, optional): Enable multiline mode
- `dotAll` (bool, optional): Enable dot-all mode

#### Example Usage

```go
// Validate email format
validator := &anyi.RegexValidator{
    Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
    Flags: "i",
}
```

### CustomValidator

Base interface for creating custom validators.

#### Interface

```go
type Validator interface {
    Validate(context *FlowContext) error
}
```

#### Implementation Example

```go
type WordCountValidator struct {
    MinWords int
    MaxWords int
}

func (v *WordCountValidator) Validate(context *FlowContext) error {
    words := strings.Fields(context.Text)
    count := len(words)

    if v.MinWords > 0 && count < v.MinWords {
        return fmt.Errorf("text has %d words, minimum required: %d", count, v.MinWords)
    }

    if v.MaxWords > 0 && count > v.MaxWords {
        return fmt.Errorf("text has %d words, maximum allowed: %d", count, v.MaxWords)
    }

    return nil
}
```

## Template System

### Template Syntax

Anyi uses Go's `text/template` package with additional functions.

#### Basic Syntax

```go
// Variable access
{{.Text}}
{{.Memory.FieldName}}

// Conditionals
{{if .Text}}Text exists{{end}}
{{if eq .Memory.Status "ready"}}Ready!{{else}}Not ready{{end}}

// Loops
{{range .Memory.Items}}
- {{.}}
{{end}}

// Functions
{{len .Text}}
{{contains .Text "keyword"}}
```

#### Custom Functions

Additional functions available in templates:

- `contains`: Check if string contains substring
- `hasPrefix`: Check if string has prefix
- `hasSuffix`: Check if string has suffix
- `upper`: Convert to uppercase
- `lower`: Convert to lowercase
- `title`: Convert to title case
- `trim`: Trim whitespace
- `join`: Join array elements with separator
- `split`: Split string into array
- `replace`: Replace substring

#### Example Templates

```yaml
# Analysis template
template: |
  Analyze the following {{.Memory.ContentType}}:

  {{.Text}}

  Please provide:
  {{range .Memory.Requirements}}
  - {{.}}
  {{end}}

  {{if .Think}}
  Previous analysis: {{.Think}}
  {{end}}

# Conditional processing
template: |
  {{if gt (len .Text) 1000}}
  This is a long document. Please provide a detailed analysis.
  {{else}}
  This is a short document. Please provide a concise analysis.
  {{end}}

  Content: {{.Text}}
```

## Error Handling

### Built-in Error Types

Common error types returned by components:

- `ValidationError`: Validation failed
- `ExecutionError`: Execution failed
- `TemplateError`: Template processing failed
- `ConfigurationError`: Invalid configuration

### Error Handling Patterns

```go
// Check for specific error types
if validationErr, ok := err.(*anyi.ValidationError); ok {
    log.Printf("Validation failed: %s", validationErr.Message)
    // Handle validation failure
}

// Retry logic with exponential backoff
maxRetries := 3
backoff := time.Second

for i := 0; i < maxRetries; i++ {
    result, err := executor.Execute(context)
    if err == nil {
        break
    }

    if i < maxRetries-1 {
        time.Sleep(backoff)
        backoff *= 2
    }
}
```

## Best Practices

### Executor Design

1. **Keep executors focused**: Each executor should have a single responsibility
2. **Use appropriate timeouts**: Set reasonable timeouts for external calls
3. **Implement proper error handling**: Handle both expected and unexpected errors
4. **Validate inputs**: Always validate inputs before processing
5. **Use templates effectively**: Leverage templates for dynamic content generation

### Validator Design

1. **Fail fast**: Validate as early as possible in the workflow
2. **Provide clear error messages**: Help users understand what went wrong
3. **Use multiple validators**: Combine validators for comprehensive validation
4. **Consider performance**: Avoid expensive validation operations when possible

### Template Best Practices

1. **Keep templates readable**: Use proper formatting and comments
2. **Handle missing data**: Use conditional checks for optional data
3. **Escape special characters**: Be careful with user-provided content
4. **Test templates thoroughly**: Test with various input scenarios

### Configuration Management

1. **Use environment variables**: Keep sensitive data out of configuration files
2. **Validate configuration**: Check configuration validity at startup
3. **Document options**: Provide clear documentation for all configuration options
4. **Use defaults wisely**: Provide sensible defaults for optional parameters

This components reference provides detailed information about all built-in components in Anyi. For practical examples and implementation patterns, see the tutorials and how-to guides.
