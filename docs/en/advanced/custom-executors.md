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

All executors must implement the `StepExecutor` interface:

```go
type StepExecutor interface {
    Init() error
    Run(flowContext FlowContext, Step *Step) (*FlowContext, error)
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

    "github.com/jieliu2000/anyi/flow"
)

type MathCalculatorExecutor struct {
    Precision int // Number of decimal places
}

func (e *MathCalculatorExecutor) Init() error {
    return nil
}

func (e *MathCalculatorExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Extract mathematical expressions from the text
    expression := strings.TrimSpace(flowContext.Text)

    // Simple calculator logic (you could use a proper expression parser)
    result, err := e.evaluateExpression(expression)
    if err != nil {
        return nil, fmt.Errorf("calculation error: %v", err)
    }

    // Format result with specified precision
    resultText := fmt.Sprintf("%."+strconv.Itoa(e.Precision)+"f", result)

    // Create new context with the result
    newContext := &flow.FlowContext{
        Text:   resultText,
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
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

    // Initialize the executor
    if err := calculator.Init(); err != nil {
        log.Fatalf("Failed to initialize executor: %v", err)
    }

    // Create a flow context with a math expression
    context := flow.FlowContext{
        Text: "15.5 + 24.3",
    }

    // Execute the calculation
    result, err := calculator.Run(context, nil)
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

    "github.com/jieliu2000/anyi/flow"
)

type HTTPAPIExecutor struct {
    URL     string
    Method  string
    Headers map[string]string
    Timeout time.Duration
}

func (e *HTTPAPIExecutor) Init() error {
    return nil
}

func (e *HTTPAPIExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Prepare request body
    requestBody := map[string]interface{}{
        "text":   flowContext.Text,
        "memory": flowContext.Memory,
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
    newContext := &flow.FlowContext{
        Text:   response.Text,
        Memory: response.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
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

    // Initialize the executor
    if err := executor.Init(); err != nil {
        log.Fatalf("Failed to initialize executor: %v", err)
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

    "github.com/jieliu2000/anyi/flow"
    _ "github.com/lib/pq" // PostgreSQL driver
)

type DatabaseQueryExecutor struct {
    DB    *sql.DB
    Query string
}

func (e *DatabaseQueryExecutor) Init() error {
    return nil
}

func (e *DatabaseQueryExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Execute query with context text as parameter
    rows, err := e.DB.Query(e.Query, flowContext.Text)
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
    resultJSON, err := json.Marshal(results)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal results: %v", err)
    }

    // Create new context
    newContext := &flow.FlowContext{
        Text:   string(resultJSON),
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
    }

    return newContext, nil
}

// Usage example
func main() {
    // Connect to database
    db, err := sql.Open("postgres", "user=username dbname=mydb sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create executor
    executor := &DatabaseQueryExecutor{
        DB:    db,
        Query: "SELECT * FROM products WHERE name ILIKE '%' || $1 || '%'",
    }

    // Initialize the executor
    if err := executor.Init(); err != nil {
        log.Fatalf("Failed to initialize executor: %v", err)
    }

    // Use in a workflow...
}
```

### File Processor Executor

