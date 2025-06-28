# Anyi 教程 - 快速入门

欢迎使用 Anyi！本教程将引导您了解核心概念，并帮助您使用 Anyi 框架构建第一个 AI 应用程序。

## 📚 文档结构

我们重新组织了 Anyi 文档，让您更容易找到所需内容：

### 🚀 Anyi 新手？从这里开始

**[📖 完整文档中心 →](README.md)**

要获得完整的文档体验，请访问我们的主文档中心，它为不同用户需求提供了有组织的学习路径。

### ⚡ 快速开始（5 分钟）

如果您想直接开始：

1. **[安装 →](getting-started/installation.md)** - 在您的系统上设置 Anyi
2. **[快速入门指南 →](getting-started/quickstart.md)** - 5 分钟内构建您的第一个 AI 应用
3. **[核心概念 →](getting-started/concepts.md)** - 了解 Anyi 的工作原理

### 📋 学习路径

按照这个结构化的学习路径：

#### 第一步：基础知识

- [安装和设置](getting-started/installation.md)
- [基本概念](getting-started/concepts.md)
- [您的第一个应用程序](getting-started/quickstart.md)

#### 第二步：核心技能

- [使用 LLM 客户端](tutorials/llm-clients.md) - 连接到 OpenAI、Anthropic、Ollama 等
- [构建工作流](tutorials/workflows.md) - 创建多步骤 AI 流程
- [配置管理](tutorials/configuration.md) - 组织您的设置

#### 第三步：高级功能

- [多模态应用](tutorials/multimodal.md) - 处理文本和图像
- [错误处理](how-to/error-handling.md) - 构建健壮的应用程序
- [性能优化](how-to/performance.md) - 速度和成本优化

#### 第四步：生产就绪

- [Web 集成](how-to/web-integration.md) - 与 Web 框架一起使用
- [安全最佳实践](advanced/security.md) - 保护您的应用程序
- [生产部署](advanced/deployment.md) - 部署到生产环境

## 🎯 常见用例

### 我想要...

**连接到 LLM 提供商**
→ [提供商设置指南](how-to/provider-setup.md)

**构建多步骤 AI 工作流**
→ [构建工作流教程](tutorials/workflows.md)

**处理错误和重试**
→ [错误处理指南](how-to/error-handling.md)

**处理图像和文本**
→ [多模态应用](tutorials/multimodal.md)

**与我的 Web 应用集成**
→ [Web 集成指南](how-to/web-integration.md)

**优化性能和成本**
→ [性能优化](how-to/performance.md)

**部署到生产环境**
→ [部署指南](advanced/deployment.md)

**创建自定义组件**
→ [自定义执行器指南](advanced/custom-executors.md)

## 📖 参考资料

当您需要查找特定信息时：

- **[API 参考](reference/api.md)** - 完整的 API 文档
- **[配置参考](reference/configuration.md)** - 所有配置选项
- **[组件参考](reference/components.md)** - 内置执行器和验证器
- **[常见问题](reference/faq.md)** - 常见问题解答

## 🔍 寻找原始教程？

全面的教程内容已重新组织为专注的模块化指南，以便更好地导航和学习。如果您需要原始的单页教程，可以在这里找到：

**[旧版教程 →](tutorial-legacy.md)**

但是，我们建议使用上面的新模块化结构以获得更好的学习体验。

## 💡 快速示例

### 简单聊天示例

```go
package main

import (
    "fmt"
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

    // 发送消息
    messages := []chat.Message{
        {Role: "user", Content: "你好，你好吗？"},
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        log.Fatalf("聊天失败: %v", err)
    }

    fmt.Println("回复:", response.Content)
}
```

### 基于配置的工作流

```yaml
# config.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"

flows:
  - name: "content_processor"
    clientName: "openai"
    steps:
      - name: "analyze"
        executor:
          type: "llm"
          withconfig:
            template: "分析以下文本: {{.Text}}"
        validator:
          type: "string"
          withconfig:
            minLength: 50
```

```go
// 加载配置并运行工作流
func main() {
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    flow, err := anyi.GetFlow("content_processor")
    if err != nil {
        log.Fatal(err)
    }

    result, err := flow.RunWithInput("您的文本在这里...")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("结果:", result.Text)
}
```

## 🤝 获取帮助

- **有问题？** 查看 [常见问题](reference/faq.md)
- **遇到问题？** 在 [GitHub](https://github.com/jieliu2000/anyi/issues) 上提交问题
- **需要示例？** 浏览 [examples 目录](../../examples/)

## 🚀 准备开始？

选择您的路径：

- **初学者**: 从 [安装](getting-started/installation.md) 开始
- **有经验者**: 跳转到 [LLM 客户端](tutorials/llm-clients.md)
- **参考**: 浏览 [API 文档](reference/api.md)

使用 Anyi 愉快编程！🎉
