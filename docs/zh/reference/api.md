# API 参考

本文档提供 Anyi 框架公共 API 的全面参考。涵盖构建 AI 应用程序时使用的核心接口、方法和数据结构。

## 核心接口

### Client 接口

`Client` 接口是与 LLM 提供商交互的主要方式。

```go
type Client interface {
    Chat(messages []chat.Message, options *chat.Options) (*chat.Message, chat.ResponseInfo, error)
    GetProvider() string
    GetModel() string
}
```

#### 方法

##### Chat

```go
Chat(messages []chat.Message, options *chat.Options) (*chat.Message, chat.ResponseInfo, error)
```

向 LLM 提供商发送聊天请求。

**参数：**

- `messages`: 构成对话的聊天消息数组
- `options`: 请求的可选配置

**返回值：**

- `*chat.Message`: 来自 LLM 的响应消息
- `chat.ResponseInfo`: 响应的元数据（使用的令牌、模型信息等）
- `error`: 请求失败时的错误

##### GetProvider

```go
GetProvider() string
```

返回 LLM 提供商的名称（例如 "openai"、"anthropic"）。

##### GetModel

```go
GetModel() string
```

返回正在使用的模型名称（例如 "gpt-4"、"claude-3-opus"）。

## 数据结构

### chat.Message

表示对话中的单个消息。

```go
type Message struct {
    Role     string      `json:"role"`
    Content  string      `json:"content"`
    Images   []string    `json:"images,omitempty"`
    Function *Function   `json:"function,omitempty"`
}
```

**字段：**

- `Role`: 消息发送者的角色（"user"、"assistant"、"system"）
- `Content`: 消息的文本内容
- `Images`: 多模态消息的图像 URL 数组
- `Function`: 函数调用信息（用于函数调用）

### chat.Options

聊天请求的配置选项。

```go
type Options struct {
    Temperature      *float64    `json:"temperature,omitempty"`
    MaxTokens        *int        `json:"max_tokens,omitempty"`
    TopP             *float64    `json:"top_p,omitempty"`
    FrequencyPenalty *float64    `json:"frequency_penalty,omitempty"`
    PresencePenalty  *float64    `json:"presence_penalty,omitempty"`
    Stop             []string    `json:"stop,omitempty"`
    Functions        []Function  `json:"functions,omitempty"`
}
```

**字段：**

- `Temperature`: 控制随机性（0.0-2.0，默认值因提供商而异）
- `MaxTokens`: 生成的最大令牌数
- `TopP`: 核心采样参数（0.0-1.0）
- `FrequencyPenalty`: 令牌频率惩罚（-2.0 到 2.0）
- `PresencePenalty`: 令牌存在惩罚（-2.0 到 2.0）
- `Stop`: 停止序列数组
- `Functions`: 函数调用的可用函数

### chat.ResponseInfo

包含 LLM 响应的元数据。

```go
type ResponseInfo struct {
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
    Model            string `json:"model"`
    Provider         string `json:"provider"`
}
```

**字段：**

- `PromptTokens`: 输入中的令牌数
- `CompletionTokens`: 响应中的令牌数
- `TotalTokens`: 使用的总令牌数（提示 + 完成）
- `Model`: 生成响应的模型
- `Provider`: 处理请求的提供商

## 客户端管理函数

### anyi.NewClient

```go
func NewClient(name string, config interface{}) (Client, error)
```

创建新的命名客户端并在全局注册表中注册。

**参数：**

- `name`: 客户端的唯一名称
- `config`: 提供商特定的配置

**返回值：**

- `Client`: 创建的客户端实例
- `error`: 客户端创建失败时的错误

### anyi.GetClient

```go
func GetClient(name string) (Client, error)
```

按名称检索先前注册的客户端。

**参数：**

- `name`: 要检索的客户端名称

**返回值：**

- `Client`: 客户端实例
- `error`: 找不到客户端时的错误

### anyi.ListClients

```go
func ListClients() []string
```

返回所有已注册客户端名称的列表。

## 配置函数

### anyi.Config

```go
func Config(config *AnyiConfig) error
```

应用配置以设置客户端和流程。

**参数：**

- `config`: 包含客户端和流程的配置结构

**返回值：**

- `error`: 配置失败时的错误

### anyi.ConfigFromFile

```go
func ConfigFromFile(filename string) error
```

从文件加载配置（支持 YAML、JSON、TOML）。

**参数：**

- `filename`: 配置文件的路径

**返回值：**

- `error`: 加载失败时的错误

## 流程管理

### Flow 接口

```go
type Flow interface {
    Run() (*FlowContext, error)
    RunWithInput(input interface{}) (*FlowContext, error)
    GetName() string
}
```

#### 方法

##### Run

```go
Run() (*FlowContext, error)
```

执行没有初始输入的流程。

##### RunWithInput

```go
RunWithInput(input interface{}) (*FlowContext, error)
```

使用提供的输入执行流程。

**参数：**

- `input`: 流程的初始输入（字符串或结构化数据）

**返回值：**

