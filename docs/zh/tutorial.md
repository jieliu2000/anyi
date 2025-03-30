# Anyi(安易) 编程指南和示例

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## 目录

- [快速入门](#快速入门)
- [简介](#简介)
- [安装](#安装)
- [大语言模型访问](#大语言模型访问)
  - [客户端创建方式](#客户端创建方式)
  - [客户端配置](#客户端配置)
    - [智谱AI](#智谱ai)
    - [阿里云灵积](#阿里云灵积)
    - [DeepSeek](#deepseek)
    - [Ollama](#ollama)
    - [其他提供商](#其他提供商)
- [聊天API使用](#聊天api使用)
  - [消息结构](#消息结构)
  - [返回值](#返回值)
  - [聊天选项](#聊天选项)
- [多模态模型使用](#多模态模型使用)
  - [发送图片](#发送图片)
  - [使用ContentParts](#使用contentparts)
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

Anyi提供了统一的方式与各种大语言模型（LLM）进行交互，通过一致的接口实现。这种方法使您能够在不更改应用程序逻辑的情况下轻松切换不同的提供商。

### 理解Anyi的客户端架构

在深入代码之前，了解Anyi如何组织LLM访问非常重要：

1. **提供商**：每个LLM服务（OpenAI、DeepSeek等）都有专用的提供商模块
2. **客户端**：处理与特定LLM服务通信的实例
3. **注册表**：全局存储命名客户端，方便在应用程序中检索

### 客户端创建方式

Anyi提供统一的接口访问各种大语言模型。创建客户端有两种主要方法：

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

每个LLM提供商都有自己的配置结构。了解每个提供商的具体配置选项对于优化与不同模型的交互至关重要。

#### 配置最佳实践

- 将API密钥存储在环境变量中，而不是硬编码
- 使用提供商特定的默认配置作为起点
- 考虑为生产环境设置自定义超时
- 对于自托管模型或代理服务，使用自定义基础URL

### 支持的LLM提供商

Anyi支持多种LLM提供商，以满足不同需求和用例。以下是各个支持的提供商的详细描述和示例。

#### 智谱AI

智谱AI提供GLM系列模型。通过 https://open.bigmodel.cn/ 访问API服务。

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

阿里云灵积提供通义千问系列模型。通过 https://help.aliyun.com/zh/dashscope/ 访问API服务。

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

DeepSeek提供专业的聊天和代码生成模型。通过 https://platform.deepseek.ai/ 访问API服务。

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

Ollama提供本地部署开源模型能力。通过运行本地服务器（默认地址：http://localhost:11434）访问API服务。

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

Anyi还支持其他LLM提供商：

- **OpenAI**: 通过 https://platform.openai.com 访问API服务。使用 `openai.DefaultConfig()` 配置。
- **Azure OpenAI**: 通过 https://azure.microsoft.com/products/ai-services/openai-service 访问API服务。使用 `azureopenai.NewConfig()` 配置。
- **Anthropic**: 通过 https://www.anthropic.com/claude 访问API服务。使用 `anthropic.NewConfig()` 配置。
- **SiliconCloud**: 企业级AI解决方案。使用 `siliconcloud.DefaultConfig()` 配置。

### 如何选择合适的LLM提供商

选择合适的LLM提供商应考虑以下因素：

1. **任务类型**：根据任务选择合适的模型，如通义千问Max（复杂问题）、Llama（本地部署）等
2. **语言需求**：中文处理优先考虑智谱AI、阿里云灵积等
3. **隐私要求**：对于敏感数据，建议使用Ollama在本地部署模型
4. **预算考虑**：在功能和成本之间权衡，根据实际需求选择
5. **延迟需求**：本地部署的Ollama可能提供最低的延迟
6. **扩展性**：Azure OpenAI提供了企业级的扩展选项

通过Anyi框架，您可以轻松在这些提供商之间切换，甚至在同一应用中使用多个不同的LLM服务。