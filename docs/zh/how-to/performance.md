# 性能优化指南

本指南将介绍如何优化 Anyi 应用的性能，包括速度优化、成本控制和资源管理。

## 性能优化策略

### 1. 客户端连接优化

#### 连接池管理

```go
package main

import (
    "sync"
    "time"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

type ClientPool struct {
    clients []anyi.Client
    mutex   sync.RWMutex
    current int
}

func NewClientPool(apiKey string, poolSize int) *ClientPool {
    clients := make([]anyi.Client, poolSize)

    for i := 0; i < poolSize; i++ {
        config := openai.Config{
            APIKey:  apiKey,
            Timeout: 30 * time.Second,
        }
        client, _ := anyi.NewClient("openai", config)
        clients[i] = client
    }

    return &ClientPool{
        clients: clients,
        current: 0,
    }
}

func (cp *ClientPool) GetClient() anyi.Client {
    cp.mutex.Lock()
    defer cp.mutex.Unlock()

    client := cp.clients[cp.current]
    cp.current = (cp.current + 1) % len(cp.clients)
    return client
}

func (cp *ClientPool) Chat(messages []chat.Message) (*chat.Message, error) {
    client := cp.GetClient()
    response, _, err := client.Chat(messages, nil)
    return response, err
}
```

#### 连接复用

```go
package main

import (
    "net/http"
    "time"
    "github.com/jieliu2000/anyi/llm/openai"
)

func createOptimizedClient(apiKey string) anyi.Client {
    // 创建优化的 HTTP 客户端
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
    }

    httpClient := &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }

    config := openai.Config{
        APIKey:     apiKey,
        HTTPClient: httpClient,
    }

    client, _ := anyi.NewClient("openai", config)
    return client
}
```

### 2. 缓存策略

#### 响应缓存

```go
package main

import (
    "crypto/md5"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    "github.com/jieliu2000/anyi"
)

type CacheEntry struct {
    Response  *chat.Message `json:"response"`
    Timestamp time.Time     `json:"timestamp"`
    TTL       time.Duration `json:"ttl"`
}

type ResponseCache struct {
    cache map[string]*CacheEntry
    mutex sync.RWMutex
}

func NewResponseCache() *ResponseCache {
    cache := &ResponseCache{
        cache: make(map[string]*CacheEntry),
    }

    // 启动清理协程
    go cache.cleanup()
    return cache
}

func (rc *ResponseCache) Get(messages []chat.Message) (*chat.Message, bool) {
    key := rc.generateKey(messages)

    rc.mutex.RLock()
    defer rc.mutex.RUnlock()

    entry, exists := rc.cache[key]
    if !exists {
        return nil, false
    }

    // 检查是否过期
    if time.Since(entry.Timestamp) > entry.TTL {
        delete(rc.cache, key)
        return nil, false
    }

    return entry.Response, true
}

func (rc *ResponseCache) Set(messages []chat.Message, response *chat.Message, ttl time.Duration) {
    key := rc.generateKey(messages)

    rc.mutex.Lock()
    defer rc.mutex.Unlock()

    rc.cache[key] = &CacheEntry{
        Response:  response,
        Timestamp: time.Now(),
        TTL:       ttl,
    }
}

func (rc *ResponseCache) generateKey(messages []chat.Message) string {
    data, _ := json.Marshal(messages)
    hash := md5.Sum(data)
    return fmt.Sprintf("%x", hash)
}

func (rc *ResponseCache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        rc.mutex.Lock()
        now := time.Now()
        for key, entry := range rc.cache {
            if now.Sub(entry.Timestamp) > entry.TTL {
                delete(rc.cache, key)
            }
        }
        rc.mutex.Unlock()
    }
}

// 缓存客户端包装器
type CachedClient struct {
    client anyi.Client
    cache  *ResponseCache
}

func NewCachedClient(client anyi.Client) *CachedClient {
    return &CachedClient{
        client: client,
        cache:  NewResponseCache(),
    }
}

func (cc *CachedClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, *chat.ChatInfo, error) {
    // 检查缓存
    if response, found := cc.cache.Get(messages); found {
        return response, &chat.ChatInfo{}, nil
    }

    // 调用实际客户端
    response, info, err := cc.client.Chat(messages, options)
    if err != nil {
        return nil, nil, err
    }

    // 缓存响应（缓存 1 小时）
    cc.cache.Set(messages, response, time.Hour)

    return response, info, nil
}
```

#### Redis 缓存

