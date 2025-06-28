# Configuration Management

This tutorial covers how to use configuration files and environment variables to manage your Anyi applications. You'll learn to separate configuration from code, manage different environments, and implement secure configuration practices.

## Table of Contents

- [Why Use Configuration Files](#why-use-configuration-files)
- [Supported Configuration Formats](#supported-configuration-formats)
- [Basic Configuration Structure](#basic-configuration-structure)
- [Client Configuration](#client-configuration)
- [Flow Configuration](#flow-configuration)
- [Environment Variables](#environment-variables)
- [Advanced Configuration Patterns](#advanced-configuration-patterns)
- [Best Practices](#best-practices)

## Why Use Configuration Files

Configuration files provide several advantages over hardcoded settings:

- **Separation of Concerns**: Keep business logic separate from configuration
- **Environment Flexibility**: Different settings for development, staging, and production
- **Security**: Keep sensitive information out of source code
- **Runtime Changes**: Modify behavior without recompiling
- **Team Collaboration**: Non-developers can modify settings
- **Version Control**: Track configuration changes over time

## Supported Configuration Formats

Anyi supports three popular configuration formats:

### YAML (Recommended)

```yaml
# config.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"
```

### JSON

```json
{
  "clients": [
    {
      "name": "openai",
      "type": "openai",
      "config": {
        "model": "gpt-4",
        "apiKey": "$OPENAI_API_KEY"
      }
    }
  ]
}
```

### TOML

```toml
[[clients]]
name = "openai"
type = "openai"

[clients.config]
model = "gpt-4"
apiKey = "$OPENAI_API_KEY"
```

## Basic Configuration Structure

A typical Anyi configuration file contains two main sections:

```yaml
# Complete configuration structure
clients:
  # Define LLM clients
  - name: "client_name"
    type: "provider_type"
    config:
      # Provider-specific configuration

flows:
  # Define workflows
  - name: "flow_name"
    clientName: "default_client"
    steps:
      # Step definitions
```

## Client Configuration

### Single Client Configuration

```yaml
clients:
  - name: "gpt4"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"
      temperature: 0.7
      maxTokens: 2000
```

### Multiple Clients

```yaml
clients:
  # OpenAI for complex reasoning
  - name: "reasoning"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"
      temperature: 0.3
      maxTokens: 1500

  # Anthropic for analysis
  - name: "analysis"
    type: "anthropic"
    config:
      model: "claude-3-opus-20240229"
      apiKey: "$ANTHROPIC_API_KEY"
      temperature: 0.5
      maxTokens: 2000

  # Local model for simple tasks
  - name: "local"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434"
      temperature: 0.8

  # DeepSeek for coding
  - name: "coding"
    type: "deepseek"
    config:
      model: "deepseek-chat"
      apiKey: "$DEEPSEEK_API_KEY"
      temperature: 0.1
```

### Provider-Specific Options

Each provider supports different configuration options:

```yaml
clients:
  # OpenAI with all options
  - name: "openai_full"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"
      baseURL: "https://api.openai.com/v1" # Custom endpoint
      temperature: 0.7
      topP: 0.9
      maxTokens: 2000
      presencePenalty: 0.1
      frequencyPenalty: 0.1
      stop: ["###", "END"]

  # Azure OpenAI
  - name: "azure"
    type: "azureopenai"
    config:
      apiKey: "$AZURE_OPENAI_API_KEY"
      deploymentId: "$AZURE_DEPLOYMENT_ID"
      endpoint: "$AZURE_OPENAI_ENDPOINT"
      temperature: 0.7

  # Zhipu AI
  - name: "zhipu"
    type: "zhipu"
    config:
      model: "glm-4"
      apiKey: "$ZHIPU_API_KEY"
      baseURL: "https://open.bigmodel.cn/api/paas/v4/"
```

## Flow Configuration

### Basic Flow Structure

```yaml
flows:
  - name: "content_processor"
    clientName: "gpt4" # Default client for all steps
    steps:
      - name: "analyze_content"
        executor:
          type: "llm"
          withconfig:
            template: "Analyze the following content: {{.Text}}"
            systemMessage: "You are a content analyst."
        maxRetryTimes: 2

      - name: "generate_summary"
        executor:
          type: "llm"
          withconfig:
            template: "Create a summary of this analysis: {{.Text}}"
            systemMessage: "You are a summarization expert."
        validator:
          type: "string"
          withconfig:
            minLength: 100
            maxLength: 500
```

### Advanced Flow with Multiple Clients

```yaml
flows:
  - name: "research_workflow"
    clientName: "reasoning" # Default client
    steps:
      - name: "initial_research"
        clientName: "reasoning" # Override for this step
        executor:
          type: "llm"
          withconfig:
            template: |
              Research the topic: {{.Text}}

              Provide comprehensive information covering:
              1. Current state
              2. Key challenges
              3. Future prospects
            systemMessage: "You are a research specialist."
        validator:
          type: "string"
          withconfig:
            minLength: 500
        maxRetryTimes: 3

      - name: "analyze_findings"
        clientName: "analysis" # Different client for analysis
        executor:
          type: "llm"
          withconfig:
            template: |
              Analyze the research findings:

              {{.Text}}

              Provide critical analysis and insights.
            systemMessage: "You are a critical analyst."
        validator:
          type: "string"
          withconfig:
            minLength: 300

      - name: "generate_report"
        # Uses default client (reasoning)
        executor:
          type: "llm"
          withconfig:
            template: |
              Create a professional report based on:

              {{.Text}}

              Format as a structured document.
            systemMessage: "You are a professional report writer."
        validator:
          type: "string"
          withconfig:
            minLength: 800
```

### Conditional Flows

```yaml
flows:
  - name: "smart_routing"
    clientName: "gpt4"
    steps:
      - name: "classify_request"
        executor:
          type: "llm"
          withconfig:
            template: |
              Classify this request: {{.Text}}

              Respond with exactly one word:
              - "technical" for technical questions
              - "creative" for creative tasks
              - "analysis" for analytical work
            systemMessage: "You are a request classifier."

      - name: "route_to_specialist"
        executor:
          type: "conditional"
          withconfig:
            condition: '{{.Text | eq "technical"}}'
            trueFlow: "technical_flow"
            falseFlow: "general_flow"
```

### Using Validators

```yaml
flows:
  - name: "quality_controlled"
    clientName: "gpt4"
    steps:
      - name: "generate_json"
        executor:
          type: "llm"
          withconfig:
            template: |
              Generate a JSON object with user information for: {{.Text}}

              Required fields: name, email, age, occupation
            systemMessage: "Generate valid JSON responses."
        validator:
          type: "json"
          withconfig:
            requiredFields: ["name", "email", "age", "occupation"]
        maxRetryTimes: 3

      - name: "validate_content"
        executor:
          type: "llm"
          withconfig:
            template: "Improve this content: {{.Text}}"
            systemMessage: "You are an editor."
        validator:
          type: "string"
          withconfig:
            minLength: 200
            maxLength: 1000
            matchRegex: "\\w+@\\w+\\.\\w+" # Must contain email
```

## Environment Variables

### Using Environment Variables in Configuration

Environment variables are referenced with the `$` prefix:

```yaml
clients:
  - name: "production_openai"
    type: "openai"
    config:
      model: "$OPENAI_MODEL" # From environment
      apiKey: "$OPENAI_API_KEY" # From environment
      baseURL: "$OPENAI_BASE_URL" # Optional override
      temperature: 0.7 # Static value
```

### Environment File (.env)

Create a `.env` file for local development:

```bash
# .env file
OPENAI_API_KEY=sk-your-openai-key
OPENAI_MODEL=gpt-4
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
DEEPSEEK_API_KEY=sk-your-deepseek-key
ZHIPU_API_KEY=your-zhipu-key

# Environment-specific settings
ENVIRONMENT=development
LOG_LEVEL=debug
MAX_RETRIES=3
```

### Loading Environment Variables

```go
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv" // Popular .env loader
	"github.com/jieliu2000/anyi"
)

func main() {
	// Load .env file in development
	if os.Getenv("ENVIRONMENT") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	// Load configuration
	err := anyi.ConfigFromFile("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use configured flows
	flow, err := anyi.GetFlow("my_workflow")
	if err != nil {
		log.Fatalf("Failed to get flow: %v", err)
	}

	result, err := flow.RunWithInput("test input")
	if err != nil {
		log.Fatalf("Flow execution failed: %v", err)
	}

	log.Printf("Result: %s", result.Text)
}
```

## Advanced Configuration Patterns

### Environment-Specific Configurations

Create separate configuration files for different environments:

```yaml
# config-development.yaml
clients:
  - name: "dev_client"
    type: "ollama" # Local model for development
    config:
      model: "llama3"
      baseURL: "http://localhost:11434"

flows:
  - name: "test_flow"
    clientName: "dev_client"
    steps:
      - name: "simple_test"
        executor:
          type: "llm"
          withconfig:
            template: "Simple test: {{.Text}}"
        maxRetryTimes: 1 # Fewer retries in development
```

```yaml
# config-production.yaml
clients:
  - name: "prod_client"
    type: "openai" # Production-grade model
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"
      temperature: 0.3 # More deterministic in production

flows:
  - name: "production_flow"
    clientName: "prod_client"
    steps:
      - name: "production_task"
        executor:
          type: "llm"
          withconfig:
            template: "Production task: {{.Text}}"
        validator:
          type: "string"
          withconfig:
            minLength: 100
        maxRetryTimes: 5 # More retries in production
```

### Configuration with Includes

For complex applications, split configuration into multiple files:

```yaml
# main-config.yaml
clients:
  - name: "primary"
    type: "openai"
    config:
      model: "$PRIMARY_MODEL"
      apiKey: "$OPENAI_API_KEY"

# Include other configuration files
includes:
  - "flows/content-flows.yaml"
  - "flows/analysis-flows.yaml"
  - "clients/specialized-clients.yaml"
```

### Dynamic Configuration

Load configuration programmatically:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

func createDynamicConfig() *anyi.AnyiConfig {
	environment := os.Getenv("ENVIRONMENT")

	config := &anyi.AnyiConfig{
		Clients: []llm.ClientConfig{},
		Flows:   []anyi.FlowConfig{},
	}

	// Add clients based on environment
	if environment == "production" {
		config.Clients = append(config.Clients, llm.ClientConfig{
			Name: "prod_openai",
			Type: "openai",
			Config: map[string]interface{}{
				"model":       "gpt-4",
				"apiKey":      os.Getenv("OPENAI_API_KEY"),
				"temperature": 0.3,
			},
		})
	} else {
		config.Clients = append(config.Clients, llm.ClientConfig{
			Name: "dev_ollama",
			Type: "ollama",
			Config: map[string]interface{}{
				"model":   "llama3",
				"baseURL": "http://localhost:11434",
			},
		})
	}

	// Add flows based on requirements
	config.Flows = append(config.Flows, anyi.FlowConfig{
		Name:       "adaptive_flow",
		ClientName: config.Clients[0].Name,
		Steps: []anyi.StepConfig{
			{
				Name: "process_input",
				Executor: &anyi.ExecutorConfig{
					Type: "llm",
					WithConfig: map[string]interface{}{
						"template":      "Process: {{.Text}}",
						"systemMessage": "You are a helpful assistant.",
					},
				},
				MaxRetryTimes: getRetryCount(environment),
			},
		},
	})

	return config
}

func getRetryCount(env string) int {
	if env == "production" {
		return 5
	}
	return 2
}

func main() {
	// Create dynamic configuration
	config := createDynamicConfig()

	// Apply configuration
	err := anyi.Config(config)
	if err != nil {
		log.Fatalf("Configuration failed: %v", err)
	}

	// Use the configured flow
	flow, err := anyi.GetFlow("adaptive_flow")
	if err != nil {
		log.Fatalf("Failed to get flow: %v", err)
	}

	result, err := flow.RunWithInput("test input")
	if err != nil {
		log.Fatalf("Flow execution failed: %v", err)
	}

	log.Printf("Result: %s", result.Text)
}
```

## Best Practices

### 1. Security Best Practices

- **Never commit API keys**: Use environment variables or secret management systems
- **Use least privilege**: Only grant necessary permissions to API keys
- **Rotate keys regularly**: Implement key rotation policies
- **Validate configuration**: Check for required settings on startup

```yaml
# Good: Using environment variables
clients:
  - name: "secure_client"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY" # From environment

# Bad: Hardcoded secrets (never do this)
# clients:
#   - name: "insecure_client"
#     type: "openai"
#     config:
#       apiKey: "sk-hardcoded-key-123"  # DON'T DO THIS
```

### 2. Environment Management

```bash
# development.env
ENVIRONMENT=development
OPENAI_MODEL=gpt-4o-mini
LOG_LEVEL=debug
MAX_RETRIES=2

# production.env
ENVIRONMENT=production
OPENAI_MODEL=gpt-4
LOG_LEVEL=info
MAX_RETRIES=5
```

### 3. Configuration Validation

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jieliu2000/anyi"
)

func validateConfiguration() error {
	required := []string{
		"OPENAI_API_KEY",
		"ENVIRONMENT",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	return nil
}

func main() {
	// Validate environment before loading configuration
	if err := validateConfiguration(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Load configuration
	configFile := fmt.Sprintf("config-%s.yaml", os.Getenv("ENVIRONMENT"))
	err := anyi.ConfigFromFile(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Configuration loaded successfully")
}
```

### 4. Configuration Documentation

Document your configuration schema:

```yaml
# config-schema.yaml - Documentation template
clients:
  - name: "string" # Required: Unique client identifier
    type: "string" # Required: Provider type (openai, anthropic, etc.)
    config:
      model: "string" # Required: Model name
      apiKey: "string" # Required: API key (use environment variables)
      temperature: 0.7 # Optional: Controls randomness (0.0-2.0)
      maxTokens: 2000 # Optional: Maximum response length
      baseURL: "string" # Optional: Custom API endpoint

flows:
  - name: "string" # Required: Unique flow identifier
    clientName: "string" # Required: Default client name
    steps:
      - name: "string" # Required: Step identifier
        clientName: "string" # Optional: Override client for this step
        maxRetryTimes: 3 # Optional: Retry attempts (default: 0)
        executor:
          type: "string" # Required: Executor type (llm, conditional, etc.)
          withconfig: # Required: Executor-specific configuration
            template: "string" # For LLM: Prompt template
            systemMessage: "string" # For LLM: System message
        validator: # Optional: Output validation
          type: "string" # Validator type (string, json)
          withconfig: # Validator-specific configuration
            minLength: 100 # For string: Minimum length
```

### 5. Testing Configuration

```go
package main

import (
	"testing"

	"github.com/jieliu2000/anyi"
)

func TestConfigurationLoading(t *testing.T) {
	// Test configuration loading
	err := anyi.ConfigFromFile("./test-config.yaml")
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}

	// Test client creation
	client, err := anyi.GetClient("test_client")
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	// Test flow creation
	flow, err := anyi.GetFlow("test_flow")
	if err != nil {
		t.Fatalf("Failed to get test flow: %v", err)
	}

	if flow == nil {
		t.Fatal("Flow is nil")
	}
}
```

By following these configuration management practices, you can build flexible, secure, and maintainable Anyi applications that adapt to different environments and requirements.
