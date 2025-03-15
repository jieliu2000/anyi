# Anyi(安易) 编程指南和示例

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## 目录

- [简介](#简介)
- [安装](#安装)
- [大语言模型访问](#大语言模型访问)
  - [客户端创建方式](#客户端创建方式)
  - [客户端配置](#客户端配置)
    - [智谱AI](#智谱ai)
    - [阿里云灵积](#阿里云灵积)
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
  - [格式化器](#格式化器)
- [高级用法](#高级用法)
  - [多客户端管理](#多客户端管理)
  - [提示词模板](#提示词模板)
  - [错误处理](#错误处理)
- [最佳实践](#最佳实践)
  - [性能优化](#性能优化)
  - [成本管理](#成本管理)
  - [安全考虑](#安全考虑)

## 简介

Anyi(安易)是一个开源的Go语言自主式AI智能体框架，旨在帮助开发者构建与实际工作场景无缝集成的AI解决方案。本指南提供详细的编程说明和示例，帮助您有效地使用Anyi框架。

### Anyi的核心特性

- **统一的LLM接口**：通过一致的API访问多种LLM提供商
- **灵活的工作流系统**：构建复杂的多步骤AI处理流程，支持验证和错误处理
- **配置管理**：支持YAML、JSON和TOML格式的配置文件
- **内置组件**：提供即用型执行器和验证器用于常见任务
- **可扩展架构**：创建自定义组件以满足特定需求

### 何时使用Anyi

Anyi在以下场景特别有用：
- 需要编排多个AI模型之间的复杂交互
- 希望构建具有验证和错误处理功能的可靠AI工作流
- 需要在不更改代码的情况下切换不同的LLM提供商
- 使用Go语言构建生产级AI应用程序

## 安装

通过Go模块安装Anyi：

```bash
go get -u github.com/jieliu2000/anyi
```

Anyi需要Go 1.20或更高版本。

### 验证安装

您可以通过创建一个简单的程序来验证Anyi安装是否成功：

```go
package main

import (
	"fmt"
	"github.com/jieliu2000/anyi"
)

func main() {
	fmt.Println("Anyi版本:", anyi.Version)
}
```

## 大语言模型访问

Anyi提供了统一的方式与各种大语言模型（LLM）进行交互，通过一致的接口实现。这种方法使您能够在不更改应用程序逻辑的情况下轻松切换不同的提供商。

### 理解Anyi的客户端架构

在深入代码之前，了解Anyi如何组织LLM访问非常重要：

1. **提供商**：每个LLM服务（智谱AI、阿里云灵积等）都有专用的提供商模块
2. **客户端**：处理与特定LLM服务通信的实例
3. **注册表**：全局存储命名客户端，方便在应用程序中检索

### 客户端创建方式

Anyi提供统一的接口访问各种大语言模型。创建客户端有两种主要方法：

1. 使用`anyi.NewClient()`- 创建一个注册到全局注册表的命名客户端
2. 使用`llm.NewClient()`- 创建一个由你自己管理的未注册客户端实例

#### 何时使用命名客户端与未注册客户端

- **命名客户端**适合于以下场景：
  - 需要从应用程序的不同部分访问同一个客户端实例
  - 希望配置一次并在整个代码库中重复使用
  - 管理具有不同配置的多个客户端

- **未注册客户端**适合于以下场景：
  - 需要用于特定任务的隔离客户端实例
  - 想要避免潜在的命名冲突
  - 应用程序结构简单，与LLM交互有限

#### 命名客户端示例

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
	// 创建一个名为"glm4"的客户端
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")
	
	client, err := anyi.NewClient("glm4", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	
	// 之后可以通过名称检索此客户端
	retrievedClient, err := anyi.GetClient("glm4")
	if err != nil {
		log.Fatalf("检索客户端失败: %v", err)
	}
	
	// 使用客户端
	messages := []chat.Message{
		{Role: "user", Content: "中国有多少个省份？"},
	}
	response, _, err := retrievedClient.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}
	
	log.Printf("回答: %s", response.Content)
}
```

#### 未注册客户端示例

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 创建一个不注册的客户端
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")
	
	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	
	// 直接使用客户端
	messages := []chat.Message{
		{Role: "user", Content: "用简单的语言解释量子计算"},
	}
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}
	
	log.Printf("回答: %s", response.Content)
}
```

### 客户端配置

每个LLM提供商都有自己的配置结构。了解每个提供商的特定配置选项对于优化与不同模型的交互至关重要。

#### 配置最佳实践

- 将API密钥存储在环境变量中，而不是硬编码在代码中
- 使用提供商特定的默认配置作为起点
- 为生产环境考虑设置自定义超时
- 对于自托管模型或代理服务，使用自定义基础URL

#### 智谱AI

智谱AI提供了GLM系列的大语言模型，包括GLM-4、GLM-3-Turbo等。

```go
// 默认配置
config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")

// 使用GLM-3-Turbo模型
config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-3-turbo")

// 自定义配置
config := zhipu.NewConfig(
    os.Getenv("ZHIPU_API_KEY"),
    "glm-4-flash",
    "https://api.bigmodel.cn"
)
```

#### 阿里云灵积

阿里云灵积模型服务提供了通义千问等一系列大语言模型。

```go
// 默认配置
config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")

// 使用千问Turbo模型
config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")

// 自定义基础URL配置
config := dashscope.NewConfig(
    os.Getenv("DASHSCOPE_API_KEY"),
    "qwen-max",
    "https://your-proxy-url.com/compatible-mode/v1"
)
```

#### Ollama

Ollama提供了在本地部署开源模型的能力，支持Llama、Qwen等多种模型。

```go
// 默认配置（本地服务器）
config := ollama.DefaultConfig("qwen2:7b")

// 自定义服务器配置
config := ollama.NewConfig("llama3", "http://your-ollama-server:11434")
```

#### 其他提供商

Anyi还支持其他许多LLM提供商，包括：

- **DeepSeek**: `deepseek.DefaultConfig()`
- **百川AI**: `baichuan.DefaultConfig()`
- **OpenAI**: `openai.DefaultConfig()`
- **Azure OpenAI**: `azureopenai.NewConfig()`
- **SiliconCloud**: `siliconcloud.DefaultConfig()`

## 聊天API使用

Anyi的核心功能是通过Chat API与大语言模型交互。本节解释如何构建对话、处理响应以及自定义聊天行为。

### 了解聊天生命周期

与LLM的典型聊天交互遵循以下步骤：
1. **准备消息**：创建表示对话的消息序列
2. **配置选项**：设置温度、最大令牌数等参数
3. **发送请求**：在客户端上调用Chat方法
4. **处理响应**：处理模型的回复和任何元数据
5. **继续对话**：将响应添加到消息历史记录中以进行后续交流

### 消息结构

Anyi中的聊天消息使用`chat.Message`结构：

```go
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string // 消息的文本内容
	Name    string // 可选的名称（用于多智能体场景）
	
	// 用于多模态内容
	ContentParts []ContentPart
}
```

### 返回值详解

调用Chat方法时，您会收到三个值：
1. **响应消息**：模型的回复，作为`chat.Message`
2. **响应信息**：关于响应的元数据（使用的令牌数、模型名称等）
3. **错误**：请求过程中可能发生的任何错误

理解这些返回值有助于您实现适当的错误处理和日志记录。

### 基本聊天示例

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
	// 创建客户端
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")
	client, err := anyi.NewClient("qwen", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 创建消息历史
	messages := []chat.Message{
		{Role: "system", Content: "你是一个乐于助人的助手。"},
		{Role: "user", Content: "机器学习可以应用在哪些领域？"},
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
		Content: "能否给出在医疗领域的具体例子？",
	})
	
	// 发送后续问题
	response, _, err = client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}
	
	log.Printf("后续回答: %s", response.Content)
}
```

### 聊天选项

你可以使用`chat.ChatOptions`自定义聊天行为：

```go
options := &chat.ChatOptions{
	Temperature: 0.7,               // 控制随机性(0.0-2.0)
	MaxTokens:   1000,              // 最大响应长度
	TopP:        0.9,               // 核采样参数
	Stream:      true,              // 启用流式响应
	Stop:        []string{"停止"},   // 自定义停止序列
}

response, info, err := client.Chat(messages, options)
```

## 多模态模型使用

许多现代大语言模型支持多模态输入，允许你发送图片和文本。

### 发送图片

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
	// 创建支持多模态的千问VL客户端
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-vl-plus")
	client, err := anyi.NewClient("qwen-vision", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 创建带图片URL的消息
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "这张图片里有什么？",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg",
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

许多大语言模型支持函数调用功能，允许AI模型请求执行特定操作。

### 函数定义

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/zhipu"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

func main() {
	// 创建一个客户端
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")
	client, err := anyi.NewClient("glm4", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 定义函数
	functions := []tools.FunctionConfig{
		{
			Name:        "get_weather",
			Description: "获取指定位置的天气信息",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "城市名称，例如'北京'",
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
		{Role: "user", Content: "北京今天的天气怎么样？"},
	}
	
	// 请求函数调用
	response, _, err := client.ChatWithFunctions(messages, functions, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}
	
	log.Printf("响应类型: %s", response.FunctionCall.Name)
	log.Printf("参数: %s", response.FunctionCall.Arguments)
	
	// 在这里你应该处理函数调用，执行请求的函数
	// 并在另一条消息中发送结果
}
```

## 工作流系统

Anyi的工作流系统是其最强大的特性之一，允许您通过连接多个步骤创建复杂的AI处理流程。

### 工作流核心概念

- **流程(Flow)**：按顺序执行的步骤序列
- **步骤(Step)**：带有执行器和可选验证器的单个工作单元
- **执行器(Executor)**：执行实际工作（例如，调用LLM，设置上下文）
- **验证器(Validator)**：确保输出在继续下一步之前满足要求
- **上下文(Context)**：在步骤之间传递的共享数据

### 何时使用工作流

工作流在以下场景特别有用：
- 多步骤推理过程
- 内容生成流程
- 数据转换和丰富
- 带有条件逻辑的决策树
- 需要验证和重试的任务

### 创建工作流

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/zhipu"
	"github.com/jieliu2000/anyi/flow"
)

func main() {
	// 创建一个客户端
	config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4")
	client, err := anyi.NewClient("glm4", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	
	// 创建单个步骤
	step1, err := anyi.NewLLMStepWithTemplate(
		"以{{.Text}}为主题，生成一个短篇故事",
		"你是一位富有创造力的小说家。",
		client,
	)
	if err != nil {
		log.Fatalf("创建步骤失败: %v", err)
	}
	step1.Name = "故事生成"
	
	step2, err := anyi.NewLLMStepWithTemplate(
		"为以下故事创建一个吸引人的标题：\n\n{{.Text}}",
		"你是一位擅长创作标题的编辑。",
		client,
	)
	if err != nil {
		log.Fatalf("创建步骤失败: %v", err)
	}
	step2.Name = "标题创作"
	
	// 创建工作流
	myFlow, err := anyi.NewFlow("故事流程", client, *step1, *step2)
	if err != nil {
		log.Fatalf("创建工作流失败: %v", err)
	}
	
	// 注册工作流
	err = anyi.RegisterFlow("故事流程", myFlow)
	if err != nil {
		log.Fatalf("注册工作流失败: %v", err)
	}
	
	// 运行工作流
	result, err := myFlow.RunWithInput("未来上海的一位侦探")
	if err != nil {
		log.Fatalf("工作流执行失败: %v", err)
	}
	
	log.Printf("标题: %s", result.Text)
}
```

### 步骤和执行器详解

工作流中的每个步骤都使用执行器来执行其任务。Anyi提供了几种内置执行器：

1. **LLMExecutor**：最常用的执行器，向LLM发送带有模板的提示并捕获响应
2. **SetContextExecutor**：直接修改工作流上下文
3. **ConditionalFlowExecutor**：基于条件引导流程
4. **RunCommandExecutor**：执行系统命令并捕获其输出

步骤可以链接在一起，一个步骤的输出成为下一个步骤的输入。

## 配置系统

Anyi的配置系统允许您以集中方式管理客户端、工作流和其他设置。这种方法带来了几个好处：

- **代码和配置分离**：保持业务逻辑与配置细节的分离
- **运行时灵活性**：无需重新编译应用程序即可更改行为
- **环境特定设置**：轻松适应不同环境（开发、测试、生产）
- **集中管理**：在一个地方定义所有LLM和工作流配置

### 动态配置

动态配置允许您以编程方式定义和更新运行时设置。这在以下情况下特别有用：
- 您的配置需要根据用户输入动态生成
- 您正在构建需要动态调整行为的系统
- 您想在不重启应用程序的情况下测试不同配置

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
				Name: "dashscope",
				Type: "dashscope",
				Config: map[string]interface{}{
					"model":  "qwen-max",
					"apiKey": os.Getenv("DASHSCOPE_API_KEY"),
				},
			},
		},
		Flows: []anyi.FlowConfig{
			{
				Name: "内容处理器",
				Steps: []anyi.StepConfig{
					{
						Name: "内容摘要",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template":      "用3个要点总结以下内容：\n\n{{.Text}}",
								"systemMessage": "你是一个专业的内容摘要专家。",
							},
						},
					},
					{
						Name: "摘要翻译",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "将以下摘要翻译成英文：\n\n{{.Text}}",
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

	// 获取并运行工作流
	flow, err := anyi.GetFlow("内容处理器")
	if err != nil {
		log.Fatalf("获取工作流失败: %v", err)
	}
	
	input := "人工智能（AI）是由机器展示的智能，与人类和动物展示的自然智能相对。人工智能研究被定义为智能体的研究领域，指的是任何能够感知其环境并采取行动以最大化其实现目标机会的系统。"人工智能"一词此前被用来描述模仿和展示与人类思维相关的"人类"认知技能的机器，如"学习"和"解决问题"。这个定义已经被主要的AI研究人员拒绝，他们现在以理性和理性行动的角度描述AI，这并不限制智能可以如何表达。"
	
	result, err := flow.RunWithInput(input)
	if err != nil {
		log.Fatalf("工作流执行失败: %v", err)
	}
	
	log.Printf("结果: %s", result.Text)
}
```

### 配置文件

对于生产应用程序，使用配置文件通常是最实用的方法。Anyi支持多种文件格式（YAML、JSON、TOML）并提供了一种加载它们的简便方法。

**使用配置文件的好处：**
- 将敏感信息（如API密钥）保持在代码库之外
- 无需更改代码即可轻松切换不同配置
- 允许非开发人员修改应用程序行为
- 支持特定环境的配置

```go
package main

import (
	"log"
	"fmt"

	"github.com/jieliu2000/anyi"
)

func main() {
	// 从文件加载配置
	err := anyi.ConfigFromFile("./config/workflow.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 通过名称访问工作流
	flow, err := anyi.GetFlow("内容创作")
	if err != nil {
		log.Fatalf("获取工作流失败: %v", err)
	}
	
	// 运行工作流
	result, err := flow.RunWithInput("自动驾驶汽车")
	if err != nil {
		log.Fatalf("工作流执行失败: %v", err)
	}
	
	fmt.Println("生成的内容:", result.Text)
}
```

配置文件示例（`config/workflow.yaml`）：

```yaml
clients:
  - name: "qwen"
    type: "dashscope"
    config:
      model: "qwen-max"
      apiKey: "$DASHSCOPE_API_KEY"
  
  - name: "glm"
    type: "zhipu"
    config:
      model: "glm-4"
      apiKey: "$ZHIPU_API_KEY"

flows:
  - name: "内容创作"
    clientName: "qwen"
    steps:
      - name: "研究主题"
        executor:
          type: "llm"
          withconfig:
            template: "研究以下主题并提供关键事实和见解: {{.Text}}"
            systemMessage: "你是一个研究助手。"
        maxRetryTimes: 2
      
      - name: "撰写文章"
        clientName: "glm"
        executor:
          type: "llm"
          withconfig:
            template: "使用提供的研究内容，撰写一篇关于此主题的详细文章：\n\n{{.Text}}"
            systemMessage: "你是一位专业作家。"
        validator:
          type: "string"
          withconfig:
            minLength: 500
```

### 环境变量

Anyi支持在配置中使用环境变量，这对于以下方面特别有用：
- 密钥管理（API密钥、令牌）
- 部署特定设置
- CI/CD流程
- 容器编排环境

在配置文件中使用`$`前缀引用环境变量。例如，配置文件中的`$ZHIPU_API_KEY`将被替换为`ZHIPU_API_KEY`环境变量的值。

**环境变量的最佳实践：**
- 使用`.env`文件进行本地开发
- 将敏感信息保存在环境变量中，而不是代码或配置文件中
- 为您的环境变量使用描述性名称
- 考虑在生产环境中使用密钥管理器

## 内置组件

Anyi提供了几种内置组件，您可以将其用作AI应用程序的构建块。了解这些组件将帮助您充分利用框架的全部功能。

### 执行器

执行器是Anyi工作流系统的核心组件，它们在每个步骤中执行实际任务。

#### 内置执行器类型

1. **LLMExecutor**：最常用的执行器，向LLM发送带有模板的提示并捕获响应。
   - 支持带变量替换的模板化提示
   - 可以为不同步骤使用不同的系统消息
   - 可与任何注册的LLM客户端配合使用

2. **SetContextExecutor**：直接修改工作流上下文，无需外部调用。
   - 用于初始化变量
   - 可以覆盖或追加到现有上下文
   - 通常在工作流开始时使用

3. **ConditionalFlowExecutor**：启用工作流中的分支逻辑。
   - 基于条件路由到不同步骤
   - 可以评估简单表达式
   - 允许复杂的决策树

4. **RunCommandExecutor**：执行系统命令并捕获其输出。
   - 连接AI与系统操作
   - 用于数据处理或外部工具集成
   - 允许工作流与操作系统交互

### 验证器

验证器是Anyi工作流系统中的关键组件，确保输出在继续下一步之前满足特定标准。它们充当质量控制机制，可以：
- 防止低质量或无效输出在工作流中传播
- 当输出不符合要求时自动触发重试
- 强制执行数据架构和格式要求
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
       MatchRegex: `\d{3}-\d{2}`, // 必须包含模式（例如，身份证格式）
   }
   ```

2. **JsonValidator**：确保输出是有效的JSON，并可以根据模式进行验证。
   - 检查有效的JSON语法
   - 可以根据JSON模式进行验证
   - 对确保结构化数据很有用

   ```go
   validator := &anyi.JsonValidator{
       RequiredFields: []string{"name", "email"},
       Schema: `{"type": "object", "properties": {"name": {"type": "string"}, "email": {"type": "string", "format": "email"}}}`,
   }
   ```

#### 有效使用验证器

要充分利用验证器：
- 从更简单的验证开始，然后逐渐增加复杂性
- 将验证器与重试逻辑结合使用
- 考虑为特定业务规则创建自定义验证器
- 记录验证失败以识别常见问题

### 格式化器

Anyi包含格式化器，帮助在工作流中处理和转换文本。格式化器可以：
- 标准化输出格式
- 提取特定信息
- 在不同表示之间转换数据
- 应用一致的样式和格式

## 最佳实践

构建有效的AI应用程序不仅需要技术知识。以下是帮助您充分利用Anyi框架的全面最佳实践。

### 性能优化

优化Anyi工作流的性能可以显著改善用户体验并降低成本：

**1. 为任务选择合适的模型**
- 对于简单任务使用更小、更快的模型
- 为复杂推理保留更强大的模型
- 考虑针对专业领域使用微调模型

**2. 实现缓存**
- 缓存常见的LLM响应以避免冗余API调用
- 对多实例部署使用分布式缓存
- 设置适当的缓存过期时间

**3. 优化提示词**
- 保持提示简洁同时包含必要上下文
- 使用清晰的指令减少来回交互
- 测试和迭代提示词以最小化令牌使用

**4. 本地部署选项**
- 对于频繁、非关键任务，使用Ollama和本地模型
- 根据需求在云端和本地模型之间平衡
- 考虑在资源受限环境中使用量化模型

**5. 并行执行**
- 识别可以并行运行的工作流步骤
- 在适当情况下使用goroutines进行并发LLM调用
- 为并行步骤实现适当的错误处理

### 成本管理

使用商业LLM提供商时，管理成本至关重要：

**1. 令牌监控**
- 实现令牌计数以跟踪使用情况
- 为异常支出模式设置警报
- 定期审计您的令牌消耗

**2. 分层模型策略**
- 使用级联方法：先尝试更便宜的模型
- 只在必要时升级到更昂贵的模型
- 为服务中断实现回退机制

**3. 响应长度控制**
- 为每个用例设置适当的最大令牌限制
- 使用验证确保输出不是不必要的冗长
- 对过长输出实施截断策略

**4. 批处理请求**
- 在可能的情况下合并多个小请求
- 对非紧急处理实现队列系统
- 在非高峰时段安排批处理

**5. 成本归因**
- 按工作流、功能或用户跟踪成本
- 实施每用户配额或速率限制
- 考虑将高级功能的成本转嫁给最终用户

### 安全考虑

在构建AI系统时，安全至关重要：

**1. API密钥管理**
- 绝不在应用程序中硬编码API密钥
- 使用环境变量或秘密管理器
- 定期轮换密钥并限制其权限

**2. 输入净化**
- 验证和净化所有用户输入
- 实施速率限制以防止滥用
- 使用上下文过滤防止提示注入

**3. 输出验证**
- 在使用LLM输出之前始终进行验证
- 在可执行上下文中使用LLM输出时要谨慎
- 对面向用户的输出实施内容审核

**4. 数据隐私**
- 最小化向LLM发送敏感数据
- 实施数据保留策略
- 考虑使用本地模型处理敏感信息

**5. 审计和日志记录**
- 维护所有LLM交互的详细日志
- 对敏感内容实施适当的日志编辑
- 设置监控系统检测异常模式或安全事件

通过遵循这些最佳实践，您可以构建不仅强大而且高效、经济且安全的AI应用程序。

## 结论

Anyi提供了一个强大的框架，用于构建AI智能体和工作流。通过组合不同的大语言模型提供商、工作流步骤和验证技术，您可以创建与现有系统集成的复杂AI应用程序。

有关更多示例和最新文档，请访问[GitHub仓库](https://github.com/jieliu2000/anyi)。

### 获取帮助和贡献

如果您遇到问题或有疑问，请考虑：
- 在GitHub上开启一个issue
- 加入社区讨论
- 阅读API文档
- 向项目贡献改进

Anyi框架在不断发展，您的反馈有助于使它对每个人都更好。
