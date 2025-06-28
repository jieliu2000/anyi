# 工作流构建教程

本教程将指导您如何使用 Anyi 构建复杂的 AI 工作流，从简单的单步流程到复杂的多步骤管道。

## 工作流基础

### 什么是工作流？

工作流是一系列按顺序执行的步骤，每个步骤可以：

- 调用 LLM 生成内容
- 处理和转换数据
- 执行条件分支
- 验证输出质量
- 设置上下文变量

### 工作流组成

```
输入 → 步骤1 → 步骤2 → 步骤3 → 输出
  │      │       │       │
  │      ▼       ▼       ▼
  │   执行器1  执行器2  执行器3
  │      │       │       │
  └─── 上下文传递 ────────┘
```

## 创建第一个工作流

### 1. 简单的文本处理工作流

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

func main() {
    // 配置客户端
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    client, err := anyi.NewClient("openai", config)
    if err != nil {
        log.Fatal(err)
    }

    // 定义工作流
    flowConfig := anyi.AnyiConfig{
        Clients: []anyi.ClientConfig{
            {
                Name: "openai",
                Type: "openai",
                Config: map[string]interface{}{
                    "apiKey": os.Getenv("OPENAI_API_KEY"),
                    "model":  "gpt-3.5-turbo",
                },
            },
        },
        Flows: []anyi.FlowConfig{
            {
                Name:       "text_processor",
                ClientName: "openai",
                Steps: []anyi.StepConfig{
                    {
                        Name: "analyze",
                        Executor: &anyi.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "分析以下文本的主要观点：\n\n{{.Text}}",
                                "systemMessage": "你是一个专业的文本分析师。",
                            },
                        },
                    },
                    {
                        Name: "summarize",
                        Executor: &anyi.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "基于以下分析，生成简洁的摘要：\n\n{{.Text}}",
                                "systemMessage": "你是一个专业的摘要写作者。",
                            },
                        },
                    },
                },
            },
        },
    }

    // 应用配置
    err = anyi.Config(&flowConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 运行工作流
    flow, err := anyi.GetFlow("text_processor")
    if err != nil {
        log.Fatal(err)
    }

    input := "人工智能正在改变世界。机器学习算法使计算机能够从数据中学习，而不需要显式编程。"

    result, err := flow.RunWithInput(input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("处理结果:\n%s\n", result.Text)
}
```

### 2. 使用配置文件

创建 `workflow.yaml`：

```yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-3.5-turbo"
      temperature: 0.7

flows:
  - name: "content_creator"
    clientName: "openai"
    steps:
      - name: "brainstorm"
        executor:
          type: "llm"
          withconfig:
            template: "为以下主题生成5个创意想法：{{.Text}}"
            systemMessage: "你是一个创意专家。"

      - name: "develop_idea"
        executor:
          type: "llm"
          withconfig:
            template: "选择最好的想法并详细展开：\n\n{{.Text}}"
            systemMessage: "你是一个内容开发专家。"

      - name: "write_content"
        executor:
          type: "llm"
          withconfig:
            template: "基于以下想法写一篇300字的文章：\n\n{{.Text}}"
            systemMessage: "你是一个专业的内容写作者。"
```

加载并运行：

```go
package main

import (
    "fmt"
    "log"
    "github.com/jieliu2000/anyi"
)

func main() {
    // 从配置文件加载
    err := anyi.ConfigFromFile("workflow.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 获取并运行工作流
    flow, err := anyi.GetFlow("content_creator")
    if err != nil {
        log.Fatal(err)
    }

    result, err := flow.RunWithInput("可持续发展")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("生成的内容:\n%s\n", result.Text)
}
```

## 高级工作流功能

### 1. 带验证的工作流

```yaml
flows:
  - name: "quality_assured_writer"
    clientName: "openai"
    steps:
      - name: "write_draft"
        executor:
          type: "llm"
          withconfig:
            template: "写一篇关于{{.Text}}的文章"
            systemMessage: "你是一个专业作家。"
        validator:
          type: "string"
          withconfig:
            minLength: 200
            maxLength: 1000
            contains: "结论"
        maxRetryTimes: 3

      - name: "review_and_improve"
        executor:
          type: "llm"
          withconfig:
            template: "审查并改进以下文章：\n\n{{.Text}}"
            systemMessage: "你是一个编辑专家。"
```

### 2. 条件分支工作流

```yaml
flows:
  - name: "smart_responder"
    clientName: "openai"
    steps:
      - name: "classify_intent"
        executor:
          type: "llm"
          withconfig:
            template: "将以下请求分类为：question、complaint、或praise。只返回分类结果：{{.Text}}"
            systemMessage: "你是一个意图分类专家。"

      - name: "route_response"
        executor:
          type: "conditional_flow"
          withconfig:
            conditions:
              - condition: "{{.Text}} == 'question'"
                flow: "answer_question"
              - condition: "{{.Text}} == 'complaint'"
                flow: "handle_complaint"
            default: "general_response"

  - name: "answer_question"
    clientName: "openai"
    steps:
      - name: "provide_answer"
        executor:
          type: "llm"
          withconfig:
            template: "详细回答以下问题：{{.Memory.original_input}}"
            systemMessage: "你是一个知识渊博的助手。"

  - name: "handle_complaint"
    clientName: "openai"
    steps:
      - name: "apologize_and_solve"
        executor:
          type: "llm"
          withconfig:
            template: "对以下投诉表示歉意并提供解决方案：{{.Memory.original_input}}"
            systemMessage: "你是一个客服专家。"
```

### 3. 数据处理工作流

```go
package main

import (
    "fmt"
    "log"
    "os"
    "github.com/jieliu2000/anyi"
)

func main() {
    config := anyi.AnyiConfig{
        Clients: []anyi.ClientConfig{
            {
                Name: "openai",
                Type: "openai",
                Config: map[string]interface{}{
                    "apiKey": os.Getenv("OPENAI_API_KEY"),
                    "model":  "gpt-3.5-turbo",
                },
            },
        },
        Flows: []anyi.FlowConfig{
            {
                Name:       "data_processor",
                ClientName: "openai",
                Steps: []anyi.StepConfig{
                    {
                        Name: "extract_data",
                        Executor: &anyi.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "从以下文本中提取结构化数据，以JSON格式返回：\n\n{{.Text}}",
                                "systemMessage": "你是一个数据提取专家。",
                            },
                        },
                        Validator: &anyi.ValidatorConfig{
                            Type: "json",
                            WithConfig: map[string]interface{}{
                                "required": []string{"name", "age", "occupation"},
                            },
                        },
                    },
                    {
                        Name: "set_extracted_data",
                        Executor: &anyi.ExecutorConfig{
                            Type: "set_context",
                            WithConfig: map[string]interface{}{
                                "key":   "extracted_data",
                                "value": "{{.Text}}",
                            },
                        },
                    },
                    {
                        Name: "generate_report",
                        Executor: &anyi.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "基于以下提取的数据生成报告：\n\n{{.Memory.extracted_data}}",
                                "systemMessage": "你是一个报告生成专家。",
                            },
                        },
                    },
                },
            },
        },
    }

    err := anyi.Config(&config)
    if err != nil {
        log.Fatal(err)
    }

    flow, err := anyi.GetFlow("data_processor")
    if err != nil {
        log.Fatal(err)
    }

    input := "张三，35岁，软件工程师，在北京工作，有10年编程经验，擅长Go和Python。"

    result, err := flow.RunWithInput(input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("处理结果:\n%s\n", result.Text)
}
```

## 上下文管理

### 1. 使用 Memory 存储数据

```yaml
flows:
  - name: "context_demo"
    clientName: "openai"
    steps:
      - name: "store_user_info"
        executor:
          type: "set_context"
          withconfig:
            key: "user_name"
            value: "{{.Text}}"

      - name: "greet_user"
        executor:
          type: "llm"
          withconfig:
            template: "向{{.Memory.user_name}}问好，并询问他们今天想聊什么"
            systemMessage: "你是一个友好的助手。"

      - name: "store_topic"
        executor:
          type: "set_context"
          withconfig:
            key: "topic"
            value: "{{.Text}}"

      - name: "discuss_topic"
        executor:
          type: "llm"
          withconfig:
            template: "与{{.Memory.user_name}}讨论{{.Memory.topic}}"
            systemMessage: "你是一个知识渊博的对话伙伴。"
