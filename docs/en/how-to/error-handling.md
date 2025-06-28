# Error Handling

This guide covers best practices for implementing robust error handling in your Anyi applications. You'll learn to handle various types of errors, implement retry strategies, and build resilient AI workflows.

## Table of Contents

- [Types of Errors](#types-of-errors)
- [Basic Error Handling](#basic-error-handling)
- [Retry Strategies](#retry-strategies)
- [Validation and Quality Control](#validation-and-quality-control)
- [Graceful Degradation](#graceful-degradation)
- [Monitoring and Alerting](#monitoring-and-alerting)
- [Best Practices](#best-practices)

## Types of Errors

Understanding different error types helps you implement appropriate handling strategies:

### 1. Network Errors

- Connection timeouts
- DNS resolution failures
- Network connectivity issues
- Service unavailability

### 2. API Errors

- Rate limiting (429 status)
- Authentication failures (401/403)
- Invalid requests (400 status)
- Server errors (500+ status)

### 3. Model Errors

- Token limit exceeded
- Content policy violations
- Model overload or unavailability
- Invalid model parameters

### 4. Validation Errors

- Output doesn't meet quality criteria
- Required fields missing
- Format validation failures
- Business rule violations

### 5. Application Errors

- Configuration errors
- Resource exhaustion
- Logic errors
- Data corruption

## Basic Error Handling

### Simple Error Handling Pattern

```go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Network-related errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "temporary failure") {
		return true
	}

	// Check for specific network error types
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	// System call errors
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			if syscallErr.Err == syscall.ECONNREFUSED {
				return true
			}
		}
	}

	// Rate limiting and server errors
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") {
		return true
	}

	return false
}

func chatWithRetry(client *anyi.Client, messages []chat.Message, maxRetries int) (*chat.Message, error) {
	var lastErr error
	backoff := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, _, err := client.Chat(messages, nil)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry if error is not retryable
		if !isRetryableError(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't wait after the last attempt
		if attempt < maxRetries {
			log.Printf("Attempt %d failed: %v. Retrying in %v...", attempt+1, err, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			if backoff > 30*time.Second {
				backoff = 30 * time.Second // Cap at 30 seconds
			}
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

func main() {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "Explain quantum computing"},
	}

	response, err := chatWithRetry(client, messages, 3)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Response: %s", response.Content)
}
```

### Error Wrapping and Context

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

type ChatError struct {
	Operation string
	Client    string
	Model     string
	Err       error
}

func (e *ChatError) Error() string {
	return fmt.Sprintf("chat operation '%s' failed for client '%s' (model: %s): %v",
		e.Operation, e.Client, e.Model, e.Err)
}

func (e *ChatError) Unwrap() error {
	return e.Err
}

func performChat(clientName, model string, messages []chat.Message) (*chat.Message, error) {
	client, err := anyi.GetClient(clientName)
	if err != nil {
		return nil, &ChatError{
			Operation: "get_client",
			Client:    clientName,
			Model:     model,
			Err:       err,
		}
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		return nil, &ChatError{
			Operation: "chat_request",
			Client:    clientName,
			Model:     model,
			Err:       err,
		}
	}

	return response, nil
}

func main() {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.Model = "gpt-4"

	_, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "What is machine learning?"},
	}

	response, err := performChat("openai", "gpt-4", messages)
	if err != nil {
		// Handle specific error types
		var chatErr *ChatError
		if fmt.Errorf("%w", err); chatErr != nil {
			log.Printf("Chat error details: Operation=%s, Client=%s, Model=%s",
				chatErr.Operation, chatErr.Client, chatErr.Model)
		}
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Response: %s", response.Content)
}
```

## Retry Strategies

### Exponential Backoff with Jitter

```go
package main

import (
	"math"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	Multiplier    float64
	Jitter        bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
	}
}

func calculateDelay(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Apply maximum delay cap
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // ±10% jitter
		delay += jitter - (0.05 * delay)
	}

	return time.Duration(delay)
}

func retryWithBackoff[T any](operation func() (T, error), config RetryConfig, isRetryable func(error) bool) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			return zero, fmt.Errorf("non-retryable error on attempt %d: %w", attempt+1, err)
		}

		// Don't wait after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := calculateDelay(attempt, config)
			log.Printf("Attempt %d failed: %v. Retrying in %v...", attempt+1, err, delay)
			time.Sleep(delay)
		}
	}

	return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// Usage example
func main() {
	config := DefaultRetryConfig()
	config.MaxAttempts = 5

	result, err := retryWithBackoff(func() (string, error) {
		// Your operation here
		return performSomeOperation()
	}, config, isRetryableError)

	if err != nil {
		log.Fatalf("Operation failed: %v", err)
	}

	log.Printf("Success: %s", result)
}
```

### Circuit Breaker Pattern

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	mu            sync.RWMutex
	state         CircuitState
	failureCount  int
	lastFailTime  time.Time
	successCount  int

	maxFailures   int
	timeout       time.Duration
	maxRequests   int
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       Closed,
		maxFailures: maxFailures,
		timeout:     timeout,
		maxRequests: 10,
	}
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Open:
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = HalfOpen
			cb.successCount = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	case HalfOpen:
		if cb.successCount >= cb.maxRequests {
			cb.state = Closed
			cb.failureCount = 0
		}
	}

	err := operation()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	return err
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = Open
	}
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == HalfOpen {
		cb.successCount++
	}

	if cb.state == Closed {
		cb.failureCount = 0
	}
}

// Usage with Anyi
func chatWithCircuitBreaker(client *anyi.Client, messages []chat.Message, cb *CircuitBreaker) (*chat.Message, error) {
	var response *chat.Message
	var chatErr error

	err := cb.Execute(func() error {
		var err error
		response, _, err = client.Chat(messages, nil)
		chatErr = err
		return err
	})

	if err != nil {
		return nil, err
	}

	return response, chatErr
}
```

## Validation and Quality Control

### Output Validation in Workflows

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

// Custom validator for email extraction
type EmailValidator struct {
	RequiredCount int
}

func (v *EmailValidator) Validate(output string) error {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(output, -1)

	if len(emails) < v.RequiredCount {
		return fmt.Errorf("expected at least %d emails, found %d", v.RequiredCount, len(emails))
	}

	return nil
}

// Custom validator for JSON structure
type StructuredDataValidator struct {
	RequiredFields []string
	Schema         string
}

func (v *StructuredDataValidator) Validate(output string) error {
	var data map[string]interface{}

	// Check if output is valid JSON
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return fmt.Errorf("output is not valid JSON: %w", err)
	}

	// Check required fields
	for _, field := range v.RequiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// Additional schema validation could be added here
	return nil
}

