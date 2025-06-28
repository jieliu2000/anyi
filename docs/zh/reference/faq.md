# 常见问题解答 (FAQ)

本文档回答有关使用 Anyi 的常见问题。如果您在这里找不到您的问题，请查看我们的 [GitHub 讨论](https://github.com/jieliu2000/anyi/discussions) 或提出问题。

## 入门指南

### 问：Anyi 的系统要求是什么？

**答：** Anyi 需要：

- Go 1.20 或更高版本
- 访问 LLM API 的互联网连接（除非使用本地模型）
- 为您选择的模型提供足够的内存（特别是本地部署）

### 问：我可以在没有互联网连接的情况下使用 Anyi 吗？

**答：** 可以，但只能通过 Ollama 使用本地模型。您需要：

1. 安装 Ollama
2. 在本地下载模型
3. 在您的 Anyi 应用程序中仅使用 Ollama 提供商

### 问：我应该选择哪个 LLM 提供商？

**答：** 这取决于您的需求：

- **通用用途**：OpenAI GPT-3.5-turbo 性价比高，GPT-4 质量更高
- **中文语言**：智谱 AI 或通义千问
- **代码生成**：DeepSeek Coder 或通义千问 Qwen-Max
- **隐私/离线**：Ollama 配合本地模型
- **企业级**：Azure OpenAI 或 SiliconCloud

## 配置和设置

### 问：如何安全地管理 API 密钥？

**答：** API 密钥管理的最佳实践：

```go
// 使用环境变量
config.APIKey = os.Getenv("OPENAI_API_KEY")

// 开发环境使用 .env 文件
// 生产环境使用云端密钥管理器
```

永远不要在源代码中硬编码 API 密钥。

### 问：我可以在同一个应用程序中使用多个 LLM 提供商吗？

**答：** 可以！这是 Anyi 的关键特性之一：

```go
// 创建多个客户端
openaiClient, _ := anyi.NewClient("openai", openaiConfig)
ollamaClient, _ := anyi.NewClient("ollama", ollamaConfig)

// 为不同任务使用不同客户端
quickResponse, _, _ := openaiClient.Chat(simpleMessages, nil)
privateResponse, _, _ := ollamaClient.Chat(sensitiveMessages, nil)
```

### 问：如何在同一提供商的不同模型之间切换？

**答：** 为不同配置创建单独的客户端：

```go
// 简单任务的快速模型
fastConfig := openai.NewConfigWithModel(apiKey, "gpt-3.5-turbo")
fastClient, _ := anyi.NewClient("fast", fastConfig)

// 复杂任务的强大模型
powerConfig := openai.NewConfigWithModel(apiKey, "gpt-4")
powerClient, _ := anyi.NewClient("power", powerConfig)
```

## 工作流和开发

### 问：如何处理超过令牌限制的大文本？

**答：** 几种方法：

1. **分块处理**：将大文本分成较小的片段

```go
func processLargeText(text string, client *llm.Client) (string, error) {
    chunks := splitIntoChunks(text, 1000) // 每块约 1000 字
    var results []string

    for _, chunk := range chunks {
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: "处理：" + chunk},
        }, nil)
        if err != nil {
            return "", err
        }
        results = append(results, response.Content)
    }

    return combineResults(results), nil
}
```

2. **摘要处理**：使用多步工作流先摘要再处理
3. **流式处理**：对于支持的提供商，使用流式响应

### 问：如何确保工作流在不稳定网络条件下正常工作？

**答：** 使用 Anyi 的内置重试机制：

```go
step := flow.Step{
    Name: "robust_step",
    Executor: myExecutor,
    MaxRetryTimes: 3,  // 最多重试 3 次
    Validator: myValidator,
}
```

您也可以实现指数退避：

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
            backoff *= 2 // 指数退避
        }
    }

    return fmt.Errorf("在 %d 次重试后失败", maxRetries)
}
```

### 问：如何将 Anyi 与现有的 Go Web 框架集成？

**答：** Anyi 与任何 Go Web 框架无缝集成。Gin 示例：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jieliu2000/anyi"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 初始化 Anyi 客户端
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

## 性能和优化

### 问：如何优化性能并降低成本？

**答：** 几种策略：

1. **选择合适的模型**：简单任务使用更便宜的模型
2. **实现缓存**：缓存常见响应
3. **设置令牌限制**：使用 `MaxTokens` 控制响应长度
4. **使用本地模型**：对于频繁的非关键任务
5. **批量处理**：将相似请求组合在一起

```go
// 示例：分层方法
func processRequest(complexity string, content string) (*chat.Message, error) {
    var client *llm.Client

    switch complexity {
    case "simple":
        client, _ = anyi.GetClient("fast")     // GPT-3.5-turbo
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

### 问：如何监控令牌使用和成本？

**答：** 通过响应信息跟踪使用情况：

```go
response, info, err := client.Chat(messages, nil)
if err == nil {
    totalTokens := info.PromptTokens + info.CompletionTokens
    log.Printf("请求使用了 %d 个令牌（提示：%d，完成：%d）",
        totalTokens, info.PromptTokens, info.CompletionTokens)

    // 估算成本（OpenAI 示例）
    cost := estimateOpenAICost(info.Model, info.PromptTokens, info.CompletionTokens)
    log.Printf("估算成本：$%.4f", cost)
}
```

## 故障排除

### 问：我遇到"创建客户端失败"错误。应该检查什么？

**答：** 常见原因和解决方案：

1. **API 密钥问题**：

   - 验证 API 密钥是否正确
   - 检查密钥是否有必要权限
   - 确保已设置计费（对于付费服务）

2. **网络问题**：

   - 检查互联网连接
   - 验证防火墙设置
   - 如果可用，尝试使用不同的基础 URL

3. **配置问题**：
   - 检查是否设置了所有必需的环境变量
   - 验证模型名称是否正确
   - 对于 Azure：确保模型已部署

### 问：我的请求因速率限制错误而失败。如何处理？

**答：** 实现速率限制和退避：

```go
func handleRateLimit(client *llm.Client, messages []chat.Message) (*chat.Message, error) {
    maxRetries := 3
    baseDelay := 1 * time.Second

    for i := 0; i < maxRetries; i++ {
        response, _, err := client.Chat(messages, nil)

        if err == nil {
            return response, nil
        }

        // 检查是否是速率限制错误
        if strings.Contains(err.Error(), "rate limit") {
            delay := baseDelay * time.Duration(1<<i) // 指数退避
            log.Printf("速率受限，等待 %v 后重试 %d", delay, i+1)
            time.Sleep(delay)
            continue
        }

        return nil, err // 非速率限制错误
    }

    return nil, fmt.Errorf("在 %d 次重试后失败", maxRetries)
}
```

### 问：如何调试工作流执行问题？

**答：** 启用详细日志记录并使用步骤验证：

```go
// 为您的执行器添加日志记录
type LoggingExecutor struct {
    *anyi.LLMExecutor
    Logger *log.Logger
}

func (e *LoggingExecutor) Run(ctx flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    e.Logger.Printf("执行步骤：%s，输入：%s", step.Name, ctx.Text)

    result, err := e.LLMExecutor.Run(ctx, step)

    if err != nil {
        e.Logger.Printf("步骤 %s 失败：%v", step.Name, err)
        return nil, err
    }

    e.Logger.Printf("步骤 %s 完成：%s", step.Name, result.Text)
    return result, nil
}
```

## 高级用法

### 问：我可以创建自定义执行器吗？

**答：** 可以！实现 `Executor` 接口：

```go
type CustomExecutor struct {
    // 您的字段
}

func (e *CustomExecutor) Run(ctx flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // 您的自定义逻辑
    result := processCustomLogic(ctx.Text)

    return &flow.FlowContext{
        Text: result,
        Memory: ctx.Memory,
        // ... 其他字段
    }, nil
}

func (e *CustomExecutor) Init() error {
    // 初始化逻辑
    return nil
}
```

### 问：如何在工作流中实现条件逻辑？

**答：** 使用 `ConditionalFlowExecutor`：

```go
conditionalExecutor := &anyi.ConditionalFlowExecutor{
    Condition: "{{contains .Text '紧急'}}",
    TrueFlow:  urgentProcessingFlow,
    FalseFlow: normalProcessingFlow,
}

step := flow.Step{
    Name: "路由",
    Executor: conditionalExecutor,
}
```

### 问：我可以在微服务架构中使用 Anyi 吗？

**答：** 当然可以！每个微服务都可以有自己的 Anyi 配置：

```go
// 服务 A：文本处理服务
func main() {
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, _ := anyi.NewClient("text-processor", config)

    // 设置使用客户端的 HTTP 处理程序
    setupTextProcessingHandlers(client)
}

// 服务 B：图像分析服务
func main() {
    config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4-vision")
    client, _ := anyi.NewClient("image-analyzer", config)

    setupImageAnalysisHandlers(client)
}
```

## 社区和支持

### 问：如果我遇到困难，在哪里可以获得帮助？

**答：** 几个选项：

1. **文档**：查看此文档和 [API 参考](api.md)
2. **示例**：浏览 [示例目录](../../../examples/)
3. **GitHub 问题**：搜索现有 [问题](https://github.com/jieliu2000/anyi/issues) 或创建新问题
4. **讨论**：加入 [GitHub 讨论](https://github.com/jieliu2000/anyi/discussions)
5. **社区**：与社区中的其他用户联系

### 问：我如何为 Anyi 做贡献？

**答：** 我们欢迎贡献！您可以：

1. **报告错误**：为您发现的任何错误打开问题
2. **建议功能**：提出新功能或改进
3. **提交代码**：为错误修复或新功能创建拉取请求
4. **改进文档**：帮助使文档更好
5. **分享示例**：贡献示例应用程序

### 问：Anyi 适合生产使用吗？

**答：** 是的！Anyi 专为生产使用而设计，具有以下特性：

- 强大的错误处理和重试机制
- 支持多个 LLM 提供商
- 不同环境的配置管理
- 全面的测试和验证
- 性能优化功能

许多用户在生产环境中成功运行 Anyi。

## 性能调优

### 问：如何提高 Anyi 应用程序的响应速度？

**答：** 几种优化策略：

1. **连接池**：重用 HTTP 连接

```go
// 配置 HTTP 客户端连接池
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}

httpClient := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

2. **并发处理**：并行处理多个请求

```go
func processConcurrently(items []string, client *llm.Client) []string {
    var wg sync.WaitGroup
    results := make([]string, len(items))

    for i, item := range items {
        wg.Add(1)
        go func(index int, text string) {
            defer wg.Done()
            response, _, _ := client.Chat([]chat.Message{
                {Role: "user", Content: text},
            }, nil)
            results[index] = response.Content
        }(i, item)
    }

    wg.Wait()
    return results
}
```

3. **缓存策略**：缓存频繁使用的响应

```go
type ResponseCache struct {
    cache map[string]*chat.Message
    mutex sync.RWMutex
}

func (c *ResponseCache) Get(key string) (*chat.Message, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    msg, exists := c.cache[key]
    return msg, exists
}

func (c *ResponseCache) Set(key string, msg *chat.Message) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.cache[key] = msg
}
```

### 问：如何监控和分析 Anyi 应用程序的性能？

**答：** 实现监控和指标收集：

```go
type Metrics struct {
    RequestCount    int64
    ErrorCount      int64
    TotalTokens     int64
    AverageLatency  time.Duration
    mutex           sync.RWMutex
}

func (m *Metrics) RecordRequest(tokens int, latency time.Duration, err error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.RequestCount++
    m.TotalTokens += int64(tokens)

    if err != nil {
        m.ErrorCount++
    }

    // 计算移动平均延迟
    m.AverageLatency = (m.AverageLatency + latency) / 2
}

func (m *Metrics) GetStats() (int64, int64, int64, time.Duration) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.RequestCount, m.ErrorCount, m.TotalTokens, m.AverageLatency
}
```

## 安全性

### 问：如何确保 Anyi 应用程序的安全性？

**答：** 遵循安全最佳实践：

1. **输入验证**：始终验证用户输入

```go
func validateInput(input string) error {
    if len(input) > 10000 {
        return errors.New("输入过长")
    }

    // 检查恶意内容
    forbidden := []string{"<script>", "javascript:", "data:"}
    for _, pattern := range forbidden {
        if strings.Contains(strings.ToLower(input), pattern) {
            return errors.New("输入包含禁止内容")
        }
    }

    return nil
}
```

2. **API 密钥保护**：安全存储和轮换 API 密钥

```go
// 使用云端密钥管理服务
func getAPIKey() (string, error) {
    // 从 AWS Secrets Manager、Azure Key Vault 等获取
    return secretsManager.GetSecret("anyi-api-key")
}
```

3. **速率限制**：防止滥用

```go
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}

func (rl *RateLimiter) Allow(clientID string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    now := time.Now()
    if rl.requests[clientID] == nil {
        rl.requests[clientID] = []time.Time{}
    }

    // 清理过期请求
    var validRequests []time.Time
    for _, req := range rl.requests[clientID] {
        if now.Sub(req) < rl.window {
            validRequests = append(validRequests, req)
        }
    }

    if len(validRequests) >= rl.limit {
        return false
    }

    rl.requests[clientID] = append(validRequests, now)
    return true
}
```

## 仍有问题？

如果您的问题在这里没有得到回答：

1. 搜索 [GitHub 仓库](https://github.com/jieliu2000/anyi)
2. 查看 [GitHub 讨论](https://github.com/jieliu2000/anyi/discussions)
3. 使用详细信息打开新的 [问题](https://github.com/jieliu2000/anyi/issues/new)
4. 浏览 [示例](../../../examples/) 获取代码样本
