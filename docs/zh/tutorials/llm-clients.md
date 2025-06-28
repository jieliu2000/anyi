# LLM 客户端教程

本教程将详细介绍如何在 Anyi 中使用各种 LLM 提供商，包括配置、认证和最佳实践。

## 概述

Anyi 支持多个 LLM 提供商，提供统一的接口来访问不同的 AI 模型：

### 支持的提供商

| 提供商       | 类型 | 主要模型                  | 特点             |
| ------------ | ---- | ------------------------- | ---------------- |
| OpenAI       | 云端 | GPT-4, GPT-3.5            | 高质量，广泛支持 |
| Anthropic    | 云端 | Claude 3.5, Claude 3      | 安全，长上下文   |
| Azure OpenAI | 云端 | GPT-4, GPT-3.5            | 企业级，合规性   |
| Ollama       | 本地 | Llama, Mistral, CodeLlama | 隐私，离线使用   |
| 智谱 AI      | 云端 | GLM-4, ChatGLM            | 中文优化         |
| 通义千问     | 云端 | Qwen-Max, Qwen-Plus       | 阿里云，中文     |
| DeepSeek     | 云端 | DeepSeek-Chat             | 代码生成         |
| SiliconCloud | 云端 | 多种模型                  | 高性价比         |

## OpenAI 客户端

### 基本设置

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
    // 创建 OpenAI 配置
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

    // 可选：自定义配置
    config.Model = "gpt-4"
    config.Temperature = 0.7
    config.MaxTokens = 1000

    // 创建客户端
    client, err := anyi.NewClient("openai", config)
    if err != nil {
        log.Fatalf("创建 OpenAI 客户端失败: %v", err)
    }

    // 准备消息
    messages := []chat.Message{
        {
            Role:    "system",
            Content: "你是一个专业的编程助手。",
        },
        {
            Role:    "user",
            Content: "请解释 Go 语言的 goroutine。",
        },
    }

    // 发送请求
    response, info, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("回复: %s\n", response.Content)
    fmt.Printf("使用 tokens: %d\n", info.TotalTokens)
}
```

### 高级配置

```go
config := openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Model:       "gpt-4-turbo-preview",
    Temperature: 0.3,
    MaxTokens:   2000,
    TopP:        0.9,
    BaseURL:     "https://api.openai.com/v1", // 自定义端点
    Timeout:     30 * time.Second,
}
```

### 流式响应

```go
func streamExample() {
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, _ := anyi.NewClient("openai", config)

    messages := []chat.Message{
        {Role: "user", Content: "写一首关于编程的诗"},
    }

    // 创建流式选项
    options := &chat.ChatOptions{
        Stream: true,
        StreamCallback: func(content string) {
            fmt.Print(content) // 实时输出
        },
    }

    response, _, err := client.Chat(messages, options)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n完整回复: %s\n", response.Content)
}
```

## Anthropic Claude 客户端

### 基本设置

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/anthropic"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 创建 Anthropic 配置
    config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))
    config.Model = "claude-3-5-sonnet-20241022"
    config.MaxTokens = 1000

    // 创建客户端
    client, err := anyi.NewClient("claude", config)
    if err != nil {
        log.Fatalf("创建 Claude 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请分析一下函数式编程的优缺点。",
        },
    }

    response, info, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("Claude 回复: %s\n", response.Content)
    fmt.Printf("输入 tokens: %d, 输出 tokens: %d\n",
        info.PromptTokens, info.CompletionTokens)
}
```

### Claude 特殊功能

```go
// 长上下文处理
config := anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    Model:     "claude-3-5-sonnet-20241022",
    MaxTokens: 4000,
    // Claude 支持非常长的上下文
}

// 系统消息
messages := []chat.Message{
    {
        Role:    "system",
        Content: "你是一个专业的技术写作专家，擅长将复杂概念解释得简单易懂。",
    },
    {
        Role:    "user",
        Content: "请写一篇关于微服务架构的文章。",
    },
}
```

## Azure OpenAI 客户端