func createValidatedWorkflow() *flow.Flow {
	client, _ := anyi.GetClient("openai")

	// Step 1: Extract contact information with validation
	extractStep := &flow.Step{
		Name: "extract_contacts",
		Executor: &anyi.LLMExecutor{
			Template: `Extract contact information from the following text and format as JSON:

{{.Text}}

Required format:
{
  "contacts": [
    {
      "name": "Full Name",
      "email": "email@example.com",
      "phone": "phone number",
      "company": "Company Name"
    }
  ]
}`,
			SystemMessage: "You are a data extraction specialist. Always return valid JSON.",
		},
		Validator: &StructuredDataValidator{
			RequiredFields: []string{"contacts"},
		},
		MaxRetryTimes: 3,
	}

	// Step 2: Validate email addresses
	validateStep := &flow.Step{
		Name: "validate_emails",
		Executor: &anyi.LLMExecutor{
			Template: `Review the extracted contact information and verify email addresses:

{{.Text}}

Ensure all email addresses are properly formatted and flag any suspicious ones.`,
			SystemMessage: "You are an email validation expert.",
		},
		Validator: &EmailValidator{
			RequiredCount: 1,
		},
		MaxRetryTimes: 2,
	}

	workflow, _ := anyi.NewFlow("contact_extraction", client, *extractStep, *validateStep)
	return workflow
}

func main() {
	// Configure client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	_, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	workflow := createValidatedWorkflow()

	// Test with sample data
	sampleText := `
	Contact John Doe at john.doe@company.com or call (555) 123-4567.
	You can also reach Jane Smith at jane@example.org for technical support.
	`

	result, err := workflow.RunWithInput(sampleText)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	log.Printf("Validated result: %s", result.Text)
}
```

## Graceful Degradation

### Fallback Strategies

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/chat"
)

type ChatService struct {
	primaryClient   *anyi.Client
	fallbackClient  *anyi.Client
	localClient     *anyi.Client
}

func NewChatService() (*ChatService, error) {
	// Primary: OpenAI GPT-4
	primaryConfig := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4")
	primaryClient, err := anyi.NewClient("primary", primaryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary client: %w", err)
	}

	// Fallback: OpenAI GPT-3.5
	fallbackConfig := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-3.5-turbo")
	fallbackClient, err := anyi.NewClient("fallback", fallbackConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create fallback client: %w", err)
	}

	// Local: Ollama (for when all cloud services fail)
	localConfig := ollama.DefaultConfig("llama3")
	localClient, err := anyi.NewClient("local", localConfig)
	if err != nil {
		log.Printf("Warning: Failed to create local client: %v", err)
		// Local client is optional
	}

	return &ChatService{
		primaryClient:  primaryClient,
		fallbackClient: fallbackClient,
		localClient:    localClient,
	}, nil
}

func (cs *ChatService) Chat(messages []chat.Message) (*chat.Message, error) {
	// Try primary client first
	response, _, err := cs.primaryClient.Chat(messages, nil)
	if err == nil {
		log.Println("Used primary client (GPT-4)")
		return response, nil
	}

	log.Printf("Primary client failed: %v", err)

	// Try fallback client
	response, _, err = cs.fallbackClient.Chat(messages, nil)
	if err == nil {
		log.Println("Used fallback client (GPT-3.5)")
		return response, nil
	}

	log.Printf("Fallback client failed: %v", err)

	// Try local client as last resort
	if cs.localClient != nil {
		response, _, err = cs.localClient.Chat(messages, nil)
		if err == nil {
			log.Println("Used local client (Ollama)")
			return response, nil
		}
		log.Printf("Local client failed: %v", err)
	}

	return nil, fmt.Errorf("all chat services failed")
}
```

