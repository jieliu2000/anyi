# Anyi(安易) 编程指南和示例

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## 目录

- [快速入门](#快速入门)
- [简介](#简介)
- [安装](#安装)
- [大语言模型访问](#大语言模型访问)
  - [理解 Anyi 的客户端架构](#理解-anyi-的客户端架构)
  - [客户端创建方式](#客户端创建方式)
  - [客户端配置](#客户端配置)
  - [支持的 LLM 提供商](#支持的-llm-提供商)
    - [OpenAI](#openai)
    - [DeepSeek](#deepseek)
    - [Azure OpenAI](#azure-openai)
    - [Ollama](#ollama)
    - [其他提供商](#其他提供商)
  - [如何选择合适的 LLM 提供商](#如何选择合适的-llm-提供商)
- [聊天 API 使用](#聊天api使用)
  - [理解聊天生命周期](#理解聊天生命周期)
  - [消息结构](#消息结构)
  - [返回值解释](#返回值解释)
  - [基本聊天示例](#基本聊天示例)
  - [聊天选项](#聊天选项)
- [多模态模型使用](#多模态模型使用)
  - [发送图片](#发送图片)
- [函数调用](#函数调用)
  - [函数定义](#函数定义)
- [工作流系统](#工作流系统)
  - [核心工作流概念](#核心工作流概念)
  - [流程上下文](#流程上下文)
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

## 简介

Anyi 是一个用 Go 语言编写的开源自主 AI 代理框架，旨在帮助您构建能够与现实世界工作流程无缝集成的 AI 代理。本指南提供了使用 Anyi 框架的详细编程说明和示例。

### Anyi 的主要特性

- **统一的 LLM 接口**：通过一致的 API 访问多个 LLM 提供商
- **灵活的工作流系统**：构建具有验证和错误处理功能的复杂多步骤 AI 流程
- **配置管理**：支持 YAML、JSON 和 TOML 配置文件
- **内置组件**：为常见任务提供即用的执行器和验证器
- **可扩展架构**：创建自定义组件以满足特定需求

### 何时使用 Anyi

Anyi 特别适用于以下场景：

- 需要在多个 AI 模型之间编排复杂交互
- 希望构建具有验证和错误处理功能的可靠 AI 工作流
- 需要在不更改代码的情况下切换不同的 LLM 提供商
- 在 Go 中构建生产级 AI 应用程序

## 安装

要开始使用 Anyi，请通过 Go modules 安装：

```bash
go get -u github.com/jieliu2000/anyi
```

Anyi 需要 Go 1.20 或更高版本。

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

#### 何时使用命名客户端与非注册客户端

- **命名客户端**适用于以下情况：

  - 需要从应用程序的不同部分访问同一个客户端实例
  - 一次配置并在整个代码库中重复使用
  - 管理具有不同配置的多个客户端

- **非注册客户端**更适合于：
  - 需要为特定任务隔离的客户端实例
  - 希望避免潜在的命名冲突
  - 应用程序结构简单，LLM 交互有限

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

#### 非注册客户端示例

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 创建客户端而不注册
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.Model = openai.GPT3Dot5Turbo

	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 直接使用客户端
	messages := []chat.Message{
		{Role: "user", Content: "用简单的术语解释量子计算"},
	}
	response, _, err := client.Chat(messages, nil)
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

Anyi 支持多种 LLM 提供商，以满足不同需求和用例。以下是各个支持的提供商的详细描述和示例，从最广泛使用的选项开始。

#### OpenAI

OpenAI 是最广泛使用的 AI 服务提供商之一。通过 https://platform.openai.com 访问 API 服务。

##### 配置示例

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
	// 默认配置（gpt-3.5-turbo）
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))

	// 指定特定模型的配置
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)

	// 创建客户端和使用示例
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("创建 OpenAI 客户端失败: %v", err)
    }

    messages := []chat.Message{
		{Role: "user", Content: "法国的首都是什么？"},
    }
    response, _, err := client.Chat(messages, nil)
	if err != nil {
        log.Fatalf("请求失败: %v", err)
    }

	log.Printf("OpenAI 回答: %s", response.Content)
}
```

#### DeepSeek

DeepSeek 提供专业的聊天和代码生成模型，可通过 https://platform.deepseek.ai/ 访问。

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
    // 使用 DeepSeek Chat 模型配置
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")

    // 使用 DeepSeek Coder 模型配置
    config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-coder")

    // 创建客户端和使用示例
    client, err := llm.NewClient(config)
    if err != nil {
        log.Fatalf("创建 DeepSeek 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {Role: "user", Content: "编写一个 Go 函数来检查字符串是否为回文"},
    }
    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("请求失败: %v", err)
    }

    log.Printf("DeepSeek 回答: %s", response.Content)
}
```

#### Azure OpenAI

Azure OpenAI 提供微软托管的 OpenAI 模型，具有企业级功能和可靠性。

##### 功能和优势

- 企业级 SLA 和技术支持
- 符合各种监管标准
- 网络隔离和私有网络部署选项
- 与其他 Azure 服务的集成

##### 配置示例

```go
package main

import (
    "log"
    "os"

    "github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/azureopenai"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	config := azureopenai.NewConfig(
		os.Getenv("AZ_OPENAI_API_KEY"),
		os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"),
		os.Getenv("AZ_OPENAI_ENDPOINT")
	)

	// 创建客户端和使用示例
	client, err := anyi.NewClient("azure-openai", config)
	if err != nil {
		log.Fatalf("创建 Azure OpenAI 客户端失败: %v", err)
    }

	// 使用客户端
    messages := []chat.Message{
		{Role: "user", Content: "机器学习和深度学习的主要区别是什么？"},
    }
    response, _, err := client.Chat(messages, nil)
	if err != nil {
        log.Fatalf("请求失败: %v", err)
    }

	log.Printf("Azure OpenAI 回答: %s", response.Content)
}
```

#### Ollama

Ollama 提供本地部署开源模型的能力，非常适合需要离线处理或数据隐私的场景。

##### 功能和优势

- 本地部署无需网络连接
- 支持各种开源模型，如 Llama、Mixtral 等
- 完全控制数据流，增强隐私保护
- 无使用费用，适合大规模实验

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

	// 创建客户端和使用示例
    client, err := anyi.NewClient("local-llm", config)
    if err != nil {
		log.Fatalf("创建 Ollama 客户端失败: %v", err)
    }

	// 使用客户端进行本地推理
    messages := []chat.Message{
        {Role: "system", Content: "你是一位专注于数论的数学专家。"},
		{Role: "user", Content: "用简单的术语解释黎曼猜想"},
    }
    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("本地推理失败: %v", err)
    }

	log.Printf("Ollama 模型回答: %s", response.Content)
}
```

#### 其他提供商

Anyi 还支持各种其他 LLM 提供商：

- **智谱 AI**：通过 https://open.bigmodel.cn/ 访问 GLM 系列模型。使用 `zhipu.DefaultConfig()` 配置。
- **Dashscope（阿里巴巴）**：通过 https://help.aliyun.com/zh/dashscope/ 访问通义千问系列模型。使用 `dashscope.DefaultConfig()` 配置。
- **Anthropic**：通过 https://www.anthropic.com/claude 访问 Claude 模型。使用 `anthropic.NewConfig()` 配置。
- **SiliconCloud**：面向企业的 AI 解决方案。使用 `siliconcloud.DefaultConfig()` 配置。

### 如何选择合适的 LLM 提供商

选择合适的 LLM 提供商应考虑以下因素：

1. **任务类型**：根据任务选择合适的模型，如通义千问 Max（复杂问题）、Llama（本地部署）等
2. **语言需求**：中文处理优先考虑智谱 AI、阿里云灵积等
3. **隐私要求**：对于敏感数据，建议使用 Ollama 在本地部署模型
4. **预算考虑**：在功能和成本之间权衡，根据实际需求选择
5. **延迟需求**：本地部署的 Ollama 可能提供最低的延迟
6. **扩展性**：Azure OpenAI 提供了企业级的扩展选项

通过 Anyi 框架，您可以轻松在这些提供商之间切换，甚至在同一应用中使用多个不同的 LLM 服务。

## 聊天 API 使用

Anyi 的核心功能是通过聊天 API 与 LLM 交互。本节解释如何构建对话、处理响应和自定义聊天行为。

### 理解聊天生命周期

与 LLM 的典型聊天交互遵循以下步骤：

1. **准备消息**：创建代表对话的消息序列
2. **配置选项**：设置温度、最大 token 数等参数
3. **发送请求**：在客户端调用 Chat 方法
4. **处理响应**：处理模型的回复和任何元数据
5. **继续对话**：将响应添加到消息历史中以进行后续交互

### 消息结构

Anyi 中的聊天消息使用 `chat.Message` 结构体：

```go
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string // 消息的文本内容
	Name    string // 可选名称（用于多代理上下文）

	// 用于多模态内容
	ContentParts []ContentPart
}
```

### 返回值解释

调用 Chat 方法时，您会收到三个值：

1. **响应消息**：模型的回复作为 `chat.Message`
2. **响应信息**：关于响应的元数据（使用的 token、模型名称等）
3. **错误**：请求期间发生的任何错误

理解这些返回值有助于您实现适当的错误处理和日志记录。

### 基本聊天示例

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
	// 创建客户端
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 创建消息历史
	messages := []chat.Message{
		{Role: "system", Content: "你是一个有帮助的助手。"},
		{Role: "user", Content: "机器学习可以用于什么？"},
	}

	// 发送聊天请求
	response, info, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}

	// 处理响应
	log.Printf("模型: %s", info.Model)
	log.Printf("回答: %s", response.Content)

	// 继续对话
	messages = append(messages, *response) // 添加助手的回答
	messages = append(messages, chat.Message{
		Role: "user",
		Content: "能给出一个在医疗保健中的具体例子吗？",
	})

	// 发送后续消息
	response, _, err = client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}

	log.Printf("后续回答: %s", response.Content)
}
```

### 聊天选项

您可以使用 `chat.ChatOptions` 自定义聊天行为：

```go
options := &chat.ChatOptions{
	Format: "json", // 指定输出为 JSON 格式（对结构化数据有用）
}

response, info, err := client.Chat(messages, options)
```

目前，Anyi 框架提供了简化的 `ChatOptions`，具有以下功能：

1. **Format**：当设置为 "json" 时，它指示模型以 JSON 格式返回响应。这在需要结构化数据时特别有用。

不同的 LLM 提供商可能会根据其底层 API 的功能以不同的行为实现这些选项。

## 多模态模型使用

许多现代 LLM 支持多模态输入，允许您在文本的同时发送图片。

### 发送图片

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
	// 创建 GPT-4 Vision 客户端
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4-vision-preview")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 创建包含图片 URL 的消息
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "这张图片中有什么？",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://example.com/image.jpg",
					},
				},
			},
		},
	}

	// 发送聊天请求
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}

	log.Printf("描述: %s", response.Content)
}
```

## 函数调用

许多 LLM 支持函数调用功能，允许 AI 模型请求特定操作。

### 函数定义

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

func main() {
	// 创建客户端
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4")
	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 定义函数
	functions := []tools.FunctionConfig{
		{
			Name:        "get_weather",
			Description: "获取指定位置的当前天气",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "城市和州，例如 'San Francisco, CA'",
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"celsius", "fahrenheit"},
						"description": "温度单位",
					},
				},
				"required": []string{"location"},
			},
		},
	}

	// 创建消息
	messages := []chat.Message{
		{Role: "user", Content: "波士顿的天气怎么样？"},
	}

	// 使用函数调用的请求
	response, _, err := client.ChatWithFunctions(messages, functions, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}

	log.Printf("响应类型: %s", response.FunctionCall.Name)
	log.Printf("参数: %s", response.FunctionCall.Arguments)

	// 在这里您将处理函数调用，执行请求的函数
	// 并在另一条消息中发送结果
}
```

## 工作流系统

Anyi 框架的工作流系统是其最强大的功能之一，允许您通过连接多个步骤来创建复杂的 AI 处理管道。

### 核心工作流概念

- **Flow（流程）**：按顺序执行的步骤序列

### 流程上下文

在工作流执行期间，需要在步骤之间维护上下文。Anyi 提供了 `FlowContext` 结构来在各种工作流步骤之间传递和共享数据。`FlowContext` 包含以下关键属性：

- **Text**：字符串类型，用于存储步骤的输入和输出文本内容。在步骤运行前，此字段是输入文本；步骤运行后，它变成输出文本。
- **Memory**：任意类型（`ShortTermMemory`），用于在步骤之间传递和共享结构化数据。
- **Flow**：对当前流程的引用。
- **ImageURLs**：字符串数组，存储用于多模态内容处理的图像 URL 列表。
- **Think**：字符串类型，存储从模型输出中的 `<think>` 标签提取的内容，用于捕获模型的思考过程而不影响最终输出。

#### 使用短期内存

短期内存允许您在工作流步骤之间传递复杂的结构化数据，而不仅仅是文本。这在需要多步骤处理和状态维护的场景中特别有用。

```go
// 创建带有结构化数据的工作流上下文
type TaskData struct {
    Objective string
    Steps     []string
    Progress  int
}

taskData := TaskData{
    Objective: "创建一个网站",
    Steps:     []string{"设计界面", "开发前端", "开发后端", "测试和部署"},
    Progress:  0,
}

// 在 Memory 中初始化带有结构化数据的上下文
flowContext := anyi.NewFlowContextWithMemory(taskData)

// 您也可以同时设置文本和内存数据
flowContext := anyi.NewFlowContext("初始输入", taskData)

// 在工作流步骤中访问和修改内存数据
func (executor *MyExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    // 访问 Memory 中的数据（需要类型断言）
    taskData := flowContext.Memory.(TaskData)

    // 更新内存数据
    taskData.Progress++
    flowContext.Memory = taskData

    // 更新输出文本
    flowContext.Text = fmt.Sprintf("当前进度: %d/%d", taskData.Progress, len(taskData.Steps))

    return &flowContext, nil
}
```

#### 使用 Think 字段

Anyi 支持特殊的 `<think>` 标签，模型可以在其中表达其思考过程。此内容不会影响最终输出，但会被捕获在 `Think` 字段中。这对于支持显式思考的模型（如 DeepSeek）特别有用，但也可以用于提示其他模型使用此格式。

处理 `<think>` 标签有两种方法：

1. **自动处理**：`Flow.Run` 方法自动检测并提取 `<think>` 标签内的内容到 `FlowContext.Think` 字段，同时从 `Text` 字段中清理标签内容。
2. **使用 DeepSeekStyleResponseFilter**：专门用于处理思考标签的执行器：

```go
// 创建过滤器处理思考标签
thinkFilter := &anyi.DeepSeekStyleResponseFilter{}
err := thinkFilter.Init()
if err != nil {
    log.Fatalf("初始化失败: %v", err)
}

// 配置是否以 JSON 格式输出结果
thinkFilter.OutputJSON = true // 当为 true 时，返回思考和结果内容为 JSON 格式

// 将 DeepSeekStyleResponseFilter 用作执行器
thinkStep := flow.Step{
    Executor: thinkFilter,
}

// 处理后，思考内容存储在 flowContext.Think 中
// 如果 OutputJSON = true，flowContext.Text 将包含 JSON 格式的思考内容和结果
```

## 配置系统

Anyi 的配置系统允许您以集中方式管理客户端、流程和其他设置。这种方法带来了几个好处：

- **代码与配置分离**：将业务逻辑与配置细节分开
- **运行时灵活性**：无需重新编译应用程序即可更改行为
- **环境特定设置**：轻松适应不同环境（开发、测试、生产）
- **集中管理**：在一个地方定义所有 LLM 和工作流配置

### 动态配置

动态配置允许您在运行时以编程方式定义和更新设置。这在以下情况下很有用：

- 您的配置需要根据用户输入动态生成
- 您正在构建需要即时调整行为的系统
- 您想要测试不同的配置而无需重启应用程序

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

func main() {
	// 定义配置
	config := anyi.AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "openai",
				Type: "openai",
				Config: map[string]interface{}{
					"model":  "gpt-4",
					"apiKey": os.Getenv("OPENAI_API_KEY"),
				},
			},
		},
		Flows: []anyi.FlowConfig{
			{
				Name: "content_processor",
				Steps: []anyi.StepConfig{
					{
						Name: "summarize_content",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template":      "将以下内容总结为 3 个要点：\n\n{{.Text}}",
								"systemMessage": "你是一个专业的总结员。",
							},
						},
					},
					{
						Name: "translate_summary",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "将以下摘要翻译成法语：\n\n{{.Text}}",
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
				},
			},
		},
	}

	// 应用配置
	err := anyi.Config(&config)
	if err != nil {
		log.Fatalf("配置失败: %v", err)
	}

	// 获取并运行流程
	flow, err := anyi.GetFlow("content_processor")
	if err != nil {
		log.Fatalf("获取流程失败: %v", err)
	}

	input := "人工智能（AI）是机器展示的智能，与动物（包括人类）展示的自然智能形成对比。AI 研究被定义为对智能代理的研究领域，智能代理是指感知其环境并采取行动最大化实现目标机会的任何系统。"

	result, err := flow.RunWithInput(input)
	if err != nil {
		log.Fatalf("流程执行失败: %v", err)
	}

	log.Printf("结果: %s", result.Text)
}
```

### 配置文件

使用配置文件通常是生产应用程序最实用的方法。Anyi 支持多种文件格式（YAML、JSON、TOML）并提供了加载它们的简便方法。

**使用配置文件的好处：**

- 将敏感信息（如 API 密钥）排除在代码库之外
- 轻松在不同配置之间切换而无需更改代码
- 允许非开发人员修改应用程序行为
- 支持环境特定的配置

```go
package main

import (
	"log"
	"fmt"

	"github.com/jieliu2000/anyi"
)

func main() {
	// 从文件加载配置
	err := anyi.ConfigFromFile("./config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 通过名称访问流程
	flow, err := anyi.GetFlow("content_creator")
	if err != nil {
		log.Fatalf("获取流程失败: %v", err)
	}

	// 运行流程
	result, err := flow.RunWithInput("自动驾驶汽车")
	if err != nil {
		log.Fatalf("流程执行失败: %v", err)
	}

	fmt.Println("生成内容:", result.Text)
}
```

示例 YAML 配置文件（`config.yaml`）：

```yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"

  - name: "anthropic"
    type: "anthropic"
    config:
      model: "claude-3-opus-20240229"
      apiKey: "$ANTHROPIC_API_KEY"

flows:
  - name: "content_creator"
    clientName: "openai"
    steps:
      - name: "research_topic"
        executor:
          type: "llm"
          withconfig:
            template: "研究以下主题并提供关键事实和见解：{{.Text}}"
            systemMessage: "你是一个研究助手。"
        maxRetryTimes: 2

      - name: "draft_article"
        clientName: "anthropic"
        executor:
          type: "llm"
          withconfig:
            template: "使用提供的研究内容写一篇关于这个主题的详细文章：\n\n{{.Text}}"
            systemMessage: "你是一个专业作家。"
        validator:
          type: "string"
          withconfig:
            minLength: 500
```

### 环境变量

Anyi 支持配置中的环境变量，这对以下情况特别有用：

- 密钥管理（API 密钥、令牌）
- 部署特定设置
- CI/CD 管道
- 容器编排环境

在配置文件中使用 `$` 前缀引用环境变量。例如，配置文件中的 `$OPENAI_API_KEY` 将被替换为 `OPENAI_API_KEY` 环境变量的值。

**环境变量最佳实践：**

- 在本地开发中使用 `.env` 文件
- 将敏感信息保存在环境变量中，而不是代码或配置文件中
- 为环境变量使用描述性名称
- 考虑在生产环境中使用密钥管理器

## 内置组件

Anyi 提供了几个内置组件，您可以将其用作 AI 应用程序的构建块。理解这些组件将帮助您利用框架的全部功能。

### 执行器

执行器是 Anyi 工作流系统的主力军。它们在每个步骤中执行实际任务。

#### 内置执行器类型

1. **LLMExecutor**：最常用的执行器，它向 LLM 发送提示并捕获响应。

   - 支持基于模板的提示与变量替换
   - 可以为不同步骤使用不同的系统消息
   - 适用于任何注册的 LLM 客户端

2. **SetContextExecutor**：直接修改流程上下文而无需外部调用。

   - 对初始化变量有用
   - 可以覆盖或追加到现有上下文
   - 通常在工作流开始时使用

3. **ConditionalFlowExecutor**：在工作流中启用分支逻辑。

   - 根据条件路由到不同步骤
   - 可以评估简单表达式
   - 允许复杂的决策树

4. **RunCommandExecutor**：执行 shell 命令并捕获其输出。
   - 弥补 AI 和系统操作之间的差距
   - 对数据处理或外部工具集成有用
   - 允许工作流与操作系统交互

### 验证器

验证器是 Anyi 工作流系统中的重要组件，确保输出在进入下一步之前满足特定标准。它们作为质量控制机制，可以：

- 防止低质量或无效输出在工作流中传播
- 在输出不满足要求时自动触发重试
- 强制执行数据模式和格式要求
- 实现业务规则和逻辑检查

#### 内置验证器类型

1. **StringValidator**：基于各种标准验证文本输出。
   - 长度检查（最小和最大长度）
   - 正则表达式模式匹配
   - 内容验证

```go
   validator := &anyi.StringValidator{
       MinLength: 100,            // 最小长度
       MaxLength: 1000,           // 最大长度
       MatchRegex: `\d{3}-\d{2}`, // 必须包含模式（例如，SSN 格式）
   }
```

2. **JsonValidator**：确保输出是有效的 JSON 并可以根据模式验证。
   - 检查有效的 JSON 语法
   - 可以根据 JSON Schema 验证
   - 对确保结构化数据有用

```go
   validator := &anyi.JsonValidator{
       RequiredFields: []string{"name", "email"},
       Schema: `{"type": "object", "properties": {"name": {"type": "string"}, "email": {"type": "string", "format": "email"}}}`,
   }
```

#### 有效使用验证器

要充分利用验证器：

- 从简单验证开始，逐渐增加复杂性
- 将验证器与重试逻辑结合使用
- 考虑为特定业务规则创建自定义验证器
- 记录验证失败以识别常见问题

## 高级用法

### 多客户端管理

Anyi 允许您同时使用不同的 LLM 提供商，为不同任务选择最合适的模型。

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

func main() {
	// 为复杂任务创建 OpenAI 客户端
	openaiConfig := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	openaiClient, err := anyi.NewClient("gpt", openaiConfig)
	if err != nil {
		log.Fatalf("创建 OpenAI 客户端失败: %v", err)
	}

	// 为简单任务创建 Ollama 本地客户端
	ollamaConfig := ollama.DefaultConfig("llama3")
	ollamaClient, err := anyi.NewClient("local", ollamaConfig)
	if err != nil {
		log.Fatalf("创建 Ollama 客户端失败: %v", err)
	}

	// 使用 OpenAI 客户端进行复杂问题解决
	complexMessages := []chat.Message{
		{Role: "user", Content: "分析人工智能在未来十年对就业市场的潜在影响"},
	}

	complexResponse, _, err := openaiClient.Chat(complexMessages, nil)
	if err != nil {
		log.Fatalf("OpenAI 请求失败: %v", err)
	}

	log.Printf("复杂问题回答（GPT）: %s", complexResponse.Content)

	// 使用本地 Ollama 客户端进行简单计算
	simpleMessages := []chat.Message{
		{Role: "user", Content: "计算 342 + 781 的结果"},
	}

	simpleResponse, _, err := ollamaClient.Chat(simpleMessages, nil)
	if err != nil {
		log.Fatalf("Ollama 请求失败: %v", err)
	}

	log.Printf("简单计算回答（Ollama）: %s", simpleResponse.Content)

	// 在工作流中，您可以根据步骤要求切换客户端
	// 工作流代码...
}
```

### 提示词模板

使用模板化提示增强了 LLM 交互的灵活性和可重用性。Anyi 利用 Go 的模板系统，支持动态变量替换。

#### 在模板中使用 FlowContext 数据

在提示模板中，您可以访问 `FlowContext` 的各种属性：

1. **使用 Text 字段**：使用 `.Text` 直接访问当前上下文文本内容。

```
分析以下文本：{{.Text}}
```

2. **使用 Memory 字段**：访问结构化数据及其内部属性。

```
任务目标：{{.Memory.Objective}}
当前进度：{{.Memory.Progress}}
任务列表：
{{range .Memory.Steps}}
- {{.}}
{{end}}
```

3. **使用 Think 字段**：访问模型的思考过程（如果前一步提取了 `<think>` 标签内容）。

```
前一步的思考过程：{{.Think}}

请继续分析并提供更详细的答案。
```

4. **使用图像 URL**：如果提供了图像 URL，您可以在提示中引用它们。

一个集成内存和思考过程的实际示例：

```go
// 定义结构化数据
type AnalysisData struct {
    Topic        string
    Requirements []string
    Progress     map[string]bool
}

// 创建结构化数据
data := AnalysisData{
    Topic:        "AI 安全",
    Requirements: []string{"当前状态", "关键挑战", "未来趋势"},
    Progress:     map[string]bool{"当前状态": true, "关键挑战": false, "未来趋势": false},
}

// 创建模板文本
templateText := `
分析以下主题：{{.Memory.Topic}}

需要覆盖的要点：
{{range .Memory.Requirements}}
- {{.}}
{{end}}

当前进度：
{{range $key, $value := .Memory.Progress}}
- {{$key}}: {{if $value}}已完成{{else}}未完成{{end}}
{{end}}

{{if .Think}}
前一步的思考过程：
{{.Think}}
{{end}}

请分析尚未完成的要点。
`

// 创建带有内存的上下文
flowContext := anyi.NewFlowContextWithMemory(data)

// 前一步可能有思考内容
flowContext.Think = "<think>我应该关注关键挑战和未来趋势，因为当前状态已经完成</think>"

// 创建模板
formatter, err := chat.NewPromptTemplateFormatter(templateText)
if err != nil {
    log.Fatalf("创建模板失败: %v", err)
}

// 创建带有模板的执行器
executor := &anyi.LLMExecutor{
    TemplateFormatter: formatter,
    SystemMessage:     "你是一个专业的研究分析师",
}

// 创建并运行流程
// ...
```

### 错误处理

在与 LLM 交互的应用程序中，健壮的错误处理至关重要。以下是在 Anyi 中实现有效错误处理的一些模式：

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 创建客户端
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 准备消息
	messages := []chat.Message{
		{Role: "user", Content: "解释量子力学的基本原理"},
	}

	// 实现重试逻辑
	maxRetries := 3
	backoff := 1 * time.Second

	var response *chat.Message
	var info chat.ResponseInfo

	for i := 0; i < maxRetries; i++ {
		response, info, err = client.Chat(messages, nil)

		if err == nil {
			// 成功获得响应，跳出循环
			break
		}

		// 检查错误是否可重试（如网络错误、超时等）
		if i < maxRetries-1 {
			log.Printf("尝试 %d 失败: %v，%v 后重试", i+1, err, backoff)
			time.Sleep(backoff)
			backoff *= 2 // 指数退避
		}
	}

	if err != nil {
		log.Fatalf("经过 %d 次尝试仍然失败: %v", maxRetries, err)
	}

	// 处理成功响应
	log.Printf("响应: %s", response.Content)
	log.Printf("使用的 token: %d", info.PromptTokens + info.CompletionTokens)

	// 错误日志记录和监控
	// 在实际应用程序中，您应该实现更复杂的错误日志记录和监控
	// 例如，将错误发送到日志管理系统或监控服务
}
```

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

### 1. 如何确保工作流在不稳定的网络条件下正常工作？

Anyi 内置了重试机制。您可以为每个步骤设置 `MaxRetryTimes` 属性来实现自动重试：

```go
// 设置最大重试次数
step1.MaxRetryTimes = 3
```

### 2. 如何处理超过 token 限制的大型文本？

```go
// 实现分块文本处理
func processLargeText(text string, client *llm.Client) (string, error) {
    // 将文本分割成较小的块
    chunks := splitIntoChunks(text, 1000) // 每块约 1000 个单词

    var results []string
    // 处理每个块
    for _, chunk := range chunks {
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: "处理以下文本: " + chunk},
        }, nil)
        if err != nil {
            return "", err
        }
        results = append(results, response.Content)
    }

    // 合并结果
    return combineResults(results), nil
}
```

### 3. 如何将 Anyi 与现有的 Go Web 框架集成？

Anyi 可以与任何 Go Web 框架（如 Gin、Echo 或 Fiber）无缝集成。以下是与 Gin 的示例：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jieliu2000/anyi"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 初始化 Anyi 客户端
    // ...

    r.POST("/ask", func(c *gin.Context) {
        var req struct {
            Question string `json:"question"`
        }
        if err := c.BindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // 使用 Anyi 客户端处理请求
        response, _, err := client.Chat([]chat.Message{
            {Role: "user", Content: req.Question},
        }, nil)

        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"answer": response.Content})
    })

    return r
}
```

## 结论

Anyi 提供了构建 AI 代理和工作流的强大框架。通过结合不同的 LLM 提供商、工作流步骤和验证技术，您可以创建与现有系统无缝集成的复杂 AI 应用程序。

有关更多示例和最新文档，请访问 [GitHub 仓库](https://github.com/jieliu2000/anyi)。

### 系统要求

- Go 1.20 或更高版本
- 网络连接（用于访问 LLM API）
- 适用于所有主要操作系统（Linux、macOS、Windows）

### 获取帮助和贡献

如果您遇到问题或有疑问，请考虑：

- 在 GitHub 上开启 issue
- 加入社区讨论
- 阅读 API 文档
- 向项目贡献改进

Anyi 框架持续发展，您的反馈有助于让它对每个人都更好。