```go
package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "github.com/jieliu2000/anyi/flow"
)

type FileProcessorExecutor struct {
    InputDir  string
    OutputDir string
    Operation string // "read", "write", "list", "delete"
}

func (e *FileProcessorExecutor) Init() error {
    return nil
}

func (e *FileProcessorExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    switch e.Operation {
    case "read":
        return e.readFile(flowContext)
    case "write":
        return e.writeFile(flowContext)
    case "list":
        return e.listFiles(flowContext)
    case "delete":
        return e.deleteFile(flowContext)
    default:
        return nil, fmt.Errorf("unsupported operation: %s", e.Operation)
    }
}

func (e *FileProcessorExecutor) readFile(flowContext flow.FlowContext) (*flow.FlowContext, error) {
    filename := strings.TrimSpace(flowContext.Text)
    filepath := filepath.Join(e.InputDir, filename)

    // Safety check: prevent path traversal attacks
    if !strings.HasPrefix(filepath, e.InputDir) {
        return nil, fmt.Errorf("invalid file path: %s", filename)
    }

    content, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %v", err)
    }

    newContext := &flow.FlowContext{
        Text:   string(content),
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
    }

    return newContext, nil
}

func (e *FileProcessorExecutor) writeFile(flowContext flow.FlowContext) (*flow.FlowContext, error) {
    // Get filename from memory
    filename, ok := flowContext.Memory.(map[string]interface{})["filename"].(string)
    if !ok {
        return nil, fmt.Errorf("filename not found in memory")
    }

    filepath := filepath.Join(e.OutputDir, filename)

    // Safety check
    if !strings.HasPrefix(filepath, e.OutputDir) {
        return nil, fmt.Errorf("invalid file path: %s", filename)
    }

    // Ensure output directory exists
    if err := os.MkdirAll(e.OutputDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create output directory: %v", err)
    }

    // Write file
    if err := ioutil.WriteFile(filepath, []byte(flowContext.Text), 0644); err != nil {
        return nil, fmt.Errorf("failed to write file: %v", err)
    }

    newContext := &flow.FlowContext{
        Text:   fmt.Sprintf("File written: %s", filepath),
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
    }

    return newContext, nil
}

func (e *FileProcessorExecutor) listFiles(flowContext flow.FlowContext) (*flow.FlowContext, error) {
    files, err := ioutil.ReadDir(e.InputDir)
    if err != nil {
        return nil, fmt.Errorf("failed to list files: %v", err)
    }

    var fileList []string
    for _, file := range files {
        if !file.IsDir() {
            fileList = append(fileList, file.Name())
        }
    }

    newContext := &flow.FlowContext{
        Text:   strings.Join(fileList, "\n"),
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
    }

    return newContext, nil
}

func (e *FileProcessorExecutor) deleteFile(flowContext flow.FlowContext) (*flow.FlowContext, error) {
    filename := strings.TrimSpace(flowContext.Text)
    filepath := filepath.Join(e.InputDir, filename)

    // Safety check
    if !strings.HasPrefix(filepath, e.InputDir) {
        return nil, fmt.Errorf("invalid file path: %s", filename)
    }

    if err := os.Remove(filepath); err != nil {
        return nil, fmt.Errorf("failed to delete file: %v", err)
    }

    newContext := &flow.FlowContext{
        Text:   fmt.Sprintf("File deleted: %s", filename),
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
    }

    return newContext, nil
}

// Usage example
func main() {
    executor := &FileProcessorExecutor{
        InputDir:  "/input",
        OutputDir: "/output",
        Operation: "read",
    }

    // Initialize the executor
    if err := executor.Init(); err != nil {
        log.Fatalf("Failed to initialize executor: %v", err)
    }

    // Use in a workflow...
}
```

## Configurable Custom Executors

### Using Configuration Maps

``go
type ConfigurableFileProcessorExecutor struct {
    config map[string]interface{}
}

// EmailExecutor implements StepExecutor for sending emails
type EmailExecutor struct {
    SMTPHost     string
    SMTPPort     int
    Username     string
    Password     string
    FromAddress  string
    ToAddresses  []string
    Subject      string
    BodyTemplate string
}

func (e *EmailExecutor) Init() error {
    return nil
}

func (e *EmailExecutor) Configure(config map[string]interface{}) error {
    if host, ok := config["smtpHost"].(string); ok {
        e.SMTPHost = host
    }

    if port, ok := config["smtpPort"].(float64); ok {
        e.SMTPPort = int(port)
    }

    if username, ok := config["username"].(string); ok {
        e.Username = username
    }

    if password, ok := config["password"].(string); ok {
        e.Password = password
    }

    if from, ok := config["fromAddress"].(string); ok {
        e.FromAddress = from
    }

    if to, ok := config["toAddresses"].([]interface{}); ok {
        for _, addr := range to {
            if addrStr, ok := addr.(string); ok {
                e.ToAddresses = append(e.ToAddresses, addrStr)
            }
        }
    }

    if subject, ok := config["subject"].(string); ok {
        e.Subject = subject
    }

    if template, ok := config["bodyTemplate"].(string); ok {
        e.BodyTemplate = template
    }

    return nil
}

