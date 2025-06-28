# 自定义执行器

本指南解释如何在 Anyi 中创建自定义执行器，以扩展框架功能超越内置组件。

## 概述

自定义执行器允许您：

- 与外部 API 和服务集成
- 实现特定领域的业务逻辑
- 为您的组织创建可重用组件
- 处理专门的数据处理任务
- 将 Anyi 与现有系统桥接

## 执行器接口

所有执行器必须实现 `Executor` 接口：

```go
type Executor interface {
    Execute(context *FlowContext) (*FlowContext, error)
}
```

### FlowContext 结构

`FlowContext` 包含流经工作流的所有数据：

```go
type FlowContext struct {
    Text   string      // 当前文本内容
    Memory interface{} // 结构化数据
    Think  string      // 思考过程
    Images []string    // 图像 URL
}
```

## 创建简单的自定义执行器

### 示例：数学计算器执行器

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
    Precision int // 小数位数
}

func (e *MathCalculatorExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 从文本中提取数学表达式
    expression := strings.TrimSpace(context.Text)

    // 简单计算器逻辑（您可以使用适当的表达式解析器）
    result, err := e.evaluateExpression(expression)
    if err != nil {
        return nil, fmt.Errorf("计算错误：%v", err)
    }

    // 使用指定精度格式化结果
    resultText := fmt.Sprintf("%."+strconv.Itoa(e.Precision)+"f", result)

    // 创建包含结果的新上下文
    newContext := &anyi.FlowContext{
        Text:   resultText,
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *MathCalculatorExecutor) evaluateExpression(expr string) (float64, error) {
    // 基于正则表达式的简单计算器，用于基本运算
    // 在实际实现中，您应该使用适当的表达式解析器

    // 处理加法
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\+\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        return a + b, nil
    }

    // 处理减法
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*-\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        return a - b, nil
    }

    // 处理乘法
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\*\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        return a * b, nil
    }

    // 处理除法
    if match := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*/\s*(\d+(?:\.\d+)?)`).FindStringSubmatch(expr); match != nil {
        a, _ := strconv.ParseFloat(match[1], 64)
        b, _ := strconv.ParseFloat(match[2], 64)
        if b == 0 {
            return 0, fmt.Errorf("除零错误")
        }
        return a / b, nil
    }

    return 0, fmt.Errorf("不支持的表达式：%s", expr)
}

// 使用示例
func main() {
    // 创建执行器
    calculator := &MathCalculatorExecutor{
        Precision: 2,
    }

    // 创建包含数学表达式的流程上下文
    context := &anyi.FlowContext{
        Text: "15.5 + 24.3",
    }

    // 执行计算
    result, err := calculator.Execute(context)
    if err != nil {
        log.Fatalf("执行失败：%v", err)
    }

    fmt.Printf("结果：%s\n", result.Text) // 输出：结果：39.80
}
```

## 高级自定义执行器示例