### Quality-Based Fallbacks

```go
package main

import (
	"log"
	"strings"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

func assessResponseQuality(response string) int {
	score := 0

	// Length check
	if len(response) > 50 {
		score += 20
	}

	// Structure check
	if strings.Contains(response, ".") && strings.Contains(response, " ") {
		score += 20
	}

	// Content quality indicators
	qualityIndicators := []string{
		"because", "therefore", "however", "specifically",
		"example", "such as", "including", "particularly",
	}

	for _, indicator := range qualityIndicators {
		if strings.Contains(strings.ToLower(response), indicator) {
			score += 10
		}
	}

	// Avoid low-quality responses
	lowQualityIndicators := []string{
		"i don't know", "i'm not sure", "i cannot",
		"sorry", "unable to", "not available",
	}

	for _, indicator := range lowQualityIndicators {
		if strings.Contains(strings.ToLower(response), indicator) {
			score -= 30
		}
	}

	return score
}

func chatWithQualityFallback(messages []chat.Message) (*chat.Message, error) {
	clients := []string{"premium", "standard", "basic"}
	minQualityScore := 50

	for _, clientName := range clients {
		client, err := anyi.GetClient(clientName)
		if err != nil {
			log.Printf("Failed to get client %s: %v", clientName, err)
			continue
		}

		response, _, err := client.Chat(messages, nil)
		if err != nil {
			log.Printf("Client %s failed: %v", clientName, err)
			continue
		}

		quality := assessResponseQuality(response.Content)
		log.Printf("Client %s quality score: %d", clientName, quality)

		if quality >= minQualityScore {
			return response, nil
		}

		log.Printf("Response quality too low (%d), trying next client", quality)
	}

	return nil, fmt.Errorf("no client provided acceptable quality response")
}
```

## Monitoring and Alerting

### Error Metrics Collection

```go
package main

import (
	"log"
	"sync"
	"time"
)

type ErrorMetrics struct {
	mu              sync.RWMutex
	totalRequests   int64
	totalErrors     int64
	errorsByType    map[string]int64
	errorsByClient  map[string]int64
	lastError       time.Time
	errorRate       float64
}

func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		errorsByType:   make(map[string]int64),
		errorsByClient: make(map[string]int64),
	}
}

func (em *ErrorMetrics) RecordRequest(clientName string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.totalRequests++
}

func (em *ErrorMetrics) RecordError(clientName, errorType string, err error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.totalErrors++
	em.errorsByType[errorType]++
	em.errorsByClient[clientName]++
	em.lastError = time.Now()

	if em.totalRequests > 0 {
		em.errorRate = float64(em.totalErrors) / float64(em.totalRequests)
	}

	log.Printf("Error recorded - Client: %s, Type: %s, Error: %v", clientName, errorType, err)

	// Trigger alerts if error rate is too high
	if em.errorRate > 0.1 { // 10% error rate threshold
		em.triggerAlert()
	}
}

func (em *ErrorMetrics) triggerAlert() {
	log.Printf("ALERT: High error rate detected: %.2f%%", em.errorRate*100)
	// Send to monitoring system, email, Slack, etc.
}

func (em *ErrorMetrics) GetStats() map[string]interface{} {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return map[string]interface{}{
		"total_requests":   em.totalRequests,
		"total_errors":     em.totalErrors,
		"error_rate":       em.errorRate,
		"errors_by_type":   em.errorsByType,
		"errors_by_client": em.errorsByClient,
		"last_error":       em.lastError,
	}
}

// Global metrics instance
var globalMetrics = NewErrorMetrics()

// Wrapper function with metrics
func chatWithMetrics(clientName string, messages []chat.Message) (*chat.Message, error) {
	globalMetrics.RecordRequest(clientName)

	client, err := anyi.GetClient(clientName)
	if err != nil {
		globalMetrics.RecordError(clientName, "client_error", err)
		return nil, err
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		// Categorize error types
		errorType := categorizeError(err)
		globalMetrics.RecordError(clientName, errorType, err)
		return nil, err
	}

	return response, nil
}

func categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "timeout") {
		return "timeout"
	}
	if strings.Contains(errStr, "rate limit") {
		return "rate_limit"
	}
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") {
		return "auth_error"
	}
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return "server_error"
	}

	return "unknown"
}
```

