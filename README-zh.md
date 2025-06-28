# Anyi(安易) - 开源的自主式 AI 智能体框架

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://go.dev/)

| [English](README.md) | [中文](README-zh.md) |

Anyi(安易)是一个强大的 AI 智能体框架，通过提供统一的大语言模型接口、健壮的验证机制和灵活的工作流系统，帮助你构建能够与实际工作场景无缝集成的 AI 解决方案。

## 核心特性

- **统一的大语言模型访问** - 通过一致的 API 连接多种 LLM 提供商（智谱 AI、阿里云灵积、OpenAI 等）
- **强大的工作流系统** - 将步骤链接起来，配合验证和自动重试，构建可靠的 AI 流程
- **配置驱动开发** - 通过代码或外部配置文件（YAML、JSON、TOML）定义工作流和客户端
- **多模态支持** - 向兼容的模型同时发送文本和图像
- **Go 模板集成** - 使用 Go 的模板引擎生成动态提示词

## 何时使用 Anyi

Anyi 特别适合以下场景：

- **AI 应用开发** - 构建具有可靠错误处理和重试机制的生产级 AI 服务
- **多模型应用** - 创建能够根据成本或能力利用不同模型处理不同任务的解决方案
- **DevOps 集成** - 通过命令执行器和 API 集成将 AI 能力与现有系统连接
- **快速原型开发** - 通过配置文件配置复杂 AI 工作流，无需修改代码
- **企业级解决方案** - 保持代码和配置分离，便于在不同环境中部署

## 支持的 LLM 提供商

- **DeepSeek** - DeepSeek Chat 和 DeepSeek Coder 等模型
- **阿里云灵积** - 通义千问系列模型
- **Ollama** - 本地部署开源模型（如 Llama、Qwen 等）
- **OpenAI** - GPT 系列模型
- **Azure OpenAI** - 微软托管的 OpenAI 模型
- **Anthropic** - Claude 系列模型（包括 Claude 3 Opus、Sonnet 和 Haiku）
- **智谱 AI** - GLM 系列模型
- **SiliconCloud** - SiliconFlow 模型

## 快速开始

### 安装

```bash
go get -u github.com/jieliu2000/anyi
```

> 需要 Go 1.20 或更高版本

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

## 创建工作流

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

Anyi 支持配置驱动开发，允许你在外部文件中定义 LLM 客户端和工作流：

```yaml
# config.yaml
clients:
  - name: "ollama"
    type: "ollama"
    config:
      model: "llama3"
      ollamaApiURL: "http://localhost:11434/api" # 本地Ollama服务

  - name: "qwen"
    type: "dashscope"
    config:
      model: "qwen-max"
      apiKey: "$DASHSCOPE_API_KEY" # 引用环境变量

flows:
  - name: "故事流程"
    clientName: "ollama" # 工作流默认客户端
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
        clientName: "qwen" # 为此步骤指定不同的客户端
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

## 内置组件

### 执行器

- **LLMExecutor** - 向大语言模型发送提示词
- **SetContextExecutor** - 修改工作流上下文
- **ConditionalFlowExecutor** - 基于条件进行分支
- **RunCommandExecutor** - 执行系统命令
- **MCPExecutor** - 与模型控制协议接口，访问外部模型、资源和工具

### 验证器

- **StringValidator** - 通过正则表达式或相等性检查文本
- **JsonValidator** - 确保输出是有效的 JSON

## 文档

### 入门指南

- [安装和设置](docs/zh/getting-started/installation.md) - 系统要求和安装
- [快速开始指南](docs/zh/getting-started/quickstart.md) - 您的第一个 Anyi 应用
- [基本概念](docs/zh/getting-started/concepts.md) - 理解客户端、流程和执行器

### 教程

- [使用 LLM 客户端](docs/zh/tutorials/llm-clients.md) - 所有支持提供商的完整指南
- [构建工作流](docs/zh/tutorials/workflows.md) - 创建复杂的 AI 工作流
- [配置管理](docs/zh/tutorials/configuration.md) - 使用配置文件和环境变量
- [多模态应用](docs/zh/tutorials/multimodal.md) - 处理文本和图像

### 操作指南

- [提供商设置](docs/zh/how-to/provider-setup.md) - 每个 LLM 提供商的详细设置
- [错误处理](docs/zh/how-to/error-handling.md) - 构建健壮应用的最佳实践
- [性能优化](docs/zh/how-to/performance.md) - 速度和成本优化
- [Web 集成](docs/zh/how-to/web-integration.md) - 在 Web 框架中使用 Anyi

### 参考资料

- [API 参考](docs/zh/reference/api.md) - 完整的 API 文档
- [配置模式](docs/zh/reference/configuration.md) - 所有配置选项
- [内置组件](docs/zh/reference/components.md) - 执行器和验证器参考

### 高级主题

- [自定义执行器](docs/zh/advanced/custom-executors.md) - 构建您自己的执行器
- [安全最佳实践](docs/zh/advanced/security.md) - 保护您的 AI 应用
- [生产部署](docs/zh/advanced/deployment.md) - 生产环境考虑

## 贡献

欢迎贡献！Anyi 正在积极开发中，您的反馈有助于让它变得更好。

## 许可证

Anyi 使用 [Apache License 2.0](LICENSE) 许可证。