### HTTP API 执行器

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
    // 准备请求体
    requestBody := map[string]interface{}{
        "text":   context.Text,
        "memory": context.Memory,
    }

    jsonBody, err := json.Marshal(requestBody)
    if err != nil {
        return nil, fmt.Errorf("序列化请求失败：%v", err)
    }

    // 创建 HTTP 请求
    req, err := http.NewRequest(e.Method, e.URL, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("创建请求失败：%v", err)
    }

    // 设置请求头
    req.Header.Set("Content-Type", "application/json")
    for key, value := range e.Headers {
        req.Header.Set(key, value)
    }

    // 创建带超时的 HTTP 客户端
    client := &http.Client{
        Timeout: e.Timeout,
    }

    // 执行请求
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP 请求失败：%v", err)
    }
    defer resp.Body.Close()

    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("读取响应失败：%v", err)
    }

    // 检查状态码
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("HTTP 错误 %d：%s", resp.StatusCode, string(body))
    }

    // 解析响应
    var response struct {
        Text   string      `json:"text"`
        Memory interface{} `json:"memory"`
    }

    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("解析响应失败：%v", err)
    }

    // 创建包含响应数据的新上下文
    newContext := &anyi.FlowContext{
        Text:   response.Text,
        Memory: response.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

// 使用示例
func main() {
    executor := &HTTPAPIExecutor{
        URL:     "https://api.example.com/process",
        Method:  "POST",
        Headers: map[string]string{
            "Authorization": "Bearer " + os.Getenv("API_TOKEN"),
        },
        Timeout: 30 * time.Second,
    }

    // 在工作流中使用...
}
```

### 数据库查询执行器

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"

    _ "github.com/lib/pq" // PostgreSQL 驱动
    "github.com/jieliu2000/anyi"
)

type DatabaseQueryExecutor struct {
    DB    *sql.DB
    Query string
}

func (e *DatabaseQueryExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 执行查询（使用上下文中的文本作为参数）
    rows, err := e.DB.Query(e.Query, context.Text)
    if err != nil {
        return nil, fmt.Errorf("数据库查询失败：%v", err)
    }
    defer rows.Close()

    // 获取列名
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("获取列名失败：%v", err)
    }

    // 读取结果
    var results []map[string]interface{}
    for rows.Next() {
        // 创建值的切片
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }

        // 扫描行
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, fmt.Errorf("扫描行失败：%v", err)
        }

        // 创建结果映射
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }

    // 将结果序列化为 JSON
    resultJSON, err := json.Marshal(results)
    if err != nil {
        return nil, fmt.Errorf("序列化结果失败：%v", err)
    }

    // 创建新上下文
    newContext := &anyi.FlowContext{
        Text:   string(resultJSON),
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

// 使用示例
func main() {
    // 连接数据库
    db, err := sql.Open("postgres", "user=username dbname=mydb sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 创建执行器
    executor := &DatabaseQueryExecutor{
        DB:    db,
        Query: "SELECT * FROM products WHERE name ILIKE '%' || $1 || '%'",
    }

    // 在工作流中使用...
}
```

### 文件处理执行器

```go
package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "github.com/jieliu2000/anyi"
)

type FileProcessorExecutor struct {
    InputDir  string
    OutputDir string
    Operation string // "read", "write", "list", "delete"
}

func (e *FileProcessorExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    switch e.Operation {
    case "read":
        return e.readFile(context)
    case "write":
        return e.writeFile(context)
    case "list":
        return e.listFiles(context)
    case "delete":
        return e.deleteFile(context)
    default:
        return nil, fmt.Errorf("不支持的操作：%s", e.Operation)
    }
}

func (e *FileProcessorExecutor) readFile(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    filename := strings.TrimSpace(context.Text)
    filepath := filepath.Join(e.InputDir, filename)

    // 安全检查：防止路径遍历攻击
    if !strings.HasPrefix(filepath, e.InputDir) {
        return nil, fmt.Errorf("无效的文件路径：%s", filename)
    }

    content, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, fmt.Errorf("读取文件失败：%v", err)
    }

    newContext := &anyi.FlowContext{
        Text:   string(content),
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *FileProcessorExecutor) writeFile(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 从内存中获取文件名
    filename, ok := context.Memory.(map[string]interface{})["filename"].(string)
    if !ok {
        return nil, fmt.Errorf("内存中未找到文件名")
    }

    filepath := filepath.Join(e.OutputDir, filename)

    // 安全检查
    if !strings.HasPrefix(filepath, e.OutputDir) {
        return nil, fmt.Errorf("无效的文件路径：%s", filename)
    }

    // 确保输出目录存在
    if err := os.MkdirAll(e.OutputDir, 0755); err != nil {
        return nil, fmt.Errorf("创建输出目录失败：%v", err)
    }

    // 写入文件
    if err := ioutil.WriteFile(filepath, []byte(context.Text), 0644); err != nil {
        return nil, fmt.Errorf("写入文件失败：%v", err)
    }

    newContext := &anyi.FlowContext{
        Text:   fmt.Sprintf("文件已写入：%s", filepath),
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *FileProcessorExecutor) listFiles(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    files, err := ioutil.ReadDir(e.InputDir)
    if err != nil {
        return nil, fmt.Errorf("列出文件失败：%v", err)
    }

    var fileList []string
    for _, file := range files {
        if !file.IsDir() {
            fileList = append(fileList, file.Name())
        }
    }

    newContext := &anyi.FlowContext{
        Text:   strings.Join(fileList, "\n"),
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *FileProcessorExecutor) deleteFile(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    filename := strings.TrimSpace(context.Text)
    filepath := filepath.Join(e.InputDir, filename)

    // 安全检查
    if !strings.HasPrefix(filepath, e.InputDir) {
        return nil, fmt.Errorf("无效的文件路径：%s", filename)
    }

    if err := os.Remove(filepath); err != nil {
        return nil, fmt.Errorf("删除文件失败：%v", err)
    }

    newContext := &anyi.FlowContext{
        Text:   fmt.Sprintf("文件已删除：%s", filename),
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}
```

## 配置驱动的执行器

### 可配置执行器接口

```go
type ConfigurableExecutor interface {
    Executor
    Configure(config map[string]interface{}) error
    Validate() error
}

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
        return fmt.Errorf("SMTP 主机不能为空")
    }

    if e.SMTPPort == 0 {
        return fmt.Errorf("SMTP 端口不能为空")
    }

    if e.FromAddress == "" {
        return fmt.Errorf("发件人地址不能为空")
    }

    if len(e.ToAddresses) == 0 {
        return fmt.Errorf("收件人地址不能为空")
    }

    return nil
}

func (e *EmailExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 渲染邮件模板
    body := strings.ReplaceAll(e.BodyTemplate, "{{.Text}}", context.Text)

    // 发送邮件（使用 net/smtp 或第三方库）
    err := e.sendEmail(e.Subject, body)
    if err != nil {
        return nil, fmt.Errorf("发送邮件失败：%v", err)
    }

    newContext := &anyi.FlowContext{
        Text:   fmt.Sprintf("邮件已发送给 %d 个收件人", len(e.ToAddresses)),
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return newContext, nil
}

func (e *EmailExecutor) sendEmail(subject, body string) error {
    // 实现邮件发送逻辑
    // 这里是简化示例
    return nil
}
```

## 执行器注册和工厂模式

### 执行器工厂

```go
type ExecutorFactory func(config map[string]interface{}) (Executor, error)

var executorRegistry = make(map[string]ExecutorFactory)

func RegisterExecutor(name string, factory ExecutorFactory) {
    executorRegistry[name] = factory
}

func CreateExecutor(name string, config map[string]interface{}) (Executor, error) {
    factory, exists := executorRegistry[name]
    if !exists {
        return nil, fmt.Errorf("未知的执行器类型：%s", name)
    }

    return factory(config)
}

// 注册内置执行器
func init() {
    RegisterExecutor("math", func(config map[string]interface{}) (Executor, error) {
        executor := &MathCalculatorExecutor{
            Precision: 2, // 默认精度
        }

        if precision, ok := config["precision"].(float64); ok {
            executor.Precision = int(precision)
        }

        return executor, nil
    })

    RegisterExecutor("http", func(config map[string]interface{}) (Executor, error) {
        executor := &HTTPAPIExecutor{
            Method:  "POST",
            Timeout: 30 * time.Second,
            Headers: make(map[string]string),
        }

        if url, ok := config["url"].(string); ok {
            executor.URL = url
        }

        if method, ok := config["method"].(string); ok {
            executor.Method = method
        }

        if timeout, ok := config["timeout"].(float64); ok {
            executor.Timeout = time.Duration(timeout) * time.Second
        }

        if headers, ok := config["headers"].(map[string]interface{}); ok {
            for key, value := range headers {
                if valueStr, ok := value.(string); ok {
                    executor.Headers[key] = valueStr
                }
            }
        }

        return executor, nil
    })

    RegisterExecutor("email", func(config map[string]interface{}) (Executor, error) {
        executor := &EmailExecutor{}

        if err := executor.Configure(config); err != nil {
            return nil, err
        }

        if err := executor.Validate(); err != nil {
            return nil, err
        }

        return executor, nil
    })
}
```

### 在配置文件中使用自定义执行器

```yaml
# config.yaml
flows:
  - name: "数据处理流程"
    steps:
      - name: "计算步骤"
        executor:
          type: "math"
          withconfig:
            precision: 3

      - name: "API调用步骤"
        executor:
          type: "http"
          withconfig:
            url: "https://api.example.com/process"
            method: "POST"
            timeout: 60
            headers:
              Authorization: "Bearer {{.Memory.token}}"
              Content-Type: "application/json"

      - name: "发送通知"
        executor:
          type: "email"
          withconfig:
            smtpHost: "smtp.gmail.com"
            smtpPort: 587
            username: "{{.Memory.emailUser}}"
            password: "{{.Memory.emailPass}}"
            fromAddress: "noreply@example.com"
            toAddresses:
              - "admin@example.com"
            subject: "处理完成通知"
            bodyTemplate: "处理结果：{{.Text}}"
```

## 错误处理和重试

### 带重试的执行器包装器

```go
type RetryableExecutor struct {
    executor    Executor
    maxRetries  int
    backoffBase time.Duration
}

func NewRetryableExecutor(executor Executor, maxRetries int, backoffBase time.Duration) *RetryableExecutor {
    return &RetryableExecutor{
        executor:    executor,
        maxRetries:  maxRetries,
        backoffBase: backoffBase,
    }
}

func (re *RetryableExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    var lastErr error

    for attempt := 0; attempt <= re.maxRetries; attempt++ {
        result, err := re.executor.Execute(context)
        if err == nil {
            return result, nil
        }

        lastErr = err

        // 如果是最后一次尝试，不等待
        if attempt == re.maxRetries {
            break
        }

        // 指数退避
        backoff := re.backoffBase * time.Duration(1<<attempt)
        log.Printf("执行失败（尝试 %d/%d）：%v，%v 后重试",
            attempt+1, re.maxRetries+1, err, backoff)
        time.Sleep(backoff)
    }

    return nil, fmt.Errorf("在 %d 次尝试后失败：%v", re.maxRetries+1, lastErr)
}
```

### 断路器模式

```go
type CircuitBreakerExecutor struct {
    executor      Executor
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

func (cbe *CircuitBreakerExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    cbe.mutex.Lock()
    defer cbe.mutex.Unlock()

    // 检查断路器状态
    if cbe.state == Open {
        if time.Since(cbe.lastFailTime) > cbe.resetTimeout {
            cbe.state = HalfOpen
        } else {
            return nil, fmt.Errorf("断路器打开，服务不可用")
        }
    }

    // 执行
    result, err := cbe.executor.Execute(context)

    if err != nil {
        cbe.failureCount++
        cbe.lastFailTime = time.Now()

        if cbe.failureCount >= cbe.maxFailures {
            cbe.state = Open
        }

        return nil, err
    }

    // 成功执行
    if cbe.state == HalfOpen {
        cbe.state = Closed
    }
    cbe.failureCount = 0

    return result, nil
}
```

## 测试自定义执行器

### 单元测试

```go
package main

import (
    "testing"

    "github.com/jieliu2000/anyi"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMathCalculatorExecutor(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        precision int
        expected  string
        shouldErr bool
    }{
        {
            name:      "简单加法",
            input:     "10 + 5",
            precision: 2,
            expected:  "15.00",
            shouldErr: false,
        },
        {
            name:      "除法",
            input:     "10 / 3",
            precision: 3,
            expected:  "3.333",
            shouldErr: false,
        },
        {
            name:      "除零错误",
            input:     "10 / 0",
            precision: 2,
            expected:  "",
            shouldErr: true,
        },
        {
            name:      "无效表达式",
            input:     "invalid",
            precision: 2,
            expected:  "",
            shouldErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor := &MathCalculatorExecutor{
                Precision: tt.precision,
            }

            context := &anyi.FlowContext{
                Text: tt.input,
            }

            result, err := executor.Execute(context)

            if tt.shouldErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expected, result.Text)
            }
        })
    }
}
```

### 集成测试

```go
func TestHTTPAPIExecutorIntegration(t *testing.T) {
    // 启动测试 HTTP 服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var request struct {
            Text   string      `json:"text"`
            Memory interface{} `json:"memory"`
        }

        err := json.NewDecoder(r.Body).Decode(&request)
        require.NoError(t, err)

        response := map[string]interface{}{
            "text":   "已处理：" + request.Text,
            "memory": request.Memory,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()

    executor := &HTTPAPIExecutor{
        URL:     server.URL,
        Method:  "POST",
        Headers: map[string]string{},
        Timeout: 5 * time.Second,
    }

    context := &anyi.FlowContext{
        Text: "测试输入",
        Memory: map[string]interface{}{
            "key": "value",
        },
    }

    result, err := executor.Execute(context)
    require.NoError(t, err)
    assert.Equal(t, "已处理：测试输入", result.Text)
}
```

## 性能优化

### 连接池

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

func (e *PooledHTTPExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 从池中获取缓冲区
    buf := e.pool.Get().(*bytes.Buffer)
    defer e.pool.Put(buf)
    buf.Reset()

    // 使用缓冲区进行 JSON 编码
    if err := json.NewEncoder(buf).Encode(context); err != nil {
        return nil, err
    }

    // 执行 HTTP 请求...
    // ...

    return nil, nil
}
```

### 缓存

```go
type CachedExecutor struct {
    executor Executor
    cache    map[string]*anyi.FlowContext
    mutex    sync.RWMutex
    ttl      time.Duration
}

func (ce *CachedExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 生成缓存键
    key := ce.generateCacheKey(context)

    // 检查缓存
    ce.mutex.RLock()
    if cached, exists := ce.cache[key]; exists {
        ce.mutex.RUnlock()
        return cached, nil
    }
    ce.mutex.RUnlock()

    // 执行并缓存结果
    result, err := ce.executor.Execute(context)
    if err != nil {
        return nil, err
    }

    ce.mutex.Lock()
    ce.cache[key] = result
    ce.mutex.Unlock()

    // 设置 TTL 清理
    time.AfterFunc(ce.ttl, func() {
        ce.mutex.Lock()
        delete(ce.cache, key)
        ce.mutex.Unlock()
    })

    return result, nil
}

func (ce *CachedExecutor) generateCacheKey(context *anyi.FlowContext) string {
    // 简化的缓存键生成
    return fmt.Sprintf("%x", sha256.Sum256([]byte(context.Text)))
}
```

## 最佳实践

### 1. 错误处理

- 始终提供有意义的错误消息
- 区分临时错误和永久错误
- 实现适当的重试逻辑
- 记录详细的错误信息用于调试

### 2. 资源管理

- 正确关闭文件、数据库连接和网络连接
- 使用上下文进行超时控制
- 实现资源池以提高性能
- 避免资源泄漏

### 3. 安全性

- 验证和清理所有输入
- 使用参数化查询防止注入攻击
- 实现适当的认证和授权
- 不在日志中记录敏感信息

### 4. 可测试性

- 编写全面的单元测试
- 使用依赖注入便于测试
- 创建模拟对象用于外部依赖
- 实现集成测试验证端到端功能

### 5. 配置

- 使所有配置可外部化
- 提供合理的默认值
- 验证配置参数
- 支持配置热重载（如果需要）

通过遵循这些指导原则和示例，您可以创建强大、可靠和可维护的自定义执行器，扩展 Anyi 框架以满足您的特定需求。