## Best Practices

### 1. Error Classification

Classify errors by severity and handling strategy:

```go
type ErrorSeverity int

const (
	Low ErrorSeverity = iota
	Medium
	High
	Critical
)

type ErrorInfo struct {
	Severity    ErrorSeverity
	Retryable   bool
	UserFacing  bool
	Message     string
}

func classifyError(err error) ErrorInfo {
	errStr := strings.ToLower(err.Error())

	// Critical errors - immediate attention required
	if strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "unauthorized") {
		return ErrorInfo{
			Severity:   Critical,
			Retryable:  false,
			UserFacing: false,
			Message:    "Authentication failure - check API keys",
		}
	}

	// High severity - affects functionality
	if strings.Contains(errStr, "quota exceeded") {
		return ErrorInfo{
			Severity:   High,
			Retryable:  false,
			UserFacing: true,
			Message:    "Service temporarily unavailable",
		}
	}

	// Medium severity - temporary issues
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "rate limit") {
		return ErrorInfo{
			Severity:   Medium,
			Retryable:  true,
			UserFacing: false,
			Message:    "Temporary service delay",
		}
	}

	// Low severity - minor issues
	return ErrorInfo{
		Severity:   Low,
		Retryable:  true,
		UserFacing: false,
		Message:    "Minor service interruption",
	}
}
```

### 2. Structured Logging

```go
package main

import (
	"encoding/json"
	"log"
	"time"
)

type LogEntry struct {
	Timestamp   time.Time   `json:"timestamp"`
	Level       string      `json:"level"`
	Component   string      `json:"component"`
	Operation   string      `json:"operation"`
	ClientName  string      `json:"client_name,omitempty"`
	Model       string      `json:"model,omitempty"`
	Error       string      `json:"error,omitempty"`
	Duration    int64       `json:"duration_ms,omitempty"`
	TokensUsed  int         `json:"tokens_used,omitempty"`
	Success     bool        `json:"success"`
}

func logError(component, operation, clientName, model string, err error, duration time.Duration) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Component: component,
		Operation: operation,
		ClientName: clientName,
		Model:     model,
		Error:     err.Error(),
		Duration:  duration.Milliseconds(),
		Success:   false,
	}

	jsonLog, _ := json.Marshal(entry)
	log.Println(string(jsonLog))
}

func logSuccess(component, operation, clientName, model string, duration time.Duration, tokensUsed int) {
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      "INFO",
		Component:  component,
		Operation:  operation,
		ClientName: clientName,
		Model:      model,
		Duration:   duration.Milliseconds(),
		TokensUsed: tokensUsed,
		Success:    true,
	}

	jsonLog, _ := json.Marshal(entry)
	log.Println(string(jsonLog))
}
```

### 3. Health Checks

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

type HealthStatus struct {
	Healthy     bool      `json:"healthy"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
	ResponseTime int64    `json:"response_time_ms"`
}

func checkClientHealth(clientName string) HealthStatus {
	start := time.Now()

	client, err := anyi.GetClient(clientName)
	if err != nil {
		return HealthStatus{
			Healthy:     false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}
	}

	// Simple health check message
	messages := []chat.Message{
		{Role: "user", Content: "Hello"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Note: This assumes your client supports context (implementation dependent)
	_, _, err = client.Chat(messages, nil)

	duration := time.Since(start)

	if err != nil {
		return HealthStatus{
			Healthy:      false,
			LastChecked:  time.Now(),
			Error:        err.Error(),
			ResponseTime: duration.Milliseconds(),
		}
	}

	return HealthStatus{
		Healthy:      true,
		LastChecked:  time.Now(),
		ResponseTime: duration.Milliseconds(),
	}
}

func periodicHealthCheck(clientNames []string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, clientName := range clientNames {
				status := checkClientHealth(clientName)
				if !status.Healthy {
					log.Printf("Health check failed for %s: %s", clientName, status.Error)
				} else {
					log.Printf("Health check passed for %s (response time: %dms)",
						clientName, status.ResponseTime)
				}
			}
		}
	}
}
```

By implementing these error handling patterns and best practices, you can build robust Anyi applications that gracefully handle failures and provide reliable service to your users.