```go
package main

import (
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
    "github.com/jieliu2000/anyi"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    return &RedisCache{client: rdb}
}

func (rc *RedisCache) Get(key string) (*chat.Message, error) {
    ctx := context.Background()
    val, err := rc.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // 缓存未命中
    }
    if err != nil {
        return nil, err
    }

    var response chat.Message
    err = json.Unmarshal([]byte(val), &response)
    return &response, err
}

func (rc *RedisCache) Set(key string, response *chat.Message, ttl time.Duration) error {
    ctx := context.Background()
    data, err := json.Marshal(response)
    if err != nil {
        return err
    }

    return rc.client.Set(ctx, key, data, ttl).Err()
}
```

### 3. 并发优化

#### 并行处理

```go
package main

import (
    "sync"
    "github.com/jieliu2000/anyi"
)

type BatchProcessor struct {
    client      anyi.Client
    concurrency int
}

func NewBatchProcessor(client anyi.Client, concurrency int) *BatchProcessor {
    return &BatchProcessor{
        client:      client,
        concurrency: concurrency,
    }
}

func (bp *BatchProcessor) ProcessBatch(inputs []string) ([]string, []error) {
    jobs := make(chan string, len(inputs))
    results := make(chan struct {
        index  int
        result string
        err    error
    }, len(inputs))

    // 启动工作协程
    var wg sync.WaitGroup
    for i := 0; i < bp.concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for input := range jobs {
                result, err := bp.processItem(input)
                results <- struct {
                    index  int
                    result string
                    err    error
                }{
                    index:  getIndex(input, inputs),
                    result: result,
                    err:    err,
                }
            }
        }()
    }

    // 发送任务
    for _, input := range inputs {
        jobs <- input
    }
    close(jobs)

    // 等待完成
    go func() {
        wg.Wait()
        close(results)
    }()

    // 收集结果
    outputs := make([]string, len(inputs))
    errors := make([]error, len(inputs))

    for result := range results {
        outputs[result.index] = result.result
        errors[result.index] = result.err
    }

    return outputs, errors
}

func (bp *BatchProcessor) processItem(input string) (string, error) {
    messages := []chat.Message{
        {Role: "user", Content: input},
    }

    response, _, err := bp.client.Chat(messages, nil)
    if err != nil {
        return "", err
    }

    return response.Content, nil
}

func getIndex(item string, slice []string) int {
    for i, v := range slice {
        if v == item {
            return i
        }
    }
    return -1
}
```

#### 工作池模式

```go
package main

import (
    "sync"
    "github.com/jieliu2000/anyi"
)

type WorkerPool struct {
    client     anyi.Client
    workerSize int
    jobQueue   chan Job
    quit       chan bool
}

type Job struct {
    ID       string
    Messages []chat.Message
    Result   chan JobResult
}

type JobResult struct {
    Response *chat.Message
    Error    error
}

func NewWorkerPool(client anyi.Client, workerSize, queueSize int) *WorkerPool {
    return &WorkerPool{
        client:     client,
        workerSize: workerSize,
        jobQueue:   make(chan Job, queueSize),
        quit:       make(chan bool),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workerSize; i++ {
        go wp.worker()
    }
}

func (wp *WorkerPool) Stop() {
    close(wp.quit)
}

func (wp *WorkerPool) Submit(job Job) {
    wp.jobQueue <- job
}

func (wp *WorkerPool) worker() {
    for {
        select {
        case job := <-wp.jobQueue:
            response, _, err := wp.client.Chat(job.Messages, nil)
            job.Result <- JobResult{
                Response: response,
                Error:    err,
            }
        case <-wp.quit:
            return
        }
    }
}

// 使用示例
func main() {
    client, _ := anyi.GetClient("openai")
    pool := NewWorkerPool(client, 5, 100)
    pool.Start()
    defer pool.Stop()

    // 提交任务
    resultChan := make(chan JobResult, 1)
    job := Job{
        ID: "task1",
        Messages: []chat.Message{
            {Role: "user", Content: "Hello"},
        },
        Result: resultChan,
    }

    pool.Submit(job)

    // 等待结果
    result := <-resultChan
    if result.Error != nil {
        log.Fatal(result.Error)
    }

    fmt.Printf("结果: %s\n", result.Response.Content)
}
```

### 4. 模型选择优化

#### 智能模型路由

