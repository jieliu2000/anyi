# 错误处理指南

本指南将介绍如何在 Anyi 应用中实施有效的错误处理策略，包括重试机制、降级策略和监控。

## 错误类型

### 1. 网络错误

**常见场景：**

- 网络连接超时
- DNS 解析失败
- 连接被拒绝

**示例：**

```go
package main

import (
    "errors"
    "log"
    "net"
    "time"
    "github.com/jieliu2000/anyi"
)

func handleNetworkErrors(client anyi.Client, messages []chat.Message) (*chat.Message, error) {
    response, _, err := client.Chat(messages, nil)
    if err != nil {
        // 检查是否为网络错误
        if isNetworkError(err) {
            log.Printf("网络错误: %v", err)
            return nil, errors.New("网络连接失败，请检查网络设置")
        }
        return nil, err
    }
    return response, nil
}

func isNetworkError(err error) bool {
    if netErr, ok := err.(net.Error); ok {
        return netErr.Timeout() || netErr.Temporary()
    }
    return false
}
```

### 2. API 错误

**常见场景：**

- 401 Unauthorized (API 密钥错误)
- 429 Too Many Requests (速率限制)
- 500 Internal Server Error (服务器错误)

**示例：**

```go
package main

import (
    "fmt"
    "strings"
    "github.com/jieliu2000/anyi"
)

func handleAPIErrors(client anyi.Client, messages []chat.Message) (*chat.Message, error) {
    response, _, err := client.Chat(messages, nil)
    if err != nil {
        errorMsg := err.Error()

        switch {
        case strings.Contains(errorMsg, "401"):
            return nil, fmt.Errorf("API 密钥无效或已过期")
        case strings.Contains(errorMsg, "429"):
            return nil, fmt.Errorf("请求频率过高，请稍后重试")
        case strings.Contains(errorMsg, "500"):
            return nil, fmt.Errorf("服务器内部错误，请稍后重试")
        case strings.Contains(errorMsg, "quota"):
            return nil, fmt.Errorf("API 配额已用完")
        default:
            return nil, fmt.Errorf("API 调用失败: %v", err)
        }
    }
    return response, nil
}
```

### 3. 验证错误

**常见场景：**

- 输出长度不符合要求
- 输出格式不正确
- 内容不符合预期

**示例：**

```go
package main

import (
    "fmt"
    "strings"
    "github.com/jieliu2000/anyi"
)

func validateResponse(response *chat.Message, minLength int, requiredWords []string) error {
    // 检查长度
    if len(response.Content) < minLength {
        return fmt.Errorf("响应太短，需要至少 %d 字符", minLength)
    }

    // 检查必需词汇
    content := strings.ToLower(response.Content)
    for _, word := range requiredWords {
        if !strings.Contains(content, strings.ToLower(word)) {
            return fmt.Errorf("响应缺少必需词汇: %s", word)
        }
    }

    return nil
}
```

## 重试机制

### 1. 简单重试

```go
package main

import (
    "log"
    "time"
    "github.com/jieliu2000/anyi"
)

func chatWithRetry(client anyi.Client, messages []chat.Message, maxRetries int) (*chat.Message, error) {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        response, _, err := client.Chat(messages, nil)
        if err == nil {
            return response, nil
        }

        lastErr = err
        log.Printf("尝试 %d/%d 失败: %v", attempt+1, maxRetries, err)

        // 如果不是可重试的错误，直接返回
        if !isRetryableError(err) {
            return nil, err
        }

        // 等待后重试
        if attempt < maxRetries-1 {
            time.Sleep(time.Second * time.Duration(attempt+1))
        }
    }

    return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    errorMsg := err.Error()

    // 可重试的错误类型
    retryableErrors := []string{
        "timeout",
        "429", // 速率限制
        "500", // 服务器错误
        "502", // 网关错误
        "503", // 服务不可用
        "504", // 网关超时
    }

    for _, retryable := range retryableErrors {
        if strings.Contains(errorMsg, retryable) {
            return true
        }
    }

    return false
}
```

### 2. 指数退避重试

