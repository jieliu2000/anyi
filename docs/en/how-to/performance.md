# Performance Optimization

This guide covers strategies for optimizing the performance of your Anyi applications, including reducing latency, managing costs, and scaling efficiently.

## Table of Contents

- [Understanding Performance Metrics](#understanding-performance-metrics)
- [Latency Optimization](#latency-optimization)
- [Cost Optimization](#cost-optimization)
- [Throughput Optimization](#throughput-optimization)
- [Caching Strategies](#caching-strategies)
- [Model Selection](#model-selection)
- [Monitoring and Profiling](#monitoring-and-profiling)

## Understanding Performance Metrics

Key performance indicators for AI applications:

### 1. Latency Metrics

- **Response Time**: Time from request to response
- **Time to First Token**: Time until first token is generated
- **Tokens per Second**: Generation speed for streaming responses
- **Network Latency**: Time spent in network communication

### 2. Throughput Metrics

- **Requests per Second (RPS)**: Number of requests handled per second
- **Concurrent Users**: Maximum simultaneous users supported
- **Queue Length**: Number of pending requests

### 3. Cost Metrics

- **Cost per Request**: Average cost per API call
- **Token Efficiency**: Output quality per token used
- **Resource Utilization**: CPU, memory, and network usage

### 4. Quality Metrics

- **Success Rate**: Percentage of successful requests
- **Retry Rate**: Frequency of request retries
- **Validation Pass Rate**: Percentage of outputs meeting quality criteria

## Latency Optimization

### 1. Connection Pooling and Keep-Alive

```go
package main

import (
	"net/http"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
)

func createOptimizedClient() (*anyi.Client, error) {
	// Create custom HTTP client with optimized settings
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			// Enable HTTP/2
			ForceAttemptHTTP2: true,
		},
	}

	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.HTTPClient = httpClient

	return anyi.NewClient("optimized", config)
}
```

### 2. Request Batching

```go
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type BatchRequest struct {
	ID       string
	Messages []chat.Message
	Response chan BatchResponse
}

type BatchResponse struct {
	ID       string
	Response *chat.Message
	Error    error
}

type BatchProcessor struct {
	client      *anyi.Client
	batchSize   int
	flushTime   time.Duration
	requests    chan BatchRequest
	batch       []BatchRequest
	mu          sync.Mutex
}

func NewBatchProcessor(client *anyi.Client, batchSize int, flushTime time.Duration) *BatchProcessor {
	bp := &BatchProcessor{
		client:    client,
		batchSize: batchSize,
		flushTime: flushTime,
		requests:  make(chan BatchRequest, 100),
		batch:     make([]BatchRequest, 0, batchSize),
	}

	go bp.processBatches()
	return bp
}

func (bp *BatchProcessor) processBatches() {
	ticker := time.NewTicker(bp.flushTime)
	defer ticker.Stop()

	for {
		select {
		case req := <-bp.requests:
			bp.mu.Lock()
			bp.batch = append(bp.batch, req)
			shouldFlush := len(bp.batch) >= bp.batchSize
			bp.mu.Unlock()

			if shouldFlush {
				bp.flushBatch()
			}

		case <-ticker.C:
			bp.flushBatch()
		}
	}
}

func (bp *BatchProcessor) flushBatch() {
	bp.mu.Lock()
	if len(bp.batch) == 0 {
		bp.mu.Unlock()
		return
	}

	currentBatch := make([]BatchRequest, len(bp.batch))
	copy(currentBatch, bp.batch)
	bp.batch = bp.batch[:0] // Clear batch
	bp.mu.Unlock()

	// Process batch concurrently
	var wg sync.WaitGroup
	for _, req := range currentBatch {
		wg.Add(1)
		go func(r BatchRequest) {
			defer wg.Done()
			response, _, err := bp.client.Chat(r.Messages, nil)
			r.Response <- BatchResponse{
				ID:       r.ID,
				Response: response,
				Error:    err,
			}
		}(req)
	}
	wg.Wait()
}

func (bp *BatchProcessor) Submit(id string, messages []chat.Message) <-chan BatchResponse {
	response := make(chan BatchResponse, 1)
	bp.requests <- BatchRequest{
		ID:       id,
		Messages: messages,
		Response: response,
	}
	return response
}
```

### 3. Streaming Responses

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

// Note: This is a conceptual example. Actual streaming implementation
// depends on the specific provider's streaming API support.

type StreamingResponse struct {
	Token string
	Done  bool
	Error error
}

func streamChat(client *anyi.Client, messages []chat.Message) <-chan StreamingResponse {
	responseChan := make(chan StreamingResponse)

	go func() {
		defer close(responseChan)

		// This would use the provider's streaming API
		// For demonstration, we'll simulate streaming
		response, _, err := client.Chat(messages, nil)
		if err != nil {
			responseChan <- StreamingResponse{Error: err}
			return
		}

		// Simulate token-by-token streaming
		words := strings.Fields(response.Content)
		for i, word := range words {
			responseChan <- StreamingResponse{
				Token: word + " ",
				Done:  i == len(words)-1,
			}
			time.Sleep(50 * time.Millisecond) // Simulate streaming delay
		}
	}()

	return responseChan
}

func handleStreamingChat(client *anyi.Client, messages []chat.Message) {
	stream := streamChat(client, messages)
	var fullResponse strings.Builder

	for chunk := range stream {
		if chunk.Error != nil {
			log.Printf("Streaming error: %v", chunk.Error)
			return
		}

		fmt.Print(chunk.Token) // Display token immediately
		fullResponse.WriteString(chunk.Token)

		if chunk.Done {
			break
		}
	}

	log.Printf("\nComplete response: %s", fullResponse.String())
}
```

## Cost Optimization

### 1. Token Management

```go
package main

import (
	"log"
	"strings"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type TokenOptimizer struct {
	client        *anyi.Client
	maxTokens     int
	targetLength  int
}

func NewTokenOptimizer(client *anyi.Client, maxTokens, targetLength int) *TokenOptimizer {
	return &TokenOptimizer{
		client:       client,
		maxTokens:    maxTokens,
		targetLength: targetLength,
	}
}

func (to *TokenOptimizer) estimateTokens(text string) int {
	// Rough estimation: 1 token ≈ 4 characters for English
	return len(text) / 4
}

func (to *TokenOptimizer) optimizePrompt(prompt string) string {
	estimatedTokens := to.estimateTokens(prompt)

	if estimatedTokens <= to.maxTokens {
		return prompt
	}

	// Truncate prompt to fit within token limit
	targetChars := to.maxTokens * 4
	if len(prompt) > targetChars {
		// Try to truncate at sentence boundaries
		sentences := strings.Split(prompt, ".")
		var optimized strings.Builder

		for _, sentence := range sentences {
			if optimized.Len()+len(sentence) > targetChars {
				break
			}
			optimized.WriteString(sentence + ".")
		}

		return optimized.String()
	}

	return prompt
}

func (to *TokenOptimizer) chatWithOptimization(messages []chat.Message) (*chat.Message, error) {
	// Optimize the last user message (typically the longest)
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			messages[i].Content = to.optimizePrompt(messages[i].Content)
			break
		}
	}

	// Set maximum tokens for response
	options := &chat.ChatOptions{
		// MaxTokens: to.targetLength, // If supported by the provider
	}

	return to.client.Chat(messages, options)
}
```

### 2. Model Selection Strategy

```go
package main

import (
	"log"
	"strings"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type ModelSelector struct {
	clients map[string]*anyi.Client
	costs   map[string]float64 // Cost per 1K tokens
}

func NewModelSelector() *ModelSelector {
	return &ModelSelector{
		clients: make(map[string]*anyi.Client),
		costs: map[string]float64{
			"gpt-4o":        0.025, // $0.025 per 1K tokens
			"gpt-4o-mini":   0.002, // $0.002 per 1K tokens
			"claude-3":      0.015, // Example pricing
			"local":         0.0,   // Local model (free)
		},
	}
}

func (ms *ModelSelector) AddClient(name string, client *anyi.Client) {
	ms.clients[name] = client
}

func (ms *ModelSelector) selectOptimalModel(prompt string, qualityRequirement string) string {
	promptLength := len(prompt)

	// For simple queries, use cheaper models
	if promptLength < 100 && qualityRequirement == "low" {
		return "gpt-4o-mini"
	}

	// For complex analysis, use more powerful models
	if strings.Contains(strings.ToLower(prompt), "analyze") ||
		strings.Contains(strings.ToLower(prompt), "complex") ||
		qualityRequirement == "high" {
		return "gpt-4o"
	}

	// For code generation, prefer specialized models
	if strings.Contains(strings.ToLower(prompt), "code") ||
		strings.Contains(strings.ToLower(prompt), "programming") {
		return "claude-3"
	}

	// Default to balanced option
	return "gpt-4o-mini"
}

func (ms *ModelSelector) chatWithOptimalModel(messages []chat.Message, qualityRequirement string) (*chat.Message, error) {
	// Use the last user message to determine complexity
	var prompt string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			prompt = messages[i].Content
			break
		}
	}

	modelName := ms.selectOptimalModel(prompt, qualityRequirement)
	client, exists := ms.clients[modelName]
	if !exists {
		// Fallback to default
		client = ms.clients["gpt-4o-mini"]
		modelName = "gpt-4o-mini"
	}

	log.Printf("Selected model: %s (cost: $%.4f per 1K tokens)", modelName, ms.costs[modelName])

	return client.Chat(messages, nil)
}
```

## Throughput Optimization

### 1. Connection Pooling

```go
package main

import (
	"sync"
	"time"

	"github.com/jieliu2000/anyi"
)

type ClientPool struct {
	clients chan *anyi.Client
	factory func() (*anyi.Client, error)
	mu      sync.Mutex
	created int
	maxSize int
}

func NewClientPool(maxSize int, factory func() (*anyi.Client, error)) *ClientPool {
	return &ClientPool{
		clients: make(chan *anyi.Client, maxSize),
		factory: factory,
		maxSize: maxSize,
	}
}

func (cp *ClientPool) Get() (*anyi.Client, error) {
	select {
	case client := <-cp.clients:
		return client, nil
	default:
		cp.mu.Lock()
		defer cp.mu.Unlock()

		if cp.created < cp.maxSize {
			client, err := cp.factory()
			if err != nil {
				return nil, err
			}
			cp.created++
			return client, nil
		}

		// Wait for a client to become available
		select {
		case client := <-cp.clients:
			return client, nil
		case <-time.After(5 * time.Second):
			return nil, fmt.Errorf("timeout waiting for client")
		}
	}
}

func (cp *ClientPool) Put(client *anyi.Client) {
	select {
	case cp.clients <- client:
	default:
		// Pool is full, discard the client
	}
}

// Usage example
func createClientFactory() func() (*anyi.Client, error) {
	return func() (*anyi.Client, error) {
		config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
		return anyi.NewClient("pooled", config)
	}
}
```

### 2. Concurrent Processing

```go
package main

import (
	"context"
	"sync"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type ConcurrentProcessor struct {
	pool        *ClientPool
	workerCount int
	semaphore   chan struct{}
}

func NewConcurrentProcessor(pool *ClientPool, workerCount int) *ConcurrentProcessor {
	return &ConcurrentProcessor{
		pool:        pool,
		workerCount: workerCount,
		semaphore:   make(chan struct{}, workerCount),
	}
}

type ProcessingTask struct {
	ID       string
	Messages []chat.Message
	Result   chan ProcessingResult
}

type ProcessingResult struct {
	ID       string
	Response *chat.Message
	Error    error
	Duration time.Duration
}

func (cp *ConcurrentProcessor) ProcessConcurrent(ctx context.Context, tasks []ProcessingTask) {
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(t ProcessingTask) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case cp.semaphore <- struct{}{}:
				defer func() { <-cp.semaphore }()
			case <-ctx.Done():
				t.Result <- ProcessingResult{
					ID:    t.ID,
					Error: ctx.Err(),
				}
				return
			}

			start := time.Now()

			// Get client from pool
			client, err := cp.pool.Get()
			if err != nil {
				t.Result <- ProcessingResult{
					ID:    t.ID,
					Error: err,
				}
				return
			}
			defer cp.pool.Put(client)

			// Process the task
			response, _, err := client.Chat(t.Messages, nil)
			duration := time.Since(start)

			t.Result <- ProcessingResult{
				ID:       t.ID,
				Response: response,
				Error:    err,
				Duration: duration,
			}
		}(task)
	}

	wg.Wait()
}
```

## Caching Strategies

### 1. Response Caching

```go
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type CacheEntry struct {
	Response  *chat.Message
	Timestamp time.Time
	TTL       time.Duration
}

func (ce *CacheEntry) IsExpired() bool {
	return time.Since(ce.Timestamp) > ce.TTL
}

type ResponseCache struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	cache := &ResponseCache{
		cache: make(map[string]*CacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (rc *ResponseCache) generateKey(messages []chat.Message) string {
	data, _ := json.Marshal(messages)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func (rc *ResponseCache) Get(messages []chat.Message) (*chat.Message, bool) {
	key := rc.generateKey(messages)

	rc.mu.RLock()
	entry, exists := rc.cache[key]
	rc.mu.RUnlock()

	if !exists || entry.IsExpired() {
		return nil, false
	}

	return entry.Response, true
}

func (rc *ResponseCache) Set(messages []chat.Message, response *chat.Message) {
	key := rc.generateKey(messages)

	rc.mu.Lock()
	rc.cache[key] = &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
		TTL:       rc.ttl,
	}
	rc.mu.Unlock()
}

func (rc *ResponseCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rc.mu.Lock()
		for key, entry := range rc.cache {
			if entry.IsExpired() {
				delete(rc.cache, key)
			}
		}
		rc.mu.Unlock()
	}
}

type CachedClient struct {
	client *anyi.Client
	cache  *ResponseCache
}

func NewCachedClient(client *anyi.Client, cacheTTL time.Duration) *CachedClient {
	return &CachedClient{
		client: client,
		cache:  NewResponseCache(cacheTTL),
	}
}

func (cc *CachedClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, *chat.ResponseInfo, error) {
	// Check cache first
	if cached, found := cc.cache.Get(messages); found {
		log.Println("Cache hit")
		return cached, nil, nil
	}

	// Cache miss, call actual client
	response, info, err := cc.client.Chat(messages, options)
	if err != nil {
		return nil, info, err
	}

	// Cache the response
	cc.cache.Set(messages, response)
	log.Println("Cache miss, response cached")

	return response, info, nil
}
```

## Model Selection

### 1. Performance-Based Model Selection

```go
package main

import (
	"sync"
	"time"

	"github.com/jieliu2000/anyi"
)

type ModelMetrics struct {
	AverageLatency time.Duration
	SuccessRate    float64
	TotalRequests  int64
	Errors         int64
	LastUsed       time.Time
}

type PerformanceTracker struct {
	metrics map[string]*ModelMetrics
	mu      sync.RWMutex
}

func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics: make(map[string]*ModelMetrics),
	}
}

func (pt *PerformanceTracker) RecordRequest(modelName string, latency time.Duration, success bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.metrics[modelName] == nil {
		pt.metrics[modelName] = &ModelMetrics{}
	}

	m := pt.metrics[modelName]
	m.TotalRequests++
	m.LastUsed = time.Now()

	if success {
		// Update average latency using exponential moving average
		if m.AverageLatency == 0 {
			m.AverageLatency = latency
		} else {
			m.AverageLatency = time.Duration(float64(m.AverageLatency)*0.9 + float64(latency)*0.1)
		}
	} else {
		m.Errors++
	}

	m.SuccessRate = float64(m.TotalRequests-m.Errors) / float64(m.TotalRequests)
}

func (pt *PerformanceTracker) GetBestModel(models []string) string {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var bestModel string
	var bestScore float64

	for _, model := range models {
		metrics := pt.metrics[model]
		if metrics == nil {
			continue
		}

		// Calculate composite score (higher is better)
		latencyScore := 1.0 / (float64(metrics.AverageLatency.Milliseconds()) + 1)
		successScore := metrics.SuccessRate

		// Weight recent usage
		recencyScore := 1.0
		if time.Since(metrics.LastUsed) > time.Hour {
			recencyScore = 0.5
		}

		score := (latencyScore * 0.4) + (successScore * 0.5) + (recencyScore * 0.1)

		if score > bestScore {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}
```

## Monitoring and Profiling

### 1. Performance Monitoring

```go
package main

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type PerformanceMonitor struct {
	metrics     map[string]*OperationMetrics
	mu          sync.RWMutex
	startTime   time.Time
}

type OperationMetrics struct {
	Count           int64
	TotalDuration   time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	AverageDuration time.Duration
	Errors          int64
	LastUpdate      time.Time
}

func NewPerformanceMonitor() *PerformanceMonitor {
	pm := &PerformanceMonitor{
		metrics:   make(map[string]*OperationMetrics),
		startTime: time.Now(),
	}

	// Start monitoring goroutine
	go pm.periodicReport()

	return pm
}

func (pm *PerformanceMonitor) RecordOperation(operation string, duration time.Duration, success bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.metrics[operation] == nil {
		pm.metrics[operation] = &OperationMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
	}

	m := pm.metrics[operation]
	m.Count++
	m.TotalDuration += duration
	m.AverageDuration = time.Duration(int64(m.TotalDuration) / m.Count)
	m.LastUpdate = time.Now()

	if duration < m.MinDuration {
		m.MinDuration = duration
	}
	if duration > m.MaxDuration {
		m.MaxDuration = duration
	}

	if !success {
		m.Errors++
	}
}

func (pm *PerformanceMonitor) periodicReport() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pm.generateReport()
	}
}

func (pm *PerformanceMonitor) generateReport() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	log.Printf("=== Performance Report ===")
	log.Printf("Uptime: %v", time.Since(pm.startTime))
	log.Printf("Memory Usage: %d MB", memStats.Alloc/1024/1024)
	log.Printf("Goroutines: %d", runtime.NumGoroutine())

	for operation, metrics := range pm.metrics {
		errorRate := float64(metrics.Errors) / float64(metrics.Count) * 100
		log.Printf("Operation: %s", operation)
		log.Printf("  Count: %d", metrics.Count)
		log.Printf("  Average Duration: %v", metrics.AverageDuration)
		log.Printf("  Min/Max Duration: %v / %v", metrics.MinDuration, metrics.MaxDuration)
		log.Printf("  Error Rate: %.2f%%", errorRate)
	}
}

// Instrumented client wrapper
type InstrumentedClient struct {
	client  *anyi.Client
	monitor *PerformanceMonitor
	name    string
}

func NewInstrumentedClient(client *anyi.Client, monitor *PerformanceMonitor, name string) *InstrumentedClient {
	return &InstrumentedClient{
		client:  client,
		monitor: monitor,
		name:    name,
	}
}

func (ic *InstrumentedClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, *chat.ResponseInfo, error) {
	start := time.Now()
	response, info, err := ic.client.Chat(messages, options)
	duration := time.Since(start)

	operation := fmt.Sprintf("%s_chat", ic.name)
	ic.monitor.RecordOperation(operation, duration, err == nil)

	if info != nil {
		// Record token usage metrics
		tokenOperation := fmt.Sprintf("%s_tokens", ic.name)
		ic.monitor.RecordOperation(tokenOperation, time.Duration(info.TotalTokens)*time.Millisecond, true)
	}

	return response, info, err
}
```

### 2. Resource Usage Monitoring

```go
package main

import (
	"runtime"
	"time"
)

type ResourceMonitor struct {
	maxMemory     uint64
	maxGoroutines int
	alerts        chan string
}

func NewResourceMonitor() *ResourceMonitor {
	rm := &ResourceMonitor{
		alerts: make(chan string, 10),
	}

	go rm.monitor()
	go rm.handleAlerts()

	return rm
}

func (rm *ResourceMonitor) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		currentMemory := memStats.Alloc
		currentGoroutines := runtime.NumGoroutine()

		// Update maximums
		if currentMemory > rm.maxMemory {
			rm.maxMemory = currentMemory
		}
		if currentGoroutines > rm.maxGoroutines {
			rm.maxGoroutines = currentGoroutines
		}

		// Check thresholds
		if currentMemory > 500*1024*1024 { // 500MB
			rm.alerts <- fmt.Sprintf("High memory usage: %d MB", currentMemory/1024/1024)
		}

		if currentGoroutines > 1000 {
			rm.alerts <- fmt.Sprintf("High goroutine count: %d", currentGoroutines)
		}
	}
}

func (rm *ResourceMonitor) handleAlerts() {
	for alert := range rm.alerts {
		log.Printf("RESOURCE ALERT: %s", alert)
		// Send to monitoring system, email, etc.
	}
}
```

By implementing these performance optimization strategies, you can significantly improve the speed, efficiency, and cost-effectiveness of your Anyi applications while maintaining high quality outputs.