```go
package main

import (
    "strings"
    "github.com/jieliu2000/anyi"
)

type ModelRouter struct {
    fastClient   anyi.Client    // 快速模型（如 GPT-3.5）
    smartClient  anyi.Client    // 智能模型（如 GPT-4）
    localClient  anyi.Client    // 本地模型（如 Ollama）
}

func NewModelRouter(fastClient, smartClient, localClient anyi.Client) *ModelRouter {
    return &ModelRouter{
        fastClient:  fastClient,
        smartClient: smartClient,
        localClient: localClient,
    }
}

func (mr *ModelRouter) Chat(messages []chat.Message) (*chat.Message, error) {
    complexity := mr.assessComplexity(messages)

    switch complexity {
    case "simple":
        return mr.chatWithFallback([]anyi.Client{mr.localClient, mr.fastClient}, messages)
    case "medium":
        return mr.chatWithFallback([]anyi.Client{mr.fastClient, mr.smartClient}, messages)
    case "complex":
        return mr.chatWithFallback([]anyi.Client{mr.smartClient, mr.fastClient}, messages)
    default:
        return mr.chatWithFallback([]anyi.Client{mr.fastClient}, messages)
    }
}

func (mr *ModelRouter) assessComplexity(messages []chat.Message) string {
    content := ""
    for _, msg := range messages {
        content += msg.Content + " "
    }

    content = strings.ToLower(content)

    // 复杂任务关键词
    complexKeywords := []string{
        "分析", "推理", "解释", "比较", "评估", "创作", "设计",
        "analyze", "reasoning", "explain", "compare", "evaluate",
    }

    // 简单任务关键词
    simpleKeywords := []string{
        "翻译", "总结", "列出", "什么是", "定义",
        "translate", "summarize", "list", "what is", "define",
    }

    for _, keyword := range complexKeywords {
        if strings.Contains(content, keyword) {
            return "complex"
        }
    }

    for _, keyword := range simpleKeywords {
        if strings.Contains(content, keyword) {
            return "simple"
        }
    }

    // 根据长度判断
    if len(content) > 500 {
        return "medium"
    }

    return "simple"
}

func (mr *ModelRouter) chatWithFallback(clients []anyi.Client, messages []chat.Message) (*chat.Message, error) {
    var lastErr error

    for _, client := range clients {
        response, _, err := client.Chat(messages, nil)
        if err == nil {
            return response, nil
        }
        lastErr = err
    }

    return nil, lastErr
}
```

### 5. Token 优化

#### 提示词优化

```go
package main

import (
    "strings"
    "github.com/jieliu2000/anyi"
)

type PromptOptimizer struct {
    client anyi.Client
}

func NewPromptOptimizer(client anyi.Client) *PromptOptimizer {
    return &PromptOptimizer{client: client}
}

func (po *PromptOptimizer) OptimizePrompt(prompt string) string {
    // 移除多余的空格和换行
    optimized := strings.TrimSpace(prompt)
    optimized = strings.Join(strings.Fields(optimized), " ")

    // 替换冗长的表达
    replacements := map[string]string{
        "请你帮我":     "请",
        "能不能":      "能否",
        "我想要你":     "请",
        "你可以":      "请",
        "非常详细地":    "详细",
        "尽可能详细地":   "详细",
    }

    for old, new := range replacements {
        optimized = strings.ReplaceAll(optimized, old, new)
    }

    return optimized
}

func (po *PromptOptimizer) Chat(prompt string) (*chat.Message, error) {
    optimizedPrompt := po.OptimizePrompt(prompt)

    messages := []chat.Message{
        {Role: "user", Content: optimizedPrompt},
    }

    response, _, err := po.client.Chat(messages, nil)
    return response, err
}
```

#### 上下文管理

```go
package main

import (
    "github.com/jieliu2000/anyi"
)

type ContextManager struct {
    client     anyi.Client
    maxTokens  int
    messages   []chat.Message
}

func NewContextManager(client anyi.Client, maxTokens int) *ContextManager {
    return &ContextManager{
        client:    client,
        maxTokens: maxTokens,
        messages:  make([]chat.Message, 0),
    }
}

func (cm *ContextManager) AddMessage(role, content string) {
    message := chat.Message{
        Role:    role,
        Content: content,
    }

    cm.messages = append(cm.messages, message)
    cm.trimContext()
}

func (cm *ContextManager) Chat(userInput string) (*chat.Message, error) {
    cm.AddMessage("user", userInput)

    response, _, err := cm.client.Chat(cm.messages, nil)
    if err != nil {
        return nil, err
    }

    cm.AddMessage("assistant", response.Content)
    return response, nil
}

func (cm *ContextManager) trimContext() {
    // 简化的 token 计算（实际应该使用 tokenizer）
    totalTokens := cm.estimateTokens()

    // 保留系统消息，删除最旧的对话
    for totalTokens > cm.maxTokens && len(cm.messages) > 1 {
        if cm.messages[0].Role == "system" && len(cm.messages) > 2 {
            // 删除第二条消息（保留系统消息）
            cm.messages = append(cm.messages[:1], cm.messages[2:]...)
        } else if cm.messages[0].Role != "system" {
            // 删除第一条消息
            cm.messages = cm.messages[1:]
        } else {
            break
        }
        totalTokens = cm.estimateTokens()
    }
}

func (cm *ContextManager) estimateTokens() int {
    total := 0
    for _, msg := range cm.messages {
        // 粗略估算：1 token ≈ 4 字符
        total += len(msg.Content) / 4
    }
    return total
}
```

