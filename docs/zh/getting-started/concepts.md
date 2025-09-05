# 基本概念

本指南介绍 Anyi 框架的核心概念，帮助您理解其架构和工作原理。

## Anyi 架构概览

Anyi 采用模块化架构，主要组件包括：

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   应用程序层     │    │   工作流层       │    │   LLM 客户端层   │
│                │    │                │    │                │
│ • 业务逻辑      │    │ • 流程定义      │    │ • DeepSeek     │
│ • 用户界面      │◄──►│ • 步骤执行      │◄──►│ • Anthropic    │
│ • API 端点     │    │ • 上下文管理    │    │ • Ollama       │
│                │    │                │    │ • 其他提供商    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   组件层        │
                    │                │
                    │ • 执行器        │
                    │ • 验证器        │
                    │ • 模板系统      │
                    └─────────────────┘
```

## 核心概念

### 1. 客户端（Clients）

客户端是与 LLM 提供商通信的接口。每个客户端封装了特定提供商的 API 调用逻辑。

#### 客户端特性

- **统一接口**：所有客户端实现相同的 `Client` 接口
- **提供商抽象**：隐藏不同 API 的差异
- **配置管理**：每个客户端有自己的配置选项
- **连接管理**：处理认证、重试、错误处理

#### 示例

```go
// 创建 DeepSeek 客户端
deepseekConfig := deepseek.DefaultConfig(apiKey, "deepseek-reasoner")
deepseekClient, err := anyi.NewClient("deepseek", deepseekConfig)

// 创建 Ollama 客户端
ollamaConfig := ollama.DefaultConfig("llama3")
ollamaClient, err := anyi.NewClient("ollama", ollamaConfig)

// 两个客户端使用相同的接口
response1, _, err := deepseekClient.Chat(messages, nil)
response2, _, err := ollamaClient.Chat(messages, nil)
```

### 2. 消息（Messages）

消息是 AI 对话的基本单位，包含角色和内容信息。

#### 消息结构

```go
type Message struct {
    Role     string      // "user", "assistant", "system"
    Content  string      // 消息内容
    Images   []string    // 图像 URL（多模态）
    Function *Function   // 函数调用信息
}
```

#### 角色类型

- **user**：用户输入的消息
- **assistant**：AI 助手的回复
- **system**：系统指令，用于设定 AI 行为

#### 示例

```go
messages := []chat.Message{
    {
        Role:    "system",
        Content: "你是一个专业的编程助手。",
    },
    {
        Role:    "user",
        Content: "请解释 Go 语言的并发模型。",
    },
}
```

### 3. 流程（Flows）

流程定义了多步骤的 AI 处理管道，每个步骤可以执行不同的操作。

#### 流程组成

- **步骤（Steps）**：流程中的单个处理单元
- **执行器（Executors）**：执行具体任务的组件
- **验证器（Validators）**：验证输出质量的组件
- **上下文（Context）**：在步骤间传递的数据

#### 流程执行流程

```
输入 → 步骤1 → 验证1 → 步骤2 → 验证2 → ... → 输出
  │      │       │       │       │
  │      ▼       │       ▼       │
  │   执行器1    │    执行器2    │
  │              │               │
  └─────── 上下文传递 ──────────┘
```

#### 示例

```yaml
flows:
  - name: "content_processor"
    steps:
      - name: "analyze"
        executor:
          type: "llm"
          withconfig:
            template: "分析：{{.Text}}"
      - name: "summarize"
        executor:
          type: "llm"
          withconfig:
            template: "总结：{{.Text}}"
```

### 4. 步骤（Steps）

步骤是流程中的单个处理单元，包含执行器、验证器和重试逻辑。

#### 步骤属性

- **名称**：步骤的唯一标识符
- **执行器**：执行具体任务的组件
- **验证器**：验证输出的组件（可选）
- **重试次数**：失败时的最大重试次数
- **客户端**：指定使用的 LLM 客户端

#### 步骤执行逻辑

```go
func (step *Step) Execute(context *FlowContext) (*FlowContext, error) {
    for attempt := 0; attempt <= step.MaxRetryTimes; attempt++ {
        // 执行器处理
        result, err := step.Executor.Execute(context)
        if err != nil {
            continue // 重试
        }

        // 验证器检查
        if step.Validator != nil {
            if err := step.Validator.Validate(result); err != nil {
                continue // 重试
            }
        }

        return result, nil
    }
    return nil, errors.New("步骤执行失败")
}
```

### 5. 执行器（Executors）

执行器是执行具体任务的组件，如调用 LLM、处理数据、执行命令等。

#### 内置执行器类型

1. **LLMExecutor**：调用 LLM 生成文本
2. **SetContextExecutor**：设置或修改上下文
3. **ConditionalFlowExecutor**：条件分支执行
4. **RunCommandExecutor**：执行系统命令

#### LLMExecutor 示例

```go
executor := &anyi.LLMExecutor{
    Template: "请分析以下内容：{{.Text}}",
    SystemMessage: "你是一个专业分析师。",
    Temperature: 0.7,
    MaxTokens: 1000,
}
```

### 6. 验证器（Validators）

验证器确保执行器的输出满足特定要求，如长度、格式、内容等。

#### 内置验证器类型

1. **StringValidator**：验证字符串属性
2. **JsonValidator**：验证 JSON 格式和结构
3. **RegexValidator**：使用正则表达式验证

#### StringValidator 示例

```go
validator := &anyi.StringValidator{
    MinLength: 50,        // 最小长度
    MaxLength: 500,       // 最大长度
    Contains: "总结",      // 必须包含的内容
    NotContains: "错误",  // 不能包含的内容
}
```

### 7. 流程上下文（FlowContext）

流程上下文在步骤间传递数据，包含文本、结构化数据、思考过程等。

#### 上下文结构

```go
type FlowContext struct {
    Text   string      // 当前处理的文本
    Memory interface{} // 结构化数据
    Think  string      // AI 的思考过程
    Images []string    // 图像 URL 列表
}
```

#### 上下文使用

```go
// 创建初始上下文
context := anyi.NewFlowContext("要处理的文本")

