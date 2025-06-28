# Provider Setup Guide

This guide provides step-by-step instructions for setting up each LLM provider supported by Anyi. Follow the sections relevant to the providers you want to use.

## OpenAI Setup

### 1. Create an OpenAI Account

1. Visit [https://platform.openai.com](https://platform.openai.com)
2. Sign up for an account or log in
3. Add billing information (required for API access)

### 2. Generate API Key

1. Go to the [API Keys page](https://platform.openai.com/api-keys)
2. Click "Create new secret key"
3. Give it a descriptive name (e.g., "anyi-development")
4. Copy the key immediately (you won't see it again)

### 3. Configure Environment

```bash
export OPENAI_API_KEY="sk-your-actual-api-key-here"
```

### 4. Test Configuration

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
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, err := anyi.NewClient("openai-test", config)
    if err != nil {
        log.Fatalf("Setup failed: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "Hello, are you working?"},
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("Test failed: %v", err)
    }

    log.Printf("Success! Response: %s", response.Content)
}
```

## Anthropic Setup

### 1. Create an Anthropic Account

1. Visit [https://www.anthropic.com/claude](https://www.anthropic.com/claude)
2. Sign up for an account
3. Request API access (may require approval)

### 2. Generate API Key

1. Go to your [Console](https://console.anthropic.com/)
2. Navigate to "API Keys"
3. Create a new key
4. Copy the key securely

### 3. Configure Environment

```bash
export ANTHROPIC_API_KEY="sk-ant-your-actual-api-key-here"
```

### 4. Test Configuration

```go
config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))
client, err := anyi.NewClient("anthropic-test", config)
// Test with a simple message...
```

## Azure OpenAI Setup

### 1. Create Azure Account

1. Visit [https://azure.microsoft.com](https://azure.microsoft.com)
2. Sign up or log in to your Azure account
3. Create a new subscription if needed

### 2. Create Azure OpenAI Resource

1. In Azure Portal, search for "Azure OpenAI"
2. Click "Create" and fill in:
   - Resource group (create new if needed)
   - Region (choose based on your location)
   - Name (unique identifier)
   - Pricing tier

### 3. Deploy a Model

1. Go to your Azure OpenAI resource
2. Click "Model deployments"
3. Create a new deployment:
   - Choose a model (e.g., gpt-4)
   - Give it a deployment name
   - Set capacity as needed

### 4. Get Configuration Details

From your Azure OpenAI resource, collect:

- **Endpoint**: Found in "Keys and Endpoint" section
- **API Key**: Also in "Keys and Endpoint" section
- **Deployment ID**: The name you gave your model deployment

### 5. Configure Environment

```bash
export AZ_OPENAI_API_KEY="your-azure-openai-key"
export AZ_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZ_OPENAI_MODEL_DEPLOYMENT_ID="your-deployment-name"
```

### 6. Test Configuration

```go
config := azureopenai.NewConfig(
    os.Getenv("AZ_OPENAI_API_KEY"),
    os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"),
    os.Getenv("AZ_OPENAI_ENDPOINT"),
)
client, err := anyi.NewClient("azure-test", config)
// Test with a message...
```

## Ollama Setup

### 1. Install Ollama

**macOS:**

```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Linux:**

```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
Download from [https://ollama.ai/download](https://ollama.ai/download)

### 2. Pull a Model

```bash
# Pull Llama 3 (8B parameters)
ollama pull llama3

# Or pull a smaller model for testing
ollama pull llama3:8b

# List available models
ollama list
```

### 3. Start Ollama Server

```bash
# Start the server (usually starts automatically)
ollama serve
```

### 4. Test Ollama

```bash
# Test with command line
ollama run llama3 "Hello, how are you?"
```

### 5. Configure Anyi

```go
// Default configuration (assumes local server)
config := ollama.DefaultConfig("llama3")

// Custom server configuration
config := ollama.NewConfig("llama3", "http://your-server:11434")

client, err := anyi.NewClient("ollama-test", config)
```

## Zhipu AI Setup

### 1. Create Zhipu AI Account

1. Visit [https://open.bigmodel.cn](https://open.bigmodel.cn)
2. Register for an account
3. Complete identity verification if required

### 2. Generate API Key

1. Go to your dashboard
2. Find the API key section
3. Create a new API key
4. Copy the key securely

### 3. Configure Environment

```bash
export ZHIPU_API_KEY="your-zhipu-api-key"
```

### 4. Test Configuration

```go
config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")
client, err := anyi.NewClient("zhipu-test", config)
// Test with Chinese or English message...
```

## Dashscope (Alibaba Cloud) Setup

### 1. Create Alibaba Cloud Account

1. Visit [https://www.alibabacloud.com](https://www.alibabacloud.com)
2. Register for an account
3. Complete real-name verification

### 2. Enable Dashscope Service

1. Go to [Dashscope Console](https://dashscope.console.aliyun.com/)
2. Enable the service
3. Choose your pricing plan

### 3. Create API Key

1. In Dashscope console, go to API-KEY management
2. Create a new API key
3. Copy the key

### 4. Configure Environment

```bash
export DASHSCOPE_API_KEY="sk-your-dashscope-api-key"
```

### 5. Test Configuration

```go
config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")
client, err := anyi.NewClient("dashscope-test", config)
// Test with a message...
```

## DeepSeek Setup

### 1. Create DeepSeek Account

1. Visit [https://platform.deepseek.ai](https://platform.deepseek.ai)
2. Sign up for an account
3. Add payment method if required

### 2. Generate API Key

1. Go to API key management
2. Create a new key
3. Copy the key securely

### 3. Configure Environment

```bash
export DEEPSEEK_API_KEY="sk-your-deepseek-api-key"
```

### 4. Test Configuration

```go
config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")
client, err := anyi.NewClient("deepseek-test", config)
// Test with a coding question...
```

## SiliconCloud Setup

### 1. Create SiliconCloud Account

1. Visit [https://siliconflow.cn](https://siliconflow.cn)
2. Register for an account
3. Complete verification process

### 2. Generate API Key

1. Access your dashboard
2. Go to API key section
3. Create a new key
4. Save the key securely

### 3. Configure Environment

```bash
export SILICONCLOUD_API_KEY="sk-your-siliconcloud-key"
```

### 4. Test Configuration

```go
config := siliconcloud.DefaultConfig(os.Getenv("SILICONCLOUD_API_KEY"), "deepseek-chat")
client, err := anyi.NewClient("silicon-test", config)
// Test with a message...
```

## Environment Management

### Using .env Files

For development, create a `.env` file:

```bash
# .env
OPENAI_API_KEY=sk-your-openai-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
AZ_OPENAI_API_KEY=your-azure-key
AZ_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZ_OPENAI_MODEL_DEPLOYMENT_ID=your-deployment
ZHIPU_API_KEY=your-zhipu-key
DASHSCOPE_API_KEY=sk-your-dashscope-key
DEEPSEEK_API_KEY=sk-your-deepseek-key
SILICONCLOUD_API_KEY=sk-your-silicon-key
```

Load it in your application:

```go
import "github.com/joho/godotenv"

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found")
    }
}
```

### Production Environment

For production, use your platform's secret management:

**Docker:**

```bash
docker run -e OPENAI_API_KEY=sk-your-key your-app
```

**Kubernetes:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: llm-secrets
type: Opaque
stringData:
  openai-api-key: sk-your-key
  anthropic-api-key: sk-ant-your-key
```

**Cloud Platforms:**

- AWS: AWS Secrets Manager
- Azure: Azure Key Vault
- GCP: Secret Manager

## Troubleshooting

### Common Issues

**Authentication Errors:**

- Verify API key is correct
- Check if key has necessary permissions
- Ensure billing is set up (for paid services)

**Network Errors:**

- Check internet connectivity
- Verify firewall settings
- Try different base URLs if available

**Rate Limiting:**

- Implement exponential backoff
- Check your usage limits
- Consider upgrading your plan

**Model Not Found:**

- Verify model name spelling
- Check if model is available in your region
- For Azure: ensure model is deployed

### Getting Help

If you encounter issues:

1. Check the provider's documentation
2. Look at Anyi's [GitHub issues](https://github.com/jieliu2000/anyi/issues)
3. Join the community discussions
4. Contact the provider's support if needed

## Next Steps

Once you have your providers set up:

1. Learn about [Error Handling](error-handling.md) for robust applications
2. Explore [Performance Optimization](performance.md) techniques
3. Check out [Configuration Management](../tutorials/configuration.md) for advanced setups
