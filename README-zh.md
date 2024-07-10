# anyi(安易) - 开源的 AI 智能体(AI Agent)框架

| [English](README.md) | [中文](README-zh.md) |

## 介绍

Anyi(安易)是一个开源的 AI 智能体(AI Agent)框架，旨在帮助你构建可以和实际工作相结合的 AI 智能体。我们也提供对大模型访问的 API。

## 特性

Anyi 作为一个 Go 语言的编程框架，提供以下特性：

- 对大模型的访问，允许通过同样的接口使用不同的配置访问不同大模型

更多功能正在开发中，敬请期待。

## 快速开始

### 安装

```bash
go get github.com/jieliu2000/anyi
```

### 使用 anyi 访问大模型

以下为使用 anyi 访问 OpenAI 的一个简单示例：

```go
package main

import (
	"os"
	"log"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"
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

在上面的示例中，首先通过`openai.DefaultConfig` 创建一个 OpenAI 的 anyi 配置，然后将该配置传递给 `llm.NewClient` 创建一个 OpenAI 客户端，最后通过 `client.Chat` 发送一个聊天请求。
