# Anyi(安易) - 开源的自主式 AI 智能体框架 

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://go.dev/)

| [English](README.md) | [中文](README-zh.md) |

Anyi(安易)是一个强大的AI智能体框架，通过提供统一的大语言模型接口、健壮的验证机制和灵活的工作流系统，帮助你构建能够与实际工作场景无缝集成的AI解决方案。

> 📚 **寻找详细教程？** 查阅我们全面的[Anyi编程指南和示例](/docs/zh/tutorial.md)

## ✨ 核心特性

- **统一的大语言模型访问** - 通过一致的API连接多种LLM提供商（智谱AI、阿里云灵积、OpenAI等）
- **强大的工作流系统** - 将步骤链接起来，配合验证和自动重试，构建可靠的AI流程
- **配置驱动开发** - 通过代码或外部配置文件（YAML、JSON、TOML）定义工作流和客户端
- **多模态支持** - 向兼容的模型同时发送文本和图像
- **Go模板集成** - 使用Go的模板引擎生成动态提示词

## 🤔 何时使用Anyi

Anyi特别适合以下场景：

- **AI应用开发** - 构建具有可靠错误处理和重试机制的生产级AI服务
- **多模型应用** - 创建能够根据成本或能力利用不同模型处理不同任务的解决方案
- **DevOps集成** - 通过命令执行器和API集成将AI能力与现有系统连接
- **快速原型开发** - 通过配置文件配置复杂AI工作流，无需修改代码
- **企业级解决方案** - 保持代码和配置分离，便于在不同环境中部署

## 📋 支持的LLM提供商

- **DeepSeek** - DeepSeek Chat和DeepSeek Coder等模型
- **阿里云灵积** - 通义千问系列模型
- **Ollama** - 本地部署开源模型（如Llama、Qwen等）
- **OpenAI** - GPT系列模型
- **Azure OpenAI** - 微软托管的OpenAI模型
- **Anthropic** - Claude系列模型（包括Claude 3 Opus、Sonnet和Haiku）
- **智谱AI** - GLM系列模型
- **SiliconCloud** - SiliconFlow模型

## 🚀 快速开始

### 安装

```bash
go get -u github.com/jieliu2000/anyi
```

> ⚠️ 需要Go 1.20或更高版本

### 基本用法

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/deepseek"  // 导入你偏好的提供商
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 创建客户端 - 只需更改导入和配置即可使用不同的提供商
	config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")
	
	client, err := anyi.NewClient("deepseek", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 发送聊天请求
	messages := []chat.Message{
		{Role: "user", Content: "中国有多少个省份？"},
	}
	
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}
	
	log.Printf("回答: %s", response.Content)
}
```

## 🔄 创建工作流

### 使用代码

```go
package main

import (
	"log"
	"os"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/dashscope"
)

func main() {
	// 创建客户端
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-max")
	client, err := anyi.NewClient("qwen", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	
	// 创建两步工作流
	step1, _ := anyi.NewLLMStepWithTemplate(
		"以{{.Text}}为主题，生成一个短篇故事",
		"你是一位富有创造力的小说家。",
		client,
	)
	step1.Name = "故事生成"
	
	step2, _ := anyi.NewLLMStepWithTemplate(
		"为以下故事创建一个吸引人的标题：\n\n{{.Text}}",
		"你是一位擅长创作标题的编辑。",
		client,
	)
	step2.Name = "标题创作"
	
	// 创建并注册工作流
	myFlow, _ := anyi.NewFlow("故事流程", client, *step1, *step2)
	anyi.RegisterFlow("故事流程", myFlow)
	
	// 运行工作流
	result, _ := myFlow.RunWithInput("未来上海的一位侦探")
	
	log.Printf("标题: %s", result.Text)
}
```

### 使用配置文件

Anyi支持配置驱动开发，允许你在外部文件中定义LLM客户端和工作流：

```yaml
# config.yaml
clients:
  - name: "ollama"
    type: "ollama"
    config:
      model: "llama3"
      ollamaApiURL: "http://localhost:11434/api"  # 本地Ollama服务
  
  - name: "qwen"
    type: "dashscope"
    config:
      model: "qwen-max"
      apiKey: "$DASHSCOPE_API_KEY"  # 引用环境变量

flows:
  - name: "故事流程"
    clientName: "ollama"  # 工作流默认客户端
    steps:
      - name: "故事生成"
        executor:
          type: "llm"
          withconfig:
            template: "以{{.Text}}为主题，生成一个短篇故事"
            systemMessage: "你是一位富有创造力的小说家。"
        maxRetryTimes: 2
      
      - name: "标题创作"
        executor:
          type: "llm"
          withconfig:
            template: "为以下故事创建一个吸引人的标题：\n\n{{.Text}}"
            systemMessage: "你是一位擅长创作标题的编辑。"
        clientName: "qwen"  # 为此步骤指定不同的客户端
```

加载并使用此配置：

```go
package main

import (
	"log"
	"github.com/jieliu2000/anyi"
)

func main() {
	// 从文件加载配置
	err := anyi.ConfigFromFile("./config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 获取并运行配置好的工作流
	flow, err := anyi.GetFlow("故事流程")
	if err != nil {
		log.Fatalf("获取工作流失败: %v", err)
	}
	
	result, err := flow.RunWithInput("未来上海的一位侦探")
	if err != nil {
		log.Fatalf("工作流执行失败: %v", err)
	}
	
	log.Printf("结果: %s", result.Text)
}
```

## 🛠️ 内置组件

### 执行器

- **LLMExecutor** - 向大语言模型发送提示词
- **SetContextExecutor** - 修改工作流上下文
- **ConditionalFlowExecutor** - 基于条件进行分支
- **RunCommandExecutor** - 执行系统命令

### 验证器

- **StringValidator** - 通过正则表达式或相等性检查文本
- **JsonValidator** - 确保输出是有效的JSON

## 📖 文档

有关全面指南和详细示例，请查看我们的[编程指南](/docs/zh/tutorial.md)。

涵盖的主题包括：
- [LLM客户端配置](/docs/zh/tutorial.md#客户端配置)
- [工作流创建](/docs/zh/tutorial.md#工作流系统)
- [使用配置文件](/docs/zh/tutorial.md#配置文件)
- [最佳实践](/docs/zh/tutorial.md#最佳实践)

## 🤝 贡献

欢迎贡献！Anyi正在积极开发中，您的反馈有助于使它对每个人都更好。

## 📄 许可证

Anyi 遵循 [Apache License 2.0](LICENSE) 开源许可。
