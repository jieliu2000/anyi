# Custom Executors

This guide explains how to create custom executors in Anyi to extend the framework's functionality beyond the built-in components.

## Overview

Custom executors allow you to:

- Integrate with external APIs and services
- Implement domain-specific business logic
- Create reusable components for your organization
- Handle specialized data processing tasks
- Bridge Anyi with existing systems

## Executor Interface

All executors must implement the `Executor` interface:

```go
type Executor interface {
    Execute(context *FlowContext) (*FlowContext, error)
}
```

### FlowContext Structure

The `FlowContext` contains all data flowing through the workflow:

```go
type FlowContext struct {
    Text   string      // Current text content
    Memory interface{} // Structured data
    Think  string      // Thinking process
    Images []string    // Image URLs
}
```

## Creating a Simple Custom Executor

### Example: Math Calculator Executor

```go
package main

import (
    "fmt"
    "strconv"
    "strings"
    "regexp"

    "github.com/jieliu2000/anyi"
)

type MathCalculatorExecutor struct {
    Precision int // Number of decimal places
}

func (e *MathCalculatorExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Extract mathematical expressions from the text
    expression := strings.TrimSpace(context.Text)

    // Simple calculator logic (you could use a proper expression parser)
    result, err := e.evaluateExpression(expression)
    if err != nil {
        return nil, fmt.Errorf("calculation error: %v", err)
    }

    // Format result with specified precision
    resultText := fmt.Sprintf("%."+strconv.Itoa(e.Precision)+"f", result)

    // Create new context with the result
    newContext := &anyi.FlowContext{
        Text:   resultText,
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *MathCalculatorExecutor) evaluateExpression(expr string) (float64, error) {
    // Simple regex-based calculator for basic operations
    // In a real implementation, you'd use a proper expression parser

    // Handle addition
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\+\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        return a + b, nil
    }

    // Handle subtraction
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*-\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        return a - b, nil
    }

    // Handle multiplication
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\*\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        return a * b, nil
    }

    // Handle division
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*/\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        if b == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return a / b, nil
    }

    return 0, fmt.Errorf("unsupported expression: %s", expr)
}

// Usage example
func main() {
    // Create the executor
    calculator := &MathCalculatorExecutor{
        Precision: 2,
    }

    // Create a flow context with a math expression
    context := &anyi.FlowContext{
        Text: "15.5 + 24.3",
    }

    // Execute the calculation
    result, err := calculator.Execute(context)
    if err != nil {
        log.Fatalf("Execution failed: %v", err)
    }

    fmt.Printf("Result: %s\n", result.Text) // Output: Result: 39.80
}
```

## Advanced Custom Executor Examples

### HTTP API Executor

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/jieliu2000/anyi"
)

type HTTPAPIExecutor struct {
    URL     string
    Method  string
    Headers map[string]string
    Timeout time.Duration
}

