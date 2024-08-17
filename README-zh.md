# Anyi(安易) - 开源的 AI 智能体(AI Agent)框架

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/Anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/Anyi)

| [English](README.md) | [中文](README-zh.md) |

## 介绍

Anyi(安易)是一个开源的 AI 智能体(AI Agent)框架，旨在帮助你构建可以和实际工作相结合的 AI 智能体。我们也提供对大模型访问的 API。


## 特性

Anyi 作为一个 Go 语言的编程框架，提供以下特性：

- 对大模型的访问，允许通过同样的接口使用不同的配置访问不同大模型，目前支持的大模型接口包括：

		- OpenAI
		- Azure OpenAI
		- 阿里云模型服务灵积(Dashscope)
		- Ollama
		
- 对以上大模型，除了支持普通文本聊天外，Anyi还支持对多模态大模型发送图片进行访问。
- 支持基于go语言模板的提示词生成
- 工作流支持：允许将多个对话任务串联起来，形成一个工作流
- 工作流步骤校验：如果一个步骤的输出不符合预期，则反复执行该步骤直到输出符合预期。如果执行次数超过设定次数，则返回一个错误。
- 工作流中不同的步骤允许使用不同的大模型客户端
- 允许定义多个工作流，并根据工作流名称访问不同的工作流

更多功能正在开发中，敬请期待。

## 快速开始

### 安装

```bash
go get github.com/jieliu2000/Anyi
```

### 使用 Anyi 访问大模型

以下为使用 Anyi 访问 OpenAI 的一个简单示例：

```go
package main

import (
	"os"
	"log"
	"github.com/jieliu2000/Anyi/llm"
	"github.com/jieliu2000/Anyi/llm/openai"
	"github.com/jieliu2000/Anyi/message"
)

func main() {

	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []message.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _ := client.Chat(messages)

	log.Printf("Response: %s\n", message.Content)
}


```

在上面的示例中，首先通过`openai.DefaultConfig` 创建一个 OpenAI 的 Anyi 配置，然后将该配置传递给 `llm.NewClient` 创建一个 OpenAI 客户端，最后通过 `client.Chat` 发送一个聊天请求。

## 许可证
Anyi 遵循 [Apache License 2.0](LICENSE)。