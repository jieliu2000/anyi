# Anyi(安易) 编程指南和示例

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## 目录

- [快速入门](#快速入门)
- [简介](#简介)
- [安装](#安装)
- [大语言模型访问](#大语言模型访问)
  - [客户端创建方式](#客户端创建方式)
  - [客户端配置](#客户端配置)
    - [智谱 AI](#智谱ai)
    - [阿里云灵积](#阿里云灵积)
    - [DeepSeek](#deepseek)
    - [Ollama](#ollama)
    - [其他提供商](#其他提供商)
- [聊天 API 使用](#聊天api使用)
  - [消息结构](#消息结构)
  - [返回值](#返回值)
  - [聊天选项](#聊天选项)
- [多模态模型使用](#多模态模型使用)
  - [发送图片](#发送图片)
  - [使用 ContentParts](#使用contentparts)
- [函数调用](#函数调用)
  - [函数定义](#函数定义)
  - [函数调用示例](#函数调用示例)
- [工作流系统](#工作流系统)
  - [创建工作流](#创建工作流)
  - [步骤和执行器](#步骤和执行器)
  - [步骤间数据传递](#步骤间数据传递)
  - [验证和重试](#验证和重试)
  - [条件工作流](#条件工作流)
- [配置系统](#配置系统)
  - [动态配置](#动态配置)
  - [配置文件](#配置文件)
  - [环境变量](#环境变量)
- [内置组件](#内置组件)
  - [执行器](#执行器)
  - [验证器](#验证器)
- [高级用法](#高级用法)
  - [多客户端管理](#多客户端管理)
  - [提示词模板](#提示词模板)
  - [错误处理](#错误处理)
- [最佳实践](#最佳实践)
  - [性能优化](#性能优化)
  - [成本管理](#成本管理)
  - [安全考虑](#安全考虑)
- [常见问题解答](#常见问题解答)

## 快速入门

如果您想快速上手 Anyi 框架，以下是最基本的步骤：

```bash
# 安装 Anyi
go get -u github.com/jieliu2000/anyi
```

### 基本用法示例

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/zhipu"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 1. 创建客户端
    config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")
    client, err := anyi.NewClient("glm4", config)
    if err != nil {
        log.Fatalf("创建客户端失败: %v", err)
    }

    // 2. 发送简单请求
    messages := []chat.Message{
        {Role: "user", Content: "请简要介绍一下量子计算"},
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("请求失败: %v", err)
    }

    fmt.Println("回答:", response.Content)
}
```

这个简单的例子展示了 Anyi 的核心功能：创建客户端并发送请求。有关更详细的说明，请继续阅读完整指南。

## 大语言模型访问

Anyi 提供了统一的方式与各种大语言模型（LLM）进行交互，通过一致的接口实现。这种方法使您能够在不更改应用程序逻辑的情况下轻松切换不同的提供商。

### 理解 Anyi 的客户端架构

在深入代码之前，了解 Anyi 如何组织 LLM 访问非常重要：

1. **提供商**：每个 LLM 服务（OpenAI、DeepSeek 等）都有专用的提供商模块
2. **客户端**：处理与特定 LLM 服务通信的实例
3. **注册表**：全局存储命名客户端，方便在应用程序中检索

### 客户端创建方式

Anyi 提供统一的接口访问各种大语言模型。创建客户端有两种主要方法：

1. 使用`anyi.NewClient()`- 创建一个注册到全局注册表的命名客户端
2. 使用`llm.NewClient()`- 创建一个由你自己管理的未注册客户端实例

#### 命名客户端示例

```go
package main

import (
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 创建一个名为"gpt4"的客户端
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    config.Model = openai.GPT4o // 使用GPT-4o模型

    client, err := anyi.NewClient("gpt4", config)
    if err != nil {
        log.Fatalf("创建客户端失败: %v", err)
    }

    // 之后可以通过名称检索此客户端
    retrievedClient, err := anyi.GetClient("gpt4")
    if err != nil {
        log.Fatalf("检索客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "法国的首都是什么？"},
    }
    response, _, err := retrievedClient.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    log.Printf("回答: %s", response.Content)
}
```

### 客户端配置

每个 LLM 提供商都有自己的配置结构。了解每个提供商的具体配置选项对于优化与不同模型的交互至关重要。

#### 通用 LLM 配置选项

所有 LLM 提供商都支持一组通用的配置选项，这些选项通过 `GeneralLLMConfig` 结构体提供。了解这些选项可以帮助您优化模型输出：

```go
// 所有LLM配置结构体都嵌入了GeneralLLMConfig
type SomeProviderConfig struct {
    config.GeneralLLMConfig
    // 其他特定提供商的配置...
}
```

GeneralLLMConfig 包含以下配置选项：

- **Temperature**：控制输出的随机性。值越高，输出越随机；值越低，输出越确定。

  - 范围通常在 0.0 到 2.0 之间，默认值为 1.0
  - 示例：`config.Temperature = 0.7` // 更加确定的输出

- **TopP**：控制输出的多样性。值越高，输出越多样；值越低，输出越保守。

  - 范围通常在 0.0 到 1.0 之间，默认值为 1.0
  - 示例：`config.TopP = 0.9` // 保持一定的多样性

- **MaxTokens**：控制生成的最大 token 数量

  - 0 表示不限制
  - 示例：`config.MaxTokens = 2000` // 限制回答长度

- **PresencePenalty**：控制模型避免重复内容的程度

  - 正值会增加避免重复的可能性，负值会增加重复的可能性
  - 示例：`config.PresencePenalty = 0.5` // 适度避免重复

- **FrequencyPenalty**：控制模型避免使用常见词的程度

  - 正值会增加避免使用常见词的可能性，负值会增加使用常见词的可能性
  - 示例：`config.FrequencyPenalty = 0.5` // 适度避免常见词

- **Stop**：指定停止生成的标记列表
  - 示例：`config.Stop = []string{"###", "结束"}` // 当遇到这些标记时停止生成

##### 示例配置

```go
import (
    "github.com/jieliu2000/anyi/llm/openai"
    "github.com/jieliu2000/anyi/llm/config"
)

// 创建配置
config := openai.DefaultConfig(apiKey)

// 调整通用参数
config.Temperature = 0.7      // 较低的温度，更确定的输出
config.TopP = 0.9             // 轻微限制token选择
config.MaxTokens = 500        // 限制回答长度
config.PresencePenalty = 0.2  // 轻微抑制重复
config.FrequencyPenalty = 0.3 // 轻微抑制常见词
config.Stop = []string{"END"} // 遇到"END"时停止生成
```

#### 配置最佳实践

- 将 API 密钥存储在环境变量中，而不是硬编码
- 使用提供商特定的默认配置作为起点
- 对创意任务使用较高的 Temperature (0.7-1.0)
- 对精确任务使用较低的 Temperature (0.1-0.3)
- 考虑为生产环境设置自定义超时
- 对于自托管模型或代理服务，使用自定义基础 URL

### 支持的 LLM 提供商

Anyi 支持多种 LLM 提供商，以满足不同需求和用例。以下是各个支持的提供商的详细描述和示例。

#### 智谱 AI

智谱 AI 提供 GLM 系列模型。通过 https://open.bigmodel.cn/ 访问 API 服务。

##### 配置示例

```go
package main

import (
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/zhipu"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 默认配置
    config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")

    // 创建客户端
    client, err := anyi.NewClient("zhipu", config)
    if (err != nil) {
        log.Fatalf("创建智谱AI客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "介绍一下中国古代四大发明"},
    }
    response, _, err := client.Chat(messages, nil)
    if (err != nil) {
        log.Fatalf("请求失败: %v", err)
    }

    log.Printf("智谱AI回答: %s", response.Content)
}
```

#### 阿里云灵积

阿里云灵积提供通义千问系列模型。通过 https://help.aliyun.com/zh/dashscope/ 访问 API 服务。

##### 配置示例

```go
package main

import (
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/dashscope"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 默认配置
    config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")

    // 创建客户端
    client, err := anyi.NewClient("qwen", config)
    if (err != nil) {
        log.Fatalf("创建阿里云灵积客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "解释一下量子计算的基本原理"},
    }
    response, _, err := client.Chat(messages, nil)
    if (err != nil) {
        log.Fatalf("请求失败: %v", err)
    }

    log.Printf("阿里云灵积回答: %s", response.Content)
}
```

#### DeepSeek

DeepSeek 提供专业的聊天和代码生成模型。通过 https://platform.deepseek.ai/ 访问 API 服务。

##### 配置示例

```go
package main

import (
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 使用DeepSeek Chat模型配置
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")

    // 使用DeepSeek Coder模型配置
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-coder")

    // 创建客户端
    client, err := llm.NewClient(config)
    if (err != nil) {
        log.Fatalf("创建DeepSeek客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "编写一个Go函数来检查字符串是否为回文"},
    }
    response, _, err := client.Chat(messages, nil)
    if (err != nil) {
        log.Fatalf("请求失败: %v", err)
    }

    log.Printf("DeepSeek回答: %s", response.Content)
}
```

#### Ollama

Ollama 提供本地部署开源模型能力。通过运行本地服务器（默认地址：http://localhost:11434）访问 API 服务。

##### 配置示例

```go
package main

import (
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/ollama"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 默认配置（本地服务器）
    config := ollama.DefaultConfig("llama3")

    // 自定义服务器配置
    config := ollama.NewConfig("mixtral", "http://your-ollama-server:11434")

    // 创建客户端
    client, err := anyi.NewClient("local-llm", config)
    if err != nil {
        log.Fatalf("创建Ollama客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "system", Content: "你是一位专注于数论的数学专家。"},
        {Role: "user", Content: "解释下里曼猜想的基本概念"},
    }
    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("本地推理失败: %v", err)
    }

    log.Printf("Ollama回答: %s", response.Content)
}
```

#### 其他提供商

Anyi 还支持其他 LLM 提供商：

- **OpenAI**: 通过 https://platform.openai.com 访问 API 服务。使用 `openai.DefaultConfig()` 配置。
- **Azure OpenAI**: 通过 https://azure.microsoft.com/products/ai-services/openai-service 访问 API 服务。使用 `azureopenai.NewConfig()` 配置。
- **Anthropic**: 通过 https://www.anthropic.com/claude 访问 API 服务。使用 `anthropic.NewConfig()` 配置。
- **SiliconCloud**: 企业级 AI 解决方案。使用 `siliconcloud.DefaultConfig()` 配置。

### 如何选择合适的 LLM 提供商

选择合适的 LLM 提供商应考虑以下因素：

1. **任务类型**：根据任务选择合适的模型，如通义千问 Max（复杂问题）、Llama（本地部署）等
2. **语言需求**：中文处理优先考虑智谱 AI、阿里云灵积等
3. **隐私要求**：对于敏感数据，建议使用 Ollama 在本地部署模型
4. **预算考虑**：在功能和成本之间权衡，根据实际需求选择
5. **延迟需求**：本地部署的 Ollama 可能提供最低的延迟
6. **扩展性**：Azure OpenAI 提供了企业级的扩展选项

通过 Anyi 框架，您可以轻松在这些提供商之间切换，甚至在同一应用中使用多个不同的 LLM 服务。

## 最佳实践

构建有效的 AI 应用需要的不仅仅是技术知识。以下是一些全面的最佳实践，帮助您充分利用 Anyi 框架。

### 性能优化

优化 Anyi 工作流的性能可以显著提高用户体验并降低成本：

**1. 为任务选择合适的模型**

- 简单任务使用更小、更快的模型
- 复杂推理任务保留更强大的模型
- 考虑使用针对特定领域微调的模型

**2. 合理配置生成参数**

- 根据任务调整 `Temperature`：
  - 对于事实性回答、编程和精确任务，使用 0.1-0.3
  - 对于创意写作、头脑风暴和多样化输出，使用 0.7-1.0
- 设置适当的 `MaxTokens` 以避免不必要的长回答
- 使用 `PresencePenalty` (0.1-0.5) 减少长文本生成中的重复输出
- 应用 `FrequencyPenalty` (0.1-0.5) 鼓励更多样化的词汇
- 使用 `Stop` 标记在适当的点自动结束生成

**3. 实现缓存**

- 缓存常见的 LLM 响应以避免重复 API 调用
- 对于多实例部署使用分布式缓存
- 设置适当的缓存过期时间

**4. 优化提示词**

- 保持提示词简洁同时包含必要的上下文
- 使用清晰的指令减少来回交互
- 测试和迭代提示词以最小化 token 使用

**5. 本地部署选项**

- 对于频繁、非关键任务，使用 Ollama 与本地模型
- 基于需求在云服务和本地模型之间取得平衡
- 考虑在资源受限环境中使用量化模型

**6. 并行执行**

- 识别可以并行运行的工作流步骤
- 适当使用 goroutines 进行并发 LLM 调用
- 为并行步骤实现适当的错误处理

### 成本管理

在使用商业 LLM 提供商时，成本管理至关重要：

**1. Token 监控**

- 实现 token 计数以跟踪使用情况
- 为异常消费模式设置警报
- 定期审计 token 消耗

**2. 分层模型策略**

- 使用级联方法：先尝试更便宜的模型
- 仅在必要时升级到更昂贵的模型
- 为服务中断实现备选方案

**3. 响应长度控制**

- 为每个用例设置适当的 MaxTokens 限制
- 使用验证确保输出不会不必要地冗长
- 为过长输出实现截断策略

**4. 批量请求**

- 可能时合并多个小请求
- 为非紧急处理实现队列系统
- 在非高峰时段安排批处理

**5. 成本归因**

- 按工作流、功能或用户跟踪成本
- 实现每用户配额或速率限制
- 考虑对高级功能向最终用户转嫁成本

### 安全考虑

在构建 AI 系统时，安全至关重要：

**1. API 密钥管理**

- 永远不要在应用程序中硬编码 API 密钥
- 使用环境变量或密钥管理器
- 定期轮换密钥并限制其权限

**2. 输入净化**

- 验证并清理所有用户输入
- 实现速率限制以防止滥用
- 使用上下文过滤防止提示注入

**3. 输出验证**

- 在使用 LLM 输出前始终进行验证
- 在可执行上下文中使用 LLM 输出时要谨慎
- 为面向用户的输出实现内容审核

**4. 数据隐私**

- 最小化向 LLM 发送敏感数据
- 实施数据保留策略
- 考虑使用本地模型处理敏感信息

**5. 审计和日志记录**

- 维护所有 LLM 交互的详细日志
- 实现敏感内容的正确日志编辑
- 设置监控以发现异常模式或安全事件

通过遵循这些最佳实践，您可以构建不仅强大而且高效、经济且安全的 AI 应用程序。

## 常见问题解答