func (e *EmailExecutor) Validate() error {
    if e.SMTPHost == "" {
        return fmt.Errorf("SMTP host cannot be empty")
    }

    if e.SMTPPort == 0 {
        return fmt.Errorf("SMTP port cannot be empty")
    }

    if e.FromAddress == "" {
        return fmt.Errorf("from address cannot be empty")
    }

    if len(e.ToAddresses) == 0 {
        return fmt.Errorf("to addresses cannot be empty")
    }

    return nil
}

func (e *EmailExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Render email template
    body := strings.ReplaceAll(e.BodyTemplate, "{{.Text}}", flowContext.Text)

    // Send email (using net/smtp or third-party library)
    err := e.sendEmail(e.Subject, body)
    if err != nil {
        return nil, fmt.Errorf("failed to send email: %v", err)
    }

    newContext := &flow.FlowContext{
        Text:   fmt.Sprintf("Email sent to %d recipients", len(e.ToAddresses)),
        Memory: flowContext.Memory,
        Think:  flowContext.Think,
        Images: flowContext.Images,
    }

    return newContext, nil
}

func (e *EmailExecutor) sendEmail(subject, body string) error {
    // Implement email sending logic
    // This is a simplified example
    return nil
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

```
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

## Error Handling and Retries

### Retryable Executor Wrapper

```go
type RetryableExecutor struct {
    executor    StepExecutor
    maxRetries  int
    backoffBase time.Duration
}

func NewRetryableExecutor(executor StepExecutor, maxRetries int, backoffBase time.Duration) *RetryableExecutor {
    return &RetryableExecutor{
        executor:    executor,
        maxRetries:  maxRetries,
        backoffBase: backoffBase,
    }
}

func (re *RetryableExecutor) Init() error {
    return nil
}

func (re *RetryableExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    var lastErr error

    for attempt := 0; attempt <= re.maxRetries; attempt++ {
        result, err := re.executor.Run(flowContext, step)
        if err == nil {
            return result, nil
        }

        lastErr = err

        // If this is the last attempt, don't wait
        if attempt == re.maxRetries {
            break
        }

        // Exponential backoff
        backoff := re.backoffBase * time.Duration(1<<attempt)
        log.Printf("Execution failed (attempt %d/%d): %v, retrying after %v",
            attempt+1, re.maxRetries+1, err, backoff)
        time.Sleep(backoff)
    }

    return nil, fmt.Errorf("failed after %d attempts: %v", re.maxRetries+1, lastErr)
}
```

### Circuit Breaker Pattern

```go
type CircuitBreakerExecutor struct {
    executor      StepExecutor
    failureCount  int
    maxFailures   int
    resetTimeout  time.Duration
    lastFailTime  time.Time
    state         CircuitState
    mutex         sync.RWMutex
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func (cbe *CircuitBreakerExecutor) Init() error {
    return nil
}

func (cbe *CircuitBreakerExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    cbe.mutex.Lock()
    defer cbe.mutex.Unlock()

    // Check circuit breaker state
    if cbe.state == Open {
        if time.Since(cbe.lastFailTime) > cbe.resetTimeout {
            cbe.state = HalfOpen
        } else {
            return nil, fmt.Errorf("circuit breaker is open, service unavailable")
        }
    }

    // Execute
    result, err := cbe.executor.Run(flowContext, step)

    if err != nil {
        cbe.failureCount++
        cbe.lastFailTime = time.Now()

        if cbe.failureCount >= cbe.maxFailures {
            cbe.state = Open
        }

        return nil, err
    }

    // Successful execution
    if cbe.state == HalfOpen {
        cbe.state = Closed
    }
    cbe.failureCount = 0

    return result, nil
}
```

## Testing Custom Executors

### Unit Testing

```go
package main

import (
    "testing"

    "github.com/jieliu2000/anyi/flow"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMathCalculatorExecutor(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        precision int
        expected  string
        hasError  bool
    }{
        {
            name:      "addition",
            input:     "10 + 5",
            precision: 2,
            expected:  "15.00",
            hasError:  false,
        },
        {
            name:      "division by zero",
            input:     "10 / 0",
            precision: 2,
            expected:  "",
            hasError:  true,
        },
        {
            name:      "invalid expression",
            input:     "not a math expression",
            precision: 2,
            expected:  "",
            hasError:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor := &MathCalculatorExecutor{
                Precision: tt.precision,
            }

            // Initialize the executor
            err := executor.Init()
            require.NoError(t, err)

            context := flow.FlowContext{
                Text: tt.input,
            }

            result, err := executor.Run(context, nil)

            if tt.hasError {
                assert.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.expected, result.Text)
        })
    }
}
```

### Integration Testing

```go
func TestCustomExecutorInFlow(t *testing.T) {
    // Register custom executor
    anyi.RegisterExecutorFactory("test_calculator", func(config map[string]interface{}) (flow.StepExecutor, error) {
        executor := &MathCalculatorExecutor{Precision: 2}
        if err := executor.Init(); err != nil {
            return nil, err
        }
        return executor, nil
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
    flowInstance, err := anyi.GetFlow("test_flow")
    if err != nil {
        t.Fatalf("Failed to get flow: %v", err)
    }

    result, err := flowInstance.RunWithInput("25 + 15")
    if err != nil {
        t.Fatalf("Flow execution failed: %v", err)
    }

    if result.Text != "40.00" {
        t.Errorf("Expected 40.00, got %s", result.Text)
    }
}
```

## Performance Optimization

### Connection Pooling

```go
type PooledHTTPExecutor struct {
    client *http.Client
    pool   sync.Pool
}

func NewPooledHTTPExecutor() *PooledHTTPExecutor {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    }

    return &PooledHTTPExecutor{
        client: &http.Client{
            Transport: transport,
            Timeout:   30 * time.Second,
        },
        pool: sync.Pool{
            New: func() interface{} {
                return &bytes.Buffer{}
            },
        },
    }
}

func (e *PooledHTTPExecutor) Init() error {
    return nil
}

func (e *PooledHTTPExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Get buffer from pool
    buf := e.pool.Get().(*bytes.Buffer)
    defer e.pool.Put(buf)
    buf.Reset()

    // Use buffer for JSON encoding
    if err := json.NewEncoder(buf).Encode(flowContext); err != nil {
        return nil, err
    }

    // Execute HTTP request...
    // ...

    return &flowContext, nil
}
```

### Caching

```go
type CachedExecutor struct {
    executor StepExecutor
    cache    map[string]*flow.FlowContext
    mutex    sync.RWMutex
    ttl      time.Duration
}

func (ce *CachedExecutor) Init() error {
    return nil
}

func (ce *CachedExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // Generate cache key
    key := ce.generateCacheKey(flowContext)

    // Check cache
    ce.mutex.RLock()
    if cached, exists := ce.cache[key]; exists {
        ce.mutex.RUnlock()
        return cached, nil
    }
    ce.mutex.RUnlock()

    // Execute and cache result
    result, err := ce.executor.Run(flowContext, step)
    if err != nil {
        return nil, err
    }

    ce.mutex.Lock()
    ce.cache[key] = result
    ce.mutex.Unlock()

    // Set TTL cleanup
    time.AfterFunc(ce.ttl, func() {
        ce.mutex.Lock()
        delete(ce.cache, key)
        ce.mutex.Unlock()
    })

    return result, nil
}

func (ce *CachedExecutor) generateCacheKey(flowContext flow.FlowContext) string {
    // Simplified cache key generation
    return fmt.Sprintf("%x", sha256.Sum256([]byte(flowContext.Text)))
}
```

