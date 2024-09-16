# Anyi(安易) - 开源的 AI 智能体(AI Agent)框架

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)

| [English](README.md) | [中文](README-zh.md) |

## 介绍

Anyi(安易)是一个开源的[Go 语言](https://go.dev/)AI 智能体(AI Agent)框架，旨在帮助你构建可以和实际工作相结合的 AI 智能体。我们也提供对大语言模型访问的 API。

Anyi 需要 Go 语言[1.20](https://go.dev/doc/devel/release#go1.20)或更高版本。

## 特性

Anyi 作为一个 Go 语言的编程框架，提供以下特性：

- 对大语言模型的访问，允许通过同样的接口使用不同的配置访问不同大语言模型，目前支持的大语言模型接口包括：

  - OpenAI
  - Azure OpenAI
  - 阿里云模型服务灵积(Dashscope)
  - Ollama 本地大模型访问，通过 ollama，Anyi 可以实现对多种本地部署大模型的访问.
  - 智谱 AI 云服务(bigmodel.cn)
  - Silicon Cloud 云服务（https://siliconflow.cn/）

- 对以上大语言模型，除了支持普通文本聊天外，Anyi 还支持对多模态大语言模型发送图片进行访问。
- 支持同时访问多个不同来源的大语言模型，不同大语言模型客户端可以通过客户端名字进行区分。
- 支持基于 Go 语言模板([text/template](https://pkg.go.dev/text/template))的提示词生成
- 工作流支持：允许将多个对话任务串联起来，形成一个工作流
- 工作流步骤校验：如果一个步骤的输出不符合预期，则反复执行该步骤直到输出符合预期。如果执行次数超过设定次数，则返回一个错误。
- 工作流中不同的步骤允许使用不同的大语言模型客户端
- 允许定义多个工作流，并根据工作流名称访问不同的工作流
- 基于配置的工作流定义：允许通过程序代码动态配置工作流，或者通过静态配置文件配置工作流

## 代码和示例

详细的使用向导请参照[Anyi 使用向导和示例](/docs/zh/tutorial.md)。下面部分是一些简单的上手指南。

## 快速开始

### 安装

```bash
go get -u github.com/jieliu2000/anyi
```

### Anyi 访问大语言模型示例

以下为使用 Anyi 访问 OpenAI 的一个简单示例：

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
	// 确保你已经安装运行了ollama并通过ollama拉取了qwen2:7b模型
	config := ollama.DefaultConfig("qwen2:7b")
	client, err := llm.NewClient(config)
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

在上面的示例中，首先通过`dashscope.DefaultConfig` 创建一个 Dashscope 的 Anyi 配置，然后将该配置传递给 `anyi.NewClient` 创建一个 OpenAI 客户端，最后通过 `client.Chat` 发送一个聊天请求。

### Anyi 工作流示例

Anyi 允许你定义工作流(Flow)，然后通过工作流名称访问不同的工作流。每个工作流可以包含多个步骤(Step)，每个步骤(Step)可以定义自己的执行器（Executor）和校验器（Validator）。

工作流可以通过以下三种方式创建：

- 直接在程序中创建`flow.Flow`, `flow.Step`，`flow.StepExecutor`, `flow.StepValidator` 等实例
- 动态配置：创建一个`anyi.AnyiConfig` 实例，然后通过 `anyi.Config` 方法初始化 Anyi，Anyi 会根据配置创建`Client`，`Flow`等对象
- 静态配置：通过配置文件创建，配置文件格式可以为 toml, yaml, json 等[viper](https://github.com/spf13/viper)支持的格式。之后通过`anyi.ConfigFromFile`方法初始化 Anyi

Anyi 允许你进行**混合配置**，也就是说你可以在程序中混合使用上面的三种方式创建工作流中的各种对象和工作流本身。

在 Anyi 的工作流中，Flow, Step, StepExecutor, StepValidator 等对象都通过`flow.FlowContext`对象进行信息传递。`flow.FlowContext`对象的声明如下：

```go
type FlowContext struct {
	Text   string
	Memory ShortTermMemory
	Flow   *Flow
}
```

其中 Text 属性用来传递文本信息，Memory 属性用来传递其他结构化信息，在当前版本中 ShortTermMemory 实际上是 any 类型，因此允许你设置为任何类型的实例，Flow 属性是用来保存对 Flow 的引用。

以下为使用 Anyi 动态配置定义一个工作流的示例：

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
				Name: "smart_writer",
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
	flow, err := anyi.GetFlow("smart_writer")
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

在上面的示例中，首先创建了一个 `AnyiConfig` 配置，该配置包含`Clients`和`Flows`两个属性。正如其名字，`Clients`是用来定义 Anyi 客户端的数组，而`Flows`是用来定义工作流的数组。

```go
	config := anyi.AnyiConfig{
		Clients: []llm.ClientConfig{
			//...
		},
		Flows: []anyi.FlowConfig{
			//...
		},
	}
```

在`Clients`中，仅包含一个 dashscope `Client` 配置。由于程序中仅有一个 `Client` 也就是名为 `dashscope` 的 `Client`，Anyi 会将这个客户端注册为**默认的 Client**。Anyi 允许注册多个 Clients，在 Flow, Step 中都可以指定使用哪个 `Client` 执行任务。如果没有指定，Anyi 就会使用默认的 Client。

在`Flows`中，定义了一个名称为 `smart_writer` 的 Flow。该工作流包含两个步骤（`Step`）:

- 第一个步骤"write_scifi_novel"使用了一个 llm 类型的 Executor。llm 是 Anyi 内建的一种执行器类型，可以使用直接提示词或者基于模板的提示词调用 LLM 模型。在上面的示例中，`template` 参数指定了调用 LLM 的提示词模板，这个模板是使用 Go 语言的文本模板([text/template](https://pkg.go.dev/text/template))。
  模板中使用了{{.Text}}作为模板的参数，`.Text`是`flow.FlowContext`中的属性。在 Anyi 的 llm 执行器中，Anyi 会根据用户的初始输入设置`flow.FlowContext`的 Text 属性。之后如果执行器输出文本内容，Anyi 会将`flow.FlowContext`的 Text 属性作为输出返回。

- 第二个步骤"translate_novel"也是使用了 llm 类型的 Executor，但是用了不同的提示词模板。

在通过`anyi.Config(&config)`配置 Anyi 以后，通过 `anyi.GetFlow("smart_writer")` 可以获取到 Anyi 创建的名为`smart_writer`的工作流。然后通过 `flow.RunWithInput("月球")` 运行该工作流，在 Flow 运行之前，RunWithInput 的参数"月球"会被设置到`flow.FlowContext`的 Text 属性中，并在之后传递给第一个 Step("write_scifi_novel")的 Executor 中。、

第一个步骤"write_scifi_novel"的 Executor 会根据提示词模板和用户输入生成一个提示词，然后调用 LLM 模型进行计算。该步骤的输出是文本内容也就是小说内容会被设置到`flow.FlowContext`的 Text 属性中，之后会传递给下一个步骤"translate_novel"去进行翻译。

同样第二个步骤也使用了 go 语言的模板，在模板中的{{.Text}}会被替换为`flow.FlowContext`的 Text 属性，也就是"write_scifi_novel"的输出内容。之后，Anyi 会调用 LLM 模型进行翻译，并再次将翻译结果设置到`flow.FlowContext`的 Text 属性中。最后，Anyi 会将`flow.FlowContext`的引用作为 Flow 运行结果返回。

## 许可证

Anyi 遵循 [Apache License 2.0](LICENSE)。
