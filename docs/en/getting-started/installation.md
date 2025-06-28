# Installation and Setup

This guide will help you install Anyi and set up your development environment.

## System Requirements

- **Go**: Version 1.20 or higher
- **Operating System**: Linux, macOS, or Windows
- **Network**: Internet connectivity for accessing LLM APIs

## Installation

### Using Go Modules (Recommended)

Add Anyi to your Go project:

```bash
go get -u github.com/jieliu2000/anyi
```

### Verify Installation

Create a simple test file to verify the installation:

```go
// test_installation.go
package main

import (
    "fmt"
    "github.com/jieliu2000/anyi"
)

func main() {
    fmt.Println("Anyi installation successful!")
    fmt.Printf("Anyi version: %s\n", anyi.Version())
}
```

Run the test:

```bash
go run test_installation.go
```

## Environment Setup

### API Keys

Anyi requires API keys for various LLM providers. Set up environment variables for the providers you plan to use:

```bash
# OpenAI
export OPENAI_API_KEY="your-openai-api-key"

# Anthropic
export ANTHROPIC_API_KEY="your-anthropic-api-key"

# Azure OpenAI
export AZ_OPENAI_API_KEY="your-azure-openai-key"
export AZ_OPENAI_ENDPOINT="your-azure-endpoint"
export AZ_OPENAI_MODEL_DEPLOYMENT_ID="your-deployment-id"

# Zhipu AI
export ZHIPU_API_KEY="your-zhipu-api-key"

# Dashscope (Alibaba Cloud)
export DASHSCOPE_API_KEY="your-dashscope-api-key"

# DeepSeek
export DEEPSEEK_API_KEY="your-deepseek-api-key"
```

### Using .env Files (Development)

For local development, you can use a `.env` file:

```bash
# .env
OPENAI_API_KEY=your-openai-api-key
ANTHROPIC_API_KEY=your-anthropic-api-key
# Add other keys as needed
```

Load the `.env` file in your application:

```go
import "github.com/joho/godotenv"

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found")
    }
}
```

## Local Model Setup (Optional)

### Ollama Installation

For local model deployment using Ollama:

1. Install Ollama from [https://ollama.ai](https://ollama.ai)

2. Pull a model:

```bash
ollama pull llama3
```

3. Verify Ollama is running:

```bash
curl http://localhost:11434/api/version
```

## Project Structure

Here's a recommended project structure for Anyi applications:

```
your-project/
├── cmd/
│   └── main.go           # Application entry point
├── config/
│   ├── config.yaml       # Configuration files
│   └── .env             # Environment variables (development)
├── flows/
│   └── workflows.go     # Workflow definitions
├── handlers/
│   └── api.go          # HTTP handlers (if building web app)
├── go.mod
├── go.sum
└── README.md
```

## Development Tools

### Recommended Go Extensions

If using Visual Studio Code:

- Go extension by Google
- YAML extension for configuration files

### Testing Setup

Set up a basic test environment:

```go
// main_test.go
package main

import (
    "os"
    "testing"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

func TestBasicSetup(t *testing.T) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set")
    }

    config := openai.DefaultConfig(apiKey)
    client, err := anyi.NewClient("test", config)
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }

    if client == nil {
        t.Fatal("Client is nil")
    }
}
```

## Next Steps

Once you have Anyi installed and configured:

1. Follow the [Quick Start Guide](quickstart.md) to create your first application
2. Learn about [Basic Concepts](concepts.md) to understand Anyi's architecture
3. Explore the [LLM Clients Tutorial](../tutorials/llm-clients.md) to connect to your preferred AI provider

## Troubleshooting

### Common Issues

**"module not found" error**

- Ensure you're using Go 1.20 or higher
- Run `go mod tidy` to resolve dependencies

**API key errors**

- Verify your API keys are correctly set in environment variables
- Check that your API keys have the necessary permissions

**Network connectivity issues**

- Ensure your firewall allows outbound HTTPS connections
- Some corporate networks may require proxy configuration

### Getting Help

If you encounter issues:

- Check the [FAQ](../reference/faq.md)
- Search existing [GitHub issues](https://github.com/jieliu2000/anyi/issues)
- Create a new issue with detailed error information
