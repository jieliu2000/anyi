# 快速入门指南

欢迎来到 Anyi！本指南将在 5 分钟内带您构建第一个 AI 应用程序。

## 前提条件

开始之前，请确保您已经：

1. 安装了 Go 1.20 或更高版本
2. 完成了 [Anyi 安装](installation.md)
3. 如果使用云服务提供商（如 DeepSeek、通义千问等），需要获得相应的 API 密钥
4. 如果使用本地模型（如 Ollama），需要在本地安装并运行相应的服务

> **📌 提示：** Anyi 支持多种类型的 LLM 提供商：
> - **云服务提供商**：如 DeepSeek、通义千问、OpenAI、Anthropic 等，需要 API 密钥
> - **本地提供商**：如 Ollama，无需 API 密钥，但需要在本地安装并运行服务

## 第一步：创建项目

创建一个新的 Go 项目：

```bash
mkdir my-first-anyi-app
cd my-first-anyi-app
go mod init my-first-anyi-app
go get github.com/jieliu2000/anyi
```

## 第二步：设置环境变量（仅限云服务提供商）

如果您使用云服务提供商，创建 `.env` 文件并添加您的 API 密钥：

```bash
# .env
DEEPSEEK_API_KEY=your-deepseek-api-key-here
```

> **提示：** 
> - 如果使用 DeepSeek、通义千问等云服务提供商，请设置相应的 API 密钥环境变量
> - 如果使用 Ollama 等本地模型提供商，则无需设置 API 密钥，但需要确保本地服务正在运行

## 第三步：编写您的第一个应用

创建 `main.go` 文件：

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek" // 云服务提供商
    // "github.com/jieliu2000/anyi/llm/ollama"  // 本地提供商（如使用 Ollama 请取消注释）
    "github.com/jieliu2000/anyi/llm/chat"
    "github.com/joho/godotenv"
)

func main() {
    // 加载环境变量
    if err := godotenv.Load(); err != nil {
        log.Println("警告：未找到 .env 文件")
    }

    // 检查是否设置了 API 密钥
    apiKey := os.Getenv("DEEPSEEK_API_KEY")
    if apiKey == "" {
        log.Println("未设置 API 密钥。如果您使用云服务提供商，请设置相应的环境变量。")
    }

    // 创建客户端 - 根据需要选择提供商
    // 对于云服务提供商（需要 API 密钥）：
    config := deepseek.DefaultConfig(apiKey, "deepseek-reasoner")
    client, err := anyi.NewClient("deepseek", config)
    
    // 对于本地提供商如 Ollama（无需 API 密钥）：
    // config := ollama.DefaultConfig("llama3") // 或您喜欢的本地模型
    // client, err := anyi.NewClient("ollama", config)
    
    if err != nil {
        log.Fatalf("创建客户端失败: %v", err)
    }

    // 准备消息
    messages := []chat.Message{
        {
            Role:    "user",
            Content: "你好！请用中文简单介绍一下人工智能。",
        },
    }

    // 发送请求并获取响应
    response, info, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天请求失败: %v", err)
    }

    // 显示结果
    fmt.Println("🤖 AI 回复:")
    fmt.Println(response.Content)
    fmt.Printf("\n📊 使用统计: %d 个 token\n", info.TotalTokens)
}
```

## 第四步：安装依赖并运行

```bash
# 安装 godotenv 用于加载环境变量
go get github.com/joho/godotenv

# 运行应用
go run main.go
```

您应该看到类似这样的输出：

```
🤖 AI 回复:
人工智能（AI）是一种让计算机系统能够执行通常需要人类智能的任务的技术。它包括机器学习、自然语言处理、计算机视觉等领域，可以帮助我们解决复杂问题、自动化任务，并在医疗、教育、交通等各个领域提供智能化解决方案。

📊 使用统计: 95 个 token
```

## 第五步：使用配置文件（可选）

为了更好地管理配置，让我们使用配置文件的方式：

创建 `config.yaml`：

```yaml
clients:
  - name: "deepseek"
    type: "deepseek"
    config:
      apiKey: "$DEEPSEEK_API_KEY"
      model: "deepseek-reasoner"
      temperature: 0.7

flows:
  - name: "chat_assistant"
    clientName: "deepseek"
    steps:
      - name: "respond"
        executor:
          type: "llm"
          withconfig:
            template: "请用友好的语调回答用户的问题：{{.Text}}"
            systemMessage: "你是一个有用的中文AI助手。"
```

创建 `config_main.go`：

```go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/joho/godotenv"
)