### 基本设置

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/azureopenai"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // Azure OpenAI 配置
    config := azureopenai.Config{
        APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
        Endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
        Model:      "gpt-4", // 您在 Azure 中部署的模型名称
        APIVersion: "2024-02-15-preview",
    }

    client, err := anyi.NewClient("azure-openai", config)
    if err != nil {
        log.Fatalf("创建 Azure OpenAI 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请介绍一下 Azure 的主要服务。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("Azure OpenAI 回复: %s\n", response.Content)
}
```

### 企业配置

```go
config := azureopenai.Config{
    APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
    Endpoint:   "https://your-resource.openai.azure.com/",
    Model:      "gpt-4-32k", // 32K 上下文版本
    APIVersion: "2024-02-15-preview",

    // 企业级设置
    Temperature: 0.1, // 更保守的输出
    MaxTokens:   2000,
    TopP:        0.95,
}
```

## Ollama 本地客户端

### 安装和设置

首先安装 Ollama：

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows - 下载安装包
# https://ollama.ai/download/windows
```

启动 Ollama 并下载模型：

```bash
# 启动服务
ollama serve

# 下载模型
ollama pull llama3
ollama pull codellama
ollama pull mistral
```

### 基本使用

```go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/ollama"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // Ollama 配置
    config := ollama.DefaultConfig("llama3")
    // 可选：自定义端点
    config.BaseURL = "http://localhost:11434"

    client, err := anyi.NewClient("ollama", config)
    if err != nil {
        log.Fatalf("创建 Ollama 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请用中文解释什么是机器学习。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("Llama 回复: %s\n", response.Content)
}
```

### 代码专用模型

```go
// 使用 CodeLlama 进行代码生成
config := ollama.Config{
    Model:       "codellama",
    BaseURL:     "http://localhost:11434",
    Temperature: 0.1, // 代码生成使用较低温度
}

client, _ := anyi.NewClient("codellama", config)

messages := []chat.Message{
    {
        Role:    "user",
        Content: "请写一个 Go 语言的快速排序函数。",
    },
}

response, _, _ := client.Chat(messages, nil)
fmt.Printf("代码:\n%s\n", response.Content)
```

## 中文 LLM 提供商

### 智谱 AI

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/zhipu"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 智谱 AI 配置
    config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"))
    config.Model = "glm-4"

    client, err := anyi.NewClient("zhipu", config)
    if err != nil {
        log.Fatalf("创建智谱客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请用中文介绍一下中国的人工智能发展现状。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("智谱回复: %s\n", response.Content)
}
```

### 通义千问

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/dashscope"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 通义千问配置
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"))
    config.Model = "qwen-max"

    client, err := anyi.NewClient("qwen", config)
    if err != nil {
        log.Fatalf("创建通义千问客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请介绍一下阿里云的核心产品。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("通义千问回复: %s\n", response.Content)
}
```

## DeepSeek 客户端

### 基本设置

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // DeepSeek 配置
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"))
    config.Model = "deepseek-chat"

    client, err := anyi.NewClient("deepseek", config)
    if err != nil {
        log.Fatalf("创建 DeepSeek 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请写一个 Python 的机器学习数据预处理脚本。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("DeepSeek 回复: %s\n", response.Content)
}
```

## 多客户端管理

### 客户端池

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
    "github.com/jieliu2000/anyi/llm/anthropic"
    "github.com/jieliu2000/anyi/llm/ollama"
    "github.com/jieliu2000/anyi/llm/chat"
)

type ClientPool struct {
    clients map[string]anyi.Client
}

func NewClientPool() *ClientPool {
    pool := &ClientPool{
        clients: make(map[string]anyi.Client),
    }

    // 初始化多个客户端
    pool.initClients()
    return pool
}

func (p *ClientPool) initClients() {
    // OpenAI
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        config := openai.DefaultConfig(apiKey)
        if client, err := anyi.NewClient("openai", config); err == nil {
            p.clients["openai"] = client
        }
    }

    // Claude
    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        config := anthropic.DefaultConfig(apiKey)
        if client, err := anyi.NewClient("claude", config); err == nil {
            p.clients["claude"] = client
        }
    }

    // Ollama (本地)
    config := ollama.DefaultConfig("llama3")
    if client, err := anyi.NewClient("ollama", config); err == nil {
        p.clients["ollama"] = client
    }
}

func (p *ClientPool) GetClient(name string) (anyi.Client, bool) {
    client, exists := p.clients[name]
    return client, exists
}

func (p *ClientPool) ChatWithFallback(messages []chat.Message, providers []string) (*chat.Message, error) {
    for _, provider := range providers {
        if client, exists := p.GetClient(provider); exists {
            response, _, err := client.Chat(messages, nil)
            if err == nil {
                return response, nil
            }
            log.Printf("提供商 %s 失败: %v", provider, err)
        }
    }
    return nil, fmt.Errorf("所有提供商都失败了")
}

func main() {
    pool := NewClientPool()

    messages := []chat.Message{
        {Role: "user", Content: "请介绍一下人工智能的发展历史。"},
    }

    // 按优先级尝试不同提供商
    providers := []string{"openai", "claude", "ollama"}

    response, err := pool.ChatWithFallback(messages, providers)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("回复: %s\n", response.Content)
}
```

### 负载均衡

```go
type LoadBalancer struct {
    clients []anyi.Client
    current int
}

func NewLoadBalancer(clients []anyi.Client) *LoadBalancer {
    return &LoadBalancer{
        clients: clients,
        current: 0,
    }
}

func (lb *LoadBalancer) NextClient() anyi.Client {
    client := lb.clients[lb.current]
    lb.current = (lb.current + 1) % len(lb.clients)
    return client
}

func (lb *LoadBalancer) Chat(messages []chat.Message) (*chat.Message, error) {
    client := lb.NextClient()
    response, _, err := client.Chat(messages, nil)
    return response, err
}
```

## 配置文件管理

### YAML 配置

```yaml
# clients.yaml
clients:
  - name: "openai-gpt4"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"
      temperature: 0.7
      maxTokens: 2000

  - name: "claude-sonnet"
    type: "anthropic"
    config:
      apiKey: "$ANTHROPIC_API_KEY"
      model: "claude-3-5-sonnet-20241022"
      maxTokens: 1000

  - name: "local-llama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434"
      temperature: 0.5

  - name: "zhipu-glm4"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY"
      model: "glm-4"
```

### 加载配置

```go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 从配置文件加载
    err := anyi.ConfigFromFile("clients.yaml")
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 获取特定客户端
    client, err := anyi.GetClient("openai-gpt4")
    if err != nil {
        log.Fatalf("获取客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "你好，请介绍一下自己。"},
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Printf("回复: %s\n", response.Content)
}
```

## 最佳实践

### 1. 错误处理和重试

```go
func chatWithRetry(client anyi.Client, messages []chat.Message, maxRetries int) (*chat.Message, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        response, _, err := client.Chat(messages, nil)
        if err == nil {
            return response, nil
        }

        lastErr = err
        log.Printf("尝试 %d 失败: %v", i+1, err)

        // 指数退避
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
    }

    return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}
```

### 2. 成本优化

```go
type CostOptimizer struct {
    cheapClient    anyi.Client // 便宜的模型
    expensiveClient anyi.Client // 昂贵但质量高的模型
}

func (co *CostOptimizer) Chat(messages []chat.Message, requireHighQuality bool) (*chat.Message, error) {
    client := co.cheapClient
    if requireHighQuality {
        client = co.expensiveClient
    }

    return client.Chat(messages, nil)
}

// 使用示例
optimizer := &CostOptimizer{
    cheapClient:     gpt35Client,    // GPT-3.5
    expensiveClient: gpt4Client,     // GPT-4
}

// 简单任务使用便宜模型
response, _ := optimizer.Chat(simpleMessages, false)

// 复杂任务使用高质量模型
response, _ := optimizer.Chat(complexMessages, true)
```

### 3. 上下文管理

```go
type ConversationManager struct {
    client   anyi.Client
    messages []chat.Message
    maxTokens int
}

func NewConversationManager(client anyi.Client, maxTokens int) *ConversationManager {
    return &ConversationManager{
        client:    client,
        messages:  make([]chat.Message, 0),
        maxTokens: maxTokens,
    }
}

func (cm *ConversationManager) AddMessage(role, content string) {
    cm.messages = append(cm.messages, chat.Message{
        Role:    role,
        Content: content,
    })

    // 如果超过最大 token 限制，删除旧消息
    cm.trimMessages()
}

func (cm *ConversationManager) Chat(userInput string) (*chat.Message, error) {
    cm.AddMessage("user", userInput)

    response, _, err := cm.client.Chat(cm.messages, nil)
    if err != nil {
        return nil, err
    }

    cm.AddMessage("assistant", response.Content)
    return response, nil
}

func (cm *ConversationManager) trimMessages() {
    // 简化的 token 计算（实际应该使用 tokenizer）
    totalTokens := 0
    for _, msg := range cm.messages {
        totalTokens += len(msg.Content) / 4 // 粗略估算
    }

    // 如果超过限制，删除最旧的非系统消息
    for totalTokens > cm.maxTokens && len(cm.messages) > 1 {
        if cm.messages[0].Role != "system" {
            cm.messages = cm.messages[1:]
            totalTokens -= len(cm.messages[0].Content) / 4
        } else if len(cm.messages) > 2 {
            cm.messages = append(cm.messages[:1], cm.messages[2:]...)
            totalTokens -= len(cm.messages[1].Content) / 4
        } else {
            break
        }
    }
}
```

### 4. 监控和日志

```go
type MonitoredClient struct {
    client anyi.Client
    logger *log.Logger
    metrics *ClientMetrics
}

type ClientMetrics struct {
    TotalRequests   int64
    SuccessRequests int64
    FailedRequests  int64
    TotalTokens     int64
    TotalCost       float64
}

func (mc *MonitoredClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, *chat.ChatInfo, error) {
    start := time.Now()
    mc.metrics.TotalRequests++

    response, info, err := mc.client.Chat(messages, options)

    duration := time.Since(start)

    if err != nil {
        mc.metrics.FailedRequests++
        mc.logger.Printf("聊天失败 (耗时: %v): %v", duration, err)
        return nil, nil, err
    }

    mc.metrics.SuccessRequests++
    mc.metrics.TotalTokens += int64(info.TotalTokens)

    mc.logger.Printf("聊天成功 (耗时: %v, tokens: %d)", duration, info.TotalTokens)

    return response, info, nil
}

func (mc *MonitoredClient) GetMetrics() ClientMetrics {
    return *mc.metrics
}
```

## 故障排除

### 常见问题

1. **API 密钥错误**

   ```
   错误: 401 Unauthorized
   ```

   检查环境变量和密钥格式

2. **网络连接问题**

   ```
   错误: dial tcp: lookup api.openai.com: no such host
   ```

   检查网络连接，考虑使用代理

3. **模型不存在**

   ```
   错误: model 'gpt-5' not found
   ```

   使用正确的模型名称

4. **速率限制**
   ```
   错误: 429 Too Many Requests
   ```
   实施重试机制和请求限流

### 调试技巧

```go
// 启用详细日志
config := openai.DefaultConfig(apiKey)
config.Debug = true

// 添加请求/响应日志
client, _ := anyi.NewClient("openai", config)
```

## 下一步

现在您已经掌握了 LLM 客户端的使用，可以：

1. 学习 [工作流构建](workflows.md) 来创建复杂的 AI 应用
2. 了解 [配置管理](configuration.md) 来更好地管理您的设置
3. 探索 [多模态应用](multimodal.md) 来处理图像和文本
4. 查看 [性能优化](../how-to/performance.md) 来提升应用性能

通过合理选择和配置 LLM 客户端，您可以构建强大、可靠、成本效益高的 AI 应用程序！