```

### 2. 动态上下文更新

```go
func runContextAwareFlow() {
    // 创建带有初始上下文的流程
    context := anyi.NewFlowContext("用户输入")
    context.Memory = map[string]interface{}{
        "user_preferences": map[string]string{
            "language": "中文",
            "style":    "正式",
        },
        "session_id": "12345",
    }

    flow, _ := anyi.GetFlow("personalized_assistant")
    result, err := flow.Run(context)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("个性化回复: %s\n", result.Text)
}
```

## 错误处理和重试

### 1. 配置重试策略

```yaml
flows:
  - name: "robust_processor"
    clientName: "openai"
    steps:
      - name: "critical_step"
        executor:
          type: "llm"
          withconfig:
            template: "执行重要任务：{{.Text}}"
        validator:
          type: "string"
          withconfig:
            minLength: 50
            notContains: "错误"
        maxRetryTimes: 5
        retryDelay: "2s"
```

### 2. 自定义错误处理

```go
func runWithErrorHandling() {
    flow, _ := anyi.GetFlow("error_prone_flow")

    result, err := flow.RunWithInput("测试输入")
    if err != nil {
        // 记录错误
        log.Printf("工作流执行失败: %v", err)

        // 尝试降级策略
        fallbackFlow, _ := anyi.GetFlow("fallback_flow")
        result, err = fallbackFlow.RunWithInput("测试输入")
        if err != nil {
            log.Fatal("降级策略也失败了")
        }
    }

    fmt.Printf("最终结果: %s\n", result.Text)
}
```

## 并行处理

### 1. 并行步骤执行

```yaml
flows:
  - name: "parallel_processor"
    clientName: "openai"
    steps:
      - name: "parallel_group"
        executor:
          type: "parallel"
          withconfig:
            steps:
              - name: "analyze_sentiment"
                executor:
                  type: "llm"
                  withconfig:
                    template: "分析情感：{{.Text}}"

              - name: "extract_keywords"
                executor:
                  type: "llm"
                  withconfig:
                    template: "提取关键词：{{.Text}}"

              - name: "classify_topic"
                executor:
                  type: "llm"
                  withconfig:
                    template: "分类主题：{{.Text}}"

      - name: "combine_results"
        executor:
          type: "llm"
          withconfig:
            template: "合并分析结果：\n情感：{{.Memory.sentiment}}\n关键词：{{.Memory.keywords}}\n主题：{{.Memory.topic}}"