### 6. 成本优化

#### 成本监控

```go
package main

import (
    "log"
    "sync"
    "time"
    "github.com/jieliu2000/anyi"
)

type CostTracker struct {
    mutex      sync.RWMutex
    dailyCost  float64
    monthlyCost float64
    lastReset  time.Time
    limits     CostLimits
}

type CostLimits struct {
    DailyLimit   float64
    MonthlyLimit float64
}

func NewCostTracker(limits CostLimits) *CostTracker {
    return &CostTracker{
        limits:    limits,
        lastReset: time.Now(),
    }
}

func (ct *CostTracker) RecordUsage(tokens int, model string) error {
    cost := ct.calculateCost(tokens, model)

    ct.mutex.Lock()
    defer ct.mutex.Unlock()

    // 检查是否需要重置日计数
    now := time.Now()
    if now.Day() != ct.lastReset.Day() {
        ct.dailyCost = 0
        ct.lastReset = now
    }

    // 检查是否需要重置月计数
    if now.Month() != ct.lastReset.Month() {
        ct.monthlyCost = 0
    }

    // 检查限额
    if ct.dailyCost+cost > ct.limits.DailyLimit {
        return fmt.Errorf("超出日使用限额: %.2f", ct.limits.DailyLimit)
    }

    if ct.monthlyCost+cost > ct.limits.MonthlyLimit {
        return fmt.Errorf("超出月使用限额: %.2f", ct.limits.MonthlyLimit)
    }

    ct.dailyCost += cost
    ct.monthlyCost += cost

    log.Printf("使用成本: $%.4f, 日累计: $%.2f, 月累计: $%.2f",
        cost, ct.dailyCost, ct.monthlyCost)

    return nil
}

func (ct *CostTracker) calculateCost(tokens int, model string) float64 {
    // 简化的成本计算（实际价格可能不同）
    pricePerToken := map[string]float64{
        "gpt-4":          0.00003,
        "gpt-3.5-turbo":  0.000002,
        "claude-3-opus":  0.000015,
        "claude-3-sonnet": 0.000003,
    }

    if price, exists := pricePerToken[model]; exists {
        return float64(tokens) * price
    }

    return 0 // 本地模型或未知模型
}

func (ct *CostTracker) GetStats() (float64, float64) {
    ct.mutex.RLock()
    defer ct.mutex.RUnlock()
    return ct.dailyCost, ct.monthlyCost
}
```

#### 智能降级

```go
package main

import (
    "github.com/jieliu2000/anyi"
)

type CostAwareClient struct {
    clients     map[string]anyi.Client
    costTracker *CostTracker
    fallbackOrder []string
}

func NewCostAwareClient(clients map[string]anyi.Client, tracker *CostTracker) *CostAwareClient {
    return &CostAwareClient{
        clients:     clients,
        costTracker: tracker,
        fallbackOrder: []string{"gpt-3.5-turbo", "claude-3-haiku", "ollama"},
    }
}

func (cac *CostAwareClient) Chat(messages []chat.Message, preferredModel string) (*chat.Message, error) {
    // 尝试首选模型
    if client, exists := cac.clients[preferredModel]; exists {
        if cac.canUseModel(preferredModel, messages) {
            response, info, err := client.Chat(messages, nil)
            if err == nil {
                cac.costTracker.RecordUsage(info.TotalTokens, preferredModel)
                return response, nil
            }
        }
    }

    // 降级到更便宜的模型
    for _, model := range cac.fallbackOrder {
        if client, exists := cac.clients[model]; exists {
            if cac.canUseModel(model, messages) {
                response, info, err := client.Chat(messages, nil)
                if err == nil {
                    cac.costTracker.RecordUsage(info.TotalTokens, model)
                    log.Printf("降级使用模型: %s", model)
                    return response, nil
                }
            }
        }
    }

    return nil, fmt.Errorf("所有模型都不可用或超出预算")
}

func (cac *CostAwareClient) canUseModel(model string, messages []chat.Message) bool {
    // 估算 token 使用量
    estimatedTokens := cac.estimateTokens(messages) * 2 // 包括响应

    // 检查是否会超出预算
    cost := cac.costTracker.calculateCost(estimatedTokens, model)
    dailyCost, _ := cac.costTracker.GetStats()

    return dailyCost+cost <= cac.costTracker.limits.DailyLimit
}

func (cac *CostAwareClient) estimateTokens(messages []chat.Message) int {
    total := 0
    for _, msg := range messages {
        total += len(msg.Content) / 4 // 粗略估算
    }
    return total
}
```