```go
package main

import (
    "math"
    "time"
    "github.com/jieliu2000/anyi"
)

type RetryConfig struct {
    MaxRetries  int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func chatWithExponentialBackoff(client anyi.Client, messages []chat.Message, config RetryConfig) (*chat.Message, error) {
    var lastErr error

    for attempt := 0; attempt < config.MaxRetries; attempt++ {
        response, _, err := client.Chat(messages, nil)
        if err == nil {
            return response, nil
        }

        lastErr = err

        if !isRetryableError(err) {
            return nil, err
        }

        if attempt < config.MaxRetries-1 {
            delay := calculateDelay(attempt, config)
            log.Printf("尝试 %d 失败，%v 后重试: %v", attempt+1, delay, err)
            time.Sleep(delay)
        }
    }

    return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", config.MaxRetries, lastErr)
}

func calculateDelay(attempt int, config RetryConfig) time.Duration {
    delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))

    if delay > float64(config.MaxDelay) {
        delay = float64(config.MaxDelay)
    }

    return time.Duration(delay)
}

// 使用示例
func main() {
    client, _ := anyi.GetClient("openai")

    config := RetryConfig{
        MaxRetries: 5,
        BaseDelay:  1 * time.Second,
        MaxDelay:   30 * time.Second,
        Multiplier: 2.0,
    }

    messages := []chat.Message{
        {Role: "user", Content: "Hello"},
    }

    response, err := chatWithExponentialBackoff(client, messages, config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("响应: %s\n", response.Content)
}
```

### 3. 带抖动的重试

```go
package main

import (
    "math/rand"
    "time"
    "github.com/jieliu2000/anyi"
)

func chatWithJitter(client anyi.Client, messages []chat.Message, maxRetries int) (*chat.Message, error) {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        response, _, err := client.Chat(messages, nil)
        if err == nil {
            return response, nil
        }

        lastErr = err

        if !isRetryableError(err) {
            return nil, err
        }

        if attempt < maxRetries-1 {
            // 基础延迟 + 随机抖动
            baseDelay := time.Duration(attempt+1) * time.Second
            jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
            delay := baseDelay + jitter

            log.Printf("尝试 %d 失败，%v 后重试", attempt+1, delay)
            time.Sleep(delay)
        }
    }

    return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}
```

## 降级策略

### 1. 多提供商降级

```go
package main

import (
    "log"
    "github.com/jieliu2000/anyi"
)

type FallbackClient struct {
    clients []anyi.Client
    names   []string
}

func NewFallbackClient(clientConfigs []struct{Name string; Client anyi.Client}) *FallbackClient {
    clients := make([]anyi.Client, len(clientConfigs))
    names := make([]string, len(clientConfigs))

    for i, config := range clientConfigs {
        clients[i] = config.Client
        names[i] = config.Name
    }

    return &FallbackClient{
        clients: clients,
        names:   names,
    }
}

func (fc *FallbackClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, *chat.ChatInfo, error) {
    var lastErr error

    for i, client := range fc.clients {
        response, info, err := client.Chat(messages, options)
        if err == nil {
            if i > 0 {
                log.Printf("使用降级客户端 %s", fc.names[i])
            }
            return response, info, nil
        }

        lastErr = err
        log.Printf("客户端 %s 失败: %v", fc.names[i], err)
    }

    return nil, nil, fmt.Errorf("所有客户端都失败了，最后错误: %v", lastErr)
}

// 使用示例
func main() {
    openaiClient, _ := anyi.GetClient("openai")
    claudeClient, _ := anyi.GetClient("claude")
    ollamaClient, _ := anyi.GetClient("ollama")

    fallbackClient := NewFallbackClient([]struct{Name string; Client anyi.Client}{
        {"OpenAI", openaiClient},
        {"Claude", claudeClient},
        {"Ollama", ollamaClient},
    })

    messages := []chat.Message{
        {Role: "user", Content: "Hello"},
    }

    response, _, err := fallbackClient.Chat(messages, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("响应: %s\n", response.Content)
}
```

### 2. 模型降级

