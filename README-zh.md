# Anyi(安易) - 开源的自主式 AI 智能体(Autonomous AI Agent)框架

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)

| [English](README.md) | [中文](README-zh.md) |

## 介绍

Anyi(安易)是一个开源的[Go 语言](https://go.dev/)自主式 AI 智能体(Autonomous AI Agent)框架，旨在帮助你构建可以和实际工作相结合的 AI 智能体。我们也提供对大语言模型访问的 API。

Anyi 需要 Go 语言[1.20](https://go.dev/doc/devel/release#go1.20)或更高版本。

## 特性

Anyi 作为一个 Go 语言的编程框架，提供以下特性：

- **对大语言模型的访问**：允许通过同样的接口使用不同的配置访问不同大语言模型，目前支持的大语言模型接口包括：

  - OpenAI
  - Azure OpenAI
  - 阿里云模型服务灵积(Dashscope)
  - Ollama 本地大模型访问，通过 ollama，Anyi 可以实现对多种本地部署大模型的访问
  - DeepSeek
  - 智谱 AI 云服务(bigmodel.cn)
  - Silicon Cloud 云服务（https://siliconflow.cn/）

- **多模态模型支持**：除了支持普通文本聊天外，Anyi 还支持对多模态大语言模型发送图片进行访问
- **多客户端支持**：支持同时访问多个不同来源的大语言模型，不同大语言模型客户端可以通过客户端名字进行区分
- **基于 Go 模板的提示词生成**：支持基于 Go 语言模板([text/template](https://pkg.go.dev/text/template))的提示词生成
- **工作流支持**：允许将多个对话任务串联起来，形成一个工作流
- **工作流步骤校验**：如果一个步骤的输出不符合预期，则反复执行该步骤直到输出符合预期。如果执行次数超过设定次数，则返回一个错误
- **多客户端工作流步骤**：工作流中不同的步骤允许使用不同的大语言模型客户端
- **多工作流定义**：允许定义多个工作流，并根据工作流名称访问不同的工作流
- **基于配置的工作流定义**：允许通过程序代码动态配置工作流，或者通过静态配置文件（支持 YAML、JSON、TOML 格式）配置工作流

## 文档和示例

详细的使用向导请参照[Anyi 使用向导和示例](/docs/zh/tutorial.md)。下面部分是一些简单的上手指南。

## 快速开始

### 安装

```bash
go get -u github.com/jieliu2000/anyi
```

### 访问大语言模型示例

以下为使用 Anyi 访问 Ollama 的一个简单示例：

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/ollama"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 确保你已经安装运行了ollama并通过ollama拉取了qwen2:7b模型
	config := ollama.DefaultConfig("qwen2:7b")
	client, err := anyi.NewClient("qwen2-7b", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _, _:= client.Chat(messages, nil)

	log.Printf("Response: %s\n", message.Content)
}
```

在上面的示例中，首先通过`ollama.DefaultConfig`创建一个 Ollama 的配置，然后将该配置传递给`anyi.NewClient`创建一个客户端，最后通过`client.Chat`发送一个聊天请求。

### 工作流示例

Anyi 允许你定义工作流(Flow)，然后通过工作流名称访问不同的工作流。每个工作流可以包含多个步骤(Step)，每个步骤可以定义自己的执行器（Executor）和校验器（Validator）。

工作流可以通过以下三种方式创建：

- **直接创建实例**：在程序中直接创建`flow.Flow`、`flow.Step`、`flow.StepExecutor`、`flow.StepValidator`等实例
- **动态配置**：创建一个`anyi.AnyiConfig`实例，然后通过`anyi.Config`方法初始化 Anyi
- **静态配置文件**：通过配置文件创建，配置文件格式可以为 YAML、JSON、TOML 等[viper](https://github.com/spf13/viper)支持的格式

Anyi 允许你进行**混合配置**，也就是说你可以在程序中混合使用上面的三种方式创建工作流中的各种对象和工作流本身。

在 Anyi 的工作流中，信息通过`flow.FlowContext`对象进行传递：

```go
type FlowContext struct {
	Text      string           // 用于传递文本信息
	Memory    ShortTermMemory  // 传递结构化信息，类型为 any
	Flow      *Flow            // 保存对 Flow 的引用
	ImageURLs []string         // 用于多模态模型的图片输入
}
```

以下是使用 Anyi 动态配置定义一个工作流的示例：

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

func main() {
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
				Name: "creative_writer",
				Steps: []anyi.StepConfig{
					{
						Name: "write_scifi_novel",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "写一篇关于{{.Text}}的科幻小说",
							},
						},
					},
					{
						Name: "translate_novel",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `把下面用'''括起来的文本翻译成法语，除了翻译的结果以外，不要有任何额外输出。需要翻译的文本:'''{{.Text}}'''。翻译结果:`,
							},
						},
					},
				},
			},
		},
	}

	anyi.Config(&config)
	flow, err := anyi.GetFlow("creative_writer")
	if err != nil {
		panic(err)
	}
	context, err := flow.RunWithInput("月球")
	if err != nil {
		panic(err)
	}
	log.Printf("%s", context.Text)
}
```

在这个示例中：
1. 定义了一个 dashscope 客户端（作为默认客户端）
2. 创建了名为 "creative_writer" 的工作流，包含两个步骤：
   - "write_scifi_novel"：生成科幻小说
   - "translate_novel"：将小说翻译成法语
3. 步骤之间通过 FlowContext 的 Text 属性传递信息
4. 最终将翻译结果输出

## 配置文件支持

Anyi 支持通过配置文件来创建和管理 LLM 客户端和工作流，这种方式非常适合构建可部署的 AI 应用。

### 配置文件的优势

1. **配置与代码分离**：可以在不修改代码的情况下调整 AI 应用的行为
2. **环境适配**：可以为不同环境（开发、测试、生产）准备不同的配置文件
3. **集中管理**：可以在一个配置文件中定义所有客户端和工作流的配置
4. **版本控制**：配置可以纳入版本控制系统，便于追踪变更

### 支持的格式

Anyi 支持多种配置文件格式：
- **YAML**（最常用，易读性好）
- **JSON**（兼容性好，适合程序生成）
- **TOML**（结构化好，适合复杂配置）

配置文件可以通过`ConfigFromFile`方法加载，或使用`ConfigFromString`方法从字符串加载。

### 使用 Ollama 的示例配置文件

以下是一个完整的使用 Ollama 的配置文件示例（YAML 格式）：

```yaml
# anyi-ollama-config.yaml
clients:
  - name: "local-llama2"
    type: "ollama"
    config:
      model: "llama2"
      ollamaApiURL: "http://localhost:11434/api"
  
  - name: "local-qwen"
    type: "ollama"
    config:
      model: "qwen2:7b"
      ollamaApiURL: "http://localhost:11434/api"

flows:
  - name: "qa-flow"
    clientName: "local-qwen"  # 默认使用 qwen 模型
    steps:
      - name: "answer-question"
        executor:
          type: "llm"
          withconfig:
            template: "请详细回答以下问题: {{.Text}}。回答要全面、准确，并给出理由。"
            systemMessage: "你是一个专业的知识问答助手，擅长提供准确、详细的回答。"
        maxRetryTimes: 2
      
  - name: "creative-writing"
    clientName: "local-llama2"  # 使用 llama2 模型
    steps:
      - name: "generate-story"
        executor:
          type: "llm"
          withconfig:
            template: "写一个关于{{.Text}}的短篇故事，故事应该有起承转合的结构。"
            systemMessage: "你是一个富有创造力的故事作家。"
      
      - name: "summarize-story"
        executor:
          type: "llm"
          withconfig:
            template: "请用三句话总结以下故事的主要内容：\n\n{{.Text}}"
        clientName: "local-qwen"  # 这个步骤特别指定使用 qwen 模型
        validator:
          type: "string"
          withconfig:
            matchRegex: "^.{30,500}$"  # 确保总结的长度在合理范围内
```

### 加载和使用配置文件

```go
package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
)

func main() {
	// 加载配置文件
	err := anyi.ConfigFromFile("./anyi-ollama-config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 获取并运行问答流程
	qaFlow, err := anyi.GetFlow("qa-flow")
	if err != nil {
		log.Fatalf("获取流程失败: %v", err)
	}
	
	result, err := qaFlow.RunWithInput("人工智能的发展历程是怎样的？")
	if err != nil {
		log.Fatalf("流程执行失败: %v", err)
	}

	fmt.Println("问答结果:", result.Text)
}
```

### 配置文件最佳实践

1. **环境变量替换**：在配置文件中可以使用 `$VARIABLE_NAME` 引用环境变量，保护敏感信息
2. **配置文件模块化**：根据功能将配置拆分为多个文件，便于管理
3. **验证器使用**：为关键步骤添加验证器，确保输出符合预期
4. **步骤重试**：为不稳定的步骤设置合理的 `maxRetryTimes` 值

## 内置组件

Anyi 提供了多种内置组件来帮助你构建 AI 应用：

### 内置执行器

- **LLMExecutor**：基于 LLM 的执行器，支持模板提示词和直接提示词
- **ConditionalFlowExecutor**：条件流执行器，可以基于条件选择不同的子工作流执行
- **RunCommandExecutor**：系统命令执行器，可以执行系统命令
- **SetContextExecutor**：上下文设置执行器，可以设置工作流上下文的属性

### 内置校验器

- **StringValidator**：字符串校验器，支持相等比较和正则表达式匹配
- **JsonValidator**：JSON 校验器，验证输出是否为有效的 JSON 格式

## 许可证

Anyi 遵循 [Apache License 2.0](LICENSE)。