## 性能监控

### 基准测试

```go
package main

import (
    "testing"
    "time"
    "github.com/jieliu2000/anyi"
)

func BenchmarkChatPerformance(b *testing.B) {
    client, _ := anyi.GetClient("openai")
    messages := []chat.Message{
        {Role: "user", Content: "Hello"},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _, err := client.Chat(messages, nil)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCachedVsUncached(b *testing.B) {
    client, _ := anyi.GetClient("openai")
    cachedClient := NewCachedClient(client)

    messages := []chat.Message{
        {Role: "user", Content: "What is AI?"},
    }

    b.Run("Uncached", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _, _ = client.Chat(messages, nil)
        }
    })

    b.Run("Cached", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _, _ = cachedClient.Chat(messages, nil)
        }
    })
}
```

### 性能指标收集

```go
package main

import (
    "sync"
    "time"
    "github.com/jieliu2000/anyi"
)

type PerformanceMetrics struct {
    mutex           sync.RWMutex
    totalRequests   int64
    successRequests int64
    failedRequests  int64
    totalLatency    time.Duration
    minLatency      time.Duration
    maxLatency      time.Duration
}

func NewPerformanceMetrics() *PerformanceMetrics {
    return &PerformanceMetrics{
        minLatency: time.Hour, // 初始化为很大的值
    }
}

func (pm *PerformanceMetrics) RecordRequest(latency time.Duration, success bool) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    pm.totalRequests++
    pm.totalLatency += latency

    if success {
        pm.successRequests++
    } else {
        pm.failedRequests++
    }

    if latency < pm.minLatency {
        pm.minLatency = latency
    }

    if latency > pm.maxLatency {
        pm.maxLatency = latency
    }
}

func (pm *PerformanceMetrics) GetStats() map[string]interface{} {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    avgLatency := time.Duration(0)
    if pm.totalRequests > 0 {
        avgLatency = pm.totalLatency / time.Duration(pm.totalRequests)
    }

    successRate := float64(0)
    if pm.totalRequests > 0 {
        successRate = float64(pm.successRequests) / float64(pm.totalRequests) * 100
    }

    return map[string]interface{}{
        "total_requests":   pm.totalRequests,
        "success_requests": pm.successRequests,
        "failed_requests":  pm.failedRequests,
        "success_rate":     successRate,
        "avg_latency":      avgLatency,
        "min_latency":      pm.minLatency,
        "max_latency":      pm.maxLatency,
    }
}
```

## 最佳实践

### 1. 性能优化清单

- ✅ 使用连接池管理客户端
- ✅ 实施响应缓存
- ✅ 并行处理独立任务
- ✅ 智能选择模型
- ✅ 优化提示词长度
- ✅ 管理上下文窗口
- ✅ 监控成本和性能
- ✅ 实施降级策略

### 2. 成本控制建议

- 使用更便宜的模型处理简单任务
- 缓存重复查询的结果
- 优化提示词以减少 token 使用
- 设置使用限额和告警
- 定期审查和优化使用模式

### 3. 监控指标

- 响应时间（平均、最小、最大）
- 成功率
- 错误率和错误类型
- Token 使用量
- 成本统计
- 缓存命中率

## 下一步

现在您已经掌握了性能优化技巧，可以：

1. 学习 [Web 集成](web-integration.md) 来构建 Web 应用
2. 探索 [安全最佳实践](../advanced/security.md) 来保护您的应用
3. 查看 [部署指南](../advanced/deployment.md) 来部署到生产环境
4. 了解 [自定义执行器](../advanced/custom-executors.md) 来扩展功能

通过实施这些性能优化策略，您可以构建快速、经济、可扩展的 AI 应用程序！