```go
package main

import (
    "log"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

type ModelFallback struct {
    client anyi.Client
    models []string
}

func NewModelFallback(apiKey string, models []string) *ModelFallback {
    config := openai.DefaultConfig(apiKey)
    client, _ := anyi.NewClient("openai", config)

    return &ModelFallback{
        client: client,
        models: models,
    }
}

func (mf *ModelFallback) Chat(messages []chat.Message) (*chat.Message, error) {
    var lastErr error

    for i, model := range mf.models {
        // 更新客户端配置
        config := openai.Config{
            APIKey: os.Getenv("OPENAI_API_KEY"),
            Model:  model,
        }

        client, err := anyi.NewClient("openai", config)
        if err != nil {
            lastErr = err
            continue
        }

        response, _, err := client.Chat(messages, nil)
        if err == nil {
            if i > 0 {
                log.Printf("使用降级模型: %s", model)
            }
            return response, nil
        }

        lastErr = err
        log.Printf("模型 %s 失败: %v", model, err)
    }

    return nil, fmt.Errorf("所有模型都失败了: %v", lastErr)
}

// 使用示例
func main() {
    models := []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo"}
    fallback := NewModelFallback(os.Getenv("OPENAI_API_KEY"), models)

    messages := []chat.Message{
        {Role: "user", Content: "Hello"},
    }

    response, err := fallback.Chat(messages)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("响应: %s\n", response.Content)
}
```

### 3. 功能降级

```go
package main

import (
    "log"
    "strings"
    "github.com/jieliu2000/anyi"
)

type GracefulDegradation struct {
    client anyi.Client
}

func NewGracefulDegradation(client anyi.Client) *GracefulDegradation {
    return &GracefulDegradation{client: client}
}

func (gd *GracefulDegradation) ProcessText(text string, complexity string) (string, error) {
    var prompt string

    switch complexity {
    case "high":
        prompt = fmt.Sprintf("请详细分析以下文本，包括语义、情感、主题等多个维度：\n%s", text)
    case "medium":
        prompt = fmt.Sprintf("请分析以下文本的主要内容和情感：\n%s", text)
    case "low":
        prompt = fmt.Sprintf("请简要总结以下文本：\n%s", text)
    default:
        prompt = text
    }

    messages := []chat.Message{
        {Role: "user", Content: prompt},
    }

    response, _, err := gd.client.Chat(messages, nil)
    if err != nil {
        // 如果高复杂度失败，尝试中等复杂度
        if complexity == "high" {
            log.Println("高复杂度处理失败，降级到中等复杂度")
            return gd.ProcessText(text, "medium")
        }

        // 如果中等复杂度失败，尝试低复杂度
        if complexity == "medium" {
            log.Println("中等复杂度处理失败，降级到低复杂度")
            return gd.ProcessText(text, "low")
        }

        // 如果低复杂度也失败，返回简单处理结果
        if complexity == "low" {
            log.Println("AI 处理失败，返回简单处理结果")
            return gd.simpleTextProcessing(text), nil
        }

        return "", err
    }

    return response.Content, nil
}

func (gd *GracefulDegradation) simpleTextProcessing(text string) string {
    // 简单的文本处理逻辑
    words := strings.Fields(text)
    if len(words) > 50 {
        return strings.Join(words[:50], " ") + "..."
    }
    return text
}
```

## 监控和日志

### 1. 错误监控

```go
package main

import (
    "log"
    "sync"
    "time"
    "github.com/jieliu2000/anyi"
)

type ErrorMonitor struct {
    errors    []ErrorRecord
    mutex     sync.RWMutex
    threshold int
    window    time.Duration
}

type ErrorRecord struct {
    Timestamp time.Time
    Error     string
    Client    string
}

func NewErrorMonitor(threshold int, window time.Duration) *ErrorMonitor {
    return &ErrorMonitor{
        errors:    make([]ErrorRecord, 0),
        threshold: threshold,
        window:    window,
    }
}

func (em *ErrorMonitor) RecordError(clientName string, err error) {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    record := ErrorRecord{
        Timestamp: time.Now(),
        Error:     err.Error(),
        Client:    clientName,
    }

    em.errors = append(em.errors, record)

    // 检查是否超过阈值
    if em.getRecentErrorCount() >= em.threshold {
        em.triggerAlert()
    }
}

func (em *ErrorMonitor) getRecentErrorCount() int {
    cutoff := time.Now().Add(-em.window)
    count := 0

    for _, record := range em.errors {
        if record.Timestamp.After(cutoff) {
            count++
        }
    }

    return count
}

func (em *ErrorMonitor) triggerAlert() {
    log.Printf("警告: %v 内发生了 %d 个错误，超过阈值 %d",
        em.window, em.getRecentErrorCount(), em.threshold)

    // 这里可以发送告警通知
    // sendAlert()
}

func (em *ErrorMonitor) GetErrorStats() map[string]int {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    stats := make(map[string]int)
    cutoff := time.Now().Add(-em.window)

    for _, record := range em.errors {
        if record.Timestamp.After(cutoff) {
            stats[record.Client]++
        }
    }

    return stats
}
```

