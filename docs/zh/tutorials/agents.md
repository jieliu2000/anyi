# 使用 Anyi 构建自主式智能体

Anyi 的智能体框架使您能够创建能够通过利用可用工作流来规划和执行复杂任务的智能自主式智能体。本教程将指导您完成构建和使用自主式智能体的过程。

## 理解智能体

Anyi 中的智能体是一个智能实体，能够：

1. **规划** - 分析任务并使用可用工作流创建执行计划
2. **执行** - 按顺序运行工作流以完成目标
3. **适应** - 根据执行结果和反馈调整计划

智能体对于需要智能决策的复杂多步骤任务特别强大。

## 创建智能体

要创建智能体，请使用 [NewAgent](../../reference/api.md#NewAgent) 函数：

```go
agent, err := anyi.NewAgent(
    "研究员",                                    // 智能体名称（用于注册）
    "研究助理",                                 // 智能体角色
    "擅长研究主题并撰写报告的专家",                // 智能体背景故事
    []string{"研究流程", "分析流程"},              // 可用流程
    client,                                   // 用于规划的 LLM 客户端
)
```

### 参数说明

- **name**: 用于在全局注册表中注册智能体的可选名称
- **role**: 智能体的角色或工作职能
- **backstory**: 帮助智能体理解其目的的背景信息
- **availableFlows**: 智能体可以使用的工作流名称列表
- **client**: 用于智能规划的 LLM 客户端（可以为 nil 以使用简单规划）

## 智能体配置

可以为智能体配置特定参数来控制其行为：

```go
// 默认配置
config := agent.DefaultConfig()

// 自定义配置
config := agent.Config{
    MaxIterations: 15,    // 最大执行迭代次数
    MaxRetries:    5,     // 最大重试次数
    Timeout:       60 * time.Minute, // 执行超时时间
}

// 应用配置
myAgent.Config = config
```

## 执行任务

创建后，智能体可以使用 [Execute](../../reference/api.md#Agent.Execute) 方法执行任务：

```go
result, context, err := agent.Execute(
    "研究人工智能对医疗保健的影响并撰写综合报告",
    anyi.AgentContext{
        Variables: map[string]interface{}{
            "深度":   "详细",
            "来源":   10,
            "格式":   "markdown",
        },
    },
)
```

执行过程包括：

1. **规划** - 智能体分析任务并创建执行计划
2. **执行** - 根据计划按顺序执行工作流
3. **监控** - 监控结果以确定完成情况
4. **适应** - 根据执行结果调整计划

## 规划策略

Anyi 智能体支持多种规划策略：

### 基于 AI 的规划

当提供 LLM 客户端时，智能体使用 AI 创建智能计划：

```go
// 带 LLM 客户端用于 AI 规划的智能体
agent := anyi.NewAgentWithClient(
    "AI 研究员",
    "智能研究助理",
    []string{"网络搜索", "分析", "总结"},
    registry.Global,
    openaiClient, // 用于规划的 LLM 客户端
)
```

### 简单规划

当没有 LLM 客户端时，智能体使用简单的顺序方法：

```go
// 不带 LLM 客户端使用简单规划的智能体
agent := anyi.NewAgent(
    "基础研究员",
    "研究助理",
    []string{"研究", "分析", "总结"},
    registry.Global, // 流程获取器
)
```

## 使用上下文

智能体使用 [AgentContext](../../reference/api.md#AgentContext) 来维护状态和传递变量：

```go
context := anyi.AgentContext{
    Variables: map[string]interface{}{
        "主题": "人工智能在医疗保健中的应用",
        "语气": "专业",
        "风格": "学术",
    },
    Memory: "之前的研究结果",
    History: []string{
        "之前任务的结果",
    },
}
```

上下文通过值传递，确保线程安全并防止意外修改。

## 示例：研究智能体

以下是研究智能体的完整示例：

```go
package main

import (
    "log"
    "os"
    
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

func main() {
    // 创建 LLM 客户端
    config := openai.DefaultConfig("gpt-4")
    config.APIKey = os.Getenv("OPENAI_API_KEY")
    client, err := anyi.NewClient("研究员", config)
    if err != nil {
        log.Fatalf("创建客户端失败: %v", err)
    }

    // 创建研究智能体
    agent, err := anyi.NewAgent(
        "ai_研究员",                            // 注册名称
        "AI 研究助理",                         // 角色
        "擅长研究主题和综合信息的专家",              // 背景故事
        []string{"网络研究", "分析", "撰写报告"},      // 可用流程
        client,                               // 规划客户端
    )
    if err != nil {
        log.Fatalf("创建智能体失败: %v", err)
    }

    // 执行研究任务
    result, context, err := agent.Execute(
        "分析量子计算的最新发展及其对网络安全的潜在影响",
        anyi.AgentContext{
            Variables: map[string]interface{}{
                "深度":       "全面",
                "视角":       "技术",
                "引用":       true,
                "字数限制":    2000,
            },
        },
    )
    if err != nil {
        log.Fatalf("智能体执行失败: %v", err)
    }

    log.Printf("研究完成。结果长度: %d 字符", len(result))
    log.Printf("执行历史: %v", context.History)
}
```

## 最佳实践

### 1. 设计特定的智能体

创建具有特定角色和功能的智能体：

```go
// 好的：特定角色和功能
writerAgent, _ := anyi.NewAgent(
    "技术作家",
    "技术文档编写者",
    "专门创建清晰技术文档的专家",
    []string{"大纲", "草稿", "审阅", "定稿"},
    client,
)

// 避免：通用、无焦点的智能体
genericAgent, _ := anyi.NewAgent(
    "助手",
    "通用助理",
    "做什么都行",
    []string{"所有流程"},
    client,
)
```

### 2. 提供清晰的指令

给智能体清晰、具体的任务：

```go
// 好的：清晰、具体的任务
result, _, _ := agent.Execute(
    "创建设置 Kubernetes 集群的分步指南",
    context,
)

// 避免：模糊、不明确的任务
result, _, _ := agent.Execute(
    "帮助处理 Kubernetes",
    context,
)
```

### 3. 使用适当的流程

将流程与智能体功能匹配：

```go
// 研究智能体与研究导向的流程
researchAgent, _ := anyi.NewAgent(
    "市场研究员",
    "市场研究分析师",
    "市场分析和竞争情报专家",
    []string{
        "竞争对手分析",
        "市场趋势",
        "客户调查",
        "数据分析",
        "报告生成",
    },
    client,
)
```

## 错误处理

智能体在执行过程中可能遇到各种错误：

```go
result, context, err := agent.Execute(task, initialContext)
if err != nil {
    switch {
    case errors.Is(err, agent.ErrPlanningFailed):
        log.Printf("规划失败: %v", err)
    case errors.Is(err, agent.ErrExecutionFailed):
        log.Printf("执行失败: %v", err)
    default:
        log.Printf("未知错误: %v", err)
    }
    
    // 如有必要，从上下文中访问部分结果
    log.Printf("部分结果: %v", context.History)
}
```

## 高级功能

### 自定义流程获取器

对于高级用例，您可以实现自定义流程获取器：

```go
type CustomFlowGetter struct {
    // 自定义流程管理逻辑
}

func (c *CustomFlowGetter) GetFlow(name string) (interface{}, error) {
    // 自定义流程检索逻辑
    return flow, nil
}

agent := agent.NewAgentWithClient(
    "自定义智能体",
    "具有自定义流程管理的智能体",
    []string{"自定义流程1", "自定义流程2"},
    &CustomFlowGetter{},
    client,
)
```

### 监控和日志

为智能体活动实现自定义监控：

```go
// 执行前
log.Printf("开始执行智能体任务: %s", task)

// 执行期间（在自定义执行器中）
log.Printf("智能体正在执行流程: %s", flowName)

// 执行后
log.Printf("智能体执行完成。结果长度: %d", len(result))
```

## 下一步

- 了解 [工作流](workflows.md) 来创建智能体将使用的流程
- 探索 [配置管理](configuration.md) 以进行复杂的智能体设置
- 查看 [API 参考](../../reference/api.md#agent) 获取详细的智能体文档