```

## 性能优化

### 1. 缓存结果

```go
type CachedFlow struct {
    flow  anyi.Flow
    cache map[string]*anyi.FlowContext
}

func NewCachedFlow(flow anyi.Flow) *CachedFlow {
    return &CachedFlow{
        flow:  flow,
        cache: make(map[string]*anyi.FlowContext),
    }
}

func (cf *CachedFlow) RunWithInput(input string) (*anyi.FlowContext, error) {
    // 检查缓存
    if result, exists := cf.cache[input]; exists {
        return result, nil
    }

    // 执行流程
    result, err := cf.flow.RunWithInput(input)
    if err != nil {
        return nil, err
    }

    // 缓存结果
    cf.cache[input] = result
    return result, nil
}
```

### 2. 批处理

```go
func processBatch(inputs []string) {
    flow, _ := anyi.GetFlow("batch_processor")

    // 并发处理
    type result struct {
        index int
        output *anyi.FlowContext
        err   error
    }

    results := make(chan result, len(inputs))

    for i, input := range inputs {
        go func(index int, inp string) {
            output, err := flow.RunWithInput(inp)
            results <- result{index, output, err}
        }(i, input)
    }

    // 收集结果
    for i := 0; i < len(inputs); i++ {
        res := <-results
        if res.err != nil {
            log.Printf("输入 %d 处理失败: %v", res.index, res.err)
        } else {
            fmt.Printf("输入 %d 结果: %s\n", res.index, res.output.Text)
        }
    }
}
```

## 监控和调试

### 1. 添加日志记录

```yaml
flows:
  - name: "logged_flow"
    clientName: "openai"
    steps:
      - name: "log_input"
        executor:
          type: "set_context"
          withconfig:
            key: "start_time"
            value: "{{now}}"

      - name: "process"
        executor:
          type: "llm"
          withconfig:
            template: "处理：{{.Text}}"

      - name: "log_output"
        executor:
          type: "set_context"
          withconfig:
            key: "end_time"
            value: "{{now}}"
```

### 2. 性能监控

```go
type MonitoredFlow struct {
    flow anyi.Flow
    metrics *FlowMetrics
}

type FlowMetrics struct {
    TotalRuns     int64
    SuccessRuns   int64
    FailedRuns    int64
    AverageTime   time.Duration
    TotalTokens   int64
}

func (mf *MonitoredFlow) RunWithInput(input string) (*anyi.FlowContext, error) {
    start := time.Now()
    mf.metrics.TotalRuns++

    result, err := mf.flow.RunWithInput(input)

    duration := time.Since(start)
    mf.updateMetrics(duration, err)

    return result, err
}

func (mf *MonitoredFlow) updateMetrics(duration time.Duration, err error) {
    if err != nil {
        mf.metrics.FailedRuns++
    } else {
        mf.metrics.SuccessRuns++
    }

    // 更新平均时间
    totalDuration := mf.metrics.AverageTime * time.Duration(mf.metrics.TotalRuns-1)
    mf.metrics.AverageTime = (totalDuration + duration) / time.Duration(mf.metrics.TotalRuns)
}
```

## 最佳实践

### 1. 工作流设计原则

- **单一职责**：每个步骤只做一件事
- **错误处理**：为关键步骤添加验证和重试
- **可测试性**：设计可以独立测试的步骤
- **可维护性**：使用清晰的命名和注释

### 2. 性能优化建议

- **缓存结果**：对重复输入使用缓存
- **并行处理**：独立步骤可以并行执行
- **资源管理**：合理配置客户端连接池
- **监控指标**：跟踪性能和错误率

### 3. 安全考虑

- **输入验证**：验证所有外部输入
- **输出过滤**：过滤敏感信息
- **访问控制**：限制工作流的访问权限
- **审计日志**：记录所有重要操作

## 下一步

现在您已经掌握了工作流构建，可以：

1. 学习 [配置管理](configuration.md) 来更好地管理复杂配置
2. 探索 [多模态应用](multimodal.md) 来处理图像和文本
3. 查看 [错误处理指南](../how-to/error-handling.md) 来构建更健壮的应用
4. 了解 [性能优化](../how-to/performance.md) 来提升工作流性能

通过合理设计和优化工作流，您可以构建强大、可靠、高效的 AI 应用程序！