### 2. 性能监控

```go
package main

import (
    "log"
    "time"
    "github.com/jieliu2000/anyi"
)

type PerformanceMonitor struct {
    client anyi.Client
    name   string
}

func NewPerformanceMonitor(client anyi.Client, name string) *PerformanceMonitor {
    return &PerformanceMonitor{
        client: client,
        name:   name,
    }
}

func (pm *PerformanceMonitor) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, *chat.ChatInfo, error) {
    start := time.Now()

    response, info, err := pm.client.Chat(messages, options)

    duration := time.Since(start)

    // 记录性能指标
    if err != nil {
        log.Printf("客户端 %s 请求失败 (耗时: %v): %v", pm.name, duration, err)
    } else {
        log.Printf("客户端 %s 请求成功 (耗时: %v, tokens: %d)",
            pm.name, duration, info.TotalTokens)
    }

    // 检查性能阈值
    if duration > 30*time.Second {
        log.Printf("警告: 客户端 %s 响应时间过长: %v", pm.name, duration)
    }

    return response, info, err
}
```

## 最佳实践

### 1. 错误分类

```go
type ErrorType int

const (
    NetworkError ErrorType = iota
    AuthError
    RateLimitError
    ValidationError
    ServerError
)

func classifyError(err error) ErrorType {
    errorMsg := err.Error()

    switch {
    case strings.Contains(errorMsg, "401"):
        return AuthError
    case strings.Contains(errorMsg, "429"):
        return RateLimitError
    case strings.Contains(errorMsg, "timeout"):
        return NetworkError
    case strings.Contains(errorMsg, "500"):
        return ServerError
    default:
        return ValidationError
    }
}
```

### 2. 用户友好的错误消息

```go
func getUserFriendlyError(err error) string {
    errorType := classifyError(err)

    switch errorType {
    case NetworkError:
        return "网络连接出现问题，请检查网络设置后重试"
    case AuthError:
        return "身份验证失败，请检查 API 密钥设置"
    case RateLimitError:
        return "请求过于频繁，请稍后再试"
    case ServerError:
        return "服务暂时不可用，请稍后重试"
    default:
        return "处理请求时出现错误，请重试"
    }
}
```

### 3. 结构化错误处理

```go
type APIError struct {
    Type    ErrorType
    Message string
    Code    int
    Retry   bool
}

func (e *APIError) Error() string {
    return e.Message
}

func (e *APIError) IsRetryable() bool {
    return e.Retry
}

func handleError(err error) *APIError {
    if apiErr, ok := err.(*APIError); ok {
        return apiErr
    }

    // 将普通错误转换为结构化错误
    errorType := classifyError(err)
    return &APIError{
        Type:    errorType,
        Message: getUserFriendlyError(err),
        Retry:   errorType == NetworkError || errorType == RateLimitError || errorType == ServerError,
    }
}
```

## 下一步

现在您已经掌握了错误处理技巧，可以：

1. 学习 [性能优化](performance.md) 来提升应用性能
2. 了解 [Web 集成](web-integration.md) 来构建 Web 应用
3. 探索 [安全最佳实践](../advanced/security.md) 来保护您的应用
4. 查看 [部署指南](../advanced/deployment.md) 来部署到生产环境

通过实施有效的错误处理策略，您可以构建更可靠、用户体验更好的 AI 应用程序！