func main() {
    // 加载环境变量
    if err := godotenv.Load(); err != nil {
        log.Println("警告：未找到 .env 文件")
    }

    // 从配置文件加载
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 获取流程
    flow, err := anyi.GetFlow("chat_assistant")
    if err != nil {
        log.Fatalf("获取流程失败: %v", err)
    }

    // 运行流程
    result, err := flow.RunWithInput("什么是机器学习？")
    if err != nil {
        log.Fatalf("运行流程失败: %v", err)
    }

    // 显示结果
    fmt.Println("🤖 AI 回复:")
    fmt.Println(result.Text)
}
```

运行配置版本：

```bash
go run config_main.go
```

## 使用其他 LLM 提供商

### Ollama（本地模型）

如果您想使用完全离线的本地模型，可以使用 Ollama：

``go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/ollama"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    // 创建 Ollama 客户端（确保 Ollama 正在运行）
    config := ollama.DefaultConfig("llama3")
    client, err := anyi.NewClient("ollama", config)
    if err != nil {
        log.Fatalf("创建 Ollama 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请用中文介绍一下 Go 编程语言。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Println("🦙 Llama 回复:")
    fmt.Println(response.Content)
}
```

### Anthropic Claude

如果您有 Anthropic API 访问权限，可以使用 Claude：

``go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/anthropic"
    "github.com/jieliu2000/anyi/llm/chat"
    "github.com/joho/godotenv"
)

func main() {
    godotenv.Load()

    // 创建 Anthropic 客户端
    config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))
    client, err := anyi.NewClient("claude", config)
    if err != nil {
        log.Fatalf("创建 Claude 客户端失败: %v", err)
    }

    messages := []chat.Message{
        {
            Role:    "user",
            Content: "请解释一下什么是函数式编程。",
        },
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Println("🧠 Claude 回复:")
    fmt.Println(response.Content)
}
```

## 构建多步骤工作流

让我们创建一个更复杂的例子，展示 Anyi 的工作流功能：

``go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek"
    "github.com/joho/godotenv"
)

func main() {
    godotenv.Load()

    // 程序化配置多步骤工作流
    config := anyi.AnyiConfig{
        Clients: []anyi.ClientConfig{
            {
                Name: "deepseek",
                Type: "deepseek",
                Config: map[string]interface{}{
                    "apiKey": os.Getenv("DEEPSEEK_API_KEY"),
                    "model":  "deepseek-reasoner",
                },
            },
        },
        Flows: []anyi.FlowConfig{
            {
                Name:       "content_creator",
                ClientName: "deepseek",
                Steps: []anyi.StepConfig{
                    {
                        Name: "analyze_topic",
                        Executor: &executors.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "分析以下主题并提供3个关键点：{{.Text}}",
                                "systemMessage": "你是一个专业的内容分析师。",
                            },
                        },
                    },
                    {
                        Name: "create_content",
                        Executor: &executors.ExecutorConfig{
                            Type: "llm",
                            WithConfig: map[string]interface{}{
                                "template": "基于以下分析，写一篇200字的介绍文章：\n\n{{.Text}}",
                                "systemMessage": "你是一个专业的内容创作者。",
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

    // 运行工作流
    flow, err := anyi.GetFlow("content_creator")
    if err != nil {
        log.Fatalf("获取流程失败: %v", err)
    }

    result, err := flow.RunWithInput("区块链技术")
    if err != nil {
        log.Fatalf("运行流程失败: %v", err)
    }

    fmt.Println("📝 生成的内容:")
    fmt.Println(result.Text)
}
```

## 常见问题解决

### 问题：API 密钥错误

```
错误: 401 Unauthorized
```

**解决方案：**

1. 检查 `.env` 文件中的 API 密钥是否正确
2. 确保没有多余的空格或引号
3. 验证 API 密钥是否有效且未过期

### 问题：网络连接问题

```
错误: dial tcp: connection timeout
```

**解决方案：**

1. 检查网络连接
2. 确认 API 服务器是否可访问
3. 考虑使用本地模型（Ollama）作为替代

### 问题：模型不存在

```
错误: model 'xxx' not found
```

**解决方案：**

1. 检查模型名称是否正确
2. 使用支持的模型名称，如 `deepseek-reasoner`、`deepseek-chat`
3. 查看提供商文档了解可用模型

## 下一步

恭喜！您已经成功创建了第一个 Anyi 应用。现在您可以：

1. **学习核心概念** - 阅读 [基本概念](concepts.md) 深入理解 Anyi
2. **探索更多提供商** - 查看 [LLM 客户端教程](../tutorials/llm-clients.md)
3. **构建复杂工作流** - 学习 [工作流构建](../tutorials/workflows.md)
4. **配置管理** - 掌握 [配置管理](../tutorials/configuration.md)
5. **处理图像** - 尝试 [多模态应用](../tutorials/multimodal.md)

## 示例代码仓库

您可以在 [examples 目录](../../../examples/) 中找到更多示例代码，包括：

- 简单聊天机器人
- 文档分析器
- 代码生成器
- Web API 服务
- 批处理工具

开始您的 AI 开发之旅吧！🚀