- `*FlowContext`: 执行后的最终流程上下文
- `error`: 流程执行失败时的错误

##### GetName

```go
GetName() string
```

返回流程的名称。

### FlowContext 结构

```go
type FlowContext struct {
    Text   string                 `json:"text"`
    Memory map[string]interface{} `json:"memory"`
    Think  string                 `json:"think"`
    Images []string               `json:"images"`
}
```

**字段：**

- `Text`: 当前文本内容
- `Memory`: 结构化内存数据
- `Think`: 思考过程内容
- `Images`: 图像 URL 数组

## 步骤管理

### Step 结构

```go
type Step struct {
    Name          string
    ClientName    string
    Executor      Executor
    Validator     Validator
    MaxRetryTimes int
    VarsImmutable bool
    TextImmutable bool
    MemoryImmutable  bool
}
```

**字段：**

- `Name`: 步骤标识符
- `ClientName`: 用于此步骤的客户端
- `Executor`: 执行器实例
- `Validator`: 验证器实例
- `MaxRetryTimes`: 最大重试次数
- `VarsImmutable`: 当设置为 true 时，步骤执行过程中不会修改上下文变量
- `TextImmutable`: 当设置为 true 时，步骤执行过程中不会修改上下文文本
- `MemoryImmutable`: 当设置为 true 时，步骤执行过程中不会修改上下文内存

## 执行器接口

### Executor 接口

```go
type Executor interface {
    Execute(ctx *FlowContext, client Client) (*FlowContext, error)
}
```

所有执行器必须实现此接口。

### 内置执行器

#### LLMExecutor

用于 LLM 处理的执行器。

```go
type LLMExecutor struct {
    Template         string
    SystemMessage    string
    Temperature      *float64
    MaxTokens        *int
    TopP             *float64
    FrequencyPenalty *float64
    PresencePenalty  *float64
    Stop             []string
    ExtractThink     bool
}
```

#### SetContextExecutor

直接修改流程上下文的执行器。

```go
type SetContextExecutor struct {
    Text   string
    Memory map[string]interface{}
    Think  string
    Images []string
    Append bool
}
```

#### ConditionalFlowExecutor

实现条件分支逻辑的执行器。

```go
type ConditionalFlowExecutor struct {
    Condition  string
    TrueFlow   string
    FalseFlow  string
    TrueSteps  []Step
    FalseSteps []Step
}
```

## 验证器接口

### Validator 接口

```go
type Validator interface {
    Validate(ctx *FlowContext) error
}
```

所有验证器必须实现此接口。

### 内置验证器

#### StringValidator

验证字符串内容的验证器。

```go
type StringValidator struct {
    MinLength     int
    MaxLength     int
    Contains      string
    NotContains   string
    MatchRegex    string
    NotMatchRegex string
    StartsWith    string
    EndsWith      string
}
```

#### JSONValidator

验证 JSON 结构的验证器。

```go
type JSONValidator struct {
    RequiredFields []string
    Schema         string
}
```

## 错误处理

### 错误类型

Anyi 定义了以下错误类型：

```go
var (
    ErrClientNotFound    = errors.New("client not found")
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrExecutionFailed   = errors.New("execution failed")
    ErrValidationFailed  = errors.New("validation failed")
    ErrNetworkTimeout    = errors.New("network timeout")
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
```

### 错误处理示例

```go
response, info, err := client.Chat(messages, nil)
if err != nil {
    switch {
    case errors.Is(err, anyi.ErrNetworkTimeout):
        // 处理网络超时
        log.Println("网络超时，稍后重试")
    case errors.Is(err, anyi.ErrRateLimitExceeded):
        // 处理速率限制
        log.Println("达到速率限制，等待后重试")
    default:
        // 处理其他错误
        log.Printf("未知错误: %v", err)
    }
    return
}
```

## 最佳实践

### 客户端管理

```go
// 在应用启动时创建客户端
func initClients() error {
    // 快速模型用于简单任务
    fastConfig := openai.NewConfigWithModel(apiKey, "gpt-3.5-turbo")
    _, err := anyi.NewClient("fast", fastConfig)
    if err != nil {
        return err
    }

    // 强大模型用于复杂任务
    powerConfig := openai.NewConfigWithModel(apiKey, "gpt-4")
    _, err = anyi.NewClient("power", powerConfig)
    if err != nil {
        return err
    }

    return nil
}
```

### 错误重试

```go
func robustChat(clientName string, messages []chat.Message, maxRetries int) (*chat.Message, error) {
    client, err := anyi.GetClient(clientName)
    if err != nil {
        return nil, err
    }

    for i := 0; i < maxRetries; i++ {
        response, _, err := client.Chat(messages, nil)
        if err == nil {
            return response, nil
        }

        if i < maxRetries-1 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }

    return nil, fmt.Errorf("在 %d 次重试后失败", maxRetries)
}
```

### 资源清理

```go
// 在应用关闭时清理资源
func cleanup() {
    // Anyi 客户端会自动清理
    // 但你可以在这里添加自定义清理逻辑
}
```