func (e *HTTPAPIExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Prepare request body
    requestBody := map[string]interface{}{
        "text":   context.Text,
        "memory": context.Memory,
    }

    jsonBody, err := json.Marshal(requestBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %v", err)
    }

    // Create HTTP request
    req, err := http.NewRequest(e.Method, e.URL, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    for key, value := range e.Headers {
        req.Header.Set(key, value)
    }

    // Create HTTP client with timeout
    client := &http.Client{
        Timeout: e.Timeout,
    }

    // Execute request
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %v", err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %v", err)
    }

    // Check status code
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
    }

    // Parse response
    var response struct {
        Text   string      `json:"text"`
        Memory interface{} `json:"memory"`
    }

    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %v", err)
    }

    // Create new context with response data
    newContext := &anyi.FlowContext{
        Text:   response.Text,
        Memory: response.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

// Usage example
func main() {
    executor := &HTTPAPIExecutor{
        URL:     "https://api.example.com/process",
        Method:  "POST",
        Headers: map[string]string{
            "Authorization": "Bearer " + os.Getenv("API_TOKEN"),
        },
        Timeout: 30 * time.Second,
    }

    // Use in a workflow...
}
```

### Database Query Executor

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"

    "github.com/jieliu2000/anyi"
    _ "github.com/lib/pq" // PostgreSQL driver
)

type DatabaseQueryExecutor struct {
    DB    *sql.DB
    Query string
}

func (e *DatabaseQueryExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Execute query with context text as parameter
    rows, err := e.DB.Query(e.Query, context.Text)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %v", err)
    }
    defer rows.Close()

    // Get column names
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get columns: %v", err)
    }

    // Collect results
    var results []map[string]interface{}

    for rows.Next() {
        // Create slice to hold column values
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))

        for i := range columns {
            valuePtrs[i] = &values[i]
        }

        // Scan row into values
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, fmt.Errorf("failed to scan row: %v", err)
        }

        // Create map for this row
        row := make(map[string]interface{})
        for i, col := range columns {
            val := values[i]

            // Convert byte arrays to strings
            if b, ok := val.([]byte); ok {
                val = string(b)
            }

            row[col] = val
        }

        results = append(results, row)
    }

    // Convert results to JSON string
    resultJSON, err := json.MarshalIndent(results, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal results: %v", err)
    }

    // Create new context with results
    newContext := &anyi.FlowContext{
        Text:   string(resultJSON),
        Memory: results, // Store structured data in memory
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

// Usage example
func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgresql://user:password@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create executor
    executor := &DatabaseQueryExecutor{
        DB:    db,
        Query: "SELECT name, description FROM products WHERE name ILIKE '%' || $1 || '%' LIMIT 10",
    }

    // Use in a workflow...
}
```

## Configurable Custom Executors

### Using Configuration Maps

```go
type ConfigurableFileProcessorExecutor struct {
    config map[string]interface{}
}

func NewConfigurableFileProcessorExecutor(config map[string]interface{}) *ConfigurableFileProcessorExecutor {
    return &ConfigurableFileProcessorExecutor{
        config: config,
    }
}

func (e *ConfigurableFileProcessorExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Extract configuration
    inputDir, _ := e.config["inputDir"].(string)
    outputDir, _ := e.config["outputDir"].(string)
    filePattern, _ := e.config["filePattern"].(string)
    maxFiles, _ := e.config["maxFiles"].(float64) // JSON numbers are float64

    // Set defaults
    if inputDir == "" {
        inputDir = "./input"
    }
    if outputDir == "" {
        outputDir = "./output"
    }
    if filePattern == "" {
        filePattern = "*.txt"
    }
    if maxFiles == 0 {
        maxFiles = 10
    }

    // Process files based on configuration
    files, err := e.processFiles(inputDir, outputDir, filePattern, int(maxFiles))
    if err != nil {
        return nil, err
    }

    // Create result summary
    summary := fmt.Sprintf("Processed %d files from %s to %s", len(files), inputDir, outputDir)

    newContext := &anyi.FlowContext{
        Text:   summary,
        Memory: map[string]interface{}{
            "processedFiles": files,
            "inputDir":       inputDir,
            "outputDir":      outputDir,
        },
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *ConfigurableFileProcessorExecutor) processFiles(inputDir, outputDir, pattern string, maxFiles int) ([]string, error) {
    // Implementation details...
    return []string{}, nil
}
```

### Using in Configuration Files

```yaml
flows:
  - name: "file_processing_flow"
    steps:
      - name: "process_files"
        executor:
          type: "file_processor"
          withconfig:
            inputDir: "/data/input"
            outputDir: "/data/output"
            filePattern: "*.csv"
            maxFiles: 50
```

## Registering Custom Executors

### Factory Pattern

```go
package main

import (
    "fmt"
    "github.com/jieliu2000/anyi"
)

// Register custom executor factory
func init() {
    anyi.RegisterExecutorFactory("math_calculator", func(config map[string]interface{}) (anyi.Executor, error) {
        precision, ok := config["precision"].(float64)
        if !ok {
            precision = 2 // default precision
        }

        return &MathCalculatorExecutor{
            Precision: int(precision),
        }, nil
    })

    anyi.RegisterExecutorFactory("http_api", func(config map[string]interface{}) (anyi.Executor, error) {
        url, ok := config["url"].(string)
        if !ok {
            return nil, fmt.Errorf("url is required")
        }

        method, ok := config["method"].(string)
        if !ok {
            method = "POST"
        }

        timeout, ok := config["timeout"].(float64)
        if !ok {
            timeout = 30
        }

        headers := make(map[string]string)
        if h, ok := config["headers"].(map[string]interface{}); ok {
            for k, v := range h {
                if s, ok := v.(string); ok {
                    headers[k] = s
                }
            }
        }

        return &HTTPAPIExecutor{
            URL:     url,
            Method:  method,
            Headers: headers,
            Timeout: time.Duration(timeout) * time.Second,
        }, nil
    })
}
```

## Error Handling in Custom Executors

### Comprehensive Error Handling

```go
type RobustExecutor struct {
    maxRetries int
    retryDelay time.Duration
}

func (e *RobustExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    var lastErr error

    for attempt := 0; attempt <= e.maxRetries; attempt++ {
        result, err := e.attemptExecution(context)
        if err == nil {
            return result, nil
        }

        lastErr = err

        // Check if error is retryable
        if !e.isRetryableError(err) {
            return nil, fmt.Errorf("non-retryable error: %v", err)
        }

        // Wait before retry (except on last attempt)
        if attempt < e.maxRetries {
            time.Sleep(e.retryDelay * time.Duration(attempt+1)) // exponential backoff
        }
    }

    return nil, fmt.Errorf("execution failed after %d attempts: %v", e.maxRetries+1, lastErr)
}

func (e *RobustExecutor) attemptExecution(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Your actual execution logic here
    return nil, nil
}

func (e *RobustExecutor) isRetryableError(err error) bool {
    // Define which errors are retryable
    errStr := err.Error()
    return strings.Contains(errStr, "timeout") ||
           strings.Contains(errStr, "connection") ||
           strings.Contains(errStr, "temporary")
}
```

## Testing Custom Executors

### Unit Testing

```go
package main

import (
    "testing"
    "github.com/jieliu2000/anyi"
)

func TestMathCalculatorExecutor(t *testing.T) {
    executor := &MathCalculatorExecutor{
        Precision: 2,
    }

    tests := []struct {
        name     string
        input    string
        expected string
        hasError bool
    }{
        {
            name:     "addition",
            input:    "10 + 5",
            expected: "15.00",
            hasError: false,
        },
        {
            name:     "division by zero",
            input:    "10 / 0",
            expected: "",
            hasError: true,
        },
        {
            name:     "invalid expression",
            input:    "not a math expression",
            expected: "",
            hasError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            context := &anyi.FlowContext{
                Text: tt.input,
            }

            result, err := executor.Execute(context)

            if tt.hasError {
                if err == nil {
                    t.Errorf("expected error, but got none")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            if result.Text != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result.Text)
            }
        })
    }
}
```

### Integration Testing

```go
func TestCustomExecutorInFlow(t *testing.T) {
    // Register custom executor
    anyi.RegisterExecutorFactory("test_calculator", func(config map[string]interface{}) (anyi.Executor, error) {
        return &MathCalculatorExecutor{Precision: 2}, nil
    })

    // Configure flow with custom executor
    config := anyi.AnyiConfig{
        Flows: []anyi.FlowConfig{
            {
                Name: "test_flow",
                Steps: []anyi.StepConfig{
                    {
                        Name: "calculate",
                        Executor: &anyi.ExecutorConfig{
                            Type: "test_calculator",
                        },
                    },
                },
            },
        },
    }

    err := anyi.Config(&config)
    if err != nil {
        t.Fatalf("Failed to configure: %v", err)
    }

    // Get and run flow
    flow, err := anyi.GetFlow("test_flow")
    if err != nil {
        t.Fatalf("Failed to get flow: %v", err)
    }

    result, err := flow.RunWithInput("25 + 15")
    if err != nil {
        t.Fatalf("Flow execution failed: %v", err)
    }

    if result.Text != "40.00" {
        t.Errorf("Expected 40.00, got %s", result.Text)
    }
}
```

## Best Practices

### Design Principles

1. **Single Responsibility**: Each executor should have one clear purpose
2. **Configurable**: Make executors configurable through the config map
3. **Error Handling**: Implement comprehensive error handling and recovery
4. **Testing**: Write thorough unit and integration tests
5. **Documentation**: Document configuration options and behavior

### Performance Considerations

1. **Resource Management**: Properly manage connections, files, and other resources
2. **Timeouts**: Always implement timeouts for external operations
3. **Memory Usage**: Be mindful of memory usage, especially with large data sets
4. **Concurrency**: Design for concurrent execution when appropriate

### Security Considerations

1. **Input Validation**: Always validate inputs before processing
2. **Sanitization**: Sanitize data before using in external systems
3. **Secrets Management**: Use environment variables for sensitive configuration
4. **Permissions**: Run with minimal required permissions

Custom executors provide powerful extensibility to the Anyi framework. By following these patterns and best practices, you can create robust, reusable components that integrate seamlessly with your AI workflows.