// 添加结构化数据
context.Memory = map[string]interface{}{
    "task": "分析",
    "priority": "高",
}

// 在模板中使用
template := "任务：{{.Memory.task}}\n优先级：{{.Memory.priority}}\n内容：{{.Text}}"
```

### 8. 配置系统

Anyi 支持多种配置方式，包括代码配置、文件配置和环境变量。

#### 配置层次

1. **环境变量**：最高优先级
2. **配置文件**：中等优先级
3. **默认值**：最低优先级

#### 配置文件示例

```yaml
clients:
  - name: "deepseek"
    type: "deepseek"
    config:
      apiKey: "$DEEPSEEK_API_KEY" # 环境变量替换
      model: "deepseek-reasoner"
      temperature: 0.7

flows:
  - name: "analyzer"
    clientName: "deepseek"
    steps:
      - name: "analyze"
        executor:
          type: "llm"
          withconfig:
            template: "分析：{{.Text}}"
```

## 工作流程示例

让我们通过一个完整的示例来理解这些概念如何协作：

### 场景：文档分析工作流

```go
package main

import (
    "fmt"
    "log"
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek"
)

func main() {
    // 1. 创建客户端
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-reasoner")
    client, err := anyi.NewClient("deepseek", config)
    if err != nil {
        log.Fatal(err)
    }

    // 2. 定义工作流配置
    flowConfig := anyi.AnyiConfig{
        Flows: []anyi.FlowConfig{
            {
                Name: "document_analyzer",
                ClientName: "deepseek",
                Steps: []anyi.StepConfig{
                    {
                        Name: "extract_key_points",
                        Executor: &anyi.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "提取以下文档的关键点：\n\n{{.Text}}",
                                "systemMessage": "你是一个专业的文档分析师。",
                            },
                        },
                        Validator: &anyi.ValidatorConfig{
                            Type: "string",
                            WithConfig: map[string]interface{}{
                                "minLength": 100,
                            },
                        },
                        MaxRetryTimes: 2,
                    },
                    {
                        Name: "generate_summary",
                        Executor: &anyi.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "基于以下关键点生成执行摘要：\n\n{{.Text}}",
                                "systemMessage": "你是一个专业的摘要写作者。",
                            },
                        },
                        Validator: &anyi.ValidatorConfig{
                            Type: "string",
                            WithConfig: map[string]interface{}{
                                "minLength": 50,
                                "maxLength": 300,
                            },
                        },
                    },
                },
            },
        },
    }

    // 3. 应用配置
    err = anyi.Config(&flowConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 4. 获取并运行流程
    flow, err := anyi.GetFlow("document_analyzer")
    if err != nil {
        log.Fatal(err)
    }

    document := `
    人工智能（AI）正在快速发展，影响着各个行业。
    机器学习算法变得更加复杂，能够处理大量数据。
    自然语言处理技术使得人机交互更加自然。
    计算机视觉在图像识别方面取得了重大突破。
    这些技术的结合正在创造新的商业机会。
    `

    result, err := flow.RunWithInput(document)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("📄 文档分析结果:")
    fmt.Println(result.Text)
}
```

### 执行流程

1. **客户端创建**：建立与 DeepSeek 的连接
2. **配置加载**：定义分析工作流
3. **流程获取**：从注册表获取工作流实例
4. **步骤执行**：
   - 步骤 1：提取关键点，验证长度
   - 步骤 2：生成摘要，验证长度范围
5. **结果返回**：返回最终的分析结果

## 最佳实践

### 1. 客户端管理

```go
// ✅ 好的做法：复用客户端
var globalClient anyi.Client

func init() {
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-reasoner")
    globalClient, _ = anyi.NewClient("deepseek", config)
}

// ❌ 避免：每次都创建新客户端
func badExample() {
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-reasoner")
    client, _ := anyi.NewClient("deepseek", config) // 每次都创建
}
```

### 2. 错误处理

```go
// ✅ 好的做法：适当的错误处理
result, err := flow.RunWithInput(input)
if err != nil {
    log.Printf("流程执行失败: %v", err)
    // 实施降级策略
    return handleFallback(input)
}

// ❌ 避免：忽略错误
result, _ := flow.RunWithInput(input) // 忽略错误
```

### 3. 配置管理

```go
// ✅ 好的做法：使用环境变量
config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-reasoner")

// ❌ 避免：硬编码密钥
config := deepseek.DefaultConfig("hardcoded-key", "deepseek-reasoner") // 不安全
```

## 下一步

现在您已经理解了 Anyi 的核心概念，可以：

1. **深入学习**：阅读 [LLM 客户端教程](../tutorials/llm-clients.md)
2. **构建工作流**：学习 [工作流构建](../tutorials/workflows.md)
3. **配置管理**：掌握 [配置管理](../tutorials/configuration.md)
4. **实际应用**：查看 [操作指南](../how-to/provider-setup.md)

通过理解这些基本概念，您将能够构建强大而灵活的 AI 应用程序！